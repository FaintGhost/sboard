# Task 003 Impl: 订阅配置可消费性（GREEN）

**depends-on**: task-003-subscription-consumption-test

## Description

实现订阅消费能力：从 `/api/sub/:uuid?format=singbox` 结果组装 sb-client 可运行配置，并使 Task 003 RED 测试转为通过。

## Execution Context

**Task Number**: 003 of 010  
**Phase**: Subscription Validation (GREEN)  
**Prerequisites**: `task-003-subscription-consumption-test`

## BDD Scenario

```gherkin
Scenario: 生成订阅链接
  Given 存在用户和已配置的入站
  When 查看用户的订阅信息
  Then 能看到订阅链接

Scenario: 验证订阅内容
  Given 用户有有效的订阅链接
  When 访问订阅链接
  Then 返回有效的代理配置内容
  And 配置格式正确（sing-box JSON 或 v2ray base64）
```

**Spec Source**: `../2026-02-23-e2e-testing-design/bdd-specs.md`（订阅管理）

## Files to Modify/Create

- Modify: `e2e/tests/e2e/subscriptions.spec.ts`
- Modify: `e2e/tests/helpers/subscription-client-config.helper.ts`
- Modify: `e2e/docker-compose.e2e.yml`

## Steps

### Step 1: 落地订阅到客户端配置转换
- 增加订阅结果到 sb-client 运行配置的转换与校验逻辑。
- 确保转换结果包含客户端最小可运行配置要素，并保留 outbound 断言信息。

### Step 2: 集成 sb-client 运行前检查
- 在测试流程中加入客户端配置投递与启动前检查步骤。
- 在失败路径输出可诊断信息（配置构建失败、字段缺失、服务不可达）。

### Step 3: GREEN 验证
- 运行订阅测试并确认通过。
- 复跑生命周期主链路测试，确认配置消费能力可复用。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/subscriptions.spec.ts
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts
```

## Success Criteria

- 订阅测试通过且断言升级为“可消费”。
- sb-client 运行前检查稳定通过，无随机失败。
