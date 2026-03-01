# Task 004: E2E Alias Impl

## Description

实现 `e2e` 项目任务与 `automation:e2e*` 别名，完整保留当前 Docker Compose 运行与清理语义。

## Execution Context

**Task Number**: 008 of 012  
**Phase**: E2E  
**depends-on**: `task-004-e2e-alias-test.md`

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

- Create: `e2e/moon.yml`
- Modify: `scripts/moon.yml`

## Steps

### Step 1: Add E2E Project Tasks
- 在 `e2e` 项目中新增 `run`、`smoke`、`down`、`report`。
- 保持与当前 `Makefile` 一致的 `docker compose` 顺序、退出码透传和清理逻辑。
- 明确禁用缓存，且 `report` 默认不在 CI 中运行。

### Step 2: Add Automation Aliases
- 在 `automation` 项目中新增 `e2e`、`e2e-smoke`、`e2e-down`、`e2e-report`，仅通过依赖转发到 `e2e` 项目。

### Step 3: Verify Green State
- 先运行结构验证脚本，再在可用环境中至少完成一次 `e2e-smoke` 与一次全量 `e2e` 对照验证。

## Verification Commands

```bash
bash scripts/verify-moon-e2e.sh
moon run automation:e2e-smoke
moon run automation:e2e
moon run automation:e2e-report
```

## Success Criteria

- `automation:e2e*` 与 `e2e:*` 已可执行。
- Compose 清理语义与旧入口一致。
- `report` 任务保留本地-only 约束。
