# SBoard

SBoard 是一个基于 sing-box 的订阅管理面板与节点管理系统。

**架构概览**
- **Panel**：管理面板（Go/Gin + SQLite + React 前端），提供用户/分组/节点/入站的全功能管理
- **Node**：实际承载入站的节点服务（内嵌 sing-box），接收 Panel 下发的入站配置

**功能概览**
- 完整的管理面板：登录、仪表盘、用户、分组、节点、入站、订阅、设置
- 分组驱动的订阅与下发：用户通过分组获取节点入站
- 支持多种协议：vless、vmess、trojan、shadowsocks（含 2022）、socks、http、mixed
- 订阅格式：sing-box JSON、v2ray（Base64 分享链接），按 User-Agent 自动选择
- 入站配置通过 sing-box 模板编辑，支持 TLS/Reality/Transport
- 入站新增/更新/删除后自动触发节点同步
- OpenAPI 3.0 驱动的前后端 API 对接，代码生成保证类型一致性

## 前置条件

- Go 1.25+
- Bun（前端构建）
- Linux 服务器（推荐）

## 快速开始（裸机）

1. 启动 Panel
```bash
cd panel
PANEL_HTTP_ADDR=:8080 \
PANEL_DB_PATH=panel.db \
PANEL_CORS_ALLOW_ORIGINS=http://localhost:5173 \
PANEL_JWT_SECRET=change-me-in-prod \
GOFLAGS='-tags=with_utls' \
go run ./cmd/panel
```

2. 启动 Node
```bash
cd node
NODE_HTTP_ADDR=:3000 \
NODE_SECRET_KEY=secret \
NODE_LOG_LEVEL=info \
GOFLAGS='-tags=with_utls' \
go run ./cmd/node
```

3. 启动前端（开发模式）
```bash
cd panel/web
bun run dev
```
Vite 会把 `/api/*` 代理到 Panel（默认 `http://127.0.0.1:8080`）。自定义目标：
```bash
VITE_PROXY_TARGET=http://127.0.0.1:8080 bun run dev
```

4. 首次访问：浏览器打开 `http://localhost:5173`，如果还未初始化管理员，会出现 onboarding 页面，要求输入 Setup Token 并创建管理员账号密码。Setup Token 默认在 Panel 启动日志中打印（或通过 `PANEL_SETUP_TOKEN` 自定义）。

## Docker 部署

仓库内置 Docker Compose 配置，默认拉取 Docker Hub 预构建镜像。

**Panel**
```bash
cd panel
export PANEL_JWT_SECRET='change-me'
docker compose up -d
```
说明：
- 默认镜像为 `faintghost/sboard-panel:latest`，可用 `SBOARD_PANEL_IMAGE` 覆盖
- Panel 会在同一进程内静态托管前端（`PANEL_SERVE_WEB=true`）
- 数据库默认路径为 `/data/panel.db`（通过 volume 映射到宿主机 `panel/data/`）

**Node**
```bash
cd node
export NODE_SECRET_KEY='change-me'
docker compose up -d
```
说明：
- compose 使用 `network_mode: host`，入站直接监听宿主机端口
- Node API（默认 `:3000`）会暴露在公网，请用防火墙只允许 Panel 服务器 IP 访问

**自建镜像**
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

## 配置说明

**Panel**
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

**Node**
| 环境变量 | 说明 | 默认值 |
|---------|------|-------|
| `NODE_HTTP_ADDR` | 监听地址 | `:3000` |
| `NODE_SECRET_KEY` | API 密钥 | - |
| `NODE_LOG_LEVEL` | 日志级别 | `info` |

## API

所有 API 端点定义在 `panel/openapi.yaml`（OpenAPI 3.0.3 spec）。

**主要端点**
- 健康检查：`GET /api/health`
- 管理员登录：`POST /api/admin/login`
- 订阅：`GET /api/sub/:user_uuid`
- Users：`GET/POST /api/users`，`GET/PUT/DELETE /api/users/:id`
- Groups：`GET/POST /api/groups`，`GET/PUT/DELETE /api/groups/:id`
- Nodes：`GET/POST /api/nodes`，`GET/PUT/DELETE /api/nodes/:id`
- Inbounds：`GET/POST /api/inbounds`，`GET/PUT/DELETE /api/inbounds/:id`

**订阅行为**
- `?format=singbox`：返回 sing-box JSON
- `?format=v2ray`：返回 v2ray 风格订阅（Base64 分享链接）
- 未指定 format：按 User-Agent 自动选择

## 开发

**目录结构**
```
panel/               # Panel 后端（Gin + SQLite）+ 前端
  cmd/panel/         # 入口
  internal/api/      # HTTP handlers（oapi-codegen 生成的接口）
  web/               # React 前端（Vite + TailwindCSS v4 + shadcn/ui）
  openapi.yaml       # OpenAPI 3.0.3 spec（API 的 Single Source of Truth）
node/                # Node 服务（嵌入 sing-box）
  cmd/node/          # 入口
  internal/api/      # Node HTTP API
  internal/sync/     # sing-box 配置解析/校验
  internal/core/     # sing-box 实例管理
docs/                # 设计与规划文档
Makefile             # 代码生成与检查
```

**OpenAPI 工作流**

API 变更流程：修改 `panel/openapi.yaml` → `make generate` → 更新 handler/页面 → 测试

```bash
# 生成 Go + TS 代码
make generate

# 检查生成代码是否与 spec 同步
make check-generate
```

工具链：
- Go 后端：[oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)（生成 Gin server interface + 类型）
- TS 前端：[@hey-api/openapi-ts](https://heyapi.dev/)（生成 SDK + 类型 + Zod schemas）

**代码质量**

前端使用 Oxc（`oxlint` + `oxfmt`）：
```bash
cd panel/web
bun run lint        # 代码检查
bun run format      # 格式化
bunx tsc -b         # 类型检查
bun run test        # 单元测试
```

**交付门禁**
- `bun run lint`
- `bun run format`
- `bunx tsc -b`
- `bun run test`
- `make check-generate`
