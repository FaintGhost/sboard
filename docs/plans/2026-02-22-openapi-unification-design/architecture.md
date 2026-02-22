# 架构设计

## 文件布局（目标状态）

```
panel/
  openapi.yaml                        # OpenAPI 3.1 spec (Single Source of Truth)
  internal/
    api/
      oapi.cfg.yaml                   # oapi-codegen 配置
      oapi_types.gen.go               # 生成: 请求/响应类型 (models)
      oapi_server.gen.go              # 生成: Gin server interface + strict server wrapper
      oapi_spec.gen.go                # 生成: embedded spec (可选)
      server.go                       # 新增: StrictServerInterface 实现入口
      server_users.go                 # 新增: Users 相关 handler 实现
      server_groups.go                # 新增: Groups 相关 handler 实现
      server_nodes.go                 # 新增: Nodes 相关 handler 实现
      server_inbounds.go              # 新增: Inbounds 相关 handler 实现
      server_sync_jobs.go             # 新增: SyncJobs 相关 handler 实现
      server_traffic.go               # 新增: Traffic 相关 handler 实现
      server_system.go                # 新增: System/Auth/Bootstrap 相关 handler 实现
      router.go                       # 重构: 使用生成的 RegisterHandlers
      auth.go                         # 保留: JWT 中间件
      cors.go                         # 保留: CORS 中间件
      request_logger.go               # 保留: 请求日志中间件
      node_sync_helpers.go            # 保留: sync 辅助函数
      # 以下文件将被删除（逻辑迁移到 server_*.go）:
      # users.go, nodes.go, groups.go, inbounds.go, etc.
  web/
    openapi-ts.config.ts              # @hey-api/openapi-ts 配置
    src/lib/api/
      gen/                            # 生成目录
        types.gen.ts                  # 生成: TypeScript 类型
        sdk.gen.ts                    # 生成: API 客户端函数
        zod.gen.ts                    # 生成: Zod 校验 schema
        client.gen.ts                 # 生成: fetch client 配置
      client.ts                       # 保留+重构: 自定义 client 配置（auth、error handling）
      # 以下文件将被删除（由生成代码替代）:
      # types.ts, users.ts, nodes.ts, groups.ts, inbounds.ts,
      # sync-jobs.ts, traffic.ts, auth.ts, system.ts,
      # singbox-tools.ts, user-groups.ts, group-users.ts, pagination.ts
```

## OpenAPI Spec 结构

### spec 文件: `panel/openapi.yaml`

