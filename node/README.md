# sboard node (VPS)

`node` 是实际承载入站的节点进程，提供一个 HTTP API 供 `panel` 下发/同步配置。

## 方式 1: Docker Compose (推荐)

在 VPS 上克隆仓库后，进入 `node/` 目录:

```bash
cd sboard/node
export NODE_SECRET_KEY='change-me'
export NODE_LOG_LEVEL=info
docker compose up -d
```

说明:
  - `docker-compose.yml` 使用 `network_mode: host`，这样 `panel` 下发的入站端口(例如 443)能直接绑定到宿主机端口。
  - 同时也意味着 `NODE_HTTP_ADDR=:3000` 会暴露在公网。务必在 VPS 防火墙中只允许 `panel` 服务器 IP 访问 `3000/tcp`。
  - 默认会拉取 Docker Hub 镜像：`SBOARD_NODE_IMAGE`（默认 `faintghost/sboard-node:latest`）。

## 方式 2: 裸机运行

```bash
cd sboard/node
export NODE_SECRET_KEY='change-me'
export NODE_LOG_LEVEL=info
go run ./cmd/node
```

## 构建并推送到 Docker Hub (本机)

适用场景: VPS 性能较弱，不适合 `docker build`。

假设你的 Docker Hub 仓库是 `faintghost/sboard-node`，在本机仓库根目录执行:

```bash
cd sboard/node
export SBOARD_NODE_IMAGE="faintghost/sboard-node:latest"
docker compose -f docker-compose.yml -f docker-compose.build.yml build --no-cache
docker push "$SBOARD_NODE_IMAGE"
```

建议额外推一个不可变 tag（便于回滚），例如用 git sha:

```bash
export TAG="$(git rev-parse --short HEAD)"
export SBOARD_NODE_IMAGE="faintghost/sboard-node:$TAG"
docker compose -f docker-compose.yml -f docker-compose.build.yml build --no-cache
docker push "$SBOARD_NODE_IMAGE"
```

## 常见问题

### Docker 构建报 missing go.sum entry

本项目的 `node/` 目录早期可能没有提交 `go.sum`。`Dockerfile` 会在构建时执行 `go mod tidy` 以自动生成并补齐依赖校验和。

如果你仍然在 VPS 上遇到类似错误，可先在 `node/` 目录执行:

```bash
GOWORK=off go mod tidy
```

然后再 `docker compose -f docker-compose.yml -f docker-compose.build.yml build`。
