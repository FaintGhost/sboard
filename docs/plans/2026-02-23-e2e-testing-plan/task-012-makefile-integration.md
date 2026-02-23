# Task 012: Makefile 集成与最终验证

**depends-on**: task-003, task-004, task-005, task-006, task-007, task-008, task-009, task-010, task-011

## Description

将 E2E 测试集成到项目根目录的 Makefile 中，提供便捷的运行命令，并执行全量测试验证整体稳定性。

## Execution Context

**Task Number**: 012 of 012
**Phase**: Refinement
**Prerequisites**: 所有测试任务（003-011）已完成

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: 跨场景 — 验证完整测试套件的运行能力

## Files to Modify

- `Makefile` — 添加 e2e 相关 targets

## Steps

### Step 1: 添加 Makefile targets

在根目录 `Makefile` 中添加以下 targets：

- `e2e`: 一键运行全部测试（构建 → 启动服务 → 运行所有 smoke + e2e 测试 → 输出报告）
- `e2e-smoke`: 仅运行 smoke 测试
- `e2e-down`: 清理测试环境（停止容器 + 删除 volumes）
- `e2e-report`: 打开 HTML 测试报告

每个 target 应包含正确的 docker compose 命令，处理退出码传播。

### Step 2: 运行全量 Smoke 测试

执行 `make e2e-smoke`，验证所有 smoke 测试通过。

### Step 3: 运行全量 E2E 测试

执行 `make e2e`，验证所有 smoke + e2e 测试通过。

### Step 4: 验证清理

执行 `make e2e-down`，确认所有容器和卷被正确清理。

### Step 5: 验证报告输出

检查 `e2e/playwright-report/` 目录是否生成了 HTML 报告，验证报告内容完整。

## Verification Commands

```bash
# 运行全量 smoke
make e2e-smoke

# 清理
make e2e-down

# 运行全量测试
make e2e

# 清理
make e2e-down

# 检查报告
ls -la e2e/playwright-report/
```

## Success Criteria

- `make e2e-smoke` 全部通过
- `make e2e` 全部通过（所有 smoke + e2e 测试）
- `make e2e-down` 正确清理所有容器和卷
- HTML 报告正确生成在 `e2e/playwright-report/`
- 总测试时间：smoke < 30 秒, 全量 < 5 分钟
