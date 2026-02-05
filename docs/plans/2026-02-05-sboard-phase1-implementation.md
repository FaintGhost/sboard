# SBoard Phase 1 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 搭建 Panel/Node 的最小可运行框架，支持 Node 接收配置同步并创建 sing-box 入站。

**Architecture:** 根目录 `go.work` 统一管理 `panel` 与 `node` 两个 Go module。Panel 负责 SQLite 迁移与健康检查；Node 内嵌 sing-box `Box`，暴露 `/api/health` 与 `/api/config/sync`，配置同步后使用 Inbound 重建实现原子更新。

**Tech Stack:** Go、Gin、SQLite、golang-migrate、sing-box v1.12.19

---

### Task 1: 初始化 Go Workspace 与模块骨架

**Files:**
- Create: `go.work`
- Create: `panel/go.mod`
- Create: `node/go.mod`
- Create: `panel/cmd/panel/main.go`
- Create: `node/cmd/node/main.go`

**Step 1: 创建 `go.work`**

```txt
go 1.25

use (
  ./panel
  ./node
)
```

**Step 2: 初始化 `panel/go.mod`**

```txt
module sboard/panel

go 1.25

require (
  github.com/gin-gonic/gin v1.11.0
  github.com/golang-jwt/jwt/v5 v5.2.2
  github.com/mattn/go-sqlite3 v1.14.32
  github.com/golang-migrate/migrate/v4 v4.18.1
)

require (
  github.com/golang-migrate/migrate/v4/database/sqlite3 v4.18.1
  github.com/golang-migrate/migrate/v4/source/file v4.18.1
)
```

**Step 3: 初始化 `node/go.mod`**

```txt
module sboard/node

go 1.25

require (
  github.com/gin-gonic/gin v1.11.0
  github.com/sagernet/sing-box v1.12.19
)
```

**Step 4: 写入最小 `main.go`（占位）**

```go
package main

func main() {}
```

**Step 5: Commit**

```bash
git add go.work panel/go.mod node/go.mod panel/cmd/panel/main.go node/cmd/node/main.go
git commit -m "chore: init workspace and modules"
```

---

### Task 2: Panel 配置与数据库迁移（含测试）

**Files:**
- Create: `panel/internal/config/config.go`
- Create: `panel/internal/db/db.go`
- Create: `panel/internal/db/migrate.go`
- Create: `panel/internal/db/migrations/0001_init.up.sql`
- Create: `panel/internal/db/migrations/0001_init.down.sql`
- Create: `panel/internal/db/migrate_test.go`

**Step 1: 写失败测试（迁移可执行）**

```go
func TestMigrateUp(t *testing.T) {
  dir := t.TempDir()
  dbPath := filepath.Join(dir, "test.db")
  db, err := db.Open(dbPath)
  require.NoError(t, err)
  t.Cleanup(func() { _ = db.Close() })

  err = db.MigrateUp(db, filepath.Join("..", "..", "..", "internal", "db", "migrations"))
  require.NoError(t, err)

  rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='users'")
  require.NoError(t, err)
  defer rows.Close()
  require.True(t, rows.Next())
}
```

**Step 2: 运行测试，确认失败**

Run: `go test ./internal/db -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

`panel/internal/config/config.go`
```go
package config

import "os"

type Config struct {
  HTTPAddr string
  DBPath   string
}

func Load() Config {
  cfg := Config{
    HTTPAddr: ":8080",
    DBPath:   "panel.db",
  }
  if v := os.Getenv("PANEL_HTTP_ADDR"); v != "" {
    cfg.HTTPAddr = v
  }
  if v := os.Getenv("PANEL_DB_PATH"); v != "" {
    cfg.DBPath = v
  }
  return cfg
}
```

`panel/internal/db/db.go`
```go
package db

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
  return sql.Open("sqlite3", path)
}
```

`panel/internal/db/migrate.go`
```go
package db

