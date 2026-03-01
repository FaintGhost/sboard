# Architecture — Node→Panel 心跳注册

## 通信模型变更

### Before (单向)

```
Panel ── Health/SyncConfig/GetTraffic ──→ Node
```

### After (双向)

```
Panel ── Health/SyncConfig/GetTraffic ──→ Node     (不变)
Panel ←── Heartbeat ─────────────────── Node       (新增)
```

两个方向的通信完全独立：
- Panel→Node：Panel 主动发起，用于配置同步和监控
- Node→Panel：Node 主动发起，用于注册和状态上报
- 任一方向故障不影响另一方向

---

## 组件架构

```
┌─────────────────────────────────────────────────────────────┐
│                          Panel                               │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ ConnectRPC Server (/rpc)                              │  │
│  │                                                       │  │
│  │  AuthService (public: Bootstrap/Login)                │  │
│  │  NodeService (protected: CRUD, Approve, Reject)       │  │
│  │  NodeRegistrationService (public: Heartbeat)  ← 新增  │  │
│  │  ...其他服务                                          │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────────┐  ┌──────────────────────────────┐   │
│  │ NodesMonitor     │  │ Heartbeat Handler             │   │
│  │ (不变)           │  │ - 接收 Node 心跳              │   │
│  │ 轮询 Node Health │  │ - 匹配 UUID/SecretKey         │   │
│  │ 触发 SyncConfig  │  │ - 创建 pending 记录           │   │
│  └──────────────────┘  └──────────────────────────────┘   │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ SQLite (nodes 表)                                     │  │
│  │  status: offline | online | pending                   │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                    │                    ▲
                    │ Panel→Node         │ Node→Panel
                    │ (Health/Sync)      │ (Heartbeat)
                    ▼                    │
┌─────────────────────────────────────────────────────────────┐
│                          Node                                │
│                                                             │
│  ┌──────────────────┐  ┌──────────────────────────────┐   │
│  │ RPC Server       │  │ Heartbeat Goroutine    ← 新增 │   │
│  │ (不变)           │  │ - 每 30s 向 Panel 发送心跳    │   │
│  │ Health/SyncConfig│  │ - PANEL_URL 为空则不启动      │   │
│  │ GetTraffic       │  │ - 失败仅日志，不影响服务      │   │
│  └──────────────────┘  └──────────────────────────────┘   │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Config                                                │  │
│  │  +PANEL_URL (新增)                                    │  │
│  │  +NODE_UUID (新增，首次启动自动生成并持久化)            │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 数据流

### 正常运行时（Node 已被 Panel 管理）

```
Node Heartbeat ──→ Panel
  Panel 查找 UUID → 找到 → 验证 SecretKey → 匹配 → 更新 last_seen_at
  Panel 返回 RECOGNIZED
```

此时心跳与 NodesMonitor 的 Health 轮询**共存**：
- 心跳：Node 主动上报，更新 last_seen_at
- Health：Panel 主动探测，驱动 offline→online 状态转换和 SyncConfig 触发
- 两者互补，不冲突

### Panel 重置后

```
Timeline:
  T+0:   Admin 删除 Panel data → Panel 重启 bootstrap
  T+0:   Node 继续以旧配置运行
  T+30s: Node 发送 Heartbeat → Panel 查找 UUID → 未找到
         → Panel 创建 pending 记录
         → Panel 返回 PENDING
  T+?:   Admin 登录 Panel → 看到 "1 个待认领节点"
         → 点击审批 → 填写名称/分组
         → Panel 将 pending→offline
         → NodesMonitor 探测 Health → online → SyncConfig
         → Node 收到新配置，覆盖旧配置
```

---

## Node 状态机

```
                    ┌──────────┐
  Heartbeat ──────→ │ pending  │ ──── Admin Approve ──→ ┌─────────┐
  (未知UUID)        └──────────┘                        │ offline │
                         │                              └────┬────┘
                    Admin Reject                              │
                         │                         Monitor Health OK
                         ▼                                    │
                    [deleted]                                  ▼
                                                        ┌─────────┐
                                           ┌───────────│ online  │
                                           │            └────┬────┘
                                           │                 │
                                      Health OK         Health Fail ×2
                                           │                 │
                                           └─────────────────┘
```

---

## Proto 服务划分

| 服务 | 调用方 | 鉴权 | 说明 |
|------|--------|------|------|
| `NodeRegistrationService.Heartbeat` | Node→Panel | 公开（无 JWT） | Node 上报心跳 |
| `NodeService.ApproveNode` | Frontend→Panel | JWT（管理员） | 审批待认领节点 |
| `NodeService.RejectNode` | Frontend→Panel | JWT（管理员） | 拒绝待认领节点 |

Heartbeat 端点必须公开（Node 没有管理员 JWT），但通过以下方式防滥用：
1. 速率限制（每 IP 每分钟最多 10 次）
2. 请求体大小限制
3. pending 记录数量上限（如最多 50 个 pending 节点）

---

## 文件变更清单

### Node 侧

| 文件 | 操作 | 说明 |
|------|------|------|
| `node/internal/config/config.go` | 修改 | 新增 PanelURL、HeartbeatInterval、NodeUUID 配置 |
| `node/internal/config/uuid.go` | 新建 | UUID 生成和持久化逻辑 |
| `node/internal/heartbeat/heartbeat.go` | 新建 | 心跳 goroutine |
| `node/cmd/node/main.go` | 修改 | 启动心跳 goroutine |
| `node/go.mod` | 修改 | 可能需要 Panel proto 的 connect client |

### Panel 侧

| 文件 | 操作 | 说明 |
|------|------|------|
| `panel/proto/sboard/panel/v1/panel.proto` | 修改 | 新增 NodeRegistrationService 和消息类型 |
| `panel/internal/rpc/server.go` | 修改 | 注册新服务，添加到 public map |
| `panel/internal/rpc/node_registration.go` | 新建 | Heartbeat handler 实现 |
| `panel/internal/rpc/services_impl.go` | 修改 | 新增 ApproveNode/RejectNode 实现 |
| `panel/internal/db/nodes.go` | 修改 | 新增 GetNodeByUUID、CreatePendingNode、ApproveNode 方法 |

### Frontend

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/src/lib/node-compose.ts` | 修改 | 模板新增 PANEL_URL 和 NODE_UUID |
| `web/src/lib/node-compose.test.ts` | 修改 | 更新测试 |
| `web/src/pages/nodes-page.tsx` | 修改 | 新增待认领节点提示和审批 UI |

### Proto 生成

运行 `moon run panel:generate` 后自动生成所有 Go/TS stubs。
