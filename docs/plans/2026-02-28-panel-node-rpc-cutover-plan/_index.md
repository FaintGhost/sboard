# Panel-Node RPC Cutover Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 在保持订阅 REST 兼容的前提下，将 Panel↔Node 通信从 REST 直接切换为 Connect RPC，并完成可回归验证。

**Architecture:** 先以 BDD 场景驱动补齐失败测试（Red），再逐步实现 Node 控制面 RPC、Panel Node RPC 客户端与监控链路切换（Green），最后完成 REST 边界收口与 e2e 回归。执行中保持每个特性成对任务（test/impl），并严格使用真实技术依赖而非线性串行。

**Tech Stack:** Go 1.25, Connect RPC, Protobuf/buf, Gin, Bun, Playwright, SQLite。

**Design Support:**
- [Design Index](../2026-02-28-panel-node-rpc-cutover-design/_index.md)
- [BDD Specs](../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md)
- [Architecture](../2026-02-28-panel-node-rpc-cutover-design/architecture.md)
- [Best Practices](../2026-02-28-panel-node-rpc-cutover-design/best-practices.md)

**Execution Plan:**
- [Task 001: Sync Success Red Test](./task-001-sync-success-test.md)
- [Task 001: Sync Success Green Impl](./task-001-sync-success-impl.md)
- [Task 002: Sync Error Mapping Red Test](./task-002-sync-error-mapping-test.md)
- [Task 002: Sync Error Mapping Green Impl](./task-002-sync-error-mapping-impl.md)
- [Task 003: Monitor RPC Telemetry Red Test](./task-003-monitor-rpc-telemetry-test.md)
- [Task 003: Monitor RPC Telemetry Green Impl](./task-003-monitor-rpc-telemetry-impl.md)
- [Task 004: Concurrency Lock Red Test](./task-004-concurrency-lock-test.md)
- [Task 004: Concurrency Lock Green Impl](./task-004-concurrency-lock-impl.md)
- [Task 005: REST Boundary Red Test](./task-005-rest-boundary-test.md)
- [Task 005: REST Boundary Green Impl](./task-005-rest-boundary-impl.md)
- [Task 006: SS2022 Compatibility Red Test](./task-006-ss2022-compat-test.md)
- [Task 006: SS2022 Compatibility Green Impl](./task-006-ss2022-compat-impl.md)
- [Task 007: E2E Cutover Red Test](./task-007-e2e-cutover-test.md)
- [Task 007: E2E Cutover Green Impl](./task-007-e2e-cutover-impl.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-28-panel-node-rpc-cutover-plan/`. Execution options:

1. Orchestrated Execution (Recommended) - Use Skill tool load `superpowers:executing-plans` skill.
2. Direct Agent Team - Use Skill tool load `superpowers:agent-team-driven-development` skill.
3. BDD-Focused Execution - Use Skill tool load `superpowers:behavior-driven-development` skill for scenario-by-scenario execution.
