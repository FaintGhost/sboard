# Task 005: Panel Heartbeat Handler — 测试

**depends-on:** task-001-proto-codegen, task-004-db-impl

## Summary

为 Panel 的 Heartbeat RPC handler 编写测试，覆盖已知节点、未知节点、SecretKey 不匹配、重复 pending 四种路径。

## BDD Scenario

```gherkin
Scenario: Panel receives heartbeat from known Node
  Given Panel has a Node record with uuid="uuid-1" and secret_key="key-1"
  When Node sends Heartbeat with uuid="uuid-1" and secret_key="key-1"
  Then Panel updates last_seen_at for that Node
  And responds with status=RECOGNIZED

Scenario: Panel rejects heartbeat with mismatched secret_key
  Given Panel has a Node record with uuid="uuid-1" and secret_key="key-1"
  When Node sends Heartbeat with uuid="uuid-1" and secret_key="wrong-key"
  Then Panel responds with status=REJECTED
  And does NOT update any Node record

Scenario: Panel receives heartbeat from unknown Node (after reset)
  Given Panel has no Node record with uuid="uuid-1"
  When Node sends Heartbeat with uuid="uuid-1" and secret_key="key-1" and api_addr=":3003"
  Then Panel creates a pending Node record with status="pending"
  And responds with status=PENDING

Scenario: Duplicate heartbeat from pending Node
  Given Panel has a pending Node with uuid="uuid-1"
  When Node sends another Heartbeat with uuid="uuid-1"
  Then Panel updates last_seen_at on the pending record
  And responds with status=PENDING (no duplicate record created)
```

## Files

- **Create:** `panel/internal/rpc/node_registration_test.go`
  - `TestHeartbeat_KnownNode` — 已知节点，SecretKey 匹配 → RECOGNIZED
  - `TestHeartbeat_KnownNodeWrongKey` — 已知节点，SecretKey 不匹配 → REJECTED
  - `TestHeartbeat_UnknownNode` — 未知节点 → 创建 pending → PENDING
  - `TestHeartbeat_DuplicatePending` — 已有 pending → 更新 last_seen_at → PENDING
  - 使用 `httptest.NewServer` + `NewHandler` 构建测试服务器
  - 使用 `panelv1connect.NewNodeRegistrationServiceClient` 发起 RPC

## Steps

1. 参考 `panel/internal/rpc/server_test.go` 中已有的 `setupRPCStore` 测试辅助
2. 每个测试构建 httptest server，通过 generated Connect client 调用 Heartbeat
3. 验证响应状态和 DB 中的记录变化
4. 测试应该先 FAIL（handler 不存在）

## Verify

```bash
cd panel && go test ./internal/rpc/ -run TestHeartbeat -v
```

应该 FAIL（handler 不存在）。
