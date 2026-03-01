# Task 003: Check Generate Impl

## Description

实现 `automation:check-generate`，确保它保留“先生成，再检查 diff”的现有语义。

## Execution Context

**Task Number**: 006 of 012  
**Phase**: Task Graph  
**depends-on**: `task-003-check-generate-test.md`

## BDD Scenario

```gherkin
Scenario: Check-generate succeeds when artifacts are fresh
  Given 所有生成产物已经与 proto 规格同步
  When 执行 `moon run automation:check-generate`
  Then 任务退出码应为 0
  And 输出应包含 “Generated files are up to date.”

Scenario: Check-generate fails when artifacts are stale
  Given proto 规格已变化但生成产物未同步
  When 执行 `moon run automation:check-generate`
  Then 任务退出码应为非 0
  And 输出应包含 “Generated files are out of date.”
  And 失败原因应是 git diff 检测到生成目录变化
```

## Files to Modify/Create

- Modify: `scripts/moon.yml`

## Steps

### Step 1: Add Check Task
- 在 `automation` 项目中新增 `check-generate` 任务。
- 该任务必须依赖 `generate`，并在仓库根工作目录执行 `git diff --exit-code`。

### Step 2: Preserve Current Messages
- 保持与当前 `Makefile` 一致的成功/失败输出文案，避免开发者排障心智漂移。
- 明确该任务默认禁用缓存。

### Step 3: Verify Green State
- 运行 Task 003 的验证脚本，确认成功/失败路径均符合预期。

## Verification Commands

```bash
moon run automation:check-generate
bash scripts/verify-moon-check-generate.sh
```

## Success Criteria

- `automation:check-generate` 行为与旧入口一致。
- 会先执行生成，再执行 diff 检查。
- 成功/失败文案与退出码均符合预期。