import (
  "database/sql"
  "fmt"

  "github.com/golang-migrate/migrate/v4"
  "github.com/golang-migrate/migrate/v4/database/sqlite3"
  _ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateUp(db *sql.DB, dir string) error {
  driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
  if err != nil {
    return err
  }
  m, err := migrate.NewWithDatabaseInstance(
    "file://"+dir,
    "sqlite3",
    driver,
  )
  if err != nil {
    return err
  }
  if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    return fmt.Errorf("migrate up: %w", err)
  }
  return nil
}
```

`panel/internal/db/migrations/0001_init.up.sql`
```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT,
  traffic_limit INTEGER DEFAULT 0,
  traffic_used INTEGER DEFAULT 0,
  traffic_reset_day INTEGER DEFAULT 0,
  expire_at DATETIME,
  status TEXT DEFAULT 'active',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE nodes (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  address TEXT NOT NULL,
  port INTEGER NOT NULL,
  secret_key TEXT NOT NULL,
  status TEXT DEFAULT 'offline',
  last_seen_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inbounds (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  tag TEXT NOT NULL,
  node_id INTEGER REFERENCES nodes(id),
  protocol TEXT NOT NULL,
  listen_port INTEGER NOT NULL,
  settings JSON NOT NULL,
  tls_settings JSON,
  transport_settings JSON,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(node_id, tag)
);

CREATE TABLE user_inbounds (
  user_id INTEGER REFERENCES users(id),
  inbound_id INTEGER REFERENCES inbounds(id),
  PRIMARY KEY (user_id, inbound_id)
);

CREATE TABLE traffic_stats (
  id INTEGER PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  node_id INTEGER REFERENCES nodes(id),
  upload INTEGER DEFAULT 0,
  download INTEGER DEFAULT 0,
  recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

`panel/internal/db/migrations/0001_init.down.sql`
```sql
DROP TABLE IF EXISTS traffic_stats;
DROP TABLE IF EXISTS user_inbounds;
DROP TABLE IF EXISTS inbounds;
DROP TABLE IF EXISTS nodes;
DROP TABLE IF EXISTS users;
```

**Step 4: 运行测试，确认通过**

Run: `go test ./internal/db -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/config panel/internal/db panel/internal/db/migrations
git commit -m "feat(panel): add config and migrations"
```

---

### Task 3: Panel Health API（含测试）

**Files:**
- Create: `panel/internal/api/router.go`
- Create: `panel/internal/api/health.go`
- Create: `panel/internal/api/health_test.go`

**Step 1: 写失败测试**

```go
func TestHealth(t *testing.T) {
  r := api.NewRouter()
  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/health", nil)

  r.ServeHTTP(w, req)

  require.Equal(t, http.StatusOK, w.Code)
  require.Contains(t, w.Body.String(), "ok")
}
```

**Step 2: 运行测试，确认失败**

Run: `go test ./internal/api -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

`panel/internal/api/router.go`
```go
package api

import "github.com/gin-gonic/gin"

func NewRouter() *gin.Engine {
  r := gin.New()
  r.GET("/api/health", Health)
  return r
}
```

`panel/internal/api/health.go`
```go
package api

import "github.com/gin-gonic/gin"

func Health(c *gin.Context) {
  c.JSON(200, gin.H{"status": "ok"})
}
```

**Step 4: 运行测试，确认通过**

Run: `go test ./internal/api -v`
Expected: PASS

**Step 5: Commit**

```bash
git add panel/internal/api
git commit -m "feat(panel): add health endpoint"
```

---

### Task 4: Node 配置解析与校验（含测试）

**Files:**
- Create: `node/internal/config/config.go`
- Create: `node/internal/sync/parse.go`
- Create: `node/internal/sync/parse_test.go`

**Step 1: 写失败测试（解析 + 校验）**

```go
func TestParseAndValidateInbounds(t *testing.T) {
  ctx := sync.NewSingboxContext()
  body := []byte(`{
    "inbounds": [
      {"type":"mixed","tag":"m1","listen":"0.0.0.0","listen_port":1080}
    ]
  }`)

  inbounds, err := sync.ParseAndValidateInbounds(ctx, body)
  require.NoError(t, err)
  require.Len(t, inbounds, 1)
  require.Equal(t, "mixed", inbounds[0].Type)
  require.Equal(t, "m1", inbounds[0].Tag)
}
```

**Step 2: 运行测试，确认失败**

Run: `go test ./internal/sync -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

`node/internal/config/config.go`
```go
package config

import "os"

type Config struct {
  HTTPAddr  string
  SecretKey string
  LogLevel  string
}

func Load() Config {
  cfg := Config{
    HTTPAddr:  ":3000",
    SecretKey: "",
    LogLevel:  "info",
  }
  if v := os.Getenv("NODE_HTTP_ADDR"); v != "" {
    cfg.HTTPAddr = v
  }
  if v := os.Getenv("NODE_SECRET_KEY"); v != "" {
    cfg.SecretKey = v
  }
  if v := os.Getenv("NODE_LOG_LEVEL"); v != "" {
    cfg.LogLevel = v
  }
  return cfg
}
```

`node/internal/sync/parse.go`
```go
package sync

import (
  "context"
  "encoding/json"
  "fmt"

  sbjson "github.com/sagernet/sing-box/common/json"
  "github.com/sagernet/sing-box/option"
  "github.com/sagernet/sing/service"
)

type inboundMeta struct {
  Tag        string `json:"tag"`
  Type       string `json:"type"`
  Listen     string `json:"listen"`
  ListenPort int    `json:"listen_port"`
}

type syncRequest struct {
  Inbounds []json.RawMessage `json:"inbounds"`
}

func NewSingboxContext() context.Context {
  return service.ContextWithDefaultRegistry(context.Background())
}

func ParseAndValidateInbounds(ctx context.Context, body []byte) ([]option.Inbound, error) {
  var req syncRequest
  if err := json.Unmarshal(body, &req); err != nil {
    return nil, err
  }

  inbounds := make([]option.Inbound, 0, len(req.Inbounds))
  seen := map[string]struct{}{}
  for i, raw := range req.Inbounds {
    var meta inboundMeta
    if err := json.Unmarshal(raw, &meta); err != nil {
      return nil, err
    }
    if meta.Tag == "" {
      return nil, fmt.Errorf("inbounds[%d].tag required", i)
    }
    if meta.Type == "" {
      return nil, fmt.Errorf("inbounds[%d].type required", i)
    }
    if meta.ListenPort <= 0 || meta.ListenPort > 65535 {
      return nil, fmt.Errorf("inbounds[%d].listen_port invalid", i)
    }
    if _, ok := seen[meta.Tag]; ok {
      return nil, fmt.Errorf("inbounds[%d].tag duplicated", i)
    }
    seen[meta.Tag] = struct{}{}

    var inb option.Inbound
    if err := sbjson.UnmarshalContext(ctx, raw, &inb); err != nil {
      return nil, err
    }
    inbounds = append(inbounds, inb)
  }
  return inbounds, nil
}
```

**Step 4: 运行测试，确认通过**

Run: `go test ./internal/sync -v`
Expected: PASS

**Step 5: Commit**

```bash
git add node/internal/config node/internal/sync
git commit -m "feat(node): parse and validate config sync"
```

---

### Task 5: Node Core 入站应用逻辑（含测试）

**Files:**
- Create: `node/internal/core/apply.go`
- Create: `node/internal/core/apply_test.go`

**Step 1: 写失败测试**

```go
type fakeInboundManager struct {
  calls int
}

func (f *fakeInboundManager) Create(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag, inboundType string, options any) error {
  f.calls++
  return nil
}

func TestApplyInbounds(t *testing.T) {
  mgr := &fakeInboundManager{}
  inbounds := []option.Inbound{{Type: "mixed", Tag: "m1", Options: struct{}{}}}
  err := core.ApplyInbounds(context.Background(), nil, nil, mgr, inbounds)
  require.NoError(t, err)
  require.Equal(t, 1, mgr.calls)
}
```

**Step 2: 运行测试，确认失败**

Run: `go test ./internal/core -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

`node/internal/core/apply.go`
```go
package core

import (
  "context"

  "github.com/sagernet/sing-box/adapter"
  "github.com/sagernet/sing-box/log"
  "github.com/sagernet/sing-box/option"
)

type InboundCreator interface {
  Create(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag, inboundType string, options any) error
}

type LoggerFactory func(typ, tag string) log.ContextLogger

func ApplyInbounds(ctx context.Context, router adapter.Router, loggerFactory LoggerFactory, mgr InboundCreator, inbounds []option.Inbound) error {
  for i := range inbounds {
    inb := inbounds[i]
    var lg log.ContextLogger
    if loggerFactory != nil {
      lg = loggerFactory(inb.Type, inb.Tag)
    }
    if err := mgr.Create(ctx, router, lg, inb.Tag, inb.Type, inb.Options); err != nil {
      return err
    }
  }
  return nil
}
```

**Step 4: 运行测试，确认通过**

Run: `go test ./internal/core -v`
Expected: PASS

**Step 5: Commit**

```bash
git add node/internal/core/apply.go node/internal/core/apply_test.go
git commit -m "feat(node): add inbound apply helper"
```

---

### Task 6: Node API（Health + Config Sync）含测试

**Files:**
- Create: `node/internal/api/router.go`
- Create: `node/internal/api/health.go`
- Create: `node/internal/api/config_sync.go`
- Create: `node/internal/api/router_test.go`

**Step 1: 写失败测试（Health）**

```go
func TestHealth(t *testing.T) {
  r := api.NewRouter("secret", nil)
  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusOK, w.Code)
}
```

**Step 2: 写失败测试（Config Sync 401）**

```go
func TestConfigSyncAuth(t *testing.T) {
  r := api.NewRouter("secret", &fakeCore{})
  w := httptest.NewRecorder()
  req := httptest.NewRequest(http.MethodPost, "/api/config/sync", strings.NewReader(`{"inbounds":[]}`))
  r.ServeHTTP(w, req)
  require.Equal(t, http.StatusUnauthorized, w.Code)
}
```

**Step 3: 运行测试，确认失败**

Run: `go test ./internal/api -v`
Expected: FAIL（缺少实现）

**Step 4: 写最小实现**

`node/internal/api/router.go`
```go
package api

import "github.com/gin-gonic/gin"

type Core interface {
  ApplyConfig(ctx *gin.Context, body []byte) error
}

func NewRouter(secret string, core Core) *gin.Engine {
  r := gin.New()
  r.GET("/api/health", Health)
  r.POST("/api/config/sync", func(c *gin.Context) {
    ConfigSync(c, secret, core)
  })
  return r
}
```

`node/internal/api/health.go`
```go
package api

import "github.com/gin-gonic/gin"

func Health(c *gin.Context) {
  c.JSON(200, gin.H{"status": "ok"})
}
```

`node/internal/api/config_sync.go`
```go
package api

import (
  "io"
  "net/http"
  "strings"

  "github.com/gin-gonic/gin"
)

func ConfigSync(c *gin.Context, secret string, core Core) {
  auth := c.GetHeader("Authorization")
  if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != secret {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    return
  }
  body, err := io.ReadAll(c.Request.Body)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
    return
  }
  if core == nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "core not ready"})
    return
  }
  if err := core.ApplyConfig(c, body); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

