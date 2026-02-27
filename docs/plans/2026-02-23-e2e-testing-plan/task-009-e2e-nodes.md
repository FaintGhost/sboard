# Task 009: E2E - 节点管理测试

**depends-on**: task-002

## Description

实现节点管理的 E2E 测试：创建节点（指向 Docker 内的 node 容器）、查看健康状态、删除节点。

## Execution Context

**Task Number**: 009 of 012
**Phase**: Core Features (E2E)
**Prerequisites**: Task 002 完成（fixtures 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "节点管理" — 全部 3 个场景

## Files to Create

- `e2e/tests/e2e/nodes.spec.ts`

## Steps

### Step 1: 研究节点管理 UI

查看 Panel 前端的节点管理页面：
- 节点列表页面结构
- 创建节点的表单字段（名称、API 地址、API 端口、密钥、公网地址等）
- 节点健康状态的显示方式（在线/离线标识）
- 删除节点的交互方式
- 查阅 `panel/proto/sboard/panel/v1/panel.proto` 中 Nodes 相关 RPC 了解字段定义

### Step 2: 实现节点管理测试

创建 `nodes.spec.ts`，使用 `authenticatedPage` fixture。使用 `test.describe.serial`。

1. **创建节点**：导航到节点管理页面，点击创建节点按钮，填写节点信息：
   - 名称: 唯一前缀 + "test-node"
   - API 地址: `node`（Docker 内部 DNS 名称）
   - API 端口: `3000`
   - 密钥: `e2e-test-node-secret`（与 Node 容器的 NODE_SECRET_KEY 一致）
   - 公网地址: `node`

   提交后验证节点列表中显示新节点

2. **查看节点健康状态**：等待 Panel 的 node monitor 轮询检查节点状态（间隔 5s），验证节点状态显示为在线/健康

3. **删除节点**：删除测试节点，确认删除，验证节点从列表中消失

### Step 3: 在 Docker 环境中验证

运行 e2e project 中的 nodes 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
  docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=e2e tests/e2e/nodes.spec.ts
```

## Success Criteria

- 3 个节点管理场景测试全部通过
- 节点创建成功并在列表中可见
- 节点健康状态正确显示（在线）
- 节点可成功删除
