# Task 007: Panel 审批/拒绝 RPC — 实现

**depends-on:** task-007-approve-reject-test

## Summary

在 Panel RPC 层实现 `ApproveNode` 和 `RejectNode` 方法。

## BDD Scenario

(同 task-007-approve-reject-test)

## Files

- **Modify:** `panel/internal/rpc/services_impl.go`
  - 实现 `ApproveNode(ctx, req)`:
    1. 验证 name 非空
    2. 调用 `store.ApproveNode(ctx, id, params)` 将 pending→offline
    3. 返回更新后的 Node 数据
  - 实现 `RejectNode(ctx, req)`:
    1. 获取 Node，验证 status="pending"
    2. 调用 `store.DeleteNode(ctx, id)` 删除记录
    3. 返回成功

## Steps

1. 在 `services_impl.go` 中添加 `ApproveNode` 和 `RejectNode` 方法
2. ApproveNode 需要验证输入（name 必填）
3. RejectNode 只对 pending 状态的节点生效
4. 错误映射: `ErrNotFound` → `connect.CodeNotFound`, 非 pending → `connect.CodeInvalidArgument`

## Verify

```bash
cd panel && go test ./internal/rpc/ -run "TestApproveNode|TestRejectNode" -v
```
