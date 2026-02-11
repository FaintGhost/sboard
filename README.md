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
PANEL_JWT_SECRET=change-me-in-prod \
# 可选：自定义 onboarding token；不设置会在首次启动日志里生成并打印
PANEL_SETUP_TOKEN= \
GOFLAGS='-tags=with_utls' \
go run ./cmd/panel
```

2. 启动 Node
```bash
cd /root/workspace/sboard/node
NODE_HTTP_ADDR=:3000 \
NODE_SECRET_KEY=secret \
NODE_LOG_LEVEL=info \
GOFLAGS='-tags=with_utls' \
go run ./cmd/node
```

3. 配置下发（示例）
```bash
curl -X POST http://127.0.0.1:3000/api/config/sync \
  -H "Authorization: Bearer secret" \
  -d '{"inbounds":[{"type":"mixed","tag":"m1","listen":"0.0.0.0","listen_port":1080}]}'
```

**Docker 部署**
仓库内置：
- `panel/Dockerfile`、`panel/docker-compose.yml`、`panel/docker-compose.build.yml`
- `node/Dockerfile`、`node/docker-compose.yml`、`node/docker-compose.build.yml`

说明：
- `panel/docker-compose.yml` 与 `node/docker-compose.yml` 默认拉取 Docker Hub 预构建镜像（适配低配 VPS）
- 如需自建镜像，请在本地 build 后 push，再通过 `SBOARD_PANEL_IMAGE` / `SBOARD_NODE_IMAGE` 指定

## Panel Docker Compose（推荐）

仓库已内置 `panel/docker-compose.yml` 与 `panel/Dockerfile`（用于构建镜像）。  
在 VPS 上建议直接拉取预构建镜像（避免低配机器构建失败），并挂载数据目录保存 SQLite。

在服务器上：

```bash
cd sboard/panel
export PANEL_JWT_SECRET='change-me'
export PANEL_SETUP_TOKEN='' # 可选：不设置会在首次启动日志里生成并打印
docker compose up -d
```

说明：
  - 默认镜像为 `faintghost/sboard-panel:latest`，可用 `SBOARD_PANEL_IMAGE` 覆盖
  - Panel 默认会在同一进程内静态托管前端（`PANEL_SERVE_WEB=true`）
  - 数据库默认路径为 `/data/panel.db`（通过 volume 映射到宿主机 `panel/data/`）

**Node Docker Compose（推荐，海外 VPS）**
仓库已内置 `node/docker-compose.yml` 与 `node/Dockerfile`。默认 compose 直接拉取预构建镜像：
```bash
cd node
export NODE_SECRET_KEY='change-me'
docker compose up -d
```
说明：
- compose 使用 `network_mode: host`，方便入站直接监听宿主机端口（例如 443）
- 这会让 Node API（默认 `:3000`）暴露在公网：请用防火墙只允许 Panel 服务器 IP 访问 3000

如果你希望在本机 build 并推送（推荐）：
```bash
# Node（从仓库构建并推送）
cd sboard/node
export SBOARD_NODE_IMAGE="faintghost/sboard-node:latest"
docker compose -f docker-compose.yml -f docker-compose.build.yml build
docker push "$SBOARD_NODE_IMAGE"

# Panel（会同时构建 web 前端 dist 并打包进镜像）
cd ../panel
export SBOARD_PANEL_IMAGE="faintghost/sboard-panel:latest"
docker compose -f docker-compose.yml -f docker-compose.build.yml build
docker push "$SBOARD_PANEL_IMAGE"
```

构建加速建议：
- 建议开启 BuildKit：`export DOCKER_BUILDKIT=1`
- 日常增量构建尽量不要加 `--no-cache`，仅在需要强制全量重建时使用

一键构建并推送（Node + Panel）：
```bash
./scripts/docker-build-push.sh --namespace faintghost --tag latest
```

**配置说明**
- Panel
  - `PANEL_HTTP_ADDR`：监听地址，默认 `:8080`
  - `PANEL_DB_PATH`：SQLite 路径，默认 `panel.db`
  - `PANEL_SERVE_WEB`：是否由 Panel 静态托管前端，默认 `false`
  - `PANEL_WEB_DIR`：前端构建产物目录（`dist`），默认 `web/dist`
  - `PANEL_JWT_SECRET`：JWT 签名密钥（必填）
  - `PANEL_SETUP_TOKEN`：首次初始化管理员所需的一次性 token（可选；为空时首次启动会生成并打印）
  - `PANEL_CORS_ALLOW_ORIGINS`：允许的前端 Origin（逗号分隔或 `*`），默认 `http://localhost:5173`
  - `PANEL_LOG_REQUESTS`：是否打印每个 HTTP 请求（`true/false`），默认 `true`
- Node
  - `NODE_HTTP_ADDR`：监听地址，默认 `:3000`
  - `NODE_SECRET_KEY`：API 密钥，用于 `Authorization: Bearer <secret>`
  - `NODE_LOG_LEVEL`：日志级别，默认 `info`



**前端代码质量（Oxc）**
- `panel/web` 已从 ESLint 迁移到 Oxc（`oxlint` + `oxfmt`）。
- 推荐使用 `bun`：
  - `cd /root/workspace/sboard/panel/web`
  - `bun run lint`
  - `bun run lint:fix`
  - `bun run format`
  - `bun run format:check`
- 兼容 `npm run` 同名脚本。


**前后端联调（本地）**
1. 启动 Panel（如上，确保设置 `PANEL_JWT_SECRET`）
2. 启动前端：
```bash
cd /root/workspace/sboard/panel/web
npm run dev
```
3. Vite 会把 `/api/*` 代理到 Panel（默认 `http://127.0.0.1:8080`）。如需自定义目标：
```bash
VITE_PROXY_TARGET=http://127.0.0.1:8080 npm run dev
```
3. 首次访问前端：
  - 如果还未初始化管理员，会出现 onboarding 页面，要求输入 Setup Token 并创建管理员账号密码
  - Setup Token 默认会在 Panel 启动日志中打印（或你也可以通过 `PANEL_SETUP_TOKEN` 自定义）

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
