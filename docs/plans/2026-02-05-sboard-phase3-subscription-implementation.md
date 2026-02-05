# SBoard Phase 3 (Subscription - sing-box) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 新增 sing-box 订阅端点，按 user_nodes 返回用户节点下的入站 outbounds，支持 UA/format 选择与 Base64(JSON) 输出。

**Architecture:** Panel 侧新增订阅 handler + 订阅生成器；DB 增加 user_nodes 与 nodes/inbounds 字段；订阅生成从 DB 拼装 outbounds。

**Tech Stack:** Go、Gin、SQLite、golang-migrate、sing-box option schema（仅 JSON 结构）

---

### Task 1: DB 迁移（nodes/inbounds/user_nodes）

**Files:**
- Modify: `panel/internal/db/migrations/0001_init.up.sql`
- Modify: `panel/internal/db/migrations/0001_init.down.sql`
- Create: `panel/internal/db/migrations/0002_nodes_user_nodes.up.sql`
- Create: `panel/internal/db/migrations/0002_nodes_user_nodes.down.sql`

**Step 1: 写失败测试（校验新表/字段存在）**

`panel/internal/db/migrate_test.go` 新增：
```go
func TestMigrateAddsUserNodes(t *testing.T) {
  dir := t.TempDir()
  dbPath := filepath.Join(dir, "test.db")
  database, err := db.Open(dbPath)
  require.NoError(t, err)
  t.Cleanup(func() { _ = database.Close() })

  _, file, _, ok := runtime.Caller(0)
  require.True(t, ok)
  migrationsDir := filepath.Join(filepath.Dir(file), "migrations")
  err = db.MigrateUp(database, migrationsDir)
  require.NoError(t, err)

  _, err = database.Exec("SELECT user_id, node_id FROM user_nodes LIMIT 1")
  require.NoError(t, err)

  _, err = database.Exec("SELECT api_address, api_port, public_address FROM nodes LIMIT 1")
  require.NoError(t, err)

  _, err = database.Exec("SELECT public_port FROM inbounds LIMIT 1")
  require.NoError(t, err)
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/db -v`
Expected: FAIL（字段/表不存在）

