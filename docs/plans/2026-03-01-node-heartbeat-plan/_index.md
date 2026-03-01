# Node→Panel 心跳注册 — 实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to execute this plan.

**Goal:** 实现 Node→Panel 双向心跳注册功能，使 Panel 重置后能自动发现已有 Node。

**Design:** [Node→Panel 心跳注册设计](../2026-03-01-node-heartbeat-design/_index.md)

**Tech Stack:** Go (ConnectRPC), TypeScript (React), SQLite, Protobuf

---

## Architecture Summary

- Node 新增 `PANEL_URL` 配置，定期向 Panel 发送 Heartbeat RPC
- Panel 新增 `NodeRegistrationService.Heartbeat` 公开端点
- 未知 Node 创建 `status=pending` 记录，管理员手动审批
- Docker Compose 模板包含 `PANEL_URL` 和 `NODE_UUID`

## Dependency Graph

```
Tier 0 (独立):  001-proto  002-compose  003-uuid  004-db
                    │                      │         │
Tier 1:         005-heartbeat-handler  006-node-hb  007-approve
                  (001+004)            (001+003)    (001+004)
                                                       │
Tier 2:                              008-monitor    009-frontend
                                      (004)          (007)
```

## Execution Plan

### Tier 0 — Foundation (可并行)

- [Task 001: Proto 定义与代码生成](./task-001-proto-codegen.md)
- [Task 002: Docker Compose 模板 — 测试](./task-002-compose-test.md)
- [Task 002: Docker Compose 模板 — 实现](./task-002-compose-impl.md)
- [Task 003: Node UUID 持久化 — 测试](./task-003-uuid-test.md)
- [Task 003: Node UUID 持久化 — 实现](./task-003-uuid-impl.md)
- [Task 004: Panel DB 层扩展 — 测试](./task-004-db-test.md)
- [Task 004: Panel DB 层扩展 — 实现](./task-004-db-impl.md)

### Tier 1 — Core Logic (依赖 Tier 0)

- [Task 005: Panel Heartbeat Handler — 测试](./task-005-heartbeat-handler-test.md)
- [Task 005: Panel Heartbeat Handler — 实现](./task-005-heartbeat-handler-impl.md)
- [Task 006: Node 心跳 Goroutine — 测试](./task-006-node-heartbeat-test.md)
- [Task 006: Node 心跳 Goroutine — 实现](./task-006-node-heartbeat-impl.md)
- [Task 007: Panel 审批/拒绝 RPC — 测试](./task-007-approve-reject-test.md)
- [Task 007: Panel 审批/拒绝 RPC — 实现](./task-007-approve-reject-impl.md)

### Tier 2 — Adaptation & UI (依赖 Tier 1)

- [Task 008: NodesMonitor 适配 — 测试](./task-008-monitor-test.md)
- [Task 008: NodesMonitor 适配 — 实现](./task-008-monitor-impl.md)
- [Task 009: 前端待认领节点 UI](./task-009-frontend-pending-ui.md)

### Verification

- [Task 010: 集成验证](./task-010-integration-verify.md)
