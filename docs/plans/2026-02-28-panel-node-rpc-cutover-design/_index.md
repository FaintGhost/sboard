# Panel-Node 直切 RPC 设计

## 背景

当前仓库状态：
- Panel 管理面已是 Connect RPC（`/rpc/*`）。
- Panel 到 Node 的调用仍是 REST（`/api/config/sync`、`/api/stats/*`、`/api/health`）。
- 用户订阅入口是 REST：`GET /api/sub/:user_uuid`。

本次决策（已确认）：
- Panel 与 Node 通信直接切换到 RPC，不走双栈并行期。
- 用户订阅链接继续保持 REST，不受影响。

## 目标与范围

### 目标

- 将 Panel→Node 的健康检查、配置下发、流量采样统一迁移到 Connect RPC。
- 清理 Node 侧用于 Panel 通信的 REST 端点与对应客户端逻辑。
- 保持现有业务语义不变（分组过滤、用户注入、SyncJob 记录、SS 2022 规则）。

### 范围内

- 新增 Node 控制面 RPC 契约与服务实现。
- Panel 侧 Node 客户端从 REST 切换为 RPC。
- 监控链路（nodes monitor、traffic monitor）切换到 RPC。
- E2E/集成测试与文档同步更新。

### 非目标

- 不变更用户订阅 REST 入口：`GET /api/sub/:user_uuid`。
- 不新增订阅格式。
- 不重构现有 group/user/node/inbound 领域模型。

## 方案对比

| 方案 | 描述 | 优点 | 缺点 | 结论 |
|---|---|---|---|---|
| A. 直切 RPC | 一次版本切换，Panel↔Node 全部改 RPC | 代码收敛快，历史路径清晰，后续维护成本低 | 发布耦合高，需要同步升级 | 采用 |
| B. 双栈过渡 | REST+RPC 并行一段时间 | 发布风险更低 | 逻辑重复、测试面扩大、清理成本高 | 不采用 |

## 核心需求（MUST）

- Node 控制面 RPC 能覆盖当前 REST 能力：Health、SyncConfig、Traffic、InboundTraffic。
- Panel 的 Node 同步主链路继续通过 `runNodeSync` 记录 `sync_jobs/sync_attempts`。
- Node 侧认证语义保持：敏感接口需要 Bearer secret；健康检查可配置为公开。
- 错误语义从“HTTP 文本解析”升级为“Connect code + 统一错误映射”。
- 交付前通过门禁：
  - `moon run panel:check-generate`
  - `cd panel && go test ./... -count=1`
  - `cd node && go test ./... -count=1`
  - `moon run web:lint`
  - `moon run web:format`
  - `moon run web:typecheck`
  - `moon run web:test`

## 成功标准

- Panel 与 Node 不再通过 `/api/config/sync`、`/api/stats/*` 通信。
- Node 同步、健康、流量采样在 RPC 路径全部可用。
- 订阅 REST 行为与当前一致（`format=singbox|v2ray` + UA fallback）。
- 生成代码无漂移，且回归测试全部通过。

## 实施分解（高层）

1. 契约层：定义 `sboard.node.v1.NodeControlService`（proto + generate）。
2. Node 服务层：实现 RPC handler + 鉴权 interceptor + 错误码映射。
3. Panel 客户端层：替换 `panel/internal/node/client.go` 为 RPC 调用实现。
4. 业务接入层：切换 `services_impl.go` 与 monitor 调用点。
5. 清理层：移除 Node REST 同步/统计接口与旧错误解析逻辑。
6. 验证层：补齐单测、集成、e2e 与发布演练。

## 风险与约束

- 直切要求 Panel 与 Node 同步发版，不能错峰。
- E2E 夹具当前直接调用 Node REST，需要同步改为 Node RPC。
- Node 镜像健康检查脚本需要从 REST 路径改为 RPC 健康检查。

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
