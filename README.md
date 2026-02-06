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
PANEL_CORS_ALLOW_ORIGINS=http://localhost:5173 \
ADMIN_USER=admin \
ADMIN_PASS=admin123 \
PANEL_JWT_SECRET=change-me-in-prod \
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

**Docker 部署**
Node 侧已内置 Docker 部署文件（见下方 Compose）。Panel 侧暂未内置 Dockerfile，以下为参考示例，按需调整。

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

**Node Docker Compose（推荐，海外 VPS）**
仓库已内置 `node/docker-compose.yml` 与 `node/Dockerfile`，可以在海外 VPS 直接构建运行：
```bash
cd node
export NODE_SECRET_KEY='change-me'
docker compose up -d --build
```
说明：
- compose 使用 `network_mode: host`，方便入站直接监听宿主机端口（例如 443）
- 这会让 Node API（默认 `:3000`）暴露在公网：请用防火墙只允许 Panel 服务器 IP 访问 3000

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
  - `ADMIN_USER`：管理员登录用户名（必填）
  - `ADMIN_PASS`：管理员登录密码（必填）
  - `PANEL_JWT_SECRET`：JWT 签名密钥（必填）
  - `PANEL_CORS_ALLOW_ORIGINS`：允许的前端 Origin（逗号分隔或 `*`），默认 `http://localhost:5173`
  - `PANEL_LOG_REQUESTS`：是否打印每个 HTTP 请求（`true/false`），默认 `true`
- Node
  - `NODE_HTTP_ADDR`：监听地址，默认 `:3000`
  - `NODE_SECRET_KEY`：API 密钥，用于 `Authorization: Bearer <secret>`
  - `NODE_LOG_LEVEL`：日志级别，默认 `info`

**前后端联调（本地）**
1. 启动 Panel（如上，确保设置 `ADMIN_USER/ADMIN_PASS/PANEL_JWT_SECRET`）
2. 启动前端：
```bash
cd /root/workspace/sboard/panel/web
npm run dev
```
3. Vite 会把 `/api/*` 代理到 Panel（默认 `http://127.0.0.1:8080`）。如需自定义目标：
```bash
VITE_PROXY_TARGET=http://127.0.0.1:8080 npm run dev
```
3. 前端登录账号密码：使用你启动 Panel 时传入的 `ADMIN_USER` 和 `ADMIN_PASS`  
   示例即 `admin / admin123`

**基础 API**
- `GET /api/health`：Panel 与 Node 均提供健康检查
- `POST /api/config/sync`：Node 接收入站配置并创建入站
- `GET /api/sub/:user_uuid`：订阅链接（sing-box）
- `GET /api/groups`：分组列表（需要管理员 JWT）
- `PUT /api/users/:id/groups`：设置用户所属分组（需要管理员 JWT）

**订阅行为**
- `?format=singbox`：返回 sing-box JSON
- `?format=v2ray`：返回 v2ray 风格订阅（Base64(多行分享链接)）
- 未指定 `format`：
  - User-Agent 命中 `sing-box`/`SFA`/`SFI` 返回 sing-box JSON
  - 其他 User-Agent 返回 v2ray 风格订阅

**节点与入站字段**
- `nodes.api_address/api_port`：Panel ↔ Node 通信地址
- `nodes.public_address`：订阅中使用的对外地址
- `inbounds.public_port`：订阅中使用的对外端口（为空时回退 `listen_port`）
- `nodes.group_id`：节点所属分组（订阅按分组下发）

**目录结构**
- `panel/`：Panel 后端（Gin + SQLite）
- `node/`：Node 服务（嵌入 sing-box）
- `docs/`：设计与规划文档

**Roadmap**
- 用户与节点管理 API
- 入站配置管理与订阅生成
- 前端管理面板