`node/internal/api/router_test.go`
```go
type fakeCore struct{ err error }

func (f *fakeCore) ApplyConfig(ctx *gin.Context, body []byte) error { return f.err }
```

**Step 5: 运行测试，确认通过**

Run: `go test ./internal/api -v`
Expected: PASS

**Step 6: Commit**

```bash
git add node/internal/api
git commit -m "feat(node): add health and config sync endpoints"
```

---

### Task 7: Node Core 真实实现（含测试）

**Files:**
- Create: `node/internal/core/core.go`
- Create: `node/internal/core/core_test.go`
- Modify: `node/cmd/node/main.go`

**Step 1: 写失败测试（newBox 可替换）**

```go
type fakeBox struct{ started bool }

func (f *fakeBox) Start() error { f.started = true; return nil }
func (f *fakeBox) Inbound() adapter.InboundManager { return nil }
func (f *fakeBox) Router() adapter.Router { return nil }
func (f *fakeBox) Close() error { return nil }

func TestNewCoreUsesNewBox(t *testing.T) {
  old := core.NewBox
  t.Cleanup(func() { core.NewBox = old })

  core.NewBox = func(opts box.Options) (core.Box, error) {
    return &fakeBox{}, nil
  }

  c, err := core.New(context.Background(), "info")
  require.NoError(t, err)
  require.NotNil(t, c)
}
```

