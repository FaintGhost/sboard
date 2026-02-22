# OpenAPI Spec 统一前后端 API 对接

## 背景与动机

当前 sboard 项目的前后端 API 对接完全依赖手动维护：

- **后端 (Go + Gin)**：请求/响应类型以 Go struct 分散在 `panel/internal/api/*.go` 各 handler 文件中（`userDTO`, `nodeDTO`, `groupDTO`, `inboundDTO` 等）
- **前端 (React + TypeScript)**：类型定义手动维护在 `panel/web/src/lib/api/types.ts`（~197 行），API 客户端手写在 `panel/web/src/lib/api/*.ts`

**核心痛点**：
1. Go struct 与 TypeScript type 手动双写，极易漂移
2. 后端新增/修改字段后，前端必须人工同步，遗漏即 bug
3. 无契约（contract）层保证前后端一致性
4. API 参数校验逻辑前后端各写一遍，无法复用
5. 无 API 文档，新增端点时缺乏参考

## 方案选型

**Source of Truth**: 手写 OpenAPI 3.1 YAML 作为唯一事实来源

| 层 | 工具 | 生成物 |
|---|------|--------|
| OpenAPI Spec | 手写 `panel/openapi.yaml` | 所有端点的契约定义 |
| Go 后端 | `oapi-codegen` (gin-server + strict-server + models) | Gin 服务器接口、请求/响应类型、参数绑定 |
| TS 前端 | `@hey-api/openapi-ts` (SDK + Zod + TypeScript) | API 客户端函数、Zod 校验 schema、TypeScript 类型 |

**选型理由**:
- `oapi-codegen`：Go 生态最成熟的 OpenAPI 代码生成器，原生支持 Gin，strict-server 模式提供类型安全的处理函数签名
- `@hey-api/openapi-ts`：TypeScript 生态活跃维护的 OpenAPI 客户端生成器，插件体系支持 Zod 运行时校验
- OpenAPI 3.1：最新规范，完全兼容 JSON Schema，支持 `nullable` 等现代特性

## 迁移策略

**一次性全量迁移**：覆盖现有全部 ~37 个操作（约 25 条路径），一步到位替换前后端的手写代码。

理由：
- 避免新旧混用期的维护负担
- 项目规模适中，全量迁移可控
- 一次性建立完整的代码生成管线

## 需求与成功标准

### 必须满足
1. 所有现有 API 端点在 OpenAPI spec 中完整定义
2. Go 后端使用 oapi-codegen 生成的类型和接口，手写 handler 实现 `StrictServerInterface`
3. TS 前端使用生成的 SDK 客户端 + Zod schema，替换手写 API 层
4. CI 管线校验生成代码的新鲜度（`make generate && git diff --exit-code`）
5. 所有现有测试通过（可调整以适配新接口）
6. 不改变现有 API 的外部行为（请求/响应格式兼容）

### 可选
- API 文档生成（Swagger UI 或 ReDoc）
- 前端 Zod 自动校验 SDK 请求/响应

## 设计文档

- [架构设计](./architecture.md) - 系统架构、文件布局、代码生成配置、迁移路径
- [BDD 规格](./bdd-specs.md) - 行为场景和测试策略
- [最佳实践](./best-practices.md) - 安全、性能和代码质量指南
