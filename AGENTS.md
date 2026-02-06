# Agent Notes (sboard)

## 约定
- 统一用中文沟通
- 代码/文档里的列表缩进使用 2 个空格

## GitHub 操作
- 当前环境已配置 GitHub CLI（`gh`）
- 如需 GitHub 仓库相关操作，优先使用 `gh`（例如 PR、release、issue 等）
- 但本项目协作约定：涉及网络的 git/gh 操作（pull/push/fetch/clone 等）由用户执行，避免误操作与代理环境差异

## 项目结构
- `panel/`
  - Go 后端：`panel/cmd/panel`，路由与业务：`panel/internal/*`
  - Web 前端：`panel/web/`（React + Vite + TailwindCSS v4 + shadcn/ui）
- `node/`
  - Go 节点：`node/cmd/node`
  - 对外 HTTP API：`node/internal/api`
  - sing-box 配置解析/校验：`node/internal/sync`
  - sing-box 实例管理与入站应用：`node/internal/core`
- `docs/`：设计与说明文档
- `go.work`：本仓库使用 Go workspace 同时开发 `panel` 与 `node`

## 核心概念与数据模型
- `group`：分组（用户隶属于分组）
- `user`：普通用户（拥有唯一订阅链接，通过分组拿到被分配的节点入站）
- `node`：节点（实际承载入站的 VPS 节点；`node` 只属于 1 个 `group`）
- `inbound`：入站（属于某个 `node`；`panel` 下发给 `node` 时会把分组用户注入到每个入站的 `users` 字段）

## 关键 API（后端）
- Panel（`panel/internal/api/router.go`）
  - 健康检查：`GET /api/health`
  - 管理员登录：`POST /api/admin/login`
  - 订阅：`GET /api/sub/:user_uuid`（`format=singbox|v2ray`，也会按 `User-Agent` 选择默认输出）
  - Users：`GET/POST /api/users`，`GET/PUT/DELETE /api/users/:id`
  - User Groups：`GET/PUT /api/users/:id/groups`
  - Groups：`GET/POST /api/groups`，`GET/PUT/DELETE /api/groups/:id`
  - Nodes：`GET/POST /api/nodes`，`GET/PUT/DELETE /api/nodes/:id`
  - Node 运维：`GET /api/nodes/:id/health`，`POST /api/nodes/:id/sync`
  - Inbounds：`GET/POST /api/inbounds`，`GET/PUT/DELETE /api/inbounds/:id`
- Node（`node/internal/api/router.go`）
  - 健康检查：`GET /api/health`
  - 配置同步：`POST /api/config/sync`（`Authorization: Bearer <NODE_SECRET_KEY>`）

## Panel 到 Node 的 Sync 协议（概览）
- Panel 组装 payload：`panel/internal/node/build_config.go`
- 下发规则（按分组生效）：
  - `node.group_id` 必须设置
  - 下发用户集：该分组下 `active` 且未过期的用户
  - 每个 inbound 注入 `users` 列表（协议相关字段：vless/vmess 用 uuid，trojan/ss 用 password）
- Shadowsocks 2022 处理：
  - `method` 为 `2022-*` 时，Node 侧要求 inbound 顶层 `password` 为 base64 key（否则报 `missing psk`）
  - Panel sync 时会对 2022 方法自动补全 inbound 顶层 `password`，并为 `users[].password` 生成符合长度的 base64 key（确定性派生，便于联调）

## 后端模块化/可拔插点（已落地）
- 入站 settings 校验可插拔：
  - `panel/internal/inbounds/validators.go`
  - API 层调用 `inbounds.ValidateSettings(protocol, settings)`，协议专属校验通过注册实现扩展
- Panel -> Node 客户端可替换（便于测试/替换传输层）：
  - `panel/internal/api/nodes_sync.go` 通过 `nodeClientFactory` 注入 `node.Client`

## 已完成的关键实现（里程碑）
- 分组驱动的订阅与下发：用户通过分组获取节点入站，订阅按用户过滤
- Node 同步联调体验：
  - 入站新增/更新/删除后自动触发对应节点 sync（不需要手动去节点页点同步）
  - Node sync 日志增强：打印每个 inbound 的 `tag/type/method/password_len/users`（不泄露敏感内容）
  - Node 将明显的 payload 错误映射为 HTTP 400（更易区分“配置问题”与“节点内部错误”）
- Docker 部署：
  - `node/docker-compose.yml` 默认拉取 Docker Hub 镜像（适配低配 VPS）
  - `node/docker-compose.build.yml` 用于本机构建并推送镜像

