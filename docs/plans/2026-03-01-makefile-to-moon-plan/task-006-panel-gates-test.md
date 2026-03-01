# Task 006: Panel Gates Test

## Description

为 `panel/web` 门禁接入 Moon 与现有 RPC/REST 边界回归建立失败验证，确保迁移后的根编排不破坏现有产品边界。

## Execution Context

**Task Number**: 011 of 012  
**Phase**: Expansion  
**depends-on**: `task-001-workspace-bootstrap-impl.md`

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

- Create: `scripts/verify-moon-panel-gates.sh`

## Steps

### Step 1: Define Frontend Gate Mapping
- 明确至少需要覆盖 `web-lint`、`web-typecheck`、`web-test`。
- 明确 `format` 由于会写工作区，可在第二阶段再决定是否区分为 fix/check 双任务。

### Step 2: Add Red Validation
- 新增验证入口，尝试执行 `moon run panel:web-lint`、`panel:web-typecheck`、`panel:web-test`。
- 同时保留一条回归验证，要求 `automation:e2e` 能继续覆盖 Node RPC cutover 与订阅 REST 保留边界。
- 在这些 panel 任务尚未定义前，确认其失败。

### Step 3: Verify Red State
- 运行验证入口，确认当前失败。

## Verification Commands

```bash
bash scripts/verify-moon-panel-gates.sh
moon run panel:web-lint
moon run panel:web-typecheck
moon run panel:web-test
```

## Success Criteria

- 已有覆盖前端门禁映射与边界回归的验证入口。
- 当前失败明确指向 panel Moon 任务缺失。
- 尚未实现正式 panel 任务。
