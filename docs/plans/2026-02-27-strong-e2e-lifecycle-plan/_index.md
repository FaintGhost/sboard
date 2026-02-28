# Strong E2E Lifecycle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 在现有 Playwright + Docker E2E 基础上，补齐“订阅可用（配置 + 流量）”强验收闭环，覆盖 bootstrap 到节点通信与订阅消费的完整生命周期。

**Architecture:** 维持 `panel + node + playwright` 三容器结构，新增 `sb-client`（订阅驱动 sing-box 客户端）与 `probe`（内部流量目标服务）。通过 bridge 网络完成服务互通，测试主断言以 Node `inbounds` 统计增量为准，`traffic` 统计作为辅助信号。

**Tech Stack:** Playwright Test, Bun, Docker Compose, TypeScript, sing-box

**Design Support:**
- [BDD Specs](../2026-02-23-e2e-testing-design/bdd-specs.md)
- [Architecture](../2026-02-23-e2e-testing-design/architecture.md)
- [Best Practices](../2026-02-23-e2e-testing-design/best-practices.md)
- 本次会话已确认约束：`配置+流量验证`、`bridge 网络`、`sing-box 客户端容器`

## Execution Plan

- [Task 001 Test: 运行时拓扑与健康门禁（RED）](./task-001-runtime-topology-test.md)
- [Task 001 Impl: 运行时拓扑与健康门禁（GREEN）](./task-001-runtime-topology-impl.md)
- [Task 002 Test: 生命周期资源编排与手动同步（RED）](./task-002-lifecycle-sync-test.md)
- [Task 002 Impl: 生命周期资源编排与手动同步（GREEN）](./task-002-lifecycle-sync-impl.md)
- [Task 003 Test: 订阅配置可消费性（RED）](./task-003-subscription-consumption-test.md)
- [Task 003 Impl: 订阅配置可消费性（GREEN）](./task-003-subscription-consumption-impl.md)
- [Task 004 Test: 目标入站流量归因（RED）](./task-004-traffic-attribution-test.md)
- [Task 004 Impl: 目标入站流量归因（GREEN）](./task-004-traffic-attribution-impl.md)
- [Task 005 Test: 稳定性与非 flaky 约束（RED）](./task-005-stability-guard-test.md)
- [Task 005 Impl: 稳定性与非 flaky 约束（GREEN）](./task-005-stability-guard-impl.md)

---

## Commit Boundaries

- 边界 1：Task 001（compose 与运行时门禁）
- 边界 2：Task 002（生命周期主链路）
- 边界 3：Task 003（订阅到客户端配置消费）
- 边界 4：Task 004（流量归因断言）
- 边界 5：Task 005（稳定性治理与回归门禁）

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-27-strong-e2e-lifecycle-plan/`.

To execute this plan, use `/superpowers:executing-plans docs/plans/2026-02-27-strong-e2e-lifecycle-plan/`.
