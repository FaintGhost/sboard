# Task 003 Test: 订阅配置可消费性（RED）

**depends-on**: 无

## Description

先定义失败测试，验证订阅链接不只是“可解析”，而是“可被 sb-client 消费并生成可运行客户端配置”。该任务只建立 RED 场景与失败基线。

## Execution Context

**Task Number**: 003 of 010  
**Phase**: Subscription Validation (RED)  
**Prerequisites**: 无

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
- Create: `e2e/tests/helpers/subscription-client-config.helper.ts`

## Steps

### Step 1: 声明强订阅断言
- 将“订阅返回 JSON”升级为“订阅可被客户端消费”的失败断言。
- 明确断言包括目标 outbound 基础字段完整性与运行时配置可构建性。

### Step 2: 编写 RED 测试
- 新增或改造订阅测试，先在当前基线触发失败。
- 失败应定位到“缺少消费能力或字段映射不满足”，不是语法错误。

### Step 3: 记录失败基线
- 保存失败日志与断言，作为 GREEN 阶段对照。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/subscriptions.spec.ts
```

## Success Criteria

- RED 用例稳定失败且失败原因明确。
- 订阅测试的验收口径升级为“可消费”。
