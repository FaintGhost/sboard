# RPC 迁移执行清单

## 目标
- 仅保留订阅接口 `GET /api/sub/:user_uuid` 为 REST 兼容入口。
- 其余 Panel 管理接口全面迁移到 RPC 生态。

## 任务
- [x] 建立 proto + buf 代码生成链路（Go connect-go + Web connect-es/connect-query）
- [x] 在后端接入 `/rpc` 路由、Connect handler 与鉴权/CORS 适配
- [x] 前端接入 `TransportProvider` 与 auth/error 拦截基础设施
- [x] 迁移业务 API 到 RPC（auth/system/users/groups/nodes/inbounds/sync-jobs/singbox-tools/traffic）
- [x] e2e 测试主链路切换到 RPC（订阅场景保留 REST 验证）
- [x] Makefile 生成门禁切到 RPC 生成产物检查（移除 OpenAPI 生成门禁）
- [x] 后端移除管理接口 REST 对外暴露（运行模式：存在 rpcHandler 时仅暴露 `/rpc/*` 与 `/api/sub/*`）
- [x] 清理 OpenAPI 生成链路与产物依赖（`internal/api/oapi_*.gen.go`、`web/src/lib/api/gen/*`）
- [x] 拆除 `internal/rpc/services_impl.go` 对 `internal/api` 的桥接依赖，改为直接调用 service/usecase 层
- [x] 前端收敛类型定义第一阶段完成（`lib/api/types.ts` 已改为导出 `lib/rpc/types`，存量 `lib/api/*` 作为兼容封装保留）
- [ ] 运行前后端完整门禁并修复问题

## Review
- 已完成：RPC 生成链路、前后端 RPC 主链路、关键 e2e 用例通过，且订阅 REST 兼容保留。
- e2e 续跑结果：`auth/users/groups/nodes/node-sync/subscriptions` 顺序执行全部通过，`--project=e2e` 全量 `16 passed`。
- 当前状态：
  - `panel/internal/rpc` 已不再依赖 `panel/internal/api`（`legacy` 桥接已移除）。
  - `panel/internal/api/oapi_*.gen.go`、`StrictServerInterface`、`RegisterHandlersWithOptions` 已从运行与编译链路中移除。
  - `NewRouter`（无 rpcHandler）继续保留完整 `/api/*` 手写路由，用于现有后端单测与兼容场景；`rpcHandler != nil` 时仅暴露 `/rpc/*` 与 `/api/sub/:user_uuid`。
- 本轮验证：`go test ./panel/internal/api ./panel/internal/rpc`、`go test ./panel/...` 均通过。
- 当前主要剩余：前端门禁与 e2e 门禁的完整回归（`bun run lint/format/test`、`bunx tsc -b`、`make check-generate`、`e2e --project=e2e`）。

---

## 2026-02-28 Panel-Node RPC 继续迁移（Brainstorming）

### 目标
- 评估并设计 Panel 与 Node 间从 REST 到 RPC 的迁移方案。
- 在不破坏现有线上行为的前提下，明确迁移阶段、兼容策略与验收标准。

### 任务
- [x] 完成发现阶段：梳理当前 Panel-Node 调用链、认证与错误语义
- [x] 完成方案阶段：给出可选迁移路径并确定推荐方案
- [x] 完成设计阶段：产出设计文档（`_index.md`、`architecture.md`、`bdd-specs.md`、`best-practices.md`）
- [x] 完成提交阶段：仅提交本轮设计文档目录
- [ ] 完成移交流程：提示使用 `writing-plans` 进入实现计划拆解

### Review
- 方案选择：已与用户确认采用“Panel↔Node 直切 RPC（不双栈）”，并保留订阅 REST（`GET /api/sub/:user_uuid`）。
- 设计产物：已新增 `docs/plans/2026-02-28-panel-node-rpc-cutover-design/`，包含 `_index.md`、`architecture.md`、`bdd-specs.md`、`best-practices.md`。
- 设计要点：
  - Node 控制面引入 `NodeControlService`（Health/SyncConfig/GetTraffic/GetInboundTraffic）。
  - Panel 侧保持现有业务调用面，替换 `internal/node/client` 传输层为 RPC。
  - 错误处理从 HTTP 文本解析升级为 Connect code 映射。
  - 发布采用同窗口成对发布与成对回滚策略。

---

## 2026-02-28 Panel↔Node REST 直切 RPC 架构研究

### 目标
- 基于当前代码事实梳理 Panel↔Node 调用链与边界。
- 给出可落地的 REST→RPC 直切目标态与最小侵入迁移步骤（不保留双栈）。
- 识别迁移风险并给出缓解策略。

### 任务
- [x] 盘点 Panel 启动装配、RPC 管理面入口与 Node 调用落点
- [x] 盘点 Node API 路由、鉴权、配置解析与状态采集路径
- [x] 抽取 REST→RPC 直切的最小改造面与依赖关系
- [x] 输出迁移风险与缓解清单

### Review
- 现状确认：Panel 管理面已为 RPC（`/rpc/*`），但 Panel→Node 仍由 `panel/internal/node/client.go` 走 HTTP REST（`/api/health`、`/api/config/sync`、`/api/stats/*`）。
- 影响范围确认：NodeSync（手动/自动）、节点健康探测、流量采集三条链路均依赖上述 REST 客户端。
- 迁移切口确认：优先保留 `panel/internal/node` 客户端接口语义，替换其传输层为 Connect RPC；Node 侧新增 RPC Server 复用现有 `sync.ParseAndValidateConfig` 与 `core.ApplyConfig`。

---

## 2026-02-28 Panel-Node RPC 实施计划拆解（Writing-Plans）

### 目标
- 基于 `2026-02-28-panel-node-rpc-cutover-design` 产出可执行实施计划。
- 按 BDD 场景拆分 Red/Green 成对任务，明确依赖、文件边界与验证命令。

### 任务
- [x] 校验设计目录完整性（`_index.md` + `bdd-specs.md`）
- [x] 读取并映射全部 BDD 场景到任务清单
- [x] 生成 `docs/plans/2026-02-28-panel-node-rpc-cutover-plan/` 与任务文件
- [x] 提交计划目录（docs commit）
- [x] 进入执行阶段移交（`executing-plans`）

### Review
- 已完成 14 个任务文件（7 组 Red/Green）与 `_index.md`，每个任务都内嵌完整 Gherkin 场景与验证命令。
- 任务依赖仅保留技术前置关系（契约先行、实现依赖对应 Red 任务、全链路 e2e 依赖核心实现完成）。
