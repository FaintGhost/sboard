# Task 004: Concurrency Lock Red Test

**depends-on**: task-001-sync-success-impl

## Description

为“同一节点并发同步保护”场景建立失败测试，确保切换协议后仍保持每节点串行化。

## Execution Context

**Task Number**: 007 of 014
**Phase**: Integration
**Prerequisites**: `task-001-sync-success-impl` 已提供可用的同步入口

## BDD Scenario

```gherkin
Scenario: 并发同步请求串行化
  Given 同一节点在短时间内收到两个同步请求
  When 两个请求并发执行
  Then 节点最终配置与最后一次成功请求一致
  And 不出现中间态损坏
```

**Spec Source**: `../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md`

## Files to Modify/Create

- Create: `panel/internal/rpc/node_sync_concurrency_test.go`
- Create: `panel/internal/node/client_concurrency_test.go`

## Steps

### Step 1: Verify Scenario
- 明确该任务仅验证并发串行与结果一致性，不扩大到多节点批量场景。

### Step 2: Implement Test (Red)
- 在 Panel 侧新增并发测试，模拟同一节点的两个同步请求并断言串行行为。
- 使用 test doubles 模拟 Node RPC 服务的阻塞与响应顺序，避免真实网络竞态。
- 确保失败原因是“未保持串行保护”或“结果不一致”。

### Step 3: Verify Red State
- 运行定向测试，确认失败（Red）。

## Verification Commands

```bash
cd panel && go test ./internal/rpc -run TestNodeSyncRPCSerializesPerNode -count=1
cd panel && go test ./internal/node -run TestRPCClientConcurrentSyncOrdering -count=1
```

## Success Criteria

- 并发串行场景有可重复执行的测试覆盖。
- 测试失败能稳定复现当前缺口或保护缺失。
