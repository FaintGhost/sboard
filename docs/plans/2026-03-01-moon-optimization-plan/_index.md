# Moon Monorepo Optimization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 `panel/web` 提升为顶层 `web/`、取消 automation alias 层、修复工具链双源、适配 Docker/buf 构建链、新增 CI 门禁、清理迁移残留。

**Architecture:** 先完成目录移动这一基础变更（Task 1），再按依赖链更新 Moon 配置（Task 2）、buf 生成链路（Task 3）、Docker 构建链（Task 4）、工具脚本（Task 5）、CI 工作流（Task 6）、文档（Task 7）、清理（Task 8），最后做全量回归（Task 9）。

**Tech Stack:** Moon v2, proto, Go 1.26.0, Node 22.22.0, Bun 1.3.9, Buf, Docker Compose, Playwright, GitHub Actions。

**Design Support:**
- [Design Index](../2026-03-01-moon-optimization-design/_index.md)
- [BDD Specs](../2026-03-01-moon-optimization-design/bdd-specs.md)
- [Architecture](../2026-03-01-moon-optimization-design/architecture.md)
- [Best Practices](../2026-03-01-moon-optimization-design/best-practices.md)

**Execution Plan:**
- [Task 1: Directory Move](./task-001-directory-move.md) — 基础：移动 web、迁移脚本
- [Task 2: Moon Config](./task-002-moon-config.md) — Moon 工作区和任务定义
- [Task 3: Buf Generation](./task-003-buf-generation.md) — buf 生成链路路径适配
- [Task 4: Docker Adaptation](./task-004-docker-adaptation.md) — Dockerfile 和 compose 构建上下文
- [Task 5: Toolchain Scripts](./task-005-toolchain-scripts.md) — moon-cli.sh 和 dev-panel-web.sh
- [Task 6: CI Workflow](./task-006-ci-workflow.md) — GitHub Actions 门禁
- [Task 7: Documentation](./task-007-documentation.md) — AGENTS.md 和 README 更新
- [Task 8: Cleanup](./task-008-cleanup.md) — 删除残留文件
- [Task 9: Regression](./task-009-regression.md) — 全量回归验证

**Dependencies:**
```
Task 1 → Task 2, Task 3, Task 4, Task 5
Task 2 + Task 3 → Task 9
Task 4 → Task 9
Task 6, Task 7, Task 8 可独立并行
Task 9 最后执行
```

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-03-01-moon-optimization-plan/`.

Execution options:

1. **Subagent-Driven (this session)** — Use Skill tool load `superpowers:subagent-driven-development` skill.
2. **Parallel Session (separate)** — Open new session with `superpowers:executing-plans`, batch execution with checkpoints.
