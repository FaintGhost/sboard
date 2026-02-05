# SBoard 阶段 2（用户管理）设计

## 目标
阶段 2 聚焦 Panel 用户 CRUD，并引入管理员 JWT 鉴权与登录。管理员账号使用环境变量提供，用户创建仅需 `username`，系统自动生成 `uuid`，其余字段使用默认值。用户删除采用软删除。

## API 结构
- `POST /api/admin/login`：管理员登录，返回 JWT
- `GET /api/users`：用户列表（默认不含 disabled）
- `POST /api/users`：创建用户（仅需 `username`）
- `GET /api/users/:id`：用户详情
- `PUT /api/users/:id`：更新用户
- `DELETE /api/users/:id`：软删除（`status=disabled`）

## 认证与授权
- 环境变量：`ADMIN_USER`、`ADMIN_PASS`、`PANEL_JWT_SECRET`
- JWT：HS256，`sub=admin`，`exp=24h`
- 中间件统一校验 `Authorization: Bearer <token>`，失败返回 401

## 数据模型与校验
- 表：沿用 `users`
- 创建：仅 `username` 必填，自动生成 `uuid`
- 默认值：`traffic_limit=0`、`traffic_used=0`、`traffic_reset_day=0`、`expire_at=NULL`、`status=active`
- 更新字段：`username`、`status`、`expire_at`、`traffic_limit`、`traffic_reset_day`
- 校验：
  - `username` 非空、唯一
  - `status` 取值限定 `active/disabled/expired/traffic_exceeded`
  - `traffic_limit >= 0`
  - `traffic_reset_day` ∈ [0,31]
  - `expire_at` 为合法时间或空

## 列表与分页
- 默认过滤 `status != disabled`
- `status=disabled` 时返回禁用用户
- 分页：`limit`/`offset`（默认 `limit=50`、`offset=0`）

## 错误与返回
- 成功：`{data: ...}`
- 失败：`{error: "..."}`
- 状态码：
  - 400 参数/校验错误
  - 401 未授权
  - 404 不存在
  - 409 唯一约束冲突
  - 500 内部错误

## 实现拆分
- `panel/internal/api/auth.go`：登录与鉴权中间件
- `panel/internal/api/users.go`：用户 CRUD
- `panel/internal/db/users.go`：用户 DAO 与 SQL
- `panel/internal/config/config.go`：新增 `ADMIN_USER`、`ADMIN_PASS`、`PANEL_JWT_SECRET`

## 测试
- DB：SQLite 临时库验证 CRUD 与唯一约束
- API：登录成功/失败、鉴权失败、创建/更新/禁用、分页与状态过滤
