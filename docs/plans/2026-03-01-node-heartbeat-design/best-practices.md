# Best Practices — Node→Panel 心跳注册

## 安全性

### 心跳端点防滥用

Heartbeat 是公开端点（Node 没有管理员 JWT），需要防护：

1. **速率限制**: 每 IP 每分钟最多 10 次心跳请求。超过则返回 `ResourceExhausted`
2. **Pending 上限**: 最多 50 个 pending 节点。达到上限后拒绝新的未知 Node 注册
3. **请求体大小**: 限制 Heartbeat 请求体 ≤ 4KB
4. **不泄露信息**: 对于 SecretKey 不匹配的情况，返回通用 REJECTED 而不透露具体原因

### SecretKey 处理

- 心跳中携带 SecretKey 用于身份匹配，传输层应使用 HTTPS
- Panel 存储 SecretKey 使用 constant-time 比较（`crypto/subtle.ConstantTimeCompare`）
- pending 节点的 SecretKey 在审批前即存储，审批后 Panel→Node 通信直接使用

### UUID 安全

- Node UUID 使用 `crypto/rand` 生成的 UUIDv4
- 持久化到磁盘时权限为 `0600`
- UUID 不是安全凭证，仅用于标识。真正的认证靠 SecretKey

## 可靠性

### 心跳容错

```
Node 心跳失败处理策略:
  1. 网络超时 → 下一个 ticker 周期重试（30s）
  2. Panel 返回错误 → 仅日志记录，不影响代理服务
  3. DNS 解析失败 → 同上
  4. Panel 不可达 → 同上
  关键原则: 心跳故障绝不影响 sing-box 代理服务
```

### 与 NodesMonitor 的协作

心跳和 Monitor 两个机制共存，各有分工：

| 维度 | NodesMonitor (Panel→Node) | Heartbeat (Node→Panel) |
|------|--------------------------|----------------------|
| 发起方 | Panel | Node |
| 目的 | 驱动状态转换 + 触发 SyncConfig | 注册 + 存活上报 |
| 触发同步 | 是（offline→online 时） | 否（仅更新 last_seen_at） |
| 缺失影响 | Node 不会被同步 | Panel 不知道 Node 存在 |

NodesMonitor 仍是**唯一触发 SyncConfig 的组件**。Heartbeat 的作用是：
1. 让 Panel 发现未知 Node（注册）
2. 提供额外的存活信号（last_seen_at 更新）

### Panel 重启后的收敛

Panel 重启后，两个机制的收敛时间：
- NodesMonitor: 第一个 tick（10s）开始轮询已知 Node
- Heartbeat: 最多 30s 后 Node 发来心跳

两者合力，Panel 重启后 30s 内可恢复对所有 Node 的感知。

## 性能

### 心跳频率

默认 30s，可通过 `NODE_HEARTBEAT_INTERVAL` 调整。

考虑因素：
- 30s 在 Panel 重置场景下足够快（管理员几分钟内就能看到 pending 节点）
- 30s 不会给 Panel 带来明显负载（100 个 Node = 每秒 ~3 个心跳请求）
- 对比 NodesMonitor 的 10s 轮询，心跳频率更低

### 心跳请求大小

每次心跳约 200 字节（UUID + SecretKey + Version + Addr），可忽略不计。

## 向后兼容性

### Node 侧

- `PANEL_URL` 为空时完全退化为当前行为（零心跳，纯被动）
- 旧版 Node（无心跳功能）在新 Panel 下正常工作（Panel 不依赖心跳来管理已知 Node）
- 新版 Node 在旧 Panel 下正常工作（心跳 RPC 返回 404，Node 仅日志警告）

### Panel 侧

- 现有 API 和前端完全不受影响
- NodesMonitor 逻辑不变
- `status=pending` 是新增状态，现有代码只处理 `offline/online`，pending 节点在 ListNodes 中默认不参与 Monitor 轮询

### Docker Compose 模板

- 新生成的 docker-compose.yml 包含 `PANEL_URL` 和 `NODE_UUID`
- 已部署的旧 docker-compose.yml 不包含这些变量 → Node 以旧模式运行 → 管理员可手动添加

## 代码质量

### 测试覆盖

所有新增功能必须有对应测试：
- 心跳 goroutine 的启动/停止/重试逻辑
- UUID 生成和持久化
- Panel Heartbeat handler 的四种路径（known/unknown/pending/rejected）
- 审批/拒绝流程
- Docker Compose 模板更新

### 错误处理

- 心跳失败：仅 log.Printf，不 panic，不阻塞
- UUID 文件损坏：重新生成新 UUID（Node 会被 Panel 视为新节点）
- Panel DB 写入 pending 失败：返回 Internal 错误，Node 下次重试
