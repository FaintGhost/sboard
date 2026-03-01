# Task 006: SS2022 Compatibility Red Test

## Description

为 SS2022 兼容场景建立失败测试，确保协议切换不会破坏下发配置与订阅中的密钥规则。

## Execution Context

**Task Number**: 011 of 014
**Phase**: Refinement
**Prerequisites**: 无

## BDD Scenario

```gherkin
Scenario: 同步与订阅均满足 SS 2022 密钥规则
  Given 入站协议为 shadowsocks 2022
  When Panel 生成节点下发配置与订阅
  Then 服务端 password 为有效 base64 密钥
  And 用户密码满足 psk:userKey 组合规则
```

**Spec Source**: `../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md`

## Files to Modify/Create

- Create: `panel/internal/node/ss2022_rpc_compat_test.go`
- Create: `panel/internal/api/subscription_ss2022_rpc_compat_test.go`
- Modify: `e2e/tests/e2e/subscriptions.spec.ts`

## Steps

### Step 1: Verify Scenario
- 明确该任务只验证 SS2022 规则稳定性，不承担一般协议回归。

### Step 2: Implement Test (Red)
- 新增 payload 构建测试，断言 RPC 切换后仍生成合法的服务端与用户密钥。
- 新增订阅测试，断言客户端密码仍为 `psk:userKey` 组合格式。
- 如需端到端补充，先在 e2e 中加入失败断言锁定该规则。

### Step 3: Verify Red State
- 运行定向测试，确认失败（Red）或显式暴露待调整处。

## Verification Commands

```bash
cd panel && go test ./internal/node -run TestSS2022RPCSyncPayloadCompatibility -count=1
cd panel && go test ./internal/api -run TestSubscriptionSS2022RPCCompatibility -count=1
cd e2e && bunx playwright test --project=e2e tests/e2e/subscriptions.spec.ts -g 'shadowsocks'
```

## Success Criteria

- SS2022 场景存在独立测试保护。
- 测试聚焦密钥格式与订阅输出，而不是广泛协议行为。
