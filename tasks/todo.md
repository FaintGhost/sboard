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
- [ ] 清理 OpenAPI 生成链路与产物依赖（`internal/api/oapi_*.gen.go`、`web/src/lib/api/gen/*`）
- [ ] 拆除 `internal/rpc/services_impl.go` 对 `internal/api` 的桥接依赖，改为直接调用 service/usecase 层
- [x] 前端收敛类型定义第一阶段完成（`lib/api/types.ts` 已改为导出 `lib/rpc/types`，存量 `lib/api/*` 作为兼容封装保留）
- [ ] 运行前后端完整门禁并修复问题

## Review
- 已完成：RPC 生成链路、前后端 RPC 主链路、关键 e2e 用例通过，且订阅 REST 兼容保留。
- e2e 续跑结果：`auth/users/groups/nodes/node-sync/subscriptions` 顺序执行全部通过，`--project=e2e` 全量 `16 passed`。
- 当前状态：当 `rpcHandler != nil` 时，服务仅暴露 `/rpc/*` 与 `/api/sub/:user_uuid`；`NewRouter`（无 rpcHandler）仍保留完整 `/api/*`，用于现有后端单测兼容。
- 当前主要遗漏：
  - `panel/internal/rpc/services_impl.go` 仍通过 `legacy *api.Server` 桥接调用。
  - `panel/internal/api/oapi_*.gen.go` 与 `StrictServerInterface` 仍是后端管理能力实现底座。
  - 项目文档仍有 OpenAPI/REST 管理端点叙述，需要统一到 RPC 现状。
- 下一步重点：切断 `internal/rpc -> internal/api` 适配依赖，移除 OpenAPI 代码生成与残留产物，完成真正的 RPC-only 管理面。
