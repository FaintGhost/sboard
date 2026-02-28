# Task 007: E2E Cutover Red Test

**depends-on**: task-001-sync-success-impl, task-002-sync-error-mapping-impl, task-003-monitor-rpc-telemetry-impl, task-004-concurrency-lock-impl, task-005-rest-boundary-impl, task-006-ss2022-compat-impl

## Description

在端到端层面先补齐失败测试，锁定直切发布后的整链路表现，包括 Node RPC 健康检查、同步成功与订阅兼容。

## Execution Context

**Task Number**: 013 of 014
**Phase**: Testing
**Prerequisites**: 六个核心特性已在实现层完成，进入全链路验证阶段

## BDD Scenario

```gherkin
Scenario: 同步成功
  Given Panel 与 Node 的 RPC 服务均已启动
  And Node 鉴权密钥与 Panel 中配置一致
  When 管理员触发节点同步
  Then Node 成功应用配置
  And SyncJob 状态为 success

Scenario: 节点健康检查
  Given Node RPC Health 可访问
  When Panel monitor 执行健康探测
  Then 节点状态被正确更新为 online

Scenario: 管理 REST 路径不可用
  Given 系统已完成直切发布
  When 客户端访问历史管理 REST 路径
  Then 返回 not found 或明确不可用

Scenario: 订阅 REST 保持可用
  Given 用户存在有效订阅
  When 客户端访问 GET /api/sub/:user_uuid?format=singbox
  Then 返回有效 sing-box 配置
```

**Spec Source**: `../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md`

## Files to Modify/Create

- Create: `e2e/tests/e2e/node-rpc-cutover.spec.ts`
- Modify: `e2e/tests/fixtures/api.fixture.ts`
- Modify: `e2e/tests/smoke/health.smoke.spec.ts`
- Modify: `e2e/docker-compose.e2e.yml`

## Steps

### Step 1: Verify Scenario
- 从 BDD 中选取必须由 e2e 证明的最小链路，不把全部分支堆进端到端用例。

### Step 2: Implement Test (Red)
- 新增 e2e 用例，描述 Node RPC 健康检查、同步、旧 REST 不可用与订阅可用的组合验证。
- 先调整 fixture 与容器健康检查期望，使当前代码在切换未完全收口前失败（Red）。
- 保持测试数据通过 API/页面创建，不依赖预置数据库状态。

### Step 3: Verify Red State
- 运行 smoke 和定向 e2e，确认失败（Red）。

## Verification Commands

```bash
make e2e-smoke
cd e2e && bunx playwright test --project=e2e tests/e2e/node-rpc-cutover.spec.ts
```

## Success Criteria

- 有独立 e2e 用例证明直切后的核心行为。
- Red 状态能稳定复现切换前后的差异。
