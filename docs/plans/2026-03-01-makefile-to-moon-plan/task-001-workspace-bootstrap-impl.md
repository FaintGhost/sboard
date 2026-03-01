# Task 001: Workspace Bootstrap Impl

## Description

实现 `.prototools` 与 Moon 工作区骨架，使仓库具备可解析的版本锁定与项目注册能力。

## Execution Context

**Task Number**: 002 of 012  
**Phase**: Foundation  
**depends-on**: `task-001-workspace-bootstrap-test.md`

## BDD Scenario

```gherkin
Scenario: Toolchain versions are strictly pinned
  Given 仓库根存在 `.prototools`
  When 开发者在仓库内执行 `proto use`
  Then `moon`、`go`、`node`、`bun` 均应解析到仓库锁定版本
  And 任务执行不得依赖系统全局版本漂移
```

## Files to Modify/Create

- Create: `.prototools`
- Create: `.moon/workspace.yml`
- Create: `.moon/toolchains.yml`
- Create: `.moon/tasks/all.yml`

## Steps

### Step 1: Create Version Source of Truth
- 新建 `.prototools`，精确锁定设计文档要求的 `moon`、`go`、`node`、`bun` 版本。
- 保证该文件成为仓库唯一的精确版本来源。

### Step 2: Create Workspace Skeleton
- 新建 `.moon/workspace.yml`，注册 `automation`、`panel`、`node`、`e2e` 四个项目，并设置默认项目。
- 新建 `.moon/toolchains.yml`，启用 `go`、`node`、`bun`，并将 JavaScript 包管理器固定为 `bun`。
- 新建 `.moon/tasks/all.yml`，定义共享 file groups，覆盖 RPC 生成输入/输出与 E2E 基础设施文件。

### Step 3: Verify Green State
- 运行 Task 001 的验证脚本，确认工具链与工作区均能被正确识别。

## Verification Commands

```bash
proto use
bash scripts/verify-moon-toolchain.sh
bash scripts/verify-moon-workspace.sh
```

## Success Criteria

- `.prototools` 与 `.moon/*` 已存在且可被解析。
- 版本检查与工作区检查通过。
- 未引入任何实际业务任务定义。
