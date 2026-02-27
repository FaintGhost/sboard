# Task 002: 共享 Fixtures 实现

**depends-on**: task-001

## Description

创建 Playwright 共享 fixtures，封装认证流程（Bootstrap + Login + storageState）、Panel/Node API 请求封装、测试数据工厂。这些 fixtures 是所有 smoke 和 e2e 测试的基础。

## Execution Context

**Task Number**: 002 of 012
**Phase**: Foundation
**Prerequisites**: Task 001 完成（项目基础设施就绪）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: 支撑所有场景的 "Given 管理员已登录" 前置条件

## Files to Create

- `e2e/tests/fixtures/auth.fixture.ts` — 认证 fixture（Bootstrap + Login + token 注入）
- `e2e/tests/fixtures/api.fixture.ts` — Panel API 和 Node API 请求封装
- `e2e/tests/fixtures/test-data.fixture.ts` — 测试数据工厂（唯一名称生成等）
- `e2e/tests/fixtures/index.ts` — 统一导出

## Steps

### Step 1: 研究 Panel API 接口

查阅 `panel/proto/sboard/panel/v1/panel.proto` 与 `panel/web/src/lib/rpc/gen/**` 了解以下 RPC 的请求/响应格式：
- `POST /rpc/sboard.panel.v1.AuthService/GetBootstrapStatus` — 检查是否需要初始化
- `POST /rpc/sboard.panel.v1.AuthService/Bootstrap` — 执行初始化（`setupToken` / `xSetupToken`）
- `POST /rpc/sboard.panel.v1.AuthService/Login` — 登录获取 JWT token
- `POST /rpc/sboard.panel.v1.HealthService/GetHealth` — Panel 健康检查

查阅 Panel 前端代码了解 token 存储方式：
- 检查 `panel/web/src/store/auth.ts` 或相关文件，确认 token 存储在 localStorage 的哪个 key 下
- 检查前端路由保护逻辑，确认 token 如何被注入到请求中

### Step 2: 创建 auth.fixture.ts

实现一个扩展 Playwright base test 的 fixture，提供：
- `authenticatedPage` — 自动完成 Bootstrap（如需要）+ Login + 注入 token 到浏览器的 page
- 使用 Playwright `request` API 直接调 Panel API 完成 Bootstrap 和 Login（不走 UI，提高速度和稳定性）
- 获取 JWT token 后，通过 `page.evaluate` 注入到 localStorage
- 导航到首页确认认证状态

关键注意事项：
- 检查 Bootstrap API 是否需要 `X-Setup-Token` header 还是其他认证方式
- token key 名称必须与前端代码一致
- 考虑 Bootstrap 幂等性（已初始化时跳过）

### Step 3: 创建 api.fixture.ts

封装常用 API 操作，提供：
- `panelAPI` — 带认证的 Panel RPC 请求封装（创建用户、创建节点、创建入站等）
- `nodeAPI` — Node API 请求封装（查询健康状态、查询配置等，使用 Bearer Token 鉴权）
- 环境变量：`BASE_URL`（Panel）、`NODE_API_URL`（Node）、`SETUP_TOKEN`

封装应该足够薄 — 仅处理认证 header 注入和 base URL，不要过度封装具体业务操作。

### Step 4: 创建 test-data.fixture.ts

实现测试数据工厂，提供：
- 唯一名称生成器（基于时间戳或随机数的前缀，避免测试间冲突）
- 默认测试数据模板（用户、分组、节点等的默认值）

### Step 5: 创建统一导出 index.ts

从 `index.ts` 导出组合后的 test fixture，使测试文件可以通过单个 import 使用所有 fixtures。

### Step 6: 编写验证测试

创建一个最小的测试文件 `e2e/tests/smoke/health.smoke.spec.ts`（Task 003 的占位），仅用于验证 fixture 加载无误：
- import fixtures
- 写一个空的 test 只验证 import 不报错

### Step 7: 在 Docker 环境中验证

启动 Docker 环境，运行占位测试，确认：
- Panel 容器启动并通过 healthcheck
- Node 容器启动并通过 healthcheck
- Playwright 容器能访问 Panel 和 Node
- Auth fixture 能成功完成 Bootstrap + Login

## Verification Commands

```bash
# 验证 TypeScript 编译无误
cd e2e && bunx tsc --noEmit

# 在 Docker 中运行验证（启动服务 + 运行占位测试）
cd e2e && docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright
```

## Success Criteria

- Fixture 文件 TypeScript 编译无误
- Docker 环境可正常启动（panel + node 健康检查通过）
- Auth fixture 能在 Docker 环境中成功完成 Bootstrap 和 Login
- API fixture 能成功调用 Panel 和 Node 的 health 端点
