# Task 005: REST Boundary Green Impl

**depends-on**: task-005-rest-boundary-test

## Description

收口 REST 边界：移除 Node 旧 REST 通信端点，仅保留订阅 REST，并更新调用夹具与路由期望。

## Execution Context

**Task Number**: 010 of 014
**Phase**: Refinement
**Prerequisites**: `task-005-rest-boundary-test` 已完成并处于 Red 状态

## BDD Scenario

```gherkin
Scenario: 管理 REST 路径不可用
  Given 系统已完成直切发布
  When 客户端访问历史管理 REST 路径
  Then 返回 not found 或明确不可用

Scenario: 订阅 REST 保持可用
  Given 用户存在有效订阅
  When 客户端访问 GET /api/sub/:user_uuid?format=singbox
  Then 返回有效 sing-box 配置

Scenario: 订阅默认格式按 User-Agent 选择
  Given 请求未显式指定 format
  When 不同 User-Agent 访问订阅接口
  Then 返回符合约定的默认格式
```

**Spec Source**: `../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md`

## Files to Modify/Create

- Modify: `node/internal/api/router.go`
- Modify: `node/cmd/node/main.go`
- Modify: `panel/internal/api/router.go`
- Modify: `panel/internal/api/subscription.go`
- Modify: `e2e/tests/fixtures/api.fixture.ts`

## Steps

### Step 1: Implement Logic (Green)
- 删除或关闭 Node 旧 REST 同步与统计路由，仅保留符合目标态的必要边界。
- 保持 Panel 订阅 REST 路由与格式选择逻辑不变。
- 更新测试夹具，使 Node 侧调用改为 RPC 入口或不再直接访问已删除的旧 REST 端点。

### Step 2: Verify Green State
- 重跑 `task-005` 定向测试，确认通过。
- 补跑订阅相关现有测试，确认订阅兼容未回归。

## Verification Commands

```bash
cd node && go test ./internal/api -run TestLegacyNodeRESTEndpointsUnavailable -count=1
cd panel && go test ./internal/api -run 'TestSubscriptionREST(ExplicitFormat|UserAgentDefault)' -count=1
cd e2e && bunx playwright test --project=e2e tests/e2e/subscriptions.spec.ts
```

## Success Criteria

- Node 旧 REST 通信端点不再对外暴露。
- 订阅 REST 行为保持稳定。
