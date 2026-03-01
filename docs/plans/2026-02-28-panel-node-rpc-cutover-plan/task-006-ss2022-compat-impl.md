# Task 006: SS2022 Compatibility Green Impl

**depends-on**: task-006-ss2022-compat-test

## Description

在不改变现有 SS2022 规则的前提下，修正 RPC 切换引入的序列化或映射差异，使下发配置与订阅输出继续满足规范。

## Execution Context

**Task Number**: 012 of 014
**Phase**: Refinement
**Prerequisites**: `task-006-ss2022-compat-test` 已完成并处于 Red 状态

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

- Modify: `panel/internal/node/build_config.go`
- Modify: `panel/internal/api/subscription.go`
- Modify: `panel/internal/rpc/services_impl.go`

## Steps

### Step 1: Implement Logic (Green)
- 保证 RPC 请求体映射不会破坏 SS2022 所需字段与密码格式。
- 修正任何因新契约字段命名或序列化导致的兼容偏差。
- 保持现有密钥派生模块与规则为唯一来源，避免重复逻辑。

### Step 2: Verify Green State
- 重跑 `task-006` 定向测试并确认通过。
- 补跑现有订阅与 payload 构建测试，确认无回归。

## Verification Commands

```bash
cd panel && go test ./internal/node -run TestSS2022RPCSyncPayloadCompatibility -count=1
cd panel && go test ./internal/api -run TestSubscriptionSS2022RPCCompatibility -count=1
```

## Success Criteria

- SS2022 下发与订阅规则在 RPC 切换后保持不变。
- 不引入第二套密钥派生逻辑。