**Step 3: 写最小实现**
- 在 `0001_init.up.sql` 中为 `nodes` 添加 `api_address`、`api_port`、`public_address`
- 在 `0001_init.up.sql` 中为 `inbounds` 添加 `public_port`
- 新增 `0002_nodes_user_nodes.*`，创建 `user_nodes` 表
- `down.sql` 中对应删除

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/db -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/db/migrations/0001_init.up.sql panel/internal/db/migrations/0001_init.down.sql panel/internal/db/migrations/0002_nodes_user_nodes.up.sql panel/internal/db/migrations/0002_nodes_user_nodes.down.sql panel/internal/db/migrate_test.go
git commit -m "feat(panel): add node public fields and user_nodes"
```

---

### Task 2: DB 访问层（Nodes/Inbounds/UserNodes 读取）

**Files:**
- Create: `panel/internal/db/subscription.go`
- Create: `panel/internal/db/subscription_test.go`

**Step 1: 写失败测试（查询用户节点与入站）**

```go
func TestSubscriptionQuery(t *testing.T) {
  store := setupStore(t)
  ctx := context.Background()

  nodeID := insertNode(t, store, "node-a", "api.local", 2222, "a.example.com")
  inboundID := insertInbound(t, store, nodeID, "vless", 443, 0)
  userID := insertUser(t, store, "alice")
  bindUserNode(t, store, userID, nodeID)

  got, err := store.ListUserInbounds(ctx, userID)
  require.NoError(t, err)
  require.Len(t, got, 1)
  require.Equal(t, "vless", got[0].InboundType)
  require.Equal(t, "a.example.com", got[0].NodePublicAddress)
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/db -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**
- 定义 `SubscriptionInbound` 结构（包含 node 公网地址、inbound 类型/端口/模板字段）
- 实现 `ListUserInbounds(ctx, userID)`，join `user_nodes -> nodes -> inbounds`
- 测试用辅助插入函数（仅测试文件内）

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/db -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/db/subscription.go panel/internal/db/subscription_test.go
git commit -m "feat(panel): add subscription db queries"
```

---

### Task 3: 订阅生成器（sing-box outbounds）

**Files:**
- Create: `panel/internal/subscription/singbox.go`
- Create: `panel/internal/subscription/singbox_test.go`

**Step 1: 写失败测试（生成 outbounds）**

```go
func TestSingboxGenerateOutbounds(t *testing.T) {
  user := subscription.User{UUID: "u-1", Username: "alice"}
  items := []subscription.Item{
    {
      InboundType: "vless",
      NodePublicAddress: "a.example.com",
      InboundListenPort: 443,
      InboundPublicPort: 0,
      Settings: json.RawMessage(`{"flow":"xtls-rprx-vision"}`),
    },
  }

  out, err := subscription.BuildSingbox(user, items)
  require.NoError(t, err)
  require.Contains(t, string(out), "a.example.com")
  require.Contains(t, string(out), "vless")
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/subscription -v`
Expected: FAIL

**Step 3: 写最小实现**
- 定义 `User` 与 `Item` 结构
- 将 `settings/tls_settings/transport_settings` 解析为 map，注入：
  - `server` = `NodePublicAddress`
  - `server_port` = `public_port > 0 ? public_port : listen_port`
  - `uuid`/`password`/`username` 按 inbound type 注入（先支持 vless/vmess/shadowsocks/trojan）
- 输出 JSON 仅包含 `outbounds`

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/subscription -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/subscription/singbox.go panel/internal/subscription/singbox_test.go
git commit -m "feat(panel): add sing-box subscription builder"
```

---

### Task 4: 订阅 API（UA/format 输出）

**Files:**
- Create: `panel/internal/api/subscription.go`
- Create: `panel/internal/api/subscription_test.go`
- Modify: `panel/internal/api/router.go`

**Step 1: 写失败测试（UA/format 覆盖）**

```go
func TestSubscriptionUAAndFormat(t *testing.T) {
  store := setupStore(t)
  userUUID := seedSubscriptionData(t, store)

  r := api.NewRouter(config.Config{}, store)

  req := httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID+"?format=singbox", nil)
  req.Header.Set("User-Agent", "clash-meta")
  w := httptest.NewRecorder()
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
  require.Contains(t, w.Header().Get("Content-Type"), "application/json")

  req = httptest.NewRequest(http.MethodGet, "/api/sub/"+userUUID, nil)
  req.Header.Set("User-Agent", "Shadowrocket")
  w = httptest.NewRecorder()
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
}
```

**Step 2: 运行测试，确认失败**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: FAIL

**Step 3: 写最小实现**
- 订阅 handler：
  - 查询 user by uuid
  - 校验 status
  - 查询 user_nodes -> inbounds
  - 调用 subscription.BuildSingbox
  - `format=singbox` 优先
  - UA 命中 `sing-box/SFA/SFI` 返回 JSON，否则 Base64
- 新增路由：`GET /api/sub/:user_uuid`

**Step 4: 运行测试，确认通过**

Run: `HTTP_PROXY=... HTTPS_PROXY=... GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go/pkg/mod go test ./internal/api -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/api/subscription.go panel/internal/api/subscription_test.go panel/internal/api/router.go
git commit -m "feat(panel): add subscription endpoint"
```

---

### Task 5: README 文档补充

**Files:**
- Modify: `README.md`

**Step 1: 写失败测试（可选，若无文档测试则跳过）**

**Step 2: 修改 README**
- 增加订阅端点说明
- UA/format 行为说明
- `nodes` 新字段与 `inbounds.public_port` 说明

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add subscription usage"
```

---

## 执行方式

Plan 完成后，进入执行：
- 使用 `superpowers:executing-plans` 在该 worktree 内逐任务执行
- 每 3 个任务一个批次回报
