# SBoard - 基于 sing-box 的订阅管理面板设计

## 项目概述

开发一款类似 remnawave 的代理订阅管理面板，使用 sing-box 1.12.19 作为核心。

### 技术栈
- **Panel**: React + Shadcn/ui + Tailwind (前端) + Golang/Gin (后端) + SQLite
- **Node**: Golang (嵌入 sing-box 作为库)
- **核心**: sing-box 1.12.19

### 功能范围
- **协议支持**: VLESS, VMess, Trojan, Shadowsocks
- **用户管理**: 流量统计、到期时间、循环重置
- **计费系统**: 暂不需要 (后续可扩展)

---

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                         Panel Server                         │
├─────────────────────────────────────────────────────────────┤
│  React Frontend  │  Gin Backend  │  SQLite  │  Sub Generator │
└────────┬────────────────┬────────────────────────────────────┘
         │                │
         │   RPC API      │
         │ (Connect + JWT)│
         ▼                ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   SBoard Node   │  │   SBoard Node   │  │   SBoard Node   │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ Node API Server │  │ Node API Server │  │ Node API Server │
│ (接收Panel指令)  │  │                 │  │                 │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ sing-box 嵌入式  │  │ sing-box 嵌入式  │  │ sing-box 嵌入式  │
│ (作为Go库运行)   │  │                 │  │                 │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

---

## 核心模块设计

### 1. Panel 后端模块

```
panel/
├── cmd/panel/main.go           # 入口
├── proto/sboard/panel/v1/      # Panel 管理面 RPC 契约（protobuf）
├── internal/
│   ├── rpc/                    # Connect RPC handlers
│   │   ├── server.go           # RPC 服务注册与鉴权拦截
│   │   ├── services_impl.go    # RPC 方法实现
│   │   └── gen/                # buf 生成代码（go + connect-go）
│   ├── api/                    # HTTP 兼容层（订阅入口、Web 托管）
│   │   ├── auth.go             # JWT 认证
│   │   ├── users.go            # 用户 CRUD
│   │   ├── nodes.go            # 节点管理
│   │   ├── inbounds.go         # 入站配置
│   │   └── subscription.go     # 订阅链接生成
│   ├── db/                     # SQLite 数据层
│   │   ├── models.go           # 数据模型
│   │   └── migrations/         # 数据库迁移
│   ├── node/                   # Node 通信客户端
│   │   └── client.go           # 调用 Node API
│   └── subscription/           # 订阅生成器
│       ├── singbox.go          # sing-box 格式
│       ├── clash.go            # Clash 格式
│       └── v2ray.go            # v2ray 格式
├── web/                        # React 前端
└── go.mod
```

### 2. Node 模块 (核心挑战)

```
node/
├── cmd/node/main.go            # 入口
├── internal/
│   ├── api/                    # Node REST API
│   │   ├── server.go           # HTTP 服务器
│   │   ├── config.go           # 配置下发接口
│   │   ├── users.go            # 用户同步接口
│   │   └── stats.go            # 流量统计接口
│   ├── core/                   # sing-box 集成
│   │   ├── box.go              # Box 实例管理
│   │   ├── inbound.go          # Inbound 管理
│   │   └── users.go            # 用户热更新
│   └── stats/                  # 流量统计收集
│       └── collector.go
└── go.mod
```

---

## sing-box 用户热更新方案

### 问题分析

sing-box 各协议的 `UpdateUsers()` 方法存在于内部 service 中，但：
- `vless.Inbound.service` 是私有字段
- `adapter.Inbound` 接口不包含 `UpdateUsers` 方法

### 解决方案：Inbound 重建法

利用 `InboundManager.Create()` 的特性：当创建同 tag 的 inbound 时，会自动关闭旧实例。

```go
// 用户更新流程
func (n *NodeCore) UpdateUsers(inboundTag string, users []User) error {
    // 1. 获取当前 inbound 配置
    config := n.getInboundConfig(inboundTag)

    // 2. 更新用户列表
    config.Users = users

    // 3. 重建 inbound (原子操作，自动关闭旧实例)
    return n.box.Inbound().Create(
        n.ctx,
        n.router,
        n.logger,
        inboundTag,
        config.Type,
        config.Options,
    )
}
```

### 优势
- 无需修改 sing-box 源码
- 利用官方 API
- 用户更新为原子操作

### 劣势
- 有短暂的连接中断 (毫秒级)
- 需要保存完整的 inbound 配置

---

## 数据模型

### Panel 数据库 (SQLite)

