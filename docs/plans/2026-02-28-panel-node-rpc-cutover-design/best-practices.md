# Best Practices

## 1. 契约治理

- Node 控制面 RPC 契约使用 Protobuf，避免手写 JSON 协议漂移。
- 所有契约变更必须走固定流程：
  - 修改 proto
  - `make generate`
  - 更新实现与测试
  - `make check-generate`
- 避免在 Panel 与 Node 各维护一份手工“等价结构体”。

## 2. 鉴权与安全

- 采用统一 Bearer secret 鉴权拦截器。
- `Health` 可按运维需求配置为匿名，其它方法默认必须鉴权。
- 严禁在日志打印明文 `password`、`uuid`、`token`、`secret`。
- 对调试日志继续采用脱敏策略（前后缀保留，中间掩码）。

## 3. 错误语义统一

- 禁止继续依赖字符串解析（如 `node sync status 500: ...`）。
- 在 Panel 侧建立 Connect code 到业务错误的统一映射层。
- 对用户可见错误与内部错误摘要分层：
  - 用户提示简短、可行动。
  - 内部摘要保留定位信息用于 `sync_jobs` 审计。

## 4. 超时与消息大小

- RPC 客户端调用统一设置 timeout，防止 monitor goroutine 长时间阻塞。
- RPC server 配置最大消息大小，至少覆盖当前 REST 的 4 MiB 上限。
- 对超大 payload 失败场景补充明确错误码与可读错误信息。

## 5. 并发与幂等

- 延续“每节点串行同步锁”策略，禁止同节点并发写配置。
- 同步失败不影响其他节点同步结果，确保部分失败可观测。
- 重试策略只针对可重试错误（`unavailable`、`deadline_exceeded`），
  不对 `invalid_argument` 重试。

## 6. 发布策略（直切）

- Panel 与 Node 必须同窗口发布，避免协议代差。
- 发布前执行最小冒烟：
  - Node RPC 健康可达。
  - 单节点手动同步成功。
  - 订阅 REST 可访问。
- 回滚必须成对回滚 Panel/Node 到同一版本代次。

## 7. 测试策略

- 单测锁定规则：鉴权、错误映射、SS 2022、并发锁。
- 集成验证协作：`runNodeSync` 与 Node RPC 服务端协同路径。
- E2E 仅保留关键链路，避免把所有分支都塞进 UI 用例。
- 每次交付前执行前后端门禁与 `make e2e`。

## 8. 文档一致性

- README/AGENTS/DESIGN 与运行时边界必须同步：
  - Panel 管理面：RPC
  - Panel↔Node：RPC（本次切换后）
  - 订阅：REST（保留）
- 清理或标注历史文档，避免“REST 管理面”误导。
