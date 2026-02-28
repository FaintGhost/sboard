# Task 004 Impl: 目标入站流量归因（GREEN）

**depends-on**: task-004-traffic-attribution-test

## Description

实现目标入站流量归因能力，使 RED 场景转为 GREEN：通过 sb-client 访问 `probe` 后，Node `inbounds` 统计在目标 `tag/user` 上出现稳定增量。

## Execution Context

**Task Number**: 004 of 010  
**Phase**: Traffic Verification (GREEN)  
**Prerequisites**: `task-004-traffic-attribution-test`

## BDD Scenario

```gherkin
Scenario: 验证 Node 接收到正确配置（API 级别）
  Given 已成功同步配置到节点
  When 通过 Node API 查询当前配置
  Then 配置内容与 Panel 下发的一致
  And sing-box 进程状态为 running

Scenario: 验证订阅内容
  Given 用户有有效的订阅链接
  When 访问订阅链接
  Then 返回有效的代理配置内容
  And 配置格式正确（sing-box JSON 或 v2ray base64）
```

**Spec Source**: `../2026-02-23-e2e-testing-design/bdd-specs.md`（配置同步与验证 + 订阅管理）

## Files to Modify/Create

- Modify: `e2e/tests/helpers/traffic-assert.helper.ts`
- Modify: `e2e/tests/e2e/lifecycle-strong.spec.ts`
- Modify: `e2e/tests/fixtures/api.fixture.ts`

## Steps

### Step 1: 实现流量归因 helper
- 在 helper 中实现重置统计、轮询、增量判定和错误信息归一化。
- 支持按 `tag/user` 过滤目标记录，输出可读诊断信息。

### Step 2: 集成到主链路测试
- 将流量归因 helper 接入生命周期强链路 spec。
- 使用显式超时与轮询间隔，避免固定 sleep 造成 flaky。

### Step 3: GREEN 验证
- 运行主链路测试并确认目标 `tag/user` 增量断言通过。
- 复跑 node-sync、subscriptions 测试，确认无行为回归。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts
cd e2e && bunx playwright test --project=e2e tests/e2e/node-sync.spec.ts tests/e2e/subscriptions.spec.ts
```

## Success Criteria

- 目标 `tag/user` 流量增量断言稳定通过。
- 测试日志可在失败时准确区分“未产流”与“归因错误”。
