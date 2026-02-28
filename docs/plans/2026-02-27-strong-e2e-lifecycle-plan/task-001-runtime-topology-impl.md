# Task 001 Impl: 运行时拓扑与健康门禁（GREEN）

**depends-on**: task-001-runtime-topology-test

## Description

实现强 E2E 运行时拓扑：在 compose 中补齐 `sb-client` 与 `probe`，并完善健康门禁，使 Task 001 的 RED 用例转为通过。

## Execution Context

**Task Number**: 001 of 010  
**Phase**: Foundation (GREEN)  
**Prerequisites**: `task-001-runtime-topology-test` 已产出稳定失败断言

## BDD Scenario

```gherkin
Scenario: 强 E2E 运行时拓扑可启动并健康
  Given E2E docker compose 环境包含 panel、node、playwright、sb-client、probe
  When 启动 E2E 基础环境并执行运行时拓扑 smoke 测试
  Then Panel 健康接口返回 status "ok"
  And Node 健康接口返回 status "ok"
  And sb-client 与 probe 均通过健康检查
  And playwright 可以通过 bridge 网络服务名访问 panel、node、probe
```

**Spec Source**: `../2026-02-23-e2e-testing-design/bdd-specs.md`（系统健康检查） + 会话已确认约束（bridge + sing-box 客户端）

## Files to Modify/Create

- Modify: `e2e/docker-compose.e2e.yml`
- Modify: `e2e/entrypoint.sh`
- Modify: `e2e/tests/smoke/runtime-topology.smoke.spec.ts`

## Steps

### Step 1: 完成运行时拓扑实现
- 在 e2e compose 中新增 `sb-client` 与 `probe` 服务定义。
- 使用 bridge 网络服务名互通，避免 `network_mode: host`。
- 为关键服务补充健康检查与健康依赖，确保测试在“服务就绪”后启动。

### Step 2: 对齐测试入口门禁
- 更新 `entrypoint.sh` 探活逻辑，覆盖本任务所需的基础服务可达性检查。
- 保证失败时输出明确错误，便于定位是服务未就绪还是网络不可达。

### Step 3: 验证 GREEN
- 运行 Task 001 的目标 smoke 测试并确认通过。
- 再次运行现有 smoke 套件，确认无回归。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=smoke tests/smoke/runtime-topology.smoke.spec.ts
cd e2e && bunx playwright test --project=smoke
```

## Success Criteria

- Task 001 RED 场景由失败转为通过。
- 运行时拓扑在 bridge 模式下稳定可用。
- 现有 smoke 用例无新增失败。
