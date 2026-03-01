# Task 003: Node UUID 持久化 — 实现

**depends-on:** task-003-uuid-test

## Summary

实现 `LoadOrGenerateUUID` 函数和 Node Config 扩展，支持 `PANEL_URL`、`NODE_HEARTBEAT_INTERVAL`、`NODE_UUID` 环境变量。

## BDD Scenario

```gherkin
Scenario: Node does NOT send heartbeat when PANEL_URL is empty
  Given a Node running without PANEL_URL configured
  When 60 seconds elapse
  Then no Heartbeat RPC is sent
  And Node operates in passive-only mode (current behavior)
```

## Files

- **Create:** `node/internal/config/uuid.go`
  - `LoadOrGenerateUUID(path string) (string, error)` — 从文件加载或生成新 UUID
  - 使用 `github.com/google/uuid` 生成 UUIDv4
  - 原子写入（临时文件 + rename），权限 `0600`
  - 空/损坏文件则重新生成

- **Modify:** `node/internal/config/config.go`
  - `Config` 结构体新增: `PanelURL string`, `HeartbeatIntervalS int`, `NodeUUID string`
  - `Load()` 函数新增环境变量读取: `PANEL_URL`, `NODE_HEARTBEAT_INTERVAL`, `NODE_UUID`
  - 新增方法 `HeartbeatInterval() time.Duration`（默认 30s，最小 5s）
  - `Load()` 中：如果 `NODE_UUID` 为空且 `PANEL_URL` 非空，调用 `LoadOrGenerateUUID` 自动生成

## Steps

1. 创建 `uuid.go`，实现 `LoadOrGenerateUUID`
2. 修改 `config.go`，扩展 Config 结构体和 Load 函数
3. UUID 持久化路径默认为 `/data/node-uuid`（与 `last_sync.json` 同目录）

## Verify

```bash
cd node && go test ./internal/config/ -v
```
