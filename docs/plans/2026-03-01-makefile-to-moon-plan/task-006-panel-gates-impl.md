# Task 006: Panel Gates Impl

## Description

在 `panel` 项目中接入前端门禁任务，并完成最终回归，确认 Moon 替换根入口后不破坏现有 RPC/REST 边界。

## Execution Context

**Task Number**: 012 of 012  
**Phase**: Expansion  
**depends-on**: `task-006-panel-gates-test.md`

## BDD Scenario

```gherkin
Scenario: Existing RPC/REST boundaries are not changed
  Given 当前系统已完成 Panel 管理面 RPC 化
  And 订阅 REST 仍作为兼容入口保留
  When 执行本次 Moon 迁移
  Then 迁移后 `GET /api/sub/:user_uuid` 行为不得改变
  And Node RPC cutover 边界断言必须继续通过

Scenario: Frontend gates can be lifted into Moon without semantic drift
  Given `panel/web` 仍以 bun scripts 承载 lint/format/test
  When 后续执行 `moon run panel:web-lint`、`panel:web-typecheck`、`panel:web-test`
  Then 它们应分别等价于 `bun run lint`、`bunx tsc -b`、`bun run test`
  And 不得改变各自退出码与错误输出结构
```

## Files to Modify/Create

- Modify: `panel/moon.yml`
- Create: `node/moon.yml`

## Steps

### Step 1: Add Panel Gate Tasks
- 在 `panel` 项目中新增 `web-lint`、`web-typecheck`、`web-test`。
- 明确它们的工作目录应指向 `panel/web`，并保留 `bun`/`bunx` 的原始输出。

### Step 2: Add Node Project Placeholder
- 创建 `node/moon.yml` 作为项目占位，确保 monorepo 结构完整，即使第一阶段不接入 node 测试/构建任务。

### Step 3: Run Final Regression
- 运行前端门禁的 Moon 任务。
- 运行 `automation:check-generate`、`automation:e2e-smoke`、`automation:e2e`，确认订阅 REST 与 Node RPC cutover 边界持续通过。

## Verification Commands

```bash
moon run panel:web-lint
moon run panel:web-typecheck
moon run panel:web-test
bash scripts/verify-moon-panel-gates.sh
moon run automation:check-generate
moon run automation:e2e-smoke
moon run automation:e2e
```

## Success Criteria

- `panel:web-lint`、`panel:web-typecheck`、`panel:web-test` 已可执行。
- 前端门禁通过 Moon 运行且语义不变。
- 最终回归确认现有 RPC/REST 边界未被破坏。
