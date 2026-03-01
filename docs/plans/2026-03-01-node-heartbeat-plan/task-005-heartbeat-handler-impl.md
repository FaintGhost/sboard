# Task 005: Panel Heartbeat Handler — 实现

**depends-on:** task-005-heartbeat-handler-test

## Summary

实现 Panel 的 `NodeRegistrationService.Heartbeat` RPC handler，并在 RPC server 中注册为公开端点。

## BDD Scenario

(同 task-005-heartbeat-handler-test)

## Files

- **Create:** `panel/internal/rpc/node_registration.go`
  - 实现 `Heartbeat` 方法，逻辑:
    1. 按 UUID 查找 Node
    2. 找到 → 验证 SecretKey → 匹配则更新 last_seen_at 返回 RECOGNIZED，不匹配返回 REJECTED
    3. 未找到 → 检查是否已有 pending → 有则更新 last_seen_at → 无则创建 pending
  - 使用 `crypto/subtle.ConstantTimeCompare` 比较 SecretKey

- **Modify:** `panel/internal/rpc/server.go`
  - 在 `NewHandler` 中注册 `NodeRegistrationService` handler
  - 在 `public` map 中添加 Heartbeat 端点路径，使其免鉴权

## Steps

1. 创建 `node_registration.go`，实现 `Heartbeat` 方法
2. 在 `server.go` 的 `NewHandler` 中注册 `NodeRegistrationServiceHandler`
3. 在 `public` map 中添加 `panelv1connect.NodeRegistrationServiceHeartbeatProcedure`

## Verify

```bash
cd panel && go test ./internal/rpc/ -run TestHeartbeat -v
```
