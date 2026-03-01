# Task 002: Generate Alias Impl

## Description

实现 `panel:generate-rpc` 与 `automation:generate`，让 Moon 成为新的生成入口。

## Execution Context

**Task Number**: 004 of 012  
**Phase**: Task Graph  
**depends-on**: `task-002-generate-alias-test.md`

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

- Create: `panel/moon.yml`
- Create: `scripts/moon.yml`

## Steps

### Step 1: Define Project Task
- 在 `panel` 项目中添加 `generate-rpc` 任务，保持与 `cd panel && go generate ./internal/rpc/...` 等价。
- 该任务必须保留 `go generate` 的原始工作目录与错误输出。

### Step 2: Define Automation Alias
- 在 `automation` 项目中添加 `generate` 别名任务，仅通过依赖转发到 `panel:generate-rpc`。
- 不要在别名层复制生成逻辑。

### Step 3: Verify Green State
- 运行 Task 002 的验证脚本。
- 与当前 `make generate` 做一次结果对照，确认生成目录副作用一致。

## Verification Commands

```bash
moon run automation:generate
bash scripts/verify-moon-generate.sh
git diff -- 'panel/internal/rpc/gen/**' 'panel/web/src/lib/rpc/gen/**' 'node/internal/rpc/gen/**'
```

## Success Criteria

- `panel:generate-rpc` 与 `automation:generate` 可执行。
- 生成行为与旧入口一致。
- 没有在别名任务中复制底层命令。
