# Task 007: E2E Cutover Green Impl

**depends-on**: task-007-e2e-cutover-test

## Description

完成 e2e 配置与夹具收口，使端到端用例在新 RPC 拓扑下通过，并执行完整回归门禁。

## Execution Context

**Task Number**: 014 of 014
**Phase**: Testing
**Prerequisites**: `task-007-e2e-cutover-test` 已完成并处于 Red 状态

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

- Modify: `e2e/tests/e2e/node-rpc-cutover.spec.ts`
- Modify: `e2e/tests/fixtures/api.fixture.ts`
- Modify: `e2e/tests/smoke/health.smoke.spec.ts`
- Modify: `e2e/docker-compose.e2e.yml`
- Modify: `README.zh.md`
- Modify: `docs/DESIGN.md`

## Steps

### Step 1: Implement Logic (Green)
- 将 e2e fixture 与 smoke 检查切换到新的 Node RPC 健康与通信路径。
- 修正容器编排中的健康检查与环境变量，使其反映新拓扑。
- 更新文档中 Panel↔Node 协议边界描述，避免继续声明为 REST。

### Step 2: Verify Green State
- 重跑 `task-007` 的定向 e2e。
- 运行全量回归门禁，确认生成、后端、前端与 e2e 全部通过。

### Step 3: Final Review
- 核对发布说明与回滚说明是否与设计一致。
- 确保仅保留订阅 REST 的对外描述。

## Verification Commands

```bash
moon run panel:check-generate
cd panel && go test ./... -count=1
cd node && go test ./... -count=1
moon run web:lint
moon run web:format
moon run web:typecheck
moon run web:test
moon run e2e:smoke
moon run e2e:run
```

## Success Criteria

- 新增 e2e 用例与 smoke 检查在 RPC 拓扑下通过。
- 全量门禁通过。
- 文档与运行时边界一致。
