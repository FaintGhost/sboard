# Task 003: Monitor RPC Telemetry Red Test

**depends-on**: task-001-sync-success-impl

## Description

为监控链路改走 RPC 建立失败测试，覆盖健康检查、网卡流量、入站统计重置三个场景。

## Execution Context

**Task Number**: 005 of 014
**Phase**: Core Features
**Prerequisites**: `task-001-sync-success-impl` 已提供 Node RPC 客户端基础能力

## BDD Scenario

```gherkin
Scenario: 节点健康检查
  Given Node RPC Health 可访问
  When Panel monitor 执行健康探测
  Then 节点状态被正确更新为 online

Scenario: 获取网卡流量
  Given Node 正常运行
  When Panel monitor 拉取 traffic 样本
  Then 返回包含 rx_bytes 和 tx_bytes 的结果

Scenario: 获取入站统计并重置
  Given Node 有入站流量数据
  When Panel 请求 InboundTraffic 且 reset=true
  Then 返回 data 与 meta
  And Node 端统计被按预期重置
```

**Spec Source**: `../2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md`

## Files to Modify/Create

- Create: `panel/internal/monitor/node_rpc_monitor_test.go`
- Create: `panel/internal/node/client_telemetry_test.go`
- Create: `node/internal/rpc/telemetry_test.go`

## Steps

### Step 1: Verify Scenario
- 将三条监控场景拆为同一特性下的三个测试分支，统一验证 RPC 监控读取行为。

### Step 2: Implement Test (Red)
- 在 Panel monitor 测试中断言健康状态更新、流量样本采集与入站 reset 行为。
- 在 Panel Node 客户端测试中断言流量/入站统计的返回字段映射。
- 在 Node 侧测试中用 test doubles 隔离真实统计提供者与系统网卡读取。

### Step 3: Verify Red State
- 运行定向测试，确认失败（Red）。

## Verification Commands

```bash
cd panel && go test ./internal/monitor -run TestNodeRPCTelemetryMonitor -count=1
cd panel && go test ./internal/node -run TestRPCClientTelemetry -count=1
cd node && go test ./internal/rpc -run TestNodeControlTelemetry -count=1
```

## Success Criteria

- 三条监控场景均被测试描述。
- 测试失败原因指向监控链路尚未切到 RPC。