```yaml
openapi: "3.1.0"
info:
  title: SBoard Panel API
  version: "1.0.0"
  description: SBoard 面板管理 API

servers:
  - url: /api
    description: Panel API

security: []  # 默认无认证

paths:
  # === 公开端点 ===
  /health:
    get: ...
  /admin/bootstrap:
    get: ...
    post: ...
  /admin/login:
    post: ...
  /sub/{user_uuid}:
    get: ...

  # === 认证端点 ===
  /admin/profile:
    get: ...
    put: ...
  /users:
    get: ...
    post: ...
  /users/{id}:
    get: ...
    put: ...
    delete: ...
  /users/{id}/groups:
    get: ...
    put: ...
  /groups:
    get: ...
    post: ...
  /groups/{id}:
    get: ...
    put: ...
    delete: ...
  /groups/{id}/users:
    get: ...
    put: ...
  /nodes:
    get: ...
    post: ...
  /nodes/{id}:
    get: ...
    put: ...
    delete: ...
  /nodes/{id}/health:
    get: ...
  /nodes/{id}/sync:
    post: ...
  /nodes/{id}/traffic:
    get: ...
  /traffic/nodes/summary:
    get: ...
  /traffic/total/summary:
    get: ...
  /traffic/timeseries:
    get: ...
  /system/info:
    get: ...
  /system/settings:
    get: ...
    put: ...
  /sync-jobs:
    get: ...
  /sync-jobs/{id}:
    get: ...
  /sync-jobs/{id}/retry:
    post: ...
  /inbounds:
    get: ...
    post: ...
  /inbounds/{id}:
    get: ...
    put: ...
    delete: ...
  /sing-box/format:
    post: ...
  /sing-box/check:
    post: ...
  /sing-box/generate:
    post: ...

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

  schemas:
    # === 通用 ===
    ErrorResponse:
      type: object
      required: [error]
      properties:
        error:
          type: string

    StatusResponse:
      type: object
      required: [status]
      properties:
        status:
          type: string

    # === 用户 ===
    User:
      type: object
      required: [id, uuid, username, group_ids, traffic_limit, traffic_used, traffic_reset_day, status]
      properties:
        id: { type: integer, format: int64 }
        uuid: { type: string }
        username: { type: string }
        group_ids: { type: array, items: { type: integer, format: int64 } }
        traffic_limit: { type: integer, format: int64 }
        traffic_used: { type: integer, format: int64 }
        traffic_reset_day: { type: integer }
        expire_at: { type: string, format: date-time, nullable: true }
        status: { $ref: '#/components/schemas/UserStatus' }

    UserStatus:
      type: string
      enum: [active, disabled, expired, traffic_exceeded]

    CreateUserRequest:
      type: object
      required: [username]
      properties:
        username: { type: string }

    UpdateUserRequest:
      type: object
      properties:
        username: { type: string }
        status: { type: string }
        expire_at: { type: string }
        traffic_limit: { type: integer, format: int64 }
        traffic_reset_day: { type: integer }

    # === 分组 ===
    Group:
      type: object
      required: [id, name, description, member_count]
      properties:
        id: { type: integer, format: int64 }
        name: { type: string }
        description: { type: string }
        member_count: { type: integer, format: int64 }

    # === 节点 ===
    Node:
      type: object
      required: [id, uuid, name, api_address, api_port, secret_key, public_address, status]
      properties:
        id: { type: integer, format: int64 }
        uuid: { type: string }
        name: { type: string }
        api_address: { type: string }
        api_port: { type: integer }
        secret_key: { type: string }
        public_address: { type: string }
        group_id: { type: integer, format: int64, nullable: true }
        status: { type: string }
        last_seen_at: { type: string, format: date-time, nullable: true }

    # === 入站 ===
    Inbound:
      type: object
      required: [id, uuid, tag, node_id, protocol, listen_port, public_port, settings]
      properties:
        id: { type: integer, format: int64 }
        uuid: { type: string }
        tag: { type: string }
        node_id: { type: integer, format: int64 }
        protocol: { type: string }
        listen_port: { type: integer }
        public_port: { type: integer }
        settings: {}  # free-form JSON
        tls_settings: { nullable: true }
        transport_settings: { nullable: true }

    # ... (SyncJob, Traffic, System 等 schema 同理)
```

### 响应格式约定

当前 API 使用多种响应格式，在 spec 中需如实反映：

| 模式 | 示例 | spec 处理 |
|------|------|-----------|
| `{"data": T}` | 大部分 list/get/create/update | 每个 response 定义 `data` 属性 |
| `{"data": T, "sync": {...}}` | inbound create/update | response 定义 `data` + `sync` 属性 |
| `{"status": "ok"}` | 部分 delete | 使用 `StatusResponse` schema |
| `{"status": "ok", "force": true, ...}` | node force delete | 专用 response schema |
| `{"message": "..."}` | hard delete user | 专用 response schema |
| `{"error": "..."}` | 所有错误 | 使用 `ErrorResponse` schema |

## 后端架构: oapi-codegen

### 配置文件: `panel/internal/api/oapi.cfg.yaml`

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/oapi-codegen/oapi-codegen/HEAD/configuration-schema.json
package: api
generate:
  gin-server: true
  strict-server: true
  models: true
  embedded-spec: false
output: oapi_server.gen.go
output-options:
  skip-prune: false
  nullable-type: true
```

> 注意：需要拆分为两个配置文件（types 和 server 分别生成），避免单文件过大。

#### 拆分配置

**types 配置** (`panel/internal/api/oapi_types.cfg.yaml`):
```yaml
package: api
generate:
  models: true
output: oapi_types.gen.go
output-options:
  nullable-type: true
```

**server 配置** (`panel/internal/api/oapi_server.cfg.yaml`):
```yaml
package: api
generate:
  gin-server: true
  strict-server: true
output: oapi_server.gen.go
output-options:
  nullable-type: true
