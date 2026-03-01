# Task 004: E2E Alias Test

## Description

为 `automation:e2e*` 与 `e2e:*` 任务建立失败验证，锁定 Docker Compose 清理语义与 report 的本地-only 约束。

## Execution Context

**Task Number**: 007 of 012  
**Phase**: E2E  
**depends-on**: `task-001-workspace-bootstrap-impl.md`

## BDD Scenario

```gherkin
Scenario: E2E run preserves cleanup semantics
  Given Docker daemon 和 docker compose 可用
  When 执行 `moon run automation:e2e`
  Then 它应先执行 compose down -v
  And 再执行 compose up --build --abort-on-container-exit --exit-code-from playwright
  And 最终无论成功失败都执行 compose down -v

Scenario: E2E smoke remains cheaper than full E2E
  Given Docker daemon 和 docker compose 可用
  When 执行 `moon run automation:e2e-smoke`
  Then 它只应启动 panel 和 node 基础服务
  And 应执行 Playwright smoke project
  And 不应替代 full e2e 的全部覆盖面

Scenario: Report task remains local-only
  Given 已存在 Playwright report 目录
  When 执行 `moon run automation:e2e-report`
  Then 它应调用 `bunx playwright show-report playwright-report`
  And 在 CI 中默认不运行
```

## Files to Modify/Create

- Create: `scripts/verify-moon-e2e.sh`

## Steps

### Step 1: Define Runtime Assertions
- 明确 `run`、`smoke`、`down`、`report` 各自必须保留的 shell 语义。
- 明确最重要的断言是“失败也清理”和“不在 alias 层复制复杂逻辑”。

### Step 2: Add Red Validation
- 新增一个验证入口，检查这些 Moon 任务存在性、调用链与关键命令片段。
- 在 `e2e/moon.yml` 尚未存在时，确认其失败。

### Step 3: Verify Red State
- 执行验证入口，确认当前失败。

## Verification Commands

```bash
bash scripts/verify-moon-e2e.sh
moon run automation:e2e
moon run automation:e2e-smoke
moon run automation:e2e-report
```

## Success Criteria

- 已有覆盖 `run/smoke/report` 的验证入口。
- 失败原因明确是 E2E 任务未定义。
- 还没有引入正式 E2E Moon 任务。
