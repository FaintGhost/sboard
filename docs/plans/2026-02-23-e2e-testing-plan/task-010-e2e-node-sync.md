# Task 010: E2E - 配置同步与验证测试

**depends-on**: task-002

## Description

实现配置同步全链路的 E2E 测试：创建入站配置并关联节点、触发同步、通过 Node API 验证配置正确性和 sing-box 运行状态。这是最关键的全链路测试。

## Execution Context

**Task Number**: 010 of 012
**Phase**: Integration
**Prerequisites**: Task 002 完成（fixtures 可用，特别是 nodeAPI fixture）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "配置同步与验证" — 全部 3 个场景

## Files to Create

- `e2e/tests/e2e/node-sync.spec.ts`

## Steps

### Step 1: 研究配置同步流程

深入了解 Panel↔Node 的配置同步机制：
- 查看 `panel/proto/sboard/panel/v1/panel.proto` 中 InboundService、SyncJobService、NodeService 相关 RPC
- 了解入站配置的创建流程和字段（协议类型、端口、关联节点等）
- 了解同步触发方式（手动触发 API 或 UI 操作）
- 查看 Node API：`POST /api/config/sync`（接收配置）、可能的配置查询端点
- 了解 sing-box 状态查询方式（Node 是否有 API 端点暴露 sing-box 运行状态）

### Step 2: 实现配置同步测试

创建 `node-sync.spec.ts`，使用 `authenticatedPage` + `panelAPI` + `nodeAPI` fixtures。使用 `test.describe.serial`。

**前置 setup**（`beforeAll`）：通过 API 快速创建：
- 一个节点（api_address=node, api_port=3000, secret_key=e2e-test-node-secret）
- 等待节点状态变为在线

测试场景：

1. **创建入站配置并同步到节点**：在 UI 中创建一个入站配置（选择支持的协议类型，如 VLESS/Shadowsocks），关联到测试节点，触发同步，验证同步状态显示成功

2. **验证 Node 接收到正确配置（API 级别）**：同步成功后，通过 Node API 查询当前状态，验证：
   - Node `/api/health` 返回正常
   - 通过适当的 Node API 端点验证 sing-box 配置已加载
   - 验证 sing-box 进程状态为 running

3. **修改入站配置后重新同步**：在 Panel UI 中修改入站配置（如改变端口或参数），重新触发同步，通过 Node API 验证配置已更新为最新版本

### Step 3: 在 Docker 环境中验证

运行 e2e project 中的 node-sync 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
  docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=e2e tests/e2e/node-sync.spec.ts
```

## Success Criteria

- 3 个配置同步场景测试全部通过
- 入站配置创建并关联节点成功
- 同步触发后 Node 端配置正确更新
- sing-box 进程处于运行状态
- 修改配置后重新同步能正确更新
