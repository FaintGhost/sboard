# SBoard Phase 2 (User Management) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Panel 增加管理员 JWT 登录与用户 CRUD（含软删除、分页与状态过滤）。

**Architecture:** Panel 增加管理员登录与 JWT 中间件，用户数据通过 SQLite DAO 访问；API 使用标准 REST 路径，并统一 JSON 返回格式。

**Tech Stack:** Go、Gin、SQLite、golang-migrate、jwt/v5

---

### Task 1: 扩展 Panel 配置与校验

**Files:**
- Modify: `panel/internal/config/config.go`
- Create: `panel/internal/config/config_test.go`
- Modify: `panel/cmd/panel/main.go`

**Step 1: 写失败测试（校验必填配置）**

```go
func TestValidateConfig(t *testing.T) {
  cfg := config.Config{}
  err := config.Validate(cfg)
  require.Error(t, err)

  cfg = config.Config{AdminUser: "admin", AdminPass: "pass", JWTSecret: "secret"}
  require.NoError(t, config.Validate(cfg))
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/config -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

`panel/internal/config/config.go`
```go
type Config struct {
  HTTPAddr  string
  DBPath    string
  AdminUser string
  AdminPass string
  JWTSecret string
}

func Load() Config {
  cfg := Config{
    HTTPAddr: ":8080",
    DBPath:   "panel.db",
  }
  if v := os.Getenv("PANEL_HTTP_ADDR"); v != "" { cfg.HTTPAddr = v }
  if v := os.Getenv("PANEL_DB_PATH"); v != "" { cfg.DBPath = v }
  if v := os.Getenv("ADMIN_USER"); v != "" { cfg.AdminUser = v }
  if v := os.Getenv("ADMIN_PASS"); v != "" { cfg.AdminPass = v }
  if v := os.Getenv("PANEL_JWT_SECRET"); v != "" { cfg.JWTSecret = v }
  return cfg
}

func Validate(cfg Config) error {
  if cfg.AdminUser == "" || cfg.AdminPass == "" || cfg.JWTSecret == "" {
    return errors.New("missing admin or jwt config")
  }
  return nil
}
```

`panel/cmd/panel/main.go`
```go
cfg := config.Load()
if err := config.Validate(cfg); err != nil { log.Fatal(err) }
```

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/config -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/config/config.go panel/internal/config/config_test.go panel/cmd/panel/main.go
git commit -m "feat(panel): add admin config validation"
```

---

### Task 2: 管理员登录与 JWT 工具

**Files:**
- Create: `panel/internal/api/auth.go`
- Create: `panel/internal/api/auth_test.go`
- Modify: `panel/internal/api/router.go`

**Step 1: 写失败测试（登录成功/失败）**

```go
func TestAdminLogin(t *testing.T) {
  cfg := config.Config{AdminUser: "admin", AdminPass: "pass", JWTSecret: "secret"}
  r := api.NewRouter(cfg, nil)

  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"pass"}`))
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)

  w = httptest.NewRecorder()
  req = httptest.NewRequest(http.MethodPost, "/api/admin/login", strings.NewReader(`{"username":"admin","password":"wrong"}`))
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusUnauthorized, w.Code)
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

`panel/internal/api/auth.go`
```go
type loginReq struct {
  Username string `json:"username"`
  Password string `json:"password"`
}

type loginResp struct {
  Token     string `json:"token"`
  ExpiresAt string `json:"expires_at"`
}

func AdminLogin(cfg config.Config) gin.HandlerFunc {
  return func(c *gin.Context) {
    var req loginReq
    if err := c.ShouldBindJSON(&req); err != nil {
      c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
      return
    }
    if req.Username != cfg.AdminUser || req.Password != cfg.AdminPass {
      c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
      return
    }
    token, exp, err := signAdminToken(cfg.JWTSecret)
    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{"error": "sign token failed"})
      return
    }
    c.JSON(http.StatusOK, gin.H{"data": loginResp{Token: token, ExpiresAt: exp.Format(time.RFC3339)}})
  }
}

func signAdminToken(secret string) (string, time.Time, error) { ... }
```

