# Task 008: NodesMonitor 适配 — 测试

**depends-on:** task-004-db-impl

## Summary

为 NodesMonitor 添加测试，验证它跳过 `status=pending` 的节点（pending 节点由心跳管理，不由 Monitor 轮询）。

## BDD Scenario

```gherkin
Scenario: NodesMonitor skips pending nodes
  Given Panel has nodes: Node-A (status="offline"), Node-B (status="pending")
  When NodesMonitor runs CheckOnce
  Then it calls Health on Node-A
  And it does NOT call Health on Node-B
  And Node-B remains in "pending" status
```

## Files

- **Create:** `panel/internal/monitor/nodes_monitor_pending_test.go`
  - `TestNodesMonitor_SkipsPendingNodes` — pending 节点不被轮询
  - `TestNodesMonitor_ApprovedNodeGetsPolled` — 审批后（offline）的节点被正常轮询
  - 使用 mock NodeClient 记录 Health 调用

## Steps

1. 参考现有 `panel/internal/monitor/` 测试模式
2. 创建包含 pending 和 offline 节点的测试数据
3. 验证 mock client 只收到 offline 节点的 Health 调用
4. 测试应该先 FAIL（Monitor 当前不过滤 pending）

## Verify

```bash
cd panel && go test ./internal/monitor/ -run TestNodesMonitor_SkipsPending -v
```

应该 FAIL（Monitor 不过滤 pending）。
