# BDD 规格

## Feature: OpenAPI Spec 代码生成

### Scenario: 从 OpenAPI spec 生成 Go 后端代码
```gherkin
Given 存在有效的 panel/openapi.yaml 文件
When 执行 go generate ./internal/api/...
Then 在 panel/internal/api/ 下生成 oapi_types.gen.go
And 在 panel/internal/api/ 下生成 oapi_server.gen.go
And 生成的 Go 代码通过编译
And 生成的 StrictServerInterface 包含所有 spec 中定义的操作
```

### Scenario: 从 OpenAPI spec 生成 TypeScript 前端代码
```gherkin
Given 存在有效的 panel/openapi.yaml 文件
And panel/web/openapi-ts.config.ts 已正确配置
When 执行 bun run generate
Then 在 panel/web/src/lib/api/gen/ 下生成 types.gen.ts
And 在 panel/web/src/lib/api/gen/ 下生成 sdk.gen.ts
And 在 panel/web/src/lib/api/gen/ 下生成 zod.gen.ts
And 生成的 TypeScript 代码通过 tsc 编译
```

### Scenario: 生成代码新鲜度检查
```gherkin
Given 代码仓库中的生成代码是最新的
When 执行 make check-generate
Then 命令成功退出（exit code 0）

Given 开发者修改了 openapi.yaml 但未重新生成
When 执行 make check-generate
Then 命令失败退出（exit code 1）
And 输出提示 "Generated code is out of date"
```

## Feature: 后端 Strict Server 行为

### Scenario: 列表用户 API
```gherkin
Given 数据库中存在 3 个活跃用户
When 发送 GET /api/users 带有 Bearer token
Then 响应状态码为 200
And 响应体为 {"data": [<3个用户对象>]}
And 每个用户对象包含 id, uuid, username, group_ids, traffic_limit, traffic_used, traffic_reset_day, status 字段
```

### Scenario: 创建用户 API - 请求校验
```gherkin
When 发送 POST /api/users 带有空 body
Then 响应状态码为 400
And 响应体包含 {"error": "..."}

When 发送 POST /api/users 带有 {"username": ""}
Then 响应状态码为 400
And 响应体包含 {"error": "invalid username"}
```

### Scenario: 未认证访问受保护端点
```gherkin
When 发送 GET /api/users 不带 Authorization header
Then 响应状态码为 401
And 响应体包含 {"error": "..."}
```

### Scenario: Inbound 创建的复合响应
```gherkin
Given 存在 node_id=1 的节点
When 发送 POST /api/inbounds 带有合法 inbound 数据
Then 响应状态码为 201
And 响应体包含 {"data": <inbound对象>, "sync": {"status": "..."}}
```

### Scenario: 分页参数
```gherkin
When 发送 GET /api/users?limit=10&offset=20
Then 后端正确接收 limit=10, offset=20 参数
And 返回对应分页的用户列表

When 发送 GET /api/users?limit=-1
Then 响应状态码为 400
```

## Feature: 前端 SDK 使用

### Scenario: 前端调用列表用户
```gherkin
Given 前端已配置 client（auth token + base URL）
When 调用 listUsers({ query: { limit: 50 } })
Then SDK 发送 GET /api/users?limit=50
And 返回值 data 类型为 { data: User[] }
```

### Scenario: 前端 Zod 校验
```gherkin
Given 生成的 Zod schema 包含 User schema
When 使用 zUser.parse(apiResponse) 校验后端响应
Then 合法数据通过校验
And 缺少 required 字段时抛出 ZodError
```

### Scenario: 前端 401 自动登出
```gherkin
Given 用户已登录（token 存在于 store）
When API 返回 401 状态码
Then client interceptor 清除 auth store 中的 token
And 用户被重定向到登录页
```

## Feature: 类型一致性

### Scenario: Go 类型与 spec 一致
```gherkin
Given openapi.yaml 中定义 User schema 有 9 个 required 字段
When oapi-codegen 生成 Go 类型
Then 生成的 User struct 包含对应的 9 个字段
And 字段类型与 spec 定义匹配（int64, string, []int64 等）
```

### Scenario: TypeScript 类型与 spec 一致
```gherkin
Given openapi.yaml 中定义 User schema
When @hey-api/openapi-ts 生成 TypeScript 类型
Then 生成的 User type 字段与 spec 完全匹配
And nullable 字段在 TS 中表示为 T | null
```

### Scenario: Spec 变更触发双端更新
```gherkin
Given openapi.yaml 中 User 新增 email 字段
When 重新运行 make generate
Then Go 的 User struct 包含 Email 字段
And TypeScript 的 User type 包含 email 字段
And Zod 的 zUser schema 包含 email 校验
```

## 测试策略

### 后端测试

- **单元测试**：对每个 `StrictServerInterface` 方法进行单元测试（纯函数式，不依赖 Gin context）
- **集成测试**：使用 `httptest` 对完整 HTTP 请求/响应进行测试（复用现有测试模式）
- **契约测试**：CI 中 `make check-generate` 确保代码与 spec 同步

### 前端测试

- **类型测试**：`tsc -b` 编译检查确保类型正确
- **Zod 测试**：对关键 schema 进行 `.parse()` / `.safeParse()` 测试
- **集成测试**：使用 MSW (Mock Service Worker) 或 vitest mock 测试 SDK 调用