`panel/internal/api/router.go`
```go
func NewRouter(cfg config.Config, store *db.Store) *gin.Engine {
  r := gin.New()
  r.GET("/api/health", Health)
  r.POST("/api/admin/login", AdminLogin(cfg))
  // users routes added later
  return r
}
```

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/api/auth.go panel/internal/api/auth_test.go panel/internal/api/router.go
git commit -m "feat(panel): add admin login"
```

---

### Task 3: JWT 鉴权中间件与路由保护

**Files:**
- Modify: `panel/internal/api/auth.go`
- Modify: `panel/internal/api/router.go`
- Modify: `panel/internal/api/health_test.go`

**Step 1: 写失败测试（未授权访问用户接口）**

```go
func TestAuthMiddleware(t *testing.T) {
  cfg := config.Config{AdminUser: "admin", AdminPass: "pass", JWTSecret: "secret"}
  r := api.NewRouter(cfg, nil)
  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusUnauthorized, w.Code)
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

在 `auth.go` 增加 `AuthMiddleware(secret string)` 校验 JWT，失败返回 401。

在 `router.go` 增加用户路由分组：
```go
auth := r.Group("/api")
auth.Use(AuthMiddleware(cfg.JWTSecret))
auth.GET("/users", UsersList(store))
```

**Step 4: 更新健康测试以适配新签名**

`health_test.go` 中 `NewRouter` 传入 `config.Config{}`。

**Step 5: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: PASS

**Step 6: Commit**

```bash
git add panel/internal/api/auth.go panel/internal/api/router.go panel/internal/api/health_test.go
 git commit -m "feat(panel): add auth middleware"
```

---

### Task 4: 用户 DAO 与数据模型（含测试）

**Files:**
- Create: `panel/internal/db/users.go`
- Create: `panel/internal/db/users_test.go`
- Modify: `panel/go.mod`

**Step 1: 写失败测试（创建与唯一约束）**

```go
func TestUserCreateAndUnique(t *testing.T) {
  db := setupDB(t)
  user, err := db.CreateUser(context.Background(), "alice")
  require.NoError(t, err)
  require.NotEmpty(t, user.UUID)

  _, err = db.CreateUser(context.Background(), "alice")
  require.Error(t, err)
  require.True(t, errors.Is(err, db.ErrConflict))
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/db -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

- 新增 `User` 结构体
- 新增 `ErrConflict`
- 使用 `github.com/google/uuid` 生成 UUID
- 实现 `CreateUser` / `GetUserByID` / `ListUsers` / `UpdateUser` / `DisableUser`

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/db -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/db/users.go panel/internal/db/users_test.go panel/go.mod
git commit -m "feat(panel): add user dao"
```

---

### Task 5: 用户 API（含测试）

**Files:**
- Create: `panel/internal/api/users.go`
- Create: `panel/internal/api/users_test.go`
- Modify: `panel/internal/api/router.go`

**Step 1: 写失败测试（创建/列表/禁用）**

```go
func TestUsersAPI(t *testing.T) {
  cfg := config.Config{JWTSecret: "secret"}
  store := setupStore(t)
  r := api.NewRouter(cfg, store)
  token := mustToken(cfg.JWTSecret)

  // create
  req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{"username":"alice"}`))
  req.Header.Set("Authorization", "Bearer "+token)
  w := httptest.NewRecorder()
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusCreated, w.Code)

  // list
  req = httptest.NewRequest(http.MethodGet, "/api/users?limit=10&offset=0", nil)
  req.Header.Set("Authorization", "Bearer "+token)
  w = httptest.NewRecorder()
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

- `POST /api/users`：创建用户
- `GET /api/users`：分页+状态过滤
- `GET /api/users/:id`：查询用户
- `PUT /api/users/:id`：更新用户
- `DELETE /api/users/:id`：软删除（status=disabled）

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/api/users.go panel/internal/api/users_test.go panel/internal/api/router.go
git commit -m "feat(panel): add user CRUD"
```

---

### Task 6: 更新 Panel 启动与路由注入（如需）

**Files:**
- Modify: `panel/cmd/panel/main.go`

**Step 1: 确认 Router 需要注入 Store**

- 将 DB 连接包装为 `db.Store`
- 调用 `api.NewRouter(cfg, store)`

**Step 2: Commit**

```bash
git add panel/cmd/panel/main.go
git commit -m "chore(panel): wire store to router"
```

---
