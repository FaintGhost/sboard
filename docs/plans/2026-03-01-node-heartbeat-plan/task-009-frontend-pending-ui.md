# Task 009: 前端待认领节点 UI

**depends-on:** task-007-approve-reject-impl

## Summary

在前端节点管理页面添加待认领节点提示和审批/拒绝功能。

## BDD Scenario

```gherkin
Scenario: Full Panel reset recovery (UI part)
  Given Panel has pending nodes from heartbeat registration
  When admin opens the nodes page
  Then admin sees "N 个节点等待认领" banner at the top
  And each pending node shows UUID, API address, last seen time
  And each pending node has "审批" and "忽略" buttons

Scenario: Admin approves pending node via UI
  Given admin sees a pending node with UUID and address
  When admin clicks "审批" button
  Then a dialog appears asking for node name (required) and group (optional)
  When admin fills in "US-West-1" and clicks confirm
  Then ApproveNode RPC is called
  And the pending node disappears from the banner
  And appears in the normal node list as "offline"
```

## Files

- **Modify:** `web/src/pages/nodes-page.tsx`
  - 在 `useQuery(["nodes"])` 结果中过滤出 `status==="pending"` 的节点
  - 在页面顶部渲染待认领节点横幅（仅当有 pending 节点时显示）
  - 每个 pending 节点显示: UUID（截断）、API 地址、最后心跳时间
  - "审批" 按钮触发对话框，包含:
    - 节点名称（必填 input）
    - 分组（可选 select）
    - 公开地址（可选 input）
  - 确认后调用 `ApproveNode` RPC mutation
  - "忽略" 按钮调用 `RejectNode` RPC mutation
  - 使用 `useMutation` 配合 `invalidateQueries(["nodes"])` 刷新列表

## Steps

1. 提取 pending 节点列表
2. 创建 `PendingNodesBanner` 组件
3. 创建 `ApproveNodeDialog` 组件（参考现有 Create Node Dialog 的模式）
4. 连接 RPC mutations
5. 验证 UI 交互正确

## Verify

```bash
cd web && bunx tsc -b && bun run test
```
