# Task 001: Sync Success Green Impl

**depends-on**: task-001-sync-success-test

## Description

实现支撑“同步成功”场景的最小闭环：定义 Node 控制面 proto、生成代码、落地 Node RPC 服务与 Panel Node RPC 客户端的成功路径。

## Execution Context

**Task Number**: 002 of 014
**Phase**: Foundation
**Prerequisites**: `task-001-sync-success-test` 已完成并处于 Red 状态

## BDD Scenario

```gherkin
Scenario: 同步成功
  Given Panel 与 Node 的 RPC 服务均已启动
  And Node 鉴权密钥与 Panel 中配置一致
  When 管理员触发节点同步
  Then Node 成功应用配置
  And SyncJob 状态为 success
```

**Spec Source**: `../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md`

## Files to Modify/Create

- Create: `panel/proto/sboard/node/v1/node.proto`
- Create: `node/internal/rpc/server.go`
- Create: `node/internal/rpc/services_impl.go`
- Modify: `panel/buf.gen.yaml`
- Modify: `panel/internal/node/client.go`
- Modify: `panel/internal/rpc/services_impl.go`
- Modify: `node/cmd/node/main.go`

## Steps

### Step 1: Implement Logic (Green)
- 新增 `NodeControlService` 契约，并将其纳入生成链路。
- 在 Node 侧注册 RPC handler，复用现有配置解析与应用能力，打通成功同步路径。
- 在 Panel 侧将 Node 客户端切换为 RPC 调用，并保持现有 `Health/SyncConfig/Traffic/InboundTraffic*` 方法签名尽量稳定。
- 让 `runNodeSync` 成功调用新客户端并继续写入成功的 `sync_jobs/sync_attempts`。

### Step 2: Verify Green State
- 重新运行 `task-001` 的定向测试，确认通过。
- 运行生成检查，确认生成产物已纳入门禁。

### Step 3: Refactor Safely
- 仅在不改变场景结果的前提下整理重复逻辑与命名。

## Verification Commands

```bash
make check-generate
cd panel && go test ./internal/rpc -run TestNodeSyncRPCSuccess -count=1
cd node && go test ./internal/rpc -run TestNodeControlSyncConfigSuccess -count=1
```

## Success Criteria

- `NodeControlService` 契约进入生成链路。
- Panel 与 Node 的同步成功路径改由 RPC 打通。
- `task-001` 新增测试通过。
