# Architecture

## 1. 当前架构（As-Is）

### 1.1 调用边界

- Panel 管理面：Connect RPC（`/rpc/sboard.panel.v1.*`）。
- Panel→Node：REST（`/api/health`、`/api/config/sync`、`/api/stats/traffic`、`/api/stats/inbounds`）。
- 用户订阅：REST（`GET /api/sub/:user_uuid`）。

### 1.2 关键调用链

1. `NodeService.SyncNode` 进入 `runNodeSync`。
2. `runNodeSync` 组装 payload（`BuildSyncPayload`）。
3. 通过 `panel/internal/node/client.go` 打 REST 到 Node。
4. Node `ConfigSync` 解析并下发到 core。
5. Panel 更新 `sync_jobs` 与节点在线状态。

## 2. 目标架构（To-Be）

### 2.1 目标边界

- Panel 管理面：保持 RPC，不变。
- Panel→Node：切为 Node 控制面 RPC（Connect）。
- 用户订阅：保持 REST，不变。

### 2.2 目标数据流

1. Panel 触发同步/监控。
2. Panel Node RPC Client 调 `NodeControlService`。
3. Node RPC Server 执行鉴权、参数校验、core 应用。
4. Node 返回结构化 RPC 错误码与响应体。
5. Panel 统一映射为 `sync_jobs`/`sync_attempts` 结果。

## 3. 契约设计

## 3.1 新增服务

```proto
service NodeControlService {
  rpc Health(HealthRequest) returns (HealthResponse);
  rpc SyncConfig(SyncConfigRequest) returns (SyncConfigResponse);
  rpc GetTraffic(GetTrafficRequest) returns (GetTrafficResponse);
  rpc GetInboundTraffic(GetInboundTrafficRequest) returns (GetInboundTrafficResponse);
}
```

## 3.2 消息设计要点

- `SyncConfigRequest` 直接携带配置结构化字段（建议保持与现有 payload 字段同构），避免二次字符串编码。
- `GetTrafficRequest` 保留 `interface` 过滤参数（兼容当前行为）。
- `GetInboundTrafficRequest` 保留 `reset` 标记（兼容当前行为）。
- 响应字段命名与现有 Panel 侧结构对齐，降低迁移成本。

## 3.3 代码生成策略

- 契约源：新增 `panel/proto/sboard/node/v1/node.proto`。
- 生成目标：
  - Panel：生成 Node RPC Client 所需代码。
  - Node：生成 Node RPC Server 所需代码。
- 要求 `make generate` 一次生成两侧产物，并纳入 `make check-generate`。

注：实现阶段优先采用 buf 的 managed/override 机制解决两侧 `go_package` 路径差异，避免手工维护两份 proto。

## 4. 组件改造

### 4.1 Node 侧

- 新增 `node/internal/rpc`：
  - `server.go`：注册 `NodeControlService`。
  - `services_impl.go`：复用现有 `sync.ParseAndValidateConfig` 与 core 逻辑。
  - `auth_interceptor.go`：统一 Bearer secret 认证。
- 路由改造：
  - 增加 `/rpc/*`。
  - 移除用于 Panel 通信的 `/api/config/sync` 与 `/api/stats/*` REST 路由。

### 4.2 Panel 侧

- 重写 `panel/internal/node/client.go` 为 RPC 客户端实现：
  - 对外方法签名尽量保持不变（`Health`、`SyncConfig`、`Traffic`、`InboundTraffic*`）。
- 更新 `panel/internal/rpc/services_impl.go`：
  - 去掉 HTTP 状态文本解析（如 `node sync status xxx`）。
  - 使用 Connect code 统一映射错误类型（`unauthenticated`、`unavailable`、`invalid_argument` 等）。
- 更新 `panel/internal/monitor/*`：
  - 监控采样与健康检查改走 RPC 客户端。

## 5. 错误与状态映射

| Node RPC code | Panel 归类 | SyncJob 期望行为 |
|---|---|---|
| `invalid_argument` | 配置错误 | 立即失败，不重试 |
| `unauthenticated` | 鉴权错误 | 立即失败，节点标记异常 |
| `unavailable` | 节点不可达 | 失败，可按策略重试 |
| `deadline_exceeded` | 超时 | 失败，可按策略重试 |
| `internal` | 节点内部错误 | 失败，记录摘要 |

## 6. 安全与可观测性

- 认证：除 `Health` 外，所有 Node RPC 方法默认要求 Bearer secret。
- 限流/大小：对齐当前 4 MiB body 限制，配置 RPC 最大消息大小。
- 日志：继续保留敏感字段脱敏（password/uuid/token）。
- 指标：保留并扩展同步耗时、失败码分布、节点在线率指标。

## 7. 迁移顺序（直切）

1. 合并契约与双侧实现代码（但暂不删旧路径）。
2. 在同一发布窗口同时发布 Panel 与 Node 新版本。
3. 发布后立即执行 smoke/e2e 校验。
4. 验证通过后删除 Node 旧 REST 路由与 Panel 旧 REST 客户端逻辑。

说明：虽然策略是“直切”，但代码提交可按小步进行；生产切换必须一次性完成。

## 8. 回滚策略

- 回滚单元：Panel 与 Node 必须成对回滚到同一代版本。
- 回滚触发：同步失败率超过阈值或核心场景 e2e 失败。
- 回滚后验证：
  - 节点同步恢复。
  - 订阅 REST 正常。
  - 管理面登录与核心 CRUD 正常。
