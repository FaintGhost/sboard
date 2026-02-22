# OpenAPI 统一前后端 API 对接 - 实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use Skill tool load `superpowers:executing-plans` skill to implement this plan task-by-task.

**Goal:** 使用 OpenAPI 3.1 YAML 作为唯一事实来源，通过 oapi-codegen (Go) 和 @hey-api/openapi-ts (TypeScript) 代码生成替换手写 API 类型和客户端代码，建立前后端契约层。

**Architecture:** 手写 `panel/openapi.yaml` 定义所有 ~37 个 API 操作。后端用 oapi-codegen 生成 Gin strict-server 接口和类型，handler 实现 `StrictServerInterface`。前端用 @hey-api/openapi-ts 生成 TypeScript 类型、SDK 客户端和 Zod schema。

**Tech Stack:** OpenAPI 3.1, oapi-codegen v2, @hey-api/openapi-ts, Zod, Gin, React + TanStack Query

**Design Support:**
- [BDD Specs](../2026-02-22-openapi-unification-design/bdd-specs.md)
- [Architecture](../2026-02-22-openapi-unification-design/architecture.md)
- [Best Practices](../2026-02-22-openapi-unification-design/best-practices.md)

## Execution Plan

- [Task 001: Write complete OpenAPI 3.1 spec](./task-001-write-openapi-spec.md)
- [Task 002: Setup Go oapi-codegen toolchain](./task-002-setup-go-codegen.md)
- [Task 003: Setup TypeScript @hey-api/openapi-ts toolchain](./task-003-setup-ts-codegen.md)
- [Task 004: Implement Server struct + Users/UserGroups handlers](./task-004-backend-users-handlers.md)
- [Task 005: Implement Groups/GroupUsers handlers](./task-005-backend-groups-handlers.md)
- [Task 006: Implement Nodes handlers](./task-006-backend-nodes-handlers.md)
- [Task 007: Implement Inbounds handlers](./task-007-backend-inbounds-handlers.md)
- [Task 008: Implement System/Auth/Bootstrap/Traffic/SyncJobs/SingBox handlers](./task-008-backend-remaining-handlers.md)
- [Task 009: Refactor router + integrate auth middleware](./task-009-backend-router-refactor.md)
- [Task 010: Fix and verify all backend tests](./task-010-backend-fix-tests.md)
- [Task 011: Configure frontend client + migrate all pages](./task-011-frontend-client-migration.md)
- [Task 012: Fix frontend tests + delete old API files](./task-012-frontend-fix-tests-cleanup.md)
- [Task 013: Build integration, CI, documentation](./task-013-build-ci-docs.md)

---

## Execution Handoff

Plan complete and saved to `docs/plans/2026-02-22-openapi-unification-plan/`. Execution options:

**1. Orchestrated Execution (Recommended)** - Use Skill tool load `superpowers:executing-plans` skill.

**2. Direct Agent Team** - Use Skill tool load `superpowers:agent-team-driven-development` skill.

**3. BDD-Focused Execution** - Use Skill tool load `superpowers:behavior-driven-development` skill for specific scenarios.
