# Node→Panel 心跳注册设计

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:writing-plans to create implementation plan from this design.

**Goal:** 增加 Node→Panel 的心跳/注册通道，使 Panel 重置后能自动发现已存在的 Node，管理员审批后恢复管控。

**Architecture:** 在现有 Panel→Node 单向通信基础上，增加 Node→Panel 的反向心跳。Node 定期向 Panel 注册自身信息，Panel 根据是否已知该 Node 决定接受或标记为"待认领"。

**Tech Stack:** Go, ConnectRPC, SQLite, React/TypeScript

---

## Context

### 现状

Panel↔Node 通信模型为**严格单向**：Panel 是唯一发起方，Node 是纯被动方。

```
Panel ── Health/SyncConfig/GetTraffic ──→ Node
Panel ←──────── (无) ────────────────── Node
```

Node 不知道 Panel 的存在，也没有 Panel 的地址信息。

### 问题

当 Panel 数据被重置（如删除 data 文件夹）时：

1. Panel 丢失所有 Node 记录
2. Node 继续以最后同步的配置运行（代理服务正常）
3. **Panel 无法发现已存在的 Node** — 管理员必须手动记住每个 Node 的地址和密钥，逐一重建记录
4. 在管理员手动恢复之前，Node 上的用户配置永远不会被更新

### 设计目标

1. Panel 重置后，Node 能**自动向 Panel 报到**
2. 管理员在 Panel UI 中看到"待认领节点"并手动审批
3. 审批后，Panel 立即推送最新配置覆盖旧配置
4. **不影响现有 Panel→Node 的通信模型**（Health/SyncConfig/GetTraffic 保持不变）
5. 心跳功能**可选**（未配置 `PANEL_URL` 时 Node 退化为当前纯被动模式）

---

## Requirements

### Functional

- **FR-1**: Node 配置 `PANEL_URL` 后，定期向 Panel 发送心跳（默认 30s）
- **FR-2**: 心跳携带 Node 的 UUID、SecretKey、版本信息、监听地址
- **FR-3**: Panel 收到已知 Node 的心跳后，更新 `last_seen_at` 和状态
- **FR-4**: Panel 收到未知 Node 的心跳后，创建 `status=pending` 的记录
- **FR-5**: 管理员可在 UI 中查看待认领节点并审批（填写名称、分组等）
- **FR-6**: 审批后 Node 变为 `online` 状态，立即触发 SyncConfig
- **FR-7**: 管理员可拒绝/删除待认领节点
- **FR-8**: Docker Compose 模板包含 `PANEL_URL` 环境变量

### Non-Functional

- **NFR-1**: 心跳失败不影响 Node 代理服务（fire-and-forget）
- **NFR-2**: 心跳请求不携带敏感用户数据
- **NFR-3**: Panel 侧心跳端点有速率限制，防止恶意注册
- **NFR-4**: 未配置 `PANEL_URL` 时 Node 行为与当前完全一致

---

## Rationale

### 为什么选择 Node→Panel 心跳而非其他方案

| 方案 | 优点 | 缺点 | 结论 |
|------|------|------|------|
| 保持单向，手动恢复 | 最简单，零改动 | Panel 重置后无法自动发现 Node | 不满足需求 |
| Config TTL/Lease | 不需要反向通道 | TTL 难定，Panel 维护期间可能误触发 | 风险太高 |
| **Node→Panel 心跳** | 自动发现，手动审批安全 | 需改通信模型 | **选择此方案** |

### 为什么心跳端点是公开的

Node 心跳端点不走 Panel 的 JWT auth（管理员 token），而是用 Node 自身的 `SecretKey` 做认证。原因：

1. Node 没有管理员 JWT，它不是"用户"
2. SecretKey 是 Node 级别的凭证，已在 Panel DB 中存储
3. 对于未知 Node（Panel 重置后），先接收心跳创建 pending 记录，审批后再验证 SecretKey

---

## Detailed Design

### 1. Proto 定义

在 Panel proto 中新增 `NodeRegistrationService`：

```protobuf
// panel/proto/sboard/panel/v1/panel.proto

service NodeRegistrationService {
  // Node 定期调用，上报自身状态
  rpc Heartbeat(NodeHeartbeatRequest) returns (NodeHeartbeatResponse);
}

message NodeHeartbeatRequest {
  string uuid = 1;        // Node UUID
  string secret_key = 2;  // Node 密钥（用于匹配身份）
  string version = 3;     // Node 版本号
  string api_addr = 4;    // Node 监听地址（如 ":3003"）
}

message NodeHeartbeatResponse {
  NodeHeartbeatStatus status = 1;
  string message = 2;     // 人类可读的状态描述

  enum NodeHeartbeatStatus {
    NODE_HEARTBEAT_STATUS_UNSPECIFIED = 0;
    NODE_HEARTBEAT_STATUS_RECOGNIZED = 1;   // 已知节点，正常
    NODE_HEARTBEAT_STATUS_PENDING = 2;      // 待审批
    NODE_HEARTBEAT_STATUS_REJECTED = 3;     // 被拒绝
  }
}
```

### 2. Node 侧改动

#### 2.1 配置扩展

