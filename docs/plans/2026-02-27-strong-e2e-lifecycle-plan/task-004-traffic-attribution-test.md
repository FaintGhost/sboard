# Task 004 Test: 目标入站流量归因（RED）

**depends-on**: 无

## Description

新增流量归因 RED 测试，验证“通过订阅客户端发起请求后，Node `inbounds` 统计在目标 `tag/user` 上出现增量”。

## Execution Context

**Task Number**: 004 of 010  
**Phase**: Traffic Verification (RED)  
**Prerequisites**: 无

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

- Modify: `e2e/tests/e2e/lifecycle-strong.spec.ts`
- Create: `e2e/tests/helpers/traffic-assert.helper.ts`
- Modify: `e2e/tests/fixtures/api.fixture.ts`

## Steps

### Step 1: 定义流量归因 RED 断言
- 增加“reset 统计 -> 发起代理请求 -> 轮询增量”的测试流程定义。
- 明确必须命中目标 inbound tag 与目标 user。

### Step 2: 编写 RED 场景
- 在主链路用例中加入流量归因断言并执行，确保当前基线先失败。
- 失败应可区分“未产生流量”与“流量未归因到目标对象”。

### Step 3: 记录失败基线
- 保存断言输出，作为 GREEN 阶段比对基线。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=e2e tests/e2e/lifecycle-strong.spec.ts
```

## Success Criteria

- 流量归因 RED 场景可重复失败并能定位原因。
- 断言明确要求 tag/user 维度增量，不接受仅网卡总流量变化。
