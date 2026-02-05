# SBoard

SBoard 是一个基于 sing-box 的订阅管理面板与节点管理系统。当前实现为阶段 1 基础框架，支持 Panel 与 Node 的最小可运行链路，以及 Node 配置下发与入站创建。

**架构概览**
- Panel：管理面板后端（Gin + SQLite），提供健康检查与后续扩展入口
- Node：实际承载入站的节点服务（内嵌 sing-box），接收 Panel 下发的入站配置

**前置条件**
- Go 1.25+
- Linux 服务器（推荐）
- 可写目录用于 SQLite 数据库与运行时文件

**快速开始（裸机）**
1. 启动 Panel
```bash
cd /root/workspace/sboard/panel
PANEL_HTTP_ADDR=:8080 \
PANEL_DB_PATH=panel.db \
go run ./cmd/panel
```

2. 启动 Node
```bash
cd /root/workspace/sboard/node
NODE_HTTP_ADDR=:3000 \
NODE_SECRET_KEY=secret \
NODE_LOG_LEVEL=info \
go run ./cmd/node
```

3. 配置下发（示例）
```bash
curl -X POST http://127.0.0.1:3000/api/config/sync \
  -H "Authorization: Bearer secret" \
  -d '{"inbounds":[{"type":"mixed","tag":"m1","listen":"0.0.0.0","listen_port":1080}]}'
```

**Docker 部署（示例）**
当前仓库未内置 Dockerfile，以下为参考示例，按需调整。

示例 `Dockerfile.panel`：
```Dockerfile
FROM golang:1.25 as builder
WORKDIR /app
COPY . .
WORKDIR /app/panel
RUN go build -o /out/panel ./cmd/panel

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /out/panel /app/panel
ENV PANEL_HTTP_ADDR=:8080
ENV PANEL_DB_PATH=/data/panel.db
EXPOSE 8080
CMD ["/app/panel"]
```

示例 `Dockerfile.node`：
```Dockerfile
FROM golang:1.25 as builder
WORKDIR /app
COPY . .
WORKDIR /app/node
RUN go build -o /out/node ./cmd/node

FROM debian:stable-slim
WORKDIR /app
COPY --from=builder /out/node /app/node
ENV NODE_HTTP_ADDR=:3000
ENV NODE_SECRET_KEY=secret
ENV NODE_LOG_LEVEL=info
EXPOSE 3000
CMD ["/app/node"]
```

构建与运行示例：
```bash
docker build -t sboard-panel -f Dockerfile.panel .
docker build -t sboard-node -f Dockerfile.node .

docker run -d --name sboard-panel -p 8080:8080 -v $(pwd)/data:/data sboard-panel
docker run -d --name sboard-node -p 3000:3000 sboard-node
```

**配置说明**
- Panel
  - `PANEL_HTTP_ADDR`：监听地址，默认 `:8080`
  - `PANEL_DB_PATH`：SQLite 路径，默认 `panel.db`
- Node
  - `NODE_HTTP_ADDR`：监听地址，默认 `:3000`
  - `NODE_SECRET_KEY`：API 密钥，用于 `Authorization: Bearer <secret>`
  - `NODE_LOG_LEVEL`：日志级别，默认 `info`

**基础 API**
- `GET /api/health`：Panel 与 Node 均提供健康检查
- `POST /api/config/sync`：Node 接收入站配置并创建入站

**目录结构**
- `panel/`：Panel 后端（Gin + SQLite）
- `node/`：Node 服务（嵌入 sing-box）
- `docs/`：设计与规划文档

**Roadmap**
- 用户与节点管理 API
- 入站配置管理与订阅生成
- 前端管理面板
