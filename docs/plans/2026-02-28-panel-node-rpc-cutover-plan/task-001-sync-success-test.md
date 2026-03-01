# Task 001: Sync Success Red Test

## Description

为“同步成功”场景先建立失败测试，锁定新的 Node RPC 契约与主同步链路的预期行为。该任务只创建或改造测试，不实现功能。

## Execution Context

**Task Number**: 001 of 014
**Phase**: Foundation
**Prerequisites**: 无

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

- Create: `panel/internal/rpc/node_sync_success_test.go`
- Create: `node/internal/rpc/services_impl_test.go`
- Modify: `panel/internal/rpc/generate.go`
- Modify: `Makefile`

## Steps

### Step 1: Verify Scenario
- 确认 `bdd-specs.md` 中存在“同步成功”场景，且本任务仅覆盖该场景。

### Step 2: Implement Test (Red)
- 在 Panel 侧新增测试，描述 `NodeService.SyncNode` 通过 Node RPC 客户端成功完成一次同步时的返回值与 `sync_jobs` 成功记录。
- 在 Node 侧新增测试，描述 `NodeControlService.SyncConfig` 成功接收并应用合法 payload 的返回结果。
- 为所有单测使用 test doubles 隔离数据库、网络与 core 应用层，避免真实外部依赖。
- 确保失败原因是“缺少 Node RPC 契约/实现”或“行为未满足”，而不是无意义的导入错误。

### Step 3: Verify Red State
- 运行最小测试命令，确认新测试失败（Red）。

## Verification Commands

```bash
moon run panel:check-generate
cd panel && go test ./internal/rpc -run TestNodeSyncRPCSuccess -count=1
cd node && go test ./internal/rpc -run TestNodeControlSyncConfigSuccess -count=1
```

## Success Criteria

- 新测试已提交到仓库。
- 新测试准确映射该 BDD 场景。
- 测试因功能未实现而失败，而不是因环境配置错误失败。
