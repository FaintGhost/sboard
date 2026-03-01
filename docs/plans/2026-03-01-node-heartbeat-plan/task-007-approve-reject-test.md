# Task 007: Panel 审批/拒绝 RPC — 测试

**depends-on:** task-001-proto-codegen, task-004-db-impl

## Summary

为 Panel 的 `ApproveNode` 和 `RejectNode` RPC 编写测试。

## BDD Scenario

```gherkin
Scenario: Admin approves a pending Node
  Given Panel has a pending Node with uuid="uuid-1" and api_address="1.2.3.4" and api_port=3003
  When admin calls ApproveNode with name="US-West-1" and group_id=1
  Then Node status changes from "pending" to "offline"
  And Node name is set to "US-West-1"
  And Node group_id is set to 1
  And NodesMonitor will detect the Node as online on next check
  And SyncConfig is triggered automatically

Scenario: Admin rejects a pending Node
  Given Panel has a pending Node with uuid="uuid-1"
  When admin calls RejectNode for uuid="uuid-1"
  Then the pending Node record is deleted
```

## Files

- **Create:** `panel/internal/rpc/node_approval_test.go`
  - `TestApproveNode_Success` — pending 节点审批成功 → status="offline"、name/group_id 被设置
  - `TestApproveNode_NotPending` — 非 pending 节点审批 → 返回 InvalidArgument 错误
  - `TestApproveNode_NotFound` — 不存在的 ID → 返回 NotFound 错误
  - `TestApproveNode_EmptyName` — 空 name → 返回 InvalidArgument（name 必填）
  - `TestRejectNode_Success` — pending 节点拒绝 → 记录被删除
  - `TestRejectNode_NotFound` — 不存在的 ID → 返回 NotFound 错误
  - 使用 `httptest.NewServer` + `NewHandler` 构建测试服务器
  - 使用带 JWT token 的 Connect client（这些是受保护端点）

## Steps

1. 参考 `panel/internal/rpc/server_test.go` 中的测试辅助（`setupRPCStore`, `bearerInterceptor`, `mustJWT`）
2. 先创建 pending 节点（通过直接调用 DB 或 Heartbeat），然后测试审批/拒绝
3. 验证 DB 状态变化
4. 测试应该先 FAIL（RPC 方法未实现）

## Verify

```bash
cd panel && go test ./internal/rpc/ -run "TestApproveNode|TestRejectNode" -v
```

应该 FAIL（方法未实现）。
