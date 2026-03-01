# Task 005: Docs Cutover Impl

## Description

切换开发文档入口到 Moon，并删除根 `Makefile`。

## Execution Context

**Task Number**: 010 of 012  
**Phase**: Cutover  
**depends-on**: `task-005-docs-cutover-test.md`

## BDD Scenario

```gherkin
Scenario: Moon becomes the only documented root entrypoint
  Given 迁移完成
  When 开发者阅读 README 或项目交付门禁说明
  Then 根级任务示例应全部使用 `moon run ...`
  And 不再把 `make ...` 作为默认入口
```

## Files to Modify/Create

- Delete: `Makefile`
- Modify: `README.zh.md`
- Modify: `README.en.md`
- Modify: `AGENTS.md`

## Steps

### Step 1: Update Developer-Facing Docs
- 将所有当前推荐的根级任务示例替换为对应的 `moon run automation:<task>`。
- 明确前端门禁与生成门禁的新入口。

### Step 2: Remove Root Makefile
- 删除根 `Makefile`，避免继续保留双入口。
- 在删除前确认所有原 target 已有 Moon 对应任务。

### Step 3: Verify Green State
- 运行 Task 005 的审计脚本，确认文档已切换且 `Makefile` 不再存在。

## Verification Commands

```bash
bash scripts/verify-no-make-entrypoints.sh
test ! -f Makefile
rg -n "make (generate|check-generate|e2e|e2e-smoke|e2e-down|e2e-report)" README.zh.md README.en.md AGENTS.md
```

## Success Criteria

- 面向开发者的根入口文档已统一切到 Moon。
- 根 `Makefile` 已删除。
- 不再存在“当前推荐路径”仍指向 `make` 的情况。