**Step 2: 运行测试，确认失败**

Run: `go test ./internal/core -v`
Expected: FAIL（缺少实现）

**Step 3: 写最小实现**

`node/internal/core/core.go`
```go
package core

import (
  "context"
  "crypto/sha256"
  "encoding/hex"
  "time"

  "github.com/sagernet/sing-box/adapter"
  "github.com/sagernet/sing-box/box"
  "github.com/sagernet/sing-box/log"
  "github.com/sagernet/sing-box/option"
)

type Box interface {
  Start() error
  Inbound() adapter.InboundManager
  Router() adapter.Router
  Close() error
}

var NewBox = func(opts box.Options) (Box, error) { return box.New(opts) }

type Core struct {
  ctx        context.Context
  box        Box
  logFactory *log.Factory
  hash       string
  at         time.Time
}

func New(ctx context.Context, logLevel string) (*Core, error) {
  opts := option.Options{
    Log: &option.LogOptions{Level: logLevel},
    Route: &option.RouteOptions{Final: "direct"},
    DNS: &option.DNSOptions{Final: "direct"},
  }
  lf, err := log.New(log.Options{Context: ctx, Options: opts.Log})
  if err != nil {
    return nil, err
  }
  b, err := NewBox(box.Options{Options: opts, Context: ctx})
  if err != nil {
    return nil, err
  }
  if err := b.Start(); err != nil {
    return nil, err
  }
  return &Core{ctx: ctx, box: b, logFactory: lf}, nil
}

func (c *Core) Apply(inbounds []option.Inbound, raw []byte) error {
  loggerFactory := func(typ, tag string) log.ContextLogger {
    return c.logFactory.NewLogger(c.ctx, tag)
  }
  if err := ApplyInbounds(c.ctx, c.box.Router(), loggerFactory, c.box.Inbound(), inbounds); err != nil {
    return err
  }
  sum := sha256.Sum256(raw)
  c.hash = hex.EncodeToString(sum[:])
  c.at = time.Now()
  return nil
}
```

