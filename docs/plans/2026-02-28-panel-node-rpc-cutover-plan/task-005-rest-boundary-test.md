# Task 005: REST Boundary Red Test

## Description

为“仅保留订阅 REST”建立失败测试，覆盖管理 REST 不可用、订阅 REST 可用、UA 默认格式三个边界。

## Execution Context

**Task Number**: 009 of 014
**Phase**: Refinement
**Prerequisites**: 无

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

- Create: `node/internal/api/rest_boundary_test.go`
- Create: `panel/internal/api/subscription_boundary_test.go`
- Modify: `e2e/tests/fixtures/api.fixture.ts`

## Steps

### Step 1: Verify Scenario
- 确认本任务只关心 REST 边界，不承担 Node RPC 主链路功能实现。

### Step 2: Implement Test (Red)
- 新增 Node 侧路由测试，断言旧 Node REST 管理/统计端点不再暴露。
- 新增 Panel 侧订阅测试，断言 `format=singbox` 与 UA 默认格式行为不变。
- 在 e2e fixture 中先加入会失败的边界断言或调用点，暴露当前与目标态差异。

### Step 3: Verify Red State
- 运行定向测试，确认失败（Red）。

## Verification Commands

```bash
cd node && go test ./internal/api -run TestLegacyNodeRESTEndpointsUnavailable -count=1
cd panel && go test ./internal/api -run 'TestSubscriptionREST(ExplicitFormat|UserAgentDefault)' -count=1
cd e2e && bunx playwright test --project=e2e tests/e2e/subscriptions.spec.ts
```

## Success Criteria

- REST 边界三条场景均被测试表达。
- 当前状态在切换前能明确失败，作为后续实现基线。
