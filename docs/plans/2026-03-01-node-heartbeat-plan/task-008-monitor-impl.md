# Task 008: NodesMonitor 适配 — 实现

**depends-on:** task-008-monitor-test

## Summary

修改 NodesMonitor 的 `CheckOnce` 方法，跳过 `status=pending` 的节点。

## BDD Scenario

(同 task-008-monitor-test)

## Files

- **Modify:** `panel/internal/monitor/nodes_monitor.go`
  - 在 `CheckOnce` 的 for 循环中，添加 `if n.Status == "pending" { continue }` 跳过 pending 节点
  - 或者修改 `ListNodes` 查询，添加 `WHERE status != 'pending'` 过滤

- **选择方案:** 在 Monitor 层过滤（而非 DB 层），因为:
  1. `ListNodes` 是通用查询，前端 UI 需要显示 pending 节点
  2. Monitor 是唯一不应处理 pending 节点的组件

## Steps

1. 在 `CheckOnce` 中添加 pending 过滤
2. 确保现有测试不受影响

## Verify

```bash
cd panel && go test ./internal/monitor/ -v
```
