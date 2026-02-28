# Task 002: Sync Error Mapping Red Test

**depends-on**: task-001-sync-success-impl

## Description

为同步失败分支建立失败测试，锁定鉴权失败、配置非法、节点不可达三类错误在 Panel 侧的统一映射行为。

## Execution Context

**Task Number**: 003 of 014
**Phase**: Core Features
**Prerequisites**: `task-001-sync-success-impl` 已提供可编译的 Node RPC 契约与成功路径

## BDD Scenario

```gherkin
Scenario: 鉴权失败
  Given Node 鉴权密钥与 Panel 配置不一致
  When 管理员触发节点同步
  Then 返回 unauthenticated 错误
  And SyncJob 状态为 failed
  And 错误摘要可用于定位密钥问题

Scenario: 配置非法
  Given 入站配置包含非法字段
  When 管理员触发节点同步
  Then 返回 invalid_argument 错误
  And SyncJob 状态为 failed
  And 不进入重试

Scenario: 节点不可达
  Given Node 网络不可达
  When 管理员触发节点同步
  Then 返回 unavailable 或 deadline_exceeded
  And SyncJob 状态为 failed
```

**Spec Source**: `../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md`

## Files to Modify/Create

- Create: `panel/internal/rpc/node_sync_error_mapping_test.go`
- Create: `node/internal/rpc/auth_interceptor_test.go`
- Modify: `node/internal/rpc/services_impl_test.go`

## Steps

### Step 1: Verify Scenario
- 确认该任务仅覆盖同步错误映射相关场景，不混入成功路径。

### Step 2: Implement Test (Red)
- 新增 Panel 侧测试，分别断言 `unauthenticated`、`invalid_argument`、`unavailable/deadline_exceeded` 的错误映射与作业落库结果。
- 新增 Node 侧测试，断言鉴权拦截器与输入校验的返回码。
- 使用 test doubles 模拟远端不可达、错误码返回和持久化依赖，避免真实网络。

### Step 3: Verify Red State
- 运行定向测试，确认失败（Red）。

## Verification Commands

```bash
cd panel && go test ./internal/rpc -run TestNodeSyncRPCErrorMapping -count=1
cd node && go test ./internal/rpc -run 'TestNodeControlAuth|TestNodeControlInvalidArgument' -count=1
```

## Success Criteria

- 三类错误场景均有明确测试覆盖。
- 测试失败点指向缺失的错误映射，而不是基础设施问题。
