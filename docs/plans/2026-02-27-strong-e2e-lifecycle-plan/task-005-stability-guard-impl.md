# Task 005 Impl: 稳定性与非 flaky 约束（GREEN）

**depends-on**: task-005-stability-guard-test

## Description

实现稳定性治理：用条件轮询替代固定 sleep，统一等待策略，并通过连续复跑验证强链路用例稳定性。

## Execution Context

**Task Number**: 005 of 010  
**Phase**: Reliability (GREEN)  
**Prerequisites**: `task-005-stability-guard-test`

## BDD Scenario

```gherkin
Scenario: 查看节点健康状态
  Given 存在一个已配置的节点
  When 查看节点详情或列表
  Then 节点状态显示为在线（健康检查通过）

Scenario Outline: 核心页面加载
  Given 管理员已登录
  When 导航到 <路由>
  Then 页面正常加载（无错误弹窗或白屏）
  And 页面标题或关键元素可见
```

**Spec Source**: `../2026-02-23-e2e-testing-design/bdd-specs.md`（节点管理 + 核心页面导航）

## Files to Modify/Create

- Modify: `e2e/tests/e2e/lifecycle-strong.spec.ts`
- Modify: `e2e/tests/e2e/node-sync.spec.ts`
- Modify: `e2e/tests/e2e/subscriptions.spec.ts`
- Modify: `e2e/tests/helpers/traffic-assert.helper.ts`

## Steps

### Step 1: 落地统一等待策略
- 将固定 sleep 替换为条件轮询与显式超时。
- 统一断言失败输出，便于定位网络、同步或流量阶段问题。

### Step 2: 通过稳定性验证
- 运行目标测试集连续复跑（至少 3 次）并记录结果。
- 确认失败率降为 0（在当前环境与数据隔离策略下）。

### Step 3: 执行门禁回归
- 运行 e2e 与项目约定门禁命令，确保改造后可交付。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts tests/e2e/node-sync.spec.ts tests/e2e/subscriptions.spec.ts
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts tests/e2e/node-sync.spec.ts tests/e2e/subscriptions.spec.ts
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts tests/e2e/node-sync.spec.ts tests/e2e/subscriptions.spec.ts
bun run lint
bun run format
bunx tsc -b
bun run test
make check-generate
```

## Success Criteria

- 强链路相关 E2E 连续复跑稳定通过。
- 前端与仓库门禁检查通过。
- 测试等待策略统一且可维护。
