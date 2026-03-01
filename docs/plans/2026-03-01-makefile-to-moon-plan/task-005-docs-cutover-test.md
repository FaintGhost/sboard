# Task 005: Docs Cutover Test

## Description

为“Moon 成为唯一文档化根入口”建立失败验证，锁定 README/AGENTS 必须改为 `moon run ...`，并要求根 `Makefile` 被移除。

## Execution Context

**Task Number**: 009 of 012  
**Phase**: Cutover  
**depends-on**: `task-004-e2e-alias-impl.md`

## BDD Scenario

```gherkin
Scenario: Moon becomes the only documented root entrypoint
  Given 迁移完成
  When 开发者阅读 README 或项目交付门禁说明
  Then 根级任务示例应全部使用 `moon run ...`
  And 不再把 `make ...` 作为默认入口
```

## Files to Modify/Create

- Create: `scripts/verify-no-make-entrypoints.sh`

## Steps

### Step 1: Define Documentation Scope
- 明确必须更新的开发者入口文档：`README.zh.md`、`README.en.md`、`AGENTS.md`。
- 明确历史设计文档可保留旧上下文，但当前面向开发者的入口文档必须切换。

### Step 2: Add Red Validation
- 新增一个文档审计入口，扫描上述文件中的根任务示例，验证它们仍引用旧 `make` 时应失败。
- 验证该脚本还应检查根 `Makefile` 仍存在时给出失败结果。

### Step 3: Verify Red State
- 执行审计入口，确认当前失败。

## Verification Commands

```bash
bash scripts/verify-no-make-entrypoints.sh
test ! -f Makefile
```

## Success Criteria

- 已有独立的文档/入口审计脚本。
- 当前失败清楚指向“文档尚未切换”和“Makefile 尚未删除”。
- 尚未修改正式文档内容。
