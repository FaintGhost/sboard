# E2E 测试架构设计

## 容器编排架构

```
                  docker-compose.e2e.yml
                  ┌─────────────────────────────────────────────┐
                  │            network: e2e-net                  │
                  │                                              │
   build from     │  ┌──────────────┐   ┌──────────────┐        │
   panel/         │  │    panel     │   │    node      │        │
   Dockerfile     │  │   GO + SPA   │   │  GO+sing-box │        │
                  │  │   :8080      │   │   :3000      │        │
                  │  │              │   │              │        │
                  │  │ rpc health   │   │ healthcheck: │        │
                  │  │ (entrypoint) │   │ /api/health  │        │
                  │  └──────┬───────┘   └──────┬───────┘        │
                  │         │                  │                │
                  │         │   e2e-net        │                │
                  │         │                  │                │
                  │  ┌──────┴──────────────────┴───────┐        │
                  │  │         playwright               │        │
                  │  │  mcr.microsoft.com/playwright    │        │
                  │  │                                  │        │
                  │  │  depends_on:                     │        │
                  │  │    panel: service_healthy         │        │
                  │  │    node:  service_healthy         │        │
                  │  │                                  │        │
                  │  │  env:                            │        │
                  │  │    BASE_URL=http://panel:8080    │        │
                  │  │    NODE_API_URL=http://node:3000 │        │
                  │  │    SETUP_TOKEN=...               │        │
                  │  └──────────────┬───────────────────┘        │
                  │                 │                            │
                  └─────────────────┼────────────────────────────┘
                                    │
                              volume mounts
                                    │
                        ┌───────────┴───────────┐
                        │ host: e2e/            │
                        │  playwright-report/   │
                        │  test-results/        │
                        └───────────────────────┘
```

## 服务启动顺序

```
1. panel + node 并行构建和启动
2. panel + node 进入 running
3. playwright 容器启动
4. entrypoint 等待 panel RPC 健康检查通过
5. entrypoint 等待 node /api/health 通过
6. Playwright 执行测试套件
7. 退出码传播到 docker compose
```

## 网络通信

| 来源 | 目标 | 协议 | 用途 |
|------|------|------|------|
| playwright → panel | `http://panel:8080` | HTTP | 浏览器访问 Web UI + API 调用 |
| playwright → node | `http://node:3000` | HTTP | API 级别验证 Node 配置/状态 |
| panel → node | `http://node:3000` | HTTP | 配置同步推送（Bearer Token 鉴权） |
| 浏览器(in playwright) → panel | `http://panel:8080` | HTTP | Playwright 浏览器内的网络请求 |

**关键点**：Playwright 浏览器运行在 playwright 容器内，因此浏览器请求 `panel:8080` 能通过 Docker 内部 DNS 解析。`baseURL` 在 Playwright 配置中设为 `http://panel:8080`。

## 测试执行流程

### Smoke 测试流程

```
1. API 检查 Panel/Node 健康
2. 检查 Bootstrap 状态 → 执行 Bootstrap（AuthService RPC）
3. 浏览器访问登录页 → 输入凭据 → 验证跳转到 Dashboard
4. 验证核心路由可加载（/users, /nodes, /subscriptions 等）
```

### E2E 全链路流程

```
1. [Setup] Bootstrap + Login → 保存认证状态
2. [Users] 创建测试用户 → 验证列表 → 编辑 → 删除
3. [Groups] 创建分组 → 分配用户
4. [Nodes] 创建节点（api_address=node, api_port=3000, secret_key=e2e-test-node-secret）
5. [Inbounds] 创建入站配置 → 关联节点
6. [Sync] 触发同步 → 通过 Node API 验证配置已生效 + sing-box running
7. [Subscriptions] 创建订阅 → 获取链接 → 验证内容格式
```

## 测试分层策略

```
                ┌─────────────────┐
                │   Smoke Tests   │  ← 每次部署后运行
                │  (~5 tests)     │     < 30 秒
                │  核心可用性验证  │
                └────────┬────────┘
                         │
                ┌────────┴────────┐
                │   E2E Tests     │  ← 定期 / PR 合并前运行
                │  (~20+ tests)   │     < 5 分钟
                │  完整用户流程    │
                └─────────────────┘
```

### Playwright Projects 映射

| Project | testMatch | retries | 用途 |
|---------|-----------|---------|------|
| `smoke` | `*.smoke.spec.ts` | 0 | 快速冒烟验证 |
| `e2e` | `tests/e2e/*.spec.ts` | 2 (CI) | 完整功能测试 |

## Fixture 架构

```
base Playwright test
  └─ auth.fixture.ts
       ├─ authenticatedPage  → Bootstrap + Login + 注入 token
       └─ apiContext         → 带认证的 API 请求上下文
  └─ api.fixture.ts
       ├─ panelAPI          → Panel RPC API 封装
       └─ nodeAPI           → Node REST API 封装
  └─ test-data.fixture.ts
       └─ testData          → 测试数据工厂（唯一名称生成等）
```

## 构建优化

### Panel 构建缓存

Panel Dockerfile 使用多阶段构建 + Go/Bun 缓存挂载。在 E2E 场景中：
- 首次构建较慢（Go 编译 + 前端打包）
- 后续构建利用 Docker layer cache 和 BuildKit cache mount 加速

### Playwright 镜像

使用官方预装浏览器的镜像，避免每次安装浏览器。镜像版本必须与 `@playwright/test` 版本匹配。

## 环境变量配置

### Panel 环境变量

| 变量 | 测试环境值 | 说明 |
|------|-----------|------|
| `PANEL_JWT_SECRET` | `e2e-test-jwt-secret-key-32chars!!` | JWT 签名密钥 |
| `PANEL_SETUP_TOKEN` | `e2e-test-setup-token` | 固定 setup token |
| `PANEL_DB_PATH` | `/data/panel.db` | 临时卷，每次全新 |
| `PANEL_SERVE_WEB` | `true` | 内嵌服务前端 |
| `PANEL_NODE_MONITOR_INTERVAL` | `5s` | 加快监控频率 |
| `PANEL_TRAFFIC_MONITOR_INTERVAL` | `10s` | 加快流量采样 |

### Node 环境变量

| 变量 | 测试环境值 | 说明 |
|------|-----------|------|
| `NODE_SECRET_KEY` | `e2e-test-node-secret` | 与 Panel 创建节点时填写一致 |
| `NODE_LOG_LEVEL` | `debug` | 调试级别日志 |

### Playwright 环境变量

| 变量 | 值 | 说明 |
|------|------|------|
| `CI` | `true` | 启用 CI 模式 |
| `BASE_URL` | `http://panel:8080` | Panel 地址（Docker 内部） |
| `NODE_API_URL` | `http://node:3000` | Node 地址（Docker 内部） |
| `SETUP_TOKEN` | `e2e-test-setup-token` | Bootstrap token |