```

### go generate 指令

在 `panel/internal/api/generate.go` 中添加：

```go
package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config oapi_types.cfg.yaml ../../openapi.yaml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config oapi_server.cfg.yaml ../../openapi.yaml
```

### Strict Server 实现模式

oapi-codegen 的 strict-server 模式会生成一个 `StrictServerInterface`：

```go
// 生成的接口（示例）
type StrictServerInterface interface {
    // (GET /users)
    ListUsers(ctx context.Context, request ListUsersRequestObject) (ListUsersResponseObject, error)
    // (POST /users)
    CreateUser(ctx context.Context, request CreateUserRequestObject) (CreateUserResponseObject, error)
    // (GET /users/{id})
    GetUser(ctx context.Context, request GetUserRequestObject) (GetUserResponseObject, error)
    // ...
}
```

实现代码示例：

```go
// server_users.go
type Server struct {
    store *db.Store
}

func (s *Server) ListUsers(ctx context.Context, req ListUsersRequestObject) (ListUsersResponseObject, error) {
    limit := defaultListLimit
    if req.Params.Limit != nil {
        limit = *req.Params.Limit
    }
    offset := 0
    if req.Params.Offset != nil {
        offset = *req.Params.Offset
    }
    status := ""
    if req.Params.Status != nil {
        status = string(*req.Params.Status)
    }

    users, err := listUsersForStatus(ctx, s.store, status, limit, offset)
    if err != nil {
        return ListUsers500JSONResponse{Error: "list users failed"}, nil
    }

    // batch fetch group IDs...
    dtos := buildUserDTOs(users, groupIDsMap)
    return ListUsers200JSONResponse{Data: dtos}, nil
}
```

### Router 重构

```go
// router.go (重构后)
func NewRouter(cfg config.Config, store *db.Store) *gin.Engine {
    r := gin.New()
    r.Use(RequestLogger(cfg.LogRequests))
    r.Use(gin.Recovery())
    r.Use(CORSMiddleware(cfg.CORSAllowOrigins))

    server := &Server{store: store, cfg: cfg}
    strictHandler := NewStrictHandler(server, []StrictMiddlewareFunc{})

    // 公开路由
    RegisterHandlersWithOptions(r, strictHandler, GinServerOptions{
        BaseURL: "/api",
        Middlewares: []MiddlewareFunc{},
    })

    // 注意：认证中间件需要通过 oapi-codegen 的 middleware 机制处理
    // 或使用 operationId-based 的 middleware 分组

    if cfg.ServeWeb {
        ServeWebUI(r, cfg.WebDir)
    }
    return r
}
```

### 认证中间件处理

oapi-codegen 支持 security scheme 中间件。在 spec 中标记需要认证的端点：

```yaml
paths:
  /users:
    get:
      security:
        - BearerAuth: []
```

实现认证 middleware：

```go
// auth_middleware.go
func NewAuthenticator(jwtSecret string) func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
    return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
        // JWT 校验逻辑
    }
}
```

## 前端架构: @hey-api/openapi-ts

### 配置文件: `panel/web/openapi-ts.config.ts`

```typescript
import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../openapi.yaml",
  output: "src/lib/api/gen",
  plugins: [
    "@hey-api/typescript",  // TypeScript 类型
    {
      name: "@hey-api/sdk",
      // 异步函数风格
      asClass: false,
    },
    "@hey-api/client-fetch", // Fetch API client
    {
      name: "zod",
      definitions: true,    // 为所有 schema 生成 Zod schema
      requests: true,       // 生成请求校验 schema
      responses: true,      // 生成响应校验 schema
    },
  ],
});
```

### 自定义 Client 配置

保留并重构 `panel/web/src/lib/api/client.ts`：

```typescript
import { client } from "./gen/client.gen";
import { useAuthStore } from "@/store/auth";

// 配置 base URL
client.setConfig({
  baseUrl: import.meta.env.VITE_API_BASE_URL?.trim() || window.location.origin,
});

// Auth 拦截器
client.interceptors.request.use((request) => {
  const token = useAuthStore.getState().token;
  if (token) {
    request.headers.set("Authorization", `Bearer ${token}`);
  }
  request.headers.set("Accept", "application/json");
  return request;
});

