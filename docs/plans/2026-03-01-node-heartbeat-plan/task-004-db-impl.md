# Task 004: Panel DB 层扩展 — 实现

**depends-on:** task-004-db-test

## Summary

在 Panel DB 层新增 Node 相关方法，支持按 UUID 查询、创建 pending 节点、审批节点。

## BDD Scenario

```gherkin
Scenario: Panel creates pending Node record
  Given Panel has no Node record with uuid="uuid-1"
  When CreatePendingNode is called with uuid, secret_key, api_address, api_port
  Then a new Node record is created with status="pending" and name="" (empty)

Scenario: Admin approves a pending Node
  Given Panel has a pending Node
  When ApproveNode is called with name and optional group_id
  Then Node status changes to "offline", name and group_id are set
```

## Files

- **Modify:** `panel/internal/db/nodes.go`
  - `GetNodeByUUID(ctx, uuid string) (Node, error)` — 按 UUID 查询 Node
  - `CreatePendingNode(ctx, PendingNodeParams) (Node, error)` — 创建 status="pending" 的 Node
    - `PendingNodeParams`: UUID, SecretKey, APIAddress, APIPort
    - Name 设为空字符串（审批时填写）
  - `ApproveNode(ctx, id int64, params ApproveNodeParams) (Node, error)` — pending→offline
    - `ApproveNodeParams`: Name, GroupID, PublicAddress
    - 只有 status="pending" 的节点可以被审批
  - `UpdateNodeLastSeen(ctx, uuid string, seenAt time.Time) error` — 按 UUID 更新 last_seen_at

- **Modify:** `panel/internal/db/nodes.go` 中的 `DeleteNode`
  - pending 节点没有 inbound，应该可以直接删除（现有逻辑已支持）

## Steps

1. 实现 `GetNodeByUUID` — 与 `GetNodeByID` 类似但 WHERE 条件用 uuid
2. 实现 `CreatePendingNode` — INSERT 带 status='pending'
3. 实现 `ApproveNode` — UPDATE SET status='offline', name=?, group_id=? WHERE id=? AND status='pending'
4. 实现 `UpdateNodeLastSeen` — UPDATE last_seen_at WHERE uuid=?

## Verify

```bash
cd panel && go test ./internal/db/ -run "TestGetNodeByUUID|TestCreatePendingNode|TestApproveNode" -v
```