```sql
-- 用户表
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    uuid TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT,
    traffic_limit INTEGER DEFAULT 0,      -- 0 = 无限制 (bytes)
    traffic_used INTEGER DEFAULT 0,       -- 已用流量 (bytes)
    traffic_reset_day INTEGER DEFAULT 0,  -- 每月重置日 (1-31, 0=不重置)
    expire_at DATETIME,                   -- 到期时间
    status TEXT DEFAULT 'active',         -- active/disabled/expired/traffic_exceeded
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 节点表
CREATE TABLE nodes (
    id INTEGER PRIMARY KEY,
    uuid TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    address TEXT NOT NULL,         -- Node API 地址
    port INTEGER NOT NULL,         -- Node API 端口
    secret_key TEXT NOT NULL,      -- 认证密钥
    status TEXT DEFAULT 'offline', -- online/offline
    last_seen_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 入站配置表
CREATE TABLE inbounds (
    id INTEGER PRIMARY KEY,
    uuid TEXT UNIQUE NOT NULL,
    tag TEXT NOT NULL,
    node_id INTEGER REFERENCES nodes(id),
    protocol TEXT NOT NULL,        -- vless/vmess/trojan/shadowsocks
    listen_port INTEGER NOT NULL,
    settings JSON NOT NULL,        -- 协议特定配置
    tls_settings JSON,
    transport_settings JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(node_id, tag)
);

-- 用户-入站关联表
CREATE TABLE user_inbounds (
    user_id INTEGER REFERENCES users(id),
    inbound_id INTEGER REFERENCES inbounds(id),
    PRIMARY KEY (user_id, inbound_id)
);

-- 流量统计表
CREATE TABLE traffic_stats (
    id INTEGER PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    node_id INTEGER REFERENCES nodes(id),
    upload INTEGER DEFAULT 0,
    download INTEGER DEFAULT 0,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## Panel-Node 通信协议

### 管理面与节点面的协议边界

- Panel 管理面：RPC（Connect），统一入口 `/rpc/sboard.panel.v1.<Service>/<Method>`
- Panel 订阅兼容：REST `GET /api/sub/:user_uuid`
- Node 对外接口：REST（健康检查、配置同步、统计）

### Node API 端点

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/health | 健康检查 |
| POST | /api/config/sync | 完整配置同步 |
| POST | /api/users/sync | 用户列表同步 |
| GET | /api/stats | 获取流量统计 |
| POST | /api/stats/reset | 重置流量统计 |

### 认证

```
Authorization: Bearer <SECRET_KEY>
```

### 配置同步请求体

```json
{
  "inbounds": [
    {
      "tag": "vless-in",
      "type": "vless",
      "listen": "::",
      "listen_port": 443,
      "users": [
        {"name": "user1", "uuid": "xxx", "flow": "xtls-rprx-vision"}
      ],
      "tls": { ... },
      "transport": { ... }
    }
  ]
}
```

### 流量统计响应

```json
{
  "users": [
    {
      "name": "user1",
      "upload": 1024000,
      "download": 2048000
    }
  ],
  "total": {
    "upload": 10240000,
    "download": 20480000
  }
}
```

---

## 订阅链接生成

### 支持格式
1. **sing-box** - 原生 JSON 配置
2. **Clash/Clash.Meta** - YAML 配置
3. **V2Ray/V2RayN** - Base64 编码链接
4. **Shadowrocket** - Base64 编码链接

### 订阅端点

```
GET /api/sub/{user_uuid}?format=singbox
GET /api/sub/{user_uuid}?format=clash
GET /api/sub/{user_uuid}?format=v2ray
```

---

## 实现阶段

### 阶段 1: 基础框架
- [ ] 项目结构搭建
- [ ] Panel: Gin 路由 + SQLite 初始化
- [ ] Node: sing-box 库嵌入 + Box 实例管理
- [ ] Panel-Node 基础通信

### 阶段 2: 核心功能
- [ ] 用户 CRUD API
- [ ] 节点管理 API
- [ ] 入站配置管理
- [ ] 用户同步到 Node
- [ ] 流量统计收集

### 阶段 3: 订阅系统
- [ ] 订阅链接生成 (sing-box 格式)
- [ ] Clash 格式转换
- [ ] V2Ray 格式转换

### 阶段 4: 前端开发
- [ ] React 项目搭建
- [ ] 仪表盘页面
- [ ] 用户管理页面
- [ ] 节点管理页面
- [ ] 入站配置页面

### 阶段 5: 高级功能
- [ ] 用户流量限制
- [ ] 用户到期自动禁用
- [ ] Webhook 通知
- [ ] 多协议模板

---

## 关键依赖

```go
// Panel go.mod
require (
    github.com/gin-gonic/gin v1.11.0
    github.com/mattn/go-sqlite3 v1.14.32
    github.com/golang-jwt/jwt/v5 v5.2.2
)

// Node go.mod
require (
    github.com/sagernet/sing-box v1.12.19
    github.com/gin-gonic/gin v1.11.0
)
```

---

## 验证方式

1. **单元测试**: 各模块独立测试
2. **集成测试**: Panel-Node 通信测试
3. **端到端测试**:
   - 创建用户 → 同步到 Node → 生成订阅 → 客户端连接验证
   - 用户更新 → Node 热更新验证
   - 流量统计准确性验证
