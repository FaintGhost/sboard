# Task 006: Node 心跳 Goroutine — 测试

**depends-on:** task-001-proto-codegen, task-003-uuid-impl

## Summary

为 Node 心跳 goroutine 编写测试，验证定时发送、优雅停止、失败容错行为。

## BDD Scenario

```gherkin
Scenario: Node sends heartbeat when PANEL_URL is configured
  Given a Node running with PANEL_URL="http://panel:8080" and NODE_UUID="uuid-1"
  And the Node has SECRET_KEY="key-1"
  When 30 seconds elapse
  Then Node sends a Heartbeat RPC to Panel with uuid="uuid-1", secret_key="key-1"
  And the heartbeat includes Node version and api_addr

Scenario: Node does NOT send heartbeat when PANEL_URL is empty
  Given a Node running without PANEL_URL configured
  When 60 seconds elapse
  Then no Heartbeat RPC is sent
  And Node operates in passive-only mode (current behavior)
```

## Files

- **Create:** `node/internal/heartbeat/heartbeat_test.go`
  - `TestHeartbeat_SendsOnTick` — mock Panel server，验证 heartbeat 请求到达
  - `TestHeartbeat_DisabledWhenNoPanelURL` — PanelURL="" 时不启动
  - `TestHeartbeat_GracefulShutdown` — cancel context 后 goroutine 退出
  - `TestHeartbeat_ContinuesOnError` — Panel 不可达时不 panic，继续下一个 tick
  - 使用 `httptest.NewServer` mock Panel 端点
  - 使用短间隔（如 50ms）加速测试

## Steps

1. 创建 `node/internal/heartbeat/` 目录和测试文件
2. Mock Panel server 记录收到的请求
3. 使用 `context.WithCancel` 控制 goroutine 生命周期
4. 验证请求中包含 UUID、SecretKey、ApiAddr 字段
5. 测试应该先 FAIL（heartbeat 包不存在）

## Verify

```bash
cd node && go test ./internal/heartbeat/ -run TestHeartbeat -v
```

应该 FAIL（包不存在）。