```go
// node/internal/config/config.go
type Config struct {
  HTTPAddr  string // 已有
  SecretKey string // 已有
  LogLevel  string // 已有
  StatePath string // 已有

  // 新增
  PanelURL           string // 环境变量 PANEL_URL，为空则不启用心跳
  HeartbeatIntervalS int    // 环境变量 NODE_HEARTBEAT_INTERVAL，默认 30
  NodeUUID           string // 环境变量 NODE_UUID，首次启动时自动生成并持久化
}
```

#### 2.2 UUID 持久化

Node 首次启动时生成 UUID 并写入 `/data/node-uuid`。后续启动读取已有 UUID。这确保 Node 在重启后身份不变。

#### 2.3 心跳 Goroutine

```go
// node/internal/heartbeat/heartbeat.go
func Run(ctx context.Context, cfg config.Config) {
  if cfg.PanelURL == "" {
    return // 未配置则不启动
  }
  client := panelv1connect.NewNodeRegistrationServiceClient(http.DefaultClient, cfg.PanelURL)
  ticker := time.NewTicker(cfg.HeartbeatInterval())
  for {
    select {
    case <-ctx.Done():
      return
    case <-ticker.C:
      resp, err := client.Heartbeat(ctx, &panelv1.NodeHeartbeatRequest{
        Uuid:      cfg.NodeUUID,
        SecretKey: cfg.SecretKey,
        Version:   version.String(),
        ApiAddr:   cfg.HTTPAddr,
      })
      if err != nil {
        log.Printf("[heartbeat] failed: %v", err) // 仅日志，不影响服务
        continue
      }
      log.Printf("[heartbeat] status=%s", resp.Msg.Status)
    }
  }
}
```

### 3. Panel 侧改动

#### 3.1 DB Schema 变更

新增 migration：

```sql
-- 扩展 status 枚举：offline | online | pending
-- 无需改表结构，status 已是 TEXT 类型，直接写入 "pending" 即可
```

#### 3.2 Heartbeat Handler

```go
// panel/internal/rpc/node_registration.go
func (s *Server) Heartbeat(ctx context.Context, req *connect.Request[panelv1.NodeHeartbeatRequest]) (*connect.Response[panelv1.NodeHeartbeatResponse], error) {
  uuid := req.Msg.Uuid
  secretKey := req.Msg.SecretKey

  // 1. 查找已知 Node
  node, err := s.store.GetNodeByUUID(ctx, uuid)
  if err == nil {
    // 已知节点：验证 SecretKey 匹配后更新 last_seen_at
    if node.SecretKey != secretKey {
      return rejected("secret_key mismatch")
    }
    s.store.MarkNodeOnline(ctx, node.ID, time.Now())
    return recognized()
  }

  // 2. 未知节点：检查是否已有 pending 记录
  pending, err := s.store.GetNodeByUUID(ctx, uuid)
  if err == nil && pending.Status == "pending" {
    return pendingApproval()
  }

  // 3. 创建 pending 记录
  s.store.CreatePendingNode(ctx, db.PendingNodeParams{
    UUID:       uuid,
    SecretKey:  secretKey,
    APIAddress: extractHost(req.Msg.ApiAddr),
    APIPort:    extractPort(req.Msg.ApiAddr),
  })
  return pendingApproval()
}
```

#### 3.3 RPC 服务注册

在 `server.go` 的 public map 中添加 Heartbeat 端点，使其免鉴权。

#### 3.4 审批 API

在现有 `NodeService` 中新增：

```protobuf
rpc ApproveNode(ApproveNodeRequest) returns (ApproveNodeResponse);
rpc RejectNode(RejectNodeRequest) returns (RejectNodeResponse);
```

审批时管理员提供 name、group_id 等信息，Panel 将 pending 节点状态改为 offline（随后 NodesMonitor 探测到 online 后触发 SyncConfig）。

### 4. Docker Compose 模板改动

```typescript
// web/src/lib/node-compose.ts
export type BuildNodeComposeInput = {
  port: number;
  secretKey: string;
  logLevel?: string;
  image?: string;
  containerName?: string;
  panelUrl: string;    // 新增
  nodeUuid: string;    // 新增
};
```

生成的环境变量新增：
```yaml
PANEL_URL: "http://panel.example.com:8080"
NODE_UUID: "generated-uuid"
```

### 5. 前端 UI 改动

#### 5.1 待认领节点提示

在节点列表页顶部，当存在 `status=pending` 的节点时，显示提示卡片：

```
┌─────────────────────────────────────────────────┐
│ ⚠️  2 个节点等待认领                              │
│                                                  │
│  UUID: abc123...  │  地址: 1.2.3.4:3003  │ [审批] │
│  UUID: def456...  │  地址: 5.6.7.8:3003  │ [审批] │
│                                                  │
│  [全部忽略]                                       │
└─────────────────────────────────────────────────┘
```

#### 5.2 审批对话框

点击"审批"后弹出对话框，让管理员填写：
- 节点名称（必填）
- 分组（可选）
- 公开地址（可选）

确认后调用 `ApproveNode` RPC，Panel 将节点状态从 pending 改为 offline → NodesMonitor 自动探测并同步。

---

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
