# Agent Notes (sboard)

## 约定
- 统一用中文沟通
- 代码/文档里的列表缩进使用 2 个空格


## 前端代码质量工具（Oxc）
- `panel/web` 已从 ESLint 迁移到 Oxc（`oxlint` + `oxfmt`），默认不再使用 ESLint。
- 优先使用 `bun` 执行前端质量命令（也兼容 `npm run`）：
  - `bun run lint` / `bun run lint:fix`
  - `bun run format` / `bun run format:check`
- 对前端代码风格和格式的改动，优先走 `oxfmt`；不要再引入新的 ESLint 配置或 ESLint 指令注释。

## 前端交付门禁（新增）
- 前端开发相关改动在交付前必须通过以下检查：
  - `bun run lint`
  - `bun run format`（或 `bun run format:check`）
  - `bunx tsc -b`（typecheck）
- 任一检查未通过时，不应标记为完成交付。

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
  - Node 将明显的 payload 错误映射为 HTTP 400（更易区分"配置问题"与"节点内部错误"）
- Docker 部署：
  - `node/docker-compose.yml` 默认拉取 Docker Hub 镜像（适配低配 VPS）
  - `node/docker-compose.build.yml` 用于本机构建并推送镜像
- Shadowsocks 2022 多用户模式：
  - SS2022 只支持 `2022-blake3-aes-128-gcm` 和 `2022-blake3-aes-256-gcm`（sing-box 的 multi-user 模式不支持 chacha20-poly1305）
  - Panel 侧校验 `validators.go` 会拦截不支持的 method，返回清晰错误信息
  - 前端 method 下拉框已移除 `chacha20-poly1305` 选项
- Node 入站热更新：
  - `ApplyInbounds` 先移除所有现有 inbound（释放端口），再创建新的
  - 避免了"编辑入站后再次同步报端口占用"的问题
- Phase 4 前端完成：
  - 所有管理页面已实现：登录、仪表盘、用户、分组、节点、入站、订阅、设置
  - 订阅页面：展示用户订阅链接、UA 行为说明、一键复制
  - 设置页面：系统信息、API 端点、订阅格式说明
  - 用户页面：支持禁用用户（软删除）

## Shadowsocks 2022 订阅生成（关键经验）

### 密码格式规范
- **服务端 (Inbound)**：顶层 `password` 为服务端 PSK，`users[].password` 为用户密钥
- **客户端 (Outbound)**：`password` 格式必须为 `<server_psk>:<user_key>`（冒号分隔）
- 参考规范：[Shadowsocks 2022 EIH Spec](https://github.com/Shadowsocks-NET/shadowsocks-specs/blob/main/2022-2-shadowsocks-2022-extensible-identity-headers.md)

### 密钥长度要求
| Method | Key Length | Base64 长度 |
|--------|-----------|------------|
| 2022-blake3-aes-128-gcm | 16 bytes | 24 chars |
| 2022-blake3-aes-256-gcm | 32 bytes | 44 chars |

### 实现细节
- 公共密钥派生模块：`panel/internal/sskey/sskey.go`
  - `DerivePassword(uuid, method)`：根据 UUID 和方法派生 base64 密钥
  - `Is2022Method(method)`：判断是否为 2022 方法
- 订阅生成时需要：
  1. 从 inbound UUID 派生 server PSK
  2. 从 user UUID 派生 user key
  3. 组合为 `psk:userKey` 格式
- sing-box shadowsocks outbound **不支持** `username` 字段，只有 `password`

### 常见错误
- `illegal base64 data`：密码不是有效的 base64 编码（如直接使用 UUID 字符串）
- `missing psk`：服务端缺少顶层 password
- `invalid argument`：使用了不支持 multi-user 的方法（如 chacha20-poly1305）


## 问题修复协作规则（新增）
- 修复问题时，不只修复单一页面/单一路径；应主动排查同类实现（相同组件、相同交互、相同调用链）并统一修复。
- 若暂时不能一次性全量修复，必须在回复中明确：
  - 已覆盖范围
  - 未覆盖范围
  - 后续补齐计划
- 新增或修改一个修复逻辑时，优先抽象为可复用模块（hook/util/service），避免在多个页面重复打补丁。
