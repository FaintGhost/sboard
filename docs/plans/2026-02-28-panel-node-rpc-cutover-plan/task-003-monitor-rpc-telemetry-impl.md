# Task 003: Monitor RPC Telemetry Green Impl

**depends-on**: task-003-monitor-rpc-telemetry-test

## Description

将 Panel monitor 与 Panel Node 客户端的健康/流量/入站统计读取迁移到 Node RPC，并实现 Node 侧对应服务方法。

## Execution Context

**Task Number**: 006 of 014
**Phase**: Core Features
**Prerequisites**: `task-003-monitor-rpc-telemetry-test` 已完成并处于 Red 状态

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

- Modify: `panel/internal/node/client.go`
- Modify: `panel/internal/monitor/nodes_monitor.go`
- Modify: `panel/internal/monitor/traffic_monitor.go`
- Modify: `node/internal/rpc/services_impl.go`
- Modify: `node/cmd/node/main.go`

## Steps

### Step 1: Implement Logic (Green)
- 为 Node 控制面补齐 `Health`、`GetTraffic`、`GetInboundTraffic` RPC 方法。
- 在 Panel Node 客户端中使用 RPC 返回值填充现有结构体。
- 让 `nodes_monitor` 与 `traffic_monitor` 改用新客户端，但保持现有业务判断与状态更新逻辑。

### Step 2: Verify Green State
- 重跑 `task-003` 定向测试并确认通过。
- 补跑已有 monitor 与 node 客户端相关测试，确认无回归。

## Verification Commands

```bash
cd panel && go test ./internal/monitor -run TestNodeRPCTelemetryMonitor -count=1
cd panel && go test ./internal/node -run TestRPCClientTelemetry -count=1
cd node && go test ./internal/rpc -run TestNodeControlTelemetry -count=1
```

## Success Criteria

- 监控链路不再依赖 Node REST 统计接口。
- 健康、流量、入站统计场景测试通过。
