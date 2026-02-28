# Task 002 Impl: 生命周期资源编排与手动同步（GREEN）

**depends-on**: task-002-lifecycle-sync-test

## Description

实现生命周期主链路所需 fixture 与测试编排能力，使 Task 002 的 RED 用例通过：完成 bootstrap 登录、创建并绑定资源、触发手动同步、验证 Node API 可观测状态。

## Execution Context

**Task Number**: 002 of 010  
**Phase**: Core Integration (GREEN)  
**Prerequisites**: `task-002-lifecycle-sync-test` 已有稳定 RED 结果

## BDD Scenario

```gherkin
Scenario: 创建入站配置并同步到节点
  Given 存在一个在线的节点
  When 创建一个入站配置并关联到该节点
  And 触发节点同步
  Then 同步状态显示成功

Scenario: 验证 Node 接收到正确配置（API 级别）
  Given 已成功同步配置到节点
  When 通过 Node API 查询当前配置
  Then 配置内容与 Panel 下发的一致
  And sing-box 进程状态为 running
```

**Spec Source**: `../2026-02-23-e2e-testing-design/bdd-specs.md`（配置同步与验证）

## Files to Modify/Create

- Modify: `e2e/tests/fixtures/api.fixture.ts`
- Modify: `e2e/tests/fixtures/auth.fixture.ts`
- Modify: `e2e/tests/e2e/lifecycle-strong.spec.ts`

## Steps

### Step 1: 补齐 fixture API 能力
- 增加生命周期场景必需的 RPC/REST 封装能力（用户分组绑定、同步任务查询、可 reset 的 inbounds 统计读取等）。
- 对错误响应补充一致的失败信息输出，保证调试可读性。

### Step 2: 实现生命周期编排
- 在强链路 spec 中落地资源创建、关系绑定、手动同步与 Node 可观测验证。
- 使用显式轮询替代固定 sleep，避免时序 flaky。

### Step 3: GREEN 验证
- 运行目标用例并确认 RED 转 GREEN。
- 复跑现有 node-sync 与 subscriptions 用例，确保兼容。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts
cd e2e && bunx playwright test --project=e2e tests/e2e/node-sync.spec.ts tests/e2e/subscriptions.spec.ts
```

## Success Criteria

- 生命周期主链路场景通过。
- 同步状态、节点健康、入站可观测数据均满足断言。
- 相关现有 E2E 用例无回归。
