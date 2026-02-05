# SBoard 阶段 1 设计

## 背景与目标
阶段 1 目标是搭建可运行的基础框架，覆盖 Panel 与 Node 的启动链路、最小 API、Node 的配置下发与 sing-box Inbound 创建路径，并引入标准化 SQLite 迁移流程。此阶段不实现用户/节点/入站 CRUD，不做订阅与前端，只验证“Panel 启动 → Node 配置同步”链路和 sing-box 嵌入式运行可行。

## 架构与启动流程
- 仓库根目录使用 `go.work` 聚合 `./panel` 与 `./node` 两个 module。
- `panel` 模块：`cmd/panel/main.go` 负责加载配置、初始化日志与 SQLite 连接，执行 `golang-migrate` 迁移（`panel/internal/db/migrations`），启动 Gin 服务并注册 `/api/health` 及占位路由骨架。
- `node` 模块：`cmd/node/main.go` 负责加载配置、初始化 `NodeCore`，创建最小可运行 sing-box `Box`（日志/路由/出站），启动 Gin 服务并注册 `/api/health` 与 `/api/config/sync`。
- Node 采用“仅内存配置”方案：入站配置完全由 `/api/config/sync` 下发，入站变更通过 Inbound 重建实现原子更新；首次同步前不承载流量。

## 数据流与接口
- 仅实现 Panel → Node 的配置同步链路。
- Node API：
  - `GET /api/health`：健康检查。
  - `POST /api/config/sync`：完整入站配置同步，使用 `Authorization: Bearer <SECRET_KEY>` 认证。
- `POST /api/config/sync`：
  - 结构校验：必须包含 `tag/type/listen/listen_port/users` 等字段。
  - 语义校验：端口范围、tag 重复、用户字段完整性。
  - 通过后批量创建 inbound；失败时 fail-fast 返回 `500`，记录错误日志，客户端可重试。
  - 支持幂等：同 tag inbound 重建会自动覆盖旧实例。
  - 同步成功后记录配置 hash 与更新时间（内存级），便于排查。

## 配置、错误处理与测试
- 配置：Panel/Node 均提供最小配置文件示例，支持环境变量覆盖（如 `PANEL_DB_PATH`、`PANEL_HTTP_ADDR`、`NODE_HTTP_ADDR`、`NODE_SECRET_KEY`、`NODE_LOG_LEVEL`）。
- 错误处理：
  - Panel 迁移失败直接退出，避免不一致 schema。
  - Node 初始化 Box 失败直接退出，避免暴露不可用 API。
  - `/api/config/sync` 返回清晰错误码与短错误信息，日志包含 trace id。
- 测试：
  - Panel：迁移能成功创建表的单元测试。
  - Node：配置解析/校验/重复 tag 处理的单元测试。
  - 集成：`/api/config/sync` 成功与失败路径（认证失败、字段缺失、端口非法）。
  - 不做端到端客户端连通性测试。

## 初始迁移
迁移文件包含 `users`、`nodes`、`inbounds`、`user_inbounds`、`traffic_stats` 表结构，字段与 `docs/DESIGN.md` 一致，为阶段 2 直接扩展预留。
