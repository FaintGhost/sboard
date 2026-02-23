# E2E Testing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 建立基于 Playwright + Docker 的全容器化 E2E 测试框架，覆盖 SBoard 的核心用户流程和 Panel↔Node 全链路验证。

**Architecture:** 三容器编排（panel + node + playwright），通过 docker-compose.e2e.yml 一键启动。测试分为 Smoke（快速冒烟）和 E2E（完整流程）两层，通过 Playwright Projects 区分。每次测试使用全新 SQLite 数据库，通过 fixtures 封装认证和测试数据管理。

**Tech Stack:** Playwright Test, Bun, Docker Compose, TypeScript

**Design Support:**
- [BDD Specs](../2026-02-23-e2e-testing-design/bdd-specs.md)
- [Architecture](../2026-02-23-e2e-testing-design/architecture.md)
- [Best Practices](../2026-02-23-e2e-testing-design/best-practices.md)

**Execution Plan:**
- [Task 001: 项目基础设施搭建](./task-001-project-scaffolding.md)
- [Task 002: 共享 Fixtures 实现](./task-002-shared-fixtures.md)
- [Task 003: Smoke - 健康检查测试](./task-003-smoke-health-check.md)
- [Task 004: Smoke - Bootstrap 初始化测试](./task-004-smoke-bootstrap.md)
- [Task 005: Smoke - 核心页面导航测试](./task-005-smoke-navigation.md)
- [Task 006: E2E - 认证管理测试](./task-006-e2e-auth.md)
- [Task 007: E2E - 用户管理 CRUD 测试](./task-007-e2e-users.md)
- [Task 008: E2E - 分组管理测试](./task-008-e2e-groups.md)
- [Task 009: E2E - 节点管理测试](./task-009-e2e-nodes.md)
- [Task 010: E2E - 配置同步与验证测试](./task-010-e2e-node-sync.md)
- [Task 011: E2E - 订阅管理测试](./task-011-e2e-subscriptions.md)
- [Task 012: Makefile 集成与最终验证](./task-012-makefile-integration.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-23-e2e-testing-plan/`. Execution options:

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
