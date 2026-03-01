# Task 002: Generate Alias Test

## Description

为 `moon run automation:generate` 建立失败验证，锁定它与现有 `make generate` 的行为等价性。

## Execution Context

**Task Number**: 003 of 012  
**Phase**: Task Graph  
**depends-on**: `task-001-workspace-bootstrap-impl.md`

## BDD Scenario

```gherkin
Scenario: Generate command remains behaviorally equivalent
  Given panel RPC proto and buf config are valid
  When 执行 `moon run automation:generate`
  Then 它应执行与原 `make generate` 等价的生成流程
  And `panel/internal/rpc/gen`
  And `panel/web/src/lib/rpc/gen`
  And `node/internal/rpc/gen`
  Should reflect freshly generated artifacts
```

## Files to Modify/Create

- Create: `scripts/verify-moon-generate.sh`

## Steps

### Step 1: Define Equivalence Boundary
- 明确“等价”包含：执行入口、工作目录、生成目录、副作用与退出码。
- 明确该验证不应吞掉 `go generate`、`buf generate` 的原始错误。

### Step 2: Add Red Validation
- 新增一个验证入口，尝试执行 `moon run automation:generate` 并对比生成目录状态。
- 在 `automation:generate` 与 `panel:generate-rpc` 尚未实现时，确认失败指向“任务不存在或未定义”。

### Step 3: Verify Red State
- 运行验证脚本，确认当前失败。

## Verification Commands

```bash
bash scripts/verify-moon-generate.sh
moon run automation:generate
```

## Success Criteria

- 已有针对 `automation:generate` 的独立验证入口。
- 失败原因明确为 Moon 任务未落地。
- 尚未实现正式任务映射。
