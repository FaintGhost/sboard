# Task 003: Check Generate Test

## Description

为 `moon run automation:check-generate` 的成功/失败路径建立验证，锁定它必须先生成再检查 diff。

## Execution Context

**Task Number**: 005 of 012  
**Phase**: Task Graph  
**depends-on**: `task-002-generate-alias-impl.md`

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

- Create: `scripts/verify-moon-check-generate.sh`

## Steps

### Step 1: Define Success and Failure Assertions
- 明确成功路径需要看到“up to date”提示。
- 明确失败路径需要通过制造受控 diff 来验证错误文案与退出码。

### Step 2: Add Red Validation
- 新增验证脚本，分别覆盖“产物新鲜”和“产物过期”两条路径。
- 在 `automation:check-generate` 尚未实现前，确认失败点是“任务缺失”。

### Step 3: Verify Red State
- 运行验证脚本，确认当前失败。

## Verification Commands

```bash
bash scripts/verify-moon-check-generate.sh
moon run automation:check-generate
```

## Success Criteria

- 已有同时覆盖成功/失败路径的验证入口。
- 失败原因明确指向 `check-generate` 任务未定义。
- 未提前实现正式检查逻辑。
