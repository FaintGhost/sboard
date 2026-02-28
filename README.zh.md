# SBoard（中文）

SBoard 是一个基于 sing-box 的订阅管理面板与节点管理系统。

## 架构概览

- **Panel**：管理面板（Go/Gin + SQLite + React 前端）
- **Node**：节点服务（内嵌 sing-box），接收 Panel 下发配置

## 功能概览

- 完整管理面板：登录、仪表盘、用户、分组、节点、入站、订阅、设置
- 分组驱动分发：用户通过分组获取节点入站
- 协议支持：vless、vmess、trojan、shadowsocks（含 2022）、socks、http、mixed
- 订阅格式：sing-box JSON、v2ray（Base64），并支持按 User-Agent 自动选择
- 入站支持 TLS/Reality/Transport
- 入站变更自动触发节点同步
- Panel 管理面 API 使用 RPC（Connect + Protobuf）

## 前置条件

- Go 1.25+
- Bun（前端构建）
- Linux 服务器（推荐）

## 快速开始（裸机）

1) 启动 Panel

```bash
cd panel
PANEL_HTTP_ADDR=:8080 \
PANEL_DB_PATH=panel.db \
PANEL_CORS_ALLOW_ORIGINS=http://localhost:5173 \
PANEL_JWT_SECRET=change-me-in-prod \
GOFLAGS='-tags=with_utls' \
go run ./cmd/panel
```

2) 启动 Node

```bash
cd node
NODE_HTTP_ADDR=:3000 \
NODE_SECRET_KEY=secret \
NODE_LOG_LEVEL=info \
GOFLAGS='-tags=with_utls' \
go run ./cmd/node
```

3) 启动前端（开发模式）

```bash
cd panel/web
VITE_API_BASE_URL=http://127.0.0.1:8080 bun run dev
```

说明：

- 管理面默认走 `/rpc/*`
- `/api/*` 仅保留订阅兼容入口（`/api/sub/:user_uuid`）和 Node 侧 REST

如需走 Vite 代理：

```bash
VITE_PROXY_TARGET=http://127.0.0.1:8080 bun run dev
```

4) 首次访问 `http://localhost:5173` 完成管理员初始化（onboarding）。

## Docker 部署（推荐）

### Panel

```bash
cd panel
export PANEL_JWT_SECRET='replace-with-strong-random'
docker compose up -d
```

说明：

- 默认镜像：`faintghost/sboard-panel:latest`，可用 `SBOARD_PANEL_IMAGE` 覆盖
- Panel 会托管前端静态文件（`PANEL_SERVE_WEB=true`）
- 数据库路径 `/data/panel.db`，映射到宿主机 `panel/data/`

### Node

- 不建议手工维护 Node compose。
- 推荐流程：先在 Panel 中“添加节点”，由 Panel 生成节点部署用 `docker-compose.yml`，复制到节点机器后执行 `docker compose up -d`。

## `PANEL_JWT_SECRET` 说明

`PANEL_JWT_SECRET` 用于管理员 JWT 的签名与校验：

- 登录成功后，Panel 使用该密钥签发 JWT
- 后续所有受保护请求都用该密钥校验 JWT
- 该值变更会导致已有 token 全部失效（需要重新登录）
- 该值为空会导致 Panel 配置校验失败

### 快速生成高强度密钥

推荐任选其一：

```bash
# 推荐（openssl）
openssl rand -base64 48

# Linux 通用
head -c 48 /dev/urandom | base64

# 如有 pwgen
pwgen -s 64 1
```

建议：

- 至少 32 字节随机值，推荐 48~64 字符
- 生产环境不要使用示例值（如 `change-me`）

## 自建镜像

```bash
# 一键构建并推送
./scripts/docker-build-push.sh --namespace faintghost --tag latest

# 或分别构建
cd node
SBOARD_NODE_IMAGE="faintghost/sboard-node:latest" \
  docker compose -f docker-compose.yml -f docker-compose.build.yml build

cd ../panel
SBOARD_PANEL_IMAGE="faintghost/sboard-panel:latest" \
  docker compose -f docker-compose.yml -f docker-compose.build.yml build
```

