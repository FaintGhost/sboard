# SBoard Phase 4 前端规划与技术规格（给 AI 前端与后端对接）

## 目标与范围
本规格用于指导前端实现 Phase 4 后台管理面板，读完即可直接开发。范围包括：登录、仪表盘、用户管理、分组管理、节点管理、入站管理、订阅/链接、系统设置。  
非目标：多协议订阅（当前仅 sing-box）、复杂 RBAC、DNS/Route 细节编辑、审计系统。

## 技术栈与约束
- React + Vite
- TailwindCSS v4
- shadcn/ui（最新版）
- Zustand
- TanStack Query
- Zod + react-hook-form
- Token 存储：`localStorage`
- 15s 轮询节点健康

## 布局与路由
- 左侧固定侧边栏 + 顶部工具条
- 登录页独立布局

路由：
- `/login`
- `/`（仪表盘）
- `/users`
- `/groups`
- `/nodes`
- `/inbounds`
- `/subscriptions`
- `/settings`

未登录访问受保护路由 → 跳转 `/login`。登录成功后回跳或进 `/`。

## 状态与数据流
- 服务器状态：TanStack Query
- 本地 UI 状态：Zustand（侧边栏折叠、筛选条件、弹窗开关）
- 表单校验：Zod + react-hook-form

统一请求封装：
- 自动注入 `Authorization: Bearer <token>`
- 401 自动清理 token 并跳转 `/login`
- 错误统一 toast

## 核心数据模型（前端视角）
### User
```json
{
  "id": 1,
  "uuid": "user-uuid",
  "username": "alice",
  "status": "active",
  "traffic_used": 0,
  "traffic_limit": 0,
  "traffic_reset_day": 0,
  "expire_at": "2026-02-05T12:00:00Z",
  "group_ids": [1,2]
}
```

### Group
```json
{
  "id": 1,
  "name": "VIP",
  "description": "高端用户组"
}
```

### Node
```json
{
  "id": 1,
  "uuid": "node-uuid",
  "name": "node-a",
  "api_address": "10.0.0.2",
  "api_port": 2222,
  "public_address": "a.example.com",
  "secret_key": "auto-generated",
  "group_id": 1,
  "status": "online",
  "last_seen_at": "2026-02-05T12:00:00Z"
}
```

### Inbound
```json
{
  "id": 1,
  "uuid": "inb-uuid",
  "tag": "vless-in",
  "protocol": "vless",
  "node_id": 1,
  "listen_port": 443,
  "public_port": 443,
  "settings": {},
  "tls_settings": {},
  "transport_settings": {}
}
```

## 页面与字段规格
### 登录
- 字段：`username`、`password`
- 成功后保存 token → 跳转

### 仪表盘
- 指标卡片：用户数、节点数、入站数、24h 流量（可占位）
- 最近变更（可占位）

### 用户管理
- 列表列：`username/uuid/status/traffic_used/traffic_limit/expire_at`
- 创建：仅 `username`
- 编辑：`status/expire_at/traffic_limit/traffic_reset_day`
- 分组：多选（用户可加入多个分组）

### 分组管理
- 字段：`name`、`description`
- 列表支持新增/编辑/删除
- 详情展示该分组下的节点与用户（只读列表）

### 节点管理
- 字段：`name/api_address/api_port/public_address/group`
- 创建：`secret_key` 后端自动生成并返回
- 创建成功弹窗：展示 `docker-compose.yml`（占位镜像名 `sboard-node:latest`）
- 健康检查：手动按钮 + 15s 轮询

### 入站管理
基础字段 + 高级 JSON 编辑器：
- 基础字段：`protocol/tag/node/listen_port/public_port`
- 高级：`settings/tls_settings/transport_settings`（JSON）

### 订阅与链接
- 展示每个用户订阅 URL（`/api/sub/{uuid}`）
- UA/format 行为提示
- 一键复制

### 系统设置
- 占位，后续扩展（如基础 URL、环境标识）

## 订阅行为说明
订阅由后端生成，仅展示 URL：
- `?format=singbox` 优先，返回 JSON
- UA 命中 `sing-box/SFA/SFI` 返回 JSON
- 其他 UA 返回 Base64(JSON)

## 拟定 API 契约（后端需对接实现）
### 已实现
- `POST /api/admin/login`
- `GET /api/users` / `POST /api/users`
- `GET /api/users/:id` / `PUT /api/users/:id` / `DELETE /api/users/:id`
- `GET /api/sub/:user_uuid`

### 需实现（前后端对接）
- 分组
  - `GET /api/groups`
  - `POST /api/groups`
  - `GET /api/groups/:id`
  - `PUT /api/groups/:id`
  - `DELETE /api/groups/:id`
- 节点
  - `GET /api/nodes`
  - `POST /api/nodes`（返回 `secret_key` 与 `docker_compose`）
  - `GET /api/nodes/:id`
  - `PUT /api/nodes/:id`
  - `DELETE /api/nodes/:id`
  - `GET /api/nodes/:id/health`
- 入站
  - `GET /api/inbounds`
  - `POST /api/inbounds`
  - `GET /api/inbounds/:id`
  - `PUT /api/inbounds/:id`
  - `DELETE /api/inbounds/:id`
- 关联
  - `PUT /api/users/:id/groups`（body: `{group_ids: []}`）
  - `PUT /api/nodes/:id/group`（body: `{group_id: 1}`）

## UI 交互要点
- 创建节点后弹窗展示 `docker-compose.yml`，一键复制
- 节点列表显示当前在线状态，15s 轮询刷新
- JSON 编辑器：格式校验失败则禁止提交
- 表格支持搜索/筛选/分页（默认 limit/offset）

## 交付物
- 前端工程初始化（React/Vite/Tailwind/shadcn）
- 上述页面与路由
- API hooks + 类型定义
- 基础样式与响应式布局

