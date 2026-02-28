# Task 005 Test: 稳定性与非 flaky 约束（RED）

**depends-on**: 无

## Description

新增稳定性 RED 任务，要求强链路测试移除固定 sleep 依赖，并建立可复跑的稳定性验收（至少连续多次运行）。

## Execution Context

**Task Number**: 005 of 010  
**Phase**: Reliability (RED)  
**Prerequisites**: 无

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

## Steps

### Step 1: 声明稳定性 RED 断言
- 增加“禁止固定 sleep、必须条件轮询”的测试约束。
- 增加连续执行验收定义（例如 3 次串行执行）。

### Step 2: 构建 RED 失败基线
- 在当前基线执行稳定性检查并记录失败证据。
- 失败原因须体现“存在不稳定等待策略”或“重复执行不稳定”。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts tests/e2e/node-sync.spec.ts tests/e2e/subscriptions.spec.ts
```

## Success Criteria

- 稳定性 RED 场景可复现失败。
- 失败信息可直接指导下一步改造。
