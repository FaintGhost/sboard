# sboard node (VPS)

`node` 是实际承载入站的节点进程，提供一个 HTTP API 供 `panel` 下发/同步配置。

## 方式 1: Docker Compose (推荐)

在 VPS 上克隆仓库后，进入 `node/` 目录:

```bash
cd sboard/node
export NODE_SECRET_KEY='change-me'
export NODE_LOG_LEVEL=info
docker compose up -d --build
```

说明:
  - `docker-compose.yml` 使用 `network_mode: host`，这样 `panel` 下发的入站端口(例如 443)能直接绑定到宿主机端口。
  - 同时也意味着 `NODE_HTTP_ADDR=:3000` 会暴露在公网。务必在 VPS 防火墙中只允许 `panel` 服务器 IP 访问 `3000/tcp`。

## 方式 2: 裸机运行

```bash
cd sboard/node
export NODE_SECRET_KEY='change-me'
export NODE_LOG_LEVEL=info
go run ./cmd/node
```

## 常见问题

### Docker 构建报 missing go.sum entry

本项目的 `node/` 目录早期可能没有提交 `go.sum`。`Dockerfile` 已在构建时执行 `go mod tidy -mod=mod` 以自动生成并补齐依赖校验和。

如果你仍然在 VPS 上遇到类似错误，可先在 `node/` 目录执行:

```bash
GOWORK=off go mod tidy
```

然后再 `docker compose build`。

