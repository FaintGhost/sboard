# Task 001 Test: 运行时拓扑与健康门禁（RED）

**depends-on**: 无

## Description

新增运行时拓扑 Smoke 测试，先让测试在当前基线上失败，明确表达强链路 E2E 所需的容器拓扑与健康门禁要求（包含 `sb-client` 与 `probe`）。

## Execution Context

**Task Number**: 001 of 010  
**Phase**: Foundation (RED)  
**Prerequisites**: 无

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

- Create: `e2e/tests/smoke/runtime-topology.smoke.spec.ts`
- Modify: `e2e/playwright.config.ts`

## Steps

### Step 1: 校验场景映射
- 确认本任务覆盖“系统健康检查”并扩展为强链路运行时拓扑场景。

### Step 2: 编写失败测试（RED）
- 新建运行时拓扑 smoke 用例，声明对 `sb-client` 与 `probe` 的健康与可达性断言。
- 保证该测试在当前环境下以业务断言失败（例如服务缺失或健康条件不满足），而不是语法或导入错误。

### Step 3: 记录失败证据
- 保存失败日志与失败断言信息，作为后续 GREEN 实现验收基线。

## Verification Commands

```bash
cd e2e && bunx playwright test --project=smoke tests/smoke/runtime-topology.smoke.spec.ts
```

## Success Criteria

- 新增测试文件已提交。
- 用例在当前基线稳定失败，且失败原因是“未满足拓扑/健康要求”。
