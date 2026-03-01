# Task 001: Workspace Bootstrap Test

## Description

先为“工具链严格锁定 + Moon 工作区骨架可被识别”建立失败验证。该任务只创建验证入口，不实现正式配置。

## Execution Context

**Task Number**: 001 of 012  
**Phase**: Foundation  
**depends-on**: 无

## BDD Scenario

```gherkin
Scenario: Toolchain versions are strictly pinned
  Given 仓库根存在 `.prototools`
  When 开发者在仓库内执行 `proto use`
  Then `moon`、`go`、`node`、`bun` 均应解析到仓库锁定版本
  And 任务执行不得依赖系统全局版本漂移
```

## Files to Modify/Create

- Create: `scripts/verify-moon-toolchain.sh`
- Create: `scripts/verify-moon-workspace.sh`

## Steps

### Step 1: Extract Expected Versions
- 从设计文档确认应锁定的 `moon`、`go`、`node`、`bun` 精确版本。
- 明确 `proto use` 与 `moon project` / `moon query` 需要被验证的成功条件。

### Step 2: Add Red Validation Entry Points
- 新增一个工具链验证入口，检查 `.prototools` 是否存在且版本值正确。
- 新增一个工作区验证入口，检查 Moon 能识别默认项目与注册的 projects。
- 验证脚本必须以可重复方式失败，且失败原因应是“配置缺失”，而不是无关的 shell 错误。

### Step 3: Verify Red State
- 在尚未创建 `.prototools` 与 `.moon/*` 前运行验证，确认其失败。

## Verification Commands

```bash
bash scripts/verify-moon-toolchain.sh
bash scripts/verify-moon-workspace.sh
```

## Success Criteria

- 已有明确的可执行验证入口。
- 失败状态准确指向“缺少版本锁定/工作区配置”。
- 没有提前实现正式 Moon 配置。