## GitHub Actions 自动发布镜像

工作流：`.github/workflows/docker-publish.yml`

- 触发条件：
  - push 到 `main`/`master`
  - push tag（`v*`）
  - 手动触发（`workflow_dispatch`）
- 发布目标：
  - Docker Hub：`docker.io/<DOCKERHUB_USERNAME>/sboard-node`、`docker.io/<DOCKERHUB_USERNAME>/sboard-panel`
  - GHCR：`ghcr.io/<repo_owner>/sboard-node`、`ghcr.io/<repo_owner>/sboard-panel`
- 默认标签：`branch`、`tag`、`sha`、`latest`（仅默认分支）

需配置仓库 Secrets：

- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`（建议 Docker Hub Access Token）

说明：

- GHCR 使用 `GITHUB_TOKEN` 自动登录，无需额外密钥

## 配置说明

### Panel

| 环境变量 | 说明 | 默认值 |
|---------|------|-------|
| `PANEL_HTTP_ADDR` | 监听地址 | `:8080` |
| `PANEL_DB_PATH` | SQLite 路径 | `panel.db` |
| `PANEL_JWT_SECRET` | JWT 签名密钥（必填） | - |
| `PANEL_SETUP_TOKEN` | 首次初始化 token（可选） | 自动生成 |
| `PANEL_SERVE_WEB` | 是否托管前端 | `false` |
| `PANEL_WEB_DIR` | 前端产物目录 | `web/dist` |
| `PANEL_CORS_ALLOW_ORIGINS` | 允许的 Origin | `http://localhost:5173` |
| `PANEL_LOG_REQUESTS` | 打印 HTTP 请求 | `true` |

### Node

| 环境变量 | 说明 | 默认值 |
|---------|------|-------|
| `NODE_HTTP_ADDR` | 监听地址 | `:3000` |
| `NODE_SECRET_KEY` | API 密钥 | - |
| `NODE_LOG_LEVEL` | 日志级别 | `info` |

## API

Panel 管理面 API 已迁移到 RPC（Connect，路径 `/rpc/*`）。

### Panel 管理面（RPC）

- 统一入口：`POST /rpc/sboard.panel.v1.<Service>/<Method>`
- 公共服务：`HealthService`、`AuthService`、`SystemService`
- 业务服务：`UserService`、`GroupService`、`NodeService`、`InboundService`
- 运维服务：`TrafficService`、`SyncJobService`、`SingBoxToolService`

### 兼容保留（REST）

- 订阅：`GET /api/sub/:user_uuid`

订阅行为：

- `?format=singbox`：返回 sing-box JSON
- `?format=v2ray`：返回 v2ray Base64 订阅
- 未指定 format：按 User-Agent 自动选择

## 开发

目录结构：

```text
panel/               # Panel 后端 + 前端
  cmd/panel/         # 入口
  internal/rpc/      # RPC 服务实现与生成代码
  proto/             # Protobuf 契约
  internal/api/      # HTTP 兼容层（含订阅入口）
  web/               # React 前端
node/                # Node 服务
  cmd/node/          # 入口
  internal/api/      # Node HTTP API
  internal/sync/     # sing-box 配置解析/校验
  internal/core/     # sing-box 实例管理
docs/                # 设计与规划文档
Makefile             # 代码生成与检查
```

RPC Proto 工作流：

```bash
# 生成 Go + TS 代码
make generate

# 检查生成代码是否与 spec 同步
make check-generate
```

前端质量工具（Oxc）：

```bash
cd panel/web
bun run lint
bun run format
bunx tsc -b
bun run test
```

交付门禁：

- `bun run lint`
- `bun run format`
- `bunx tsc -b`
- `bun run test`
- `make check-generate`
