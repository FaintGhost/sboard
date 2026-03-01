# Task 006: Node 心跳 Goroutine — 实现

**depends-on:** task-006-node-heartbeat-test

## Summary

实现 Node 心跳 goroutine 和 main.go 集成，使 Node 在配置 `PANEL_URL` 后定期向 Panel 发送心跳。

## BDD Scenario

(同 task-006-node-heartbeat-test)

## Files

- **Create:** `node/internal/heartbeat/heartbeat.go`
  - `Run(ctx context.Context, cfg Config)` — 心跳主循环
    - `cfg.PanelURL` 为空时直接 return（不启动）
    - 创建 `panelv1connect.NewNodeRegistrationServiceClient`
    - 启动 `time.NewTicker(cfg.HeartbeatInterval())`
    - 每个 tick 发送 `Heartbeat` RPC
    - 失败仅 `log.Printf`，不中断循环
    - `ctx.Done()` 时退出
  - 心跳 goroutine 需要的 Config 接口或结构体（可直接用 `config.Config`）

- **Modify:** `node/cmd/node/main.go`
  - 在启动 HTTP server 之前，`go heartbeat.Run(ctx, cfg)` 启动心跳 goroutine
  - ctx 与 main 的 signal handler 共享，SIGTERM 时一起退出

- **Modify:** `node/go.mod`
  - 添加对 Panel proto generated code 的依赖（`panelv1connect` client）
  - 注意: Node 需要 import Panel 的 generated connect client。方案: 将 Panel proto 的 Go 生成代码发布为可 import 的包，或在 Node 的 `go.mod` 中用 `replace` 指令引用本地路径。参考现有的 `panel/internal/rpc/gen/` 结构。

## Steps

1. 创建 `heartbeat.go`，实现 `Run` 函数
2. 在 `main.go` 中启动 goroutine
3. 处理 Node 如何 import Panel 的 generated proto code（可能需要将 proto 生成到共享位置或用 go.mod replace）

## Verify

```bash
cd node && go test ./internal/heartbeat/ -v
```