**Step 4: 运行测试，确认通过**

Run: `go test ./internal/core -v`
Expected: PASS

**Step 5: 修改 `node/cmd/node/main.go` 连接 API**

```go
package main

import (
  "log"

  "sboard/node/internal/api"
  "sboard/node/internal/config"
  "sboard/node/internal/core"
  "sboard/node/internal/sync"
  "github.com/gin-gonic/gin"
)

type coreAdapter struct{ c *core.Core }

func (a *coreAdapter) ApplyConfig(ctx *gin.Context, body []byte) error {
  sbctx := sync.NewSingboxContext()
  inbounds, err := sync.ParseAndValidateInbounds(sbctx, body)
  if err != nil {
    return err
  }
  return a.c.Apply(inbounds, body)
}

func main() {
  cfg := config.Load()
  sbctx := sync.NewSingboxContext()
  c, err := core.New(sbctx, cfg.LogLevel)
  if err != nil {
    log.Fatal(err)
  }
  r := api.NewRouter(cfg.SecretKey, &coreAdapter{c: c})
  if err := r.Run(cfg.HTTPAddr); err != nil {
    log.Fatal(err)
  }
}
```

**Step 6: Commit**

```bash
git add node/internal/core/core.go node/internal/core/core_test.go node/cmd/node/main.go
git commit -m "feat(node): wire sing-box core"
```

---

### Task 8: Panel 启动与路由连接（无测试）

**Files:**
- Modify: `panel/cmd/panel/main.go`

**Step 1: 实现 Panel 启动流程**

```go
package main

import (
  "log"

  "sboard/panel/internal/api"
  "sboard/panel/internal/config"
  "sboard/panel/internal/db"
)

func main() {
  cfg := config.Load()
  database, err := db.Open(cfg.DBPath)
  if err != nil {
    log.Fatal(err)
  }
  if err := db.MigrateUp(database, "internal/db/migrations"); err != nil {
    log.Fatal(err)
  }
  r := api.NewRouter()
  if err := r.Run(cfg.HTTPAddr); err != nil {
    log.Fatal(err)
  }
}
```

**Step 2: Commit**

```bash
git add panel/cmd/panel/main.go
git commit -m "feat(panel): wire server startup"
```

---

### Task 9: 最小集成检查（手动）

**Files:**
- None

**Step 1: 启动 Panel**

Run: `cd panel && go run ./cmd/panel`
Expected: `/api/health` 返回 200

**Step 2: 启动 Node**

Run: `cd node && NODE_SECRET_KEY=secret go run ./cmd/node`
Expected: `/api/health` 返回 200

**Step 3: Config Sync**

Run:
```bash
curl -X POST http://127.0.0.1:3000/api/config/sync \
  -H "Authorization: Bearer secret" \
  -d '{"inbounds":[{"type":"mixed","tag":"m1","listen":"0.0.0.0","listen_port":1080}]}'
```
Expected: `{"status":"ok"}`

---
