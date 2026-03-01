# Task 004: Concurrency Lock Green Impl

**depends-on**: task-004-concurrency-lock-test

## Description

确保 RPC 切换后仍沿用并正确应用“每节点串行同步锁”，避免并发同步导致配置覆盖或状态错乱。

## Execution Context

**Task Number**: 008 of 014
**Phase**: Integration
**Prerequisites**: `task-004-concurrency-lock-test` 已完成并处于 Red 状态

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

- Modify: `panel/internal/rpc/services_impl.go`
- Modify: `panel/internal/node/client.go`

## Steps

### Step 1: Implement Logic (Green)
- 确认并修正同步锁在 RPC 路径上的覆盖范围，确保同一节点请求仍被串行执行。
- 保持作业写入、错误返回与最终状态更新在并发下仍一致。
- 避免为此引入跨节点的全局串行化。

### Step 2: Verify Green State
- 重跑 `task-004` 定向测试，确认通过。
- 补跑 `task-001` 和 `task-002` 的同步相关测试，确认没有引入死锁或回归。

## Verification Commands

```bash
cd panel && go test ./internal/rpc -run 'TestNodeSyncRPC(Success|ErrorMapping|SerializesPerNode)' -count=1
cd panel && go test ./internal/node -run TestRPCClientConcurrentSyncOrdering -count=1
```

## Success Criteria

- 同一节点并发同步在 RPC 路径上仍被串行化。
- 不影响其他节点的独立同步能力。
