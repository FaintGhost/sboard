# Makefile to Moon Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 完全弃用根 `Makefile`，迁移到 Moon 作为 monorepo 唯一任务编排入口，并由 Moon 全量接管 `Go + Node + Bun` 工具链且严格锁定版本，同时保持现有生成、门禁与 E2E 行为等价。

**Architecture:** 先建立“版本锁定 + Moon 工作区骨架”并用可重复的验证脚本锁定预期行为（Red），再逐步接入 `generate` / `check-generate` / `e2e*` / `panel/web` 门禁任务（Green），最后切换文档入口并删除根 `Makefile`。执行中保持任务按真实技术依赖串联，而不是人为线性排队。

**Tech Stack:** Moon v2, proto, Go 1.26.0, Node 22.22.0, Bun 1.3.9, Buf, Docker Compose, Playwright。

**Design Support:**
- [Design Index](../2026-03-01-makefile-to-moon-design/_index.md)
- [BDD Specs](../2026-03-01-makefile-to-moon-design/bdd-specs.md)
- [Architecture](../2026-03-01-makefile-to-moon-design/architecture.md)
- [Best Practices](../2026-03-01-makefile-to-moon-design/best-practices.md)

**Execution Plan:**
- [Task 001: Workspace Bootstrap Test](./task-001-workspace-bootstrap-test.md)
- [Task 001: Workspace Bootstrap Impl](./task-001-workspace-bootstrap-impl.md)
- [Task 002: Generate Alias Test](./task-002-generate-alias-test.md)
- [Task 002: Generate Alias Impl](./task-002-generate-alias-impl.md)
- [Task 003: Check Generate Test](./task-003-check-generate-test.md)
- [Task 003: Check Generate Impl](./task-003-check-generate-impl.md)
- [Task 004: E2E Alias Test](./task-004-e2e-alias-test.md)
- [Task 004: E2E Alias Impl](./task-004-e2e-alias-impl.md)
- [Task 005: Docs Cutover Test](./task-005-docs-cutover-test.md)
- [Task 005: Docs Cutover Impl](./task-005-docs-cutover-impl.md)
- [Task 006: Panel Gates Test](./task-006-panel-gates-test.md)
- [Task 006: Panel Gates Impl](./task-006-panel-gates-impl.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-03-01-makefile-to-moon-plan/`.

Execution options:

1. Orchestrated Execution (Recommended) - Use Skill tool load `superpowers:executing-plans` skill.
2. Direct Agent Team - Use Skill tool load `superpowers:agent-team-driven-development` skill.
3. BDD-Focused Execution - Use Skill tool load `superpowers:behavior-driven-development` skill for scenario-by-scenario execution.