// 401 处理
client.interceptors.response.use((response) => {
  if (response.status === 401) {
    useAuthStore.getState().clearToken();
  }
  return response;
});

export { client };
export { ApiError } from "./gen/sdk.gen";
```

### 前端使用模式

#### 迁移前（手写）
```typescript
import { listUsers } from "@/lib/api/users";
import type { User } from "@/lib/api/types";

const users: User[] = await listUsers({ limit: 50 });
```

#### 迁移后（生成）
```typescript
import { listUsers } from "@/lib/api/gen/sdk.gen";

const { data } = await listUsers({ query: { limit: 50 } });
const users = data?.data; // data 属性来自 {"data": [...]} 响应格式
```

#### 配合 React Query
```typescript
import { listUsers } from "@/lib/api/gen/sdk.gen";
import { useQuery } from "@tanstack/react-query";

function useUsers(params?: { status?: UserStatus }) {
  return useQuery({
    queryKey: ["users", params],
    queryFn: async () => {
      const { data } = await listUsers({ query: params });
      return data?.data ?? [];
    },
  });
}
```

## 构建集成

### Makefile

```makefile
.PHONY: generate generate-go generate-ts check-generate

generate: generate-go generate-ts

generate-go:
	cd panel && go generate ./internal/api/...

generate-ts:
	cd panel/web && bun run generate

check-generate: generate
	git diff --exit-code -- '*.gen.go' '*.gen.ts' || \
		(echo "ERROR: Generated code is out of date. Run 'make generate' and commit." && exit 1)
```

### package.json 新增脚本

```json
{
  "scripts": {
    "generate": "openapi-ts"
  }
}
```

### CI 集成

在 CI 管线中增加检查步骤：

```yaml
- name: Check generated code freshness
  run: make check-generate
```

## 迁移路径

### 阶段 1: 基础设施搭建
1. 编写完整 `panel/openapi.yaml`
2. 配置 oapi-codegen（安装依赖、编写配置文件、添加 go generate 指令）
3. 配置 @hey-api/openapi-ts（安装依赖、编写配置文件、添加 npm script）
4. 运行代码生成，验证输出

### 阶段 2: 后端迁移
1. 创建 `Server` struct 实现 `StrictServerInterface`
2. 逐个迁移现有 handler 到 strict server 方法（保留业务逻辑，改变签名）
3. 重构 `router.go` 使用生成的 `RegisterHandlers`
4. 处理认证中间件集成
5. 删除旧的 handler 文件和 DTO 定义
6. 运行并修复所有后端测试

### 阶段 3: 前端迁移
1. 运行代码生成，检查生成的客户端代码
2. 配置自定义 client（auth、error handling）
3. 逐个页面替换 API 调用为生成的 SDK
4. 删除旧的手写 API 文件
5. 运行并修复所有前端测试
6. 验证 Zod schema 校验行为

### 阶段 4: 清理与验证
1. 删除所有废弃文件
2. 运行全量测试
3. 添加 CI check-generate 步骤
4. 更新 CLAUDE.md 文档

## 需要关注的复杂点

### 1. Inbound 的复合响应

Inbound create/update 返回 `{"data": inbound, "sync": syncResult}`。这需要在 spec 中定义为包含两个字段的响应对象。strict server 的返回类型需要包含 sync 信息。

### 2. 认证中间件与路由分组

当前路由用 Gin 的 `Group` + `Use(AuthMiddleware)` 实现分组认证。迁移到 oapi-codegen 后，需要使用 spec 的 `security` 声明 + oapi-codegen 的 middleware 机制。

### 3. Subscription 端点

`GET /api/sub/{user_uuid}` 根据 `format` 参数和 `User-Agent` 返回不同格式（JSON 或 plain text）。这需要在 spec 中用 `content` 的多种 media type 来表达。

### 4. 自由格式 JSON 字段

Inbound 的 `settings`、`tls_settings`、`transport_settings` 是自由格式 JSON（`json.RawMessage`）。在 OpenAPI 中用 `type: object` 或空 schema `{}` 表示。

### 5. 分页参数复用

多个 list 端点共用 `limit`/`offset` 查询参数。可以使用 OpenAPI 的 `$ref` 复用参数定义。
