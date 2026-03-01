# Task 002: Sync Error Mapping Green Impl

**depends-on**: task-002-sync-error-mapping-test

## Description

实现 Node RPC 鉴权、输入校验与 Panel 侧 Connect code 映射，使同步失败分支满足 BDD 预期。

## Execution Context

**Task Number**: 004 of 014
**Phase**: Core Features
**Prerequisites**: `task-002-sync-error-mapping-test` 已完成并处于 Red 状态

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

- Create: `node/internal/rpc/auth_interceptor.go`
- Modify: `node/internal/rpc/services_impl.go`
- Modify: `panel/internal/node/client.go`
- Modify: `panel/internal/rpc/services_impl.go`

## Steps

### Step 1: Implement Logic (Green)
- 为 Node 控制面建立统一 Bearer secret 鉴权拦截器，并对公开健康检查设定明确策略。
- 将 Node 侧校验失败映射为 `invalid_argument`，将认证失败映射为 `unauthenticated`。
- 在 Panel 侧将 Connect 错误码映射为统一业务错误摘要与 `sync_jobs/sync_attempts` 状态。
- 对可重试与不可重试错误做明确分流，避免对配置错误重试。

### Step 2: Verify Green State
- 重跑 `task-002` 定向测试并确认通过。
- 补跑 `task-001` 测试，确认成功路径无回归。

## Verification Commands

```bash
cd panel && go test ./internal/rpc -run 'TestNodeSyncRPC(Success|ErrorMapping)' -count=1
cd node && go test ./internal/rpc -run 'TestNodeControl(Auth|InvalidArgument|SyncConfigSuccess)' -count=1
```

## Success Criteria

- 三类错误能稳定映射到预期 Connect code 与作业状态。
- 不再依赖 HTTP 状态文本解析作为主判断依据。
