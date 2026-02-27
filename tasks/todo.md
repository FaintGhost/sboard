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
