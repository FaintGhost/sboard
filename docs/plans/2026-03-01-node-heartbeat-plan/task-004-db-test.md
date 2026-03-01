# Task 004: Panel DB 层扩展 — 测试

**depends-on:** (none)

## Summary

为 Panel DB 层新增的 Node 查询和创建方法编写测试，涵盖 `GetNodeByUUID`、`CreatePendingNode`、`ApproveNode`、`DeleteNode`（pending 状态）。

## BDD Scenario

```gherkin
Scenario: Panel receives heartbeat from unknown Node (after reset)
  Given Panel has no Node record with uuid="uuid-1"
  When Node sends Heartbeat with uuid="uuid-1" and secret_key="key-1" and api_addr=":3003"
  Then Panel creates a pending Node record with status="pending"

Scenario: Duplicate heartbeat from pending Node
  Given Panel has a pending Node with uuid="uuid-1"
  When Node sends another Heartbeat with uuid="uuid-1"
  Then Panel updates last_seen_at on the pending record (no duplicate)

Scenario: Admin approves a pending Node
  Given Panel has a pending Node with uuid="uuid-1" and api_address="1.2.3.4" and api_port=3003
  When admin calls ApproveNode with name="US-West-1" and group_id=1
  Then Node status changes from "pending" to "offline"
  And Node name is set to "US-West-1"
```

## Files

- **Create:** `panel/internal/db/nodes_pending_test.go`
  - `TestGetNodeByUUID_Found` — UUID 匹配时返回正确 Node
  - `TestGetNodeByUUID_NotFound` — UUID 不存在时返回 ErrNotFound
  - `TestCreatePendingNode` — 创建 pending 节点，验证 status="pending"
  - `TestCreatePendingNode_DuplicateUUID` — 重复 UUID 返回已有记录而非报错
  - `TestApproveNode` — pending→offline 状态转换，name/group_id 被设置
  - `TestApproveNode_NotPending` — 非 pending 节点审批返回错误
  - `TestDeleteNode_Pending` — pending 节点无 inbound 约束可直接删除
  - 使用 `setupStore(t)` 测试辅助函数（参考现有 `panel/internal/db/` 测试模式）

## Steps

1. 参考现有 `panel/internal/db/nodes.go` 的测试模式（如果有）或 `panel/internal/api/*_test.go` 中的 `setupStore`
2. 每个测试创建临时 SQLite 数据库
3. 测试应该先 FAIL（方法不存在）

## Verify

```bash
cd panel && go test ./internal/db/ -run TestGetNodeByUUID -v && go test ./internal/db/ -run TestCreatePendingNode -v && go test ./internal/db/ -run TestApproveNode -v
```

应该 FAIL（方法不存在）。
