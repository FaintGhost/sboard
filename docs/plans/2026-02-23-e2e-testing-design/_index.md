# E2E 测试设计：基于 Playwright + Docker 的全容器化测试方案

## 背景与目标

SBoard 已有完善的 Go 单元测试（~86 文件）和 Vitest 前端单元测试（~21 文件），但缺乏端到端测试来验证：
- Panel Web UI 的核心用户流程
- Panel ↔ Node 的配置同步全链路
- 部署后的冒烟验证

### 目标

1. 建立基于 Playwright 的 E2E 测试框架，全容器化运行（Panel + Node + Playwright）
2. 分层组织：**Smoke 测试**（快速验证核心可用性）+ **E2E 测试**（完整用户流程）
3. 覆盖：登录、用户管理 CRUD、节点管理 + 同步、订阅功能、Node 配置验证
4. 每次测试从全新数据库开始，测试数据通过 API/UI 操作创建

## 需求

### 功能需求

| ID | 需求 | 优先级 |
|----|------|--------|
| FR-1 | Docker Compose 一键启动完整测试环境 | P0 |
| FR-2 | Smoke 测试：验证 Bootstrap、登录、首页加载 | P0 |
| FR-3 | E2E 测试：用户 CRUD（创建/编辑/删除/分组管理） | P0 |
| FR-4 | E2E 测试：节点 CRUD + 配置同步触发 | P0 |
| FR-5 | E2E 测试：订阅生成与内容验证 | P0 |
| FR-6 | E2E 测试：Node 配置下发验证（API 级别验证配置正确性 + sing-box 状态） | P0 |
| FR-7 | 测试报告（HTML）和失败截图/trace 输出到宿主机 | P1 |
| FR-8 | 支持按 suite 运行（仅 smoke / 仅 e2e / 全部） | P1 |

### 非功能需求

| ID | 需求 |
|----|------|
| NFR-1 | 测试环境完全隔离，不影响开发环境 |
| NFR-2 | Smoke 测试 < 30 秒完成 |
| NFR-3 | 全量 E2E 测试 < 5 分钟完成 |
| NFR-4 | 支持 CI 集成（可后续扩展 GitHub Actions） |

## 技术决策

### Docker 架构

全容器化方案，通过 `docker-compose.e2e.yml` 编排三个服务：

```
┌─────────────────────────────────────────────┐
│            docker network: e2e-net           │
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐  │
│  │  panel    │  │  node    │  │ playwright│  │
│  │  :8080   ├──┤  :3000   │  │  runner   │  │
│  │(app+web) │  │(sing-box)│  │  (tests)  │  │
│  └──────────┘  └──────────┘  └───────────┘  │
│       ↑                            │         │
│       └────── http://panel:8080 ───┘         │
│       └────── http://node:3000  ───┘         │
└─────────────────────────────────────────────┘
         │
    volume mounts
         │
    ┌────┴────┐
    │ host fs │
    │ e2e/    │
    │  ├─ playwright-report/  │
    │  └─ test-results/       │
    └─────────┘
```

### 技术选型

| 组件 | 选择 | 理由 |
|------|------|------|
| 测试框架 | Playwright Test | 现代 E2E 框架，官方 Docker 镜像，自动等待，API 测试支持 |
| 包管理器 | Bun | 与项目前端一致 |
| Docker 镜像 | `mcr.microsoft.com/playwright:v1.50.1-noble` | 官方预装浏览器镜像 |
| 等待策略 | Docker healthcheck + `depends_on: condition: service_healthy` | 原生支持，无需额外脚本 |
| 测试分层 | Playwright Projects + 文件命名约定 | 灵活按 suite 运行 |

### 测试数据策略

- 每次测试运行使用全新 SQLite 数据库（容器临时卷）
- 通过固定 `PANEL_SETUP_TOKEN` 完成 Bootstrap
- 使用 Playwright fixtures 封装认证状态（storageState 复用）
- 测试间通过唯一前缀避免数据冲突

## 详细设计

### 目录结构

```
e2e/
├── playwright.config.ts         # Playwright 配置（定义 smoke/e2e projects）
├── package.json                 # 依赖：@playwright/test, bun
├── bun.lock
├── docker-compose.e2e.yml       # 测试环境编排
├── Dockerfile                   # Playwright 测试运行器镜像
├── .env.e2e                     # 测试环境变量
├── tests/
│   ├── smoke/                   # Smoke 测试（快速验证）
│   │   ├── health.smoke.spec.ts     # API 健康检查
│   │   ├── bootstrap.smoke.spec.ts  # Bootstrap + 首次登录
│   │   └── navigation.smoke.spec.ts # 核心页面可访问
│   ├── e2e/                     # 完整 E2E 测试
│   │   ├── auth.spec.ts             # 登录/登出/token 过期
│   │   ├── users.spec.ts            # 用户 CRUD
│   │   ├── groups.spec.ts           # 分组管理
│   │   ├── nodes.spec.ts            # 节点 CRUD + 状态
│   │   ├── node-sync.spec.ts        # 节点配置同步 + 验证
│   │   ├── inbounds.spec.ts         # 入站配置
│   │   ├── subscriptions.spec.ts    # 订阅生成与验证
│   │   └── sync-jobs.spec.ts        # 同步任务
│   └── fixtures/                # 共享 fixtures
│       ├── auth.fixture.ts          # 认证 fixture（Bootstrap + Login + storageState）
│       ├── api.fixture.ts           # API helper（直接调 Panel/Node API）
│       └── test-data.fixture.ts     # 测试数据工厂
├── playwright-report/           # HTML 报告输出（.gitignore）
└── test-results/                # 截图/trace 输出（.gitignore）
```

### Docker Compose 编排

```yaml
# e2e/docker-compose.e2e.yml
services:
  panel:
    build:
      context: ..
      dockerfile: panel/Dockerfile
    environment:
      PANEL_JWT_SECRET: e2e-test-jwt-secret-key-32chars!!
      PANEL_SETUP_TOKEN: e2e-test-setup-token
      PANEL_DB_PATH: /data/panel.db
      PANEL_SERVE_WEB: "true"
      PANEL_WEB_DIR: /app/web/dist
      PANEL_LOG_REQUESTS: "true"
      PANEL_NODE_MONITOR_INTERVAL: 5s
      PANEL_TRAFFIC_MONITOR_INTERVAL: 10s
    volumes:
      - panel-data:/data
    healthcheck:
      # panel 由 playwright entrypoint 通过 RPC 健康检查等待
      interval: 3s
      timeout: 5s
      retries: 15
      start_period: 30s
    networks:
      - e2e-net

  node:
    build:
      context: ..
      dockerfile: node/Dockerfile
    environment:
      NODE_SECRET_KEY: e2e-test-node-secret
      NODE_LOG_LEVEL: debug
    volumes:
      - node-data:/data
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:3000/api/health || exit 1"]
      interval: 3s
      timeout: 5s
      retries: 10
      start_period: 10s
    networks:
      - e2e-net

  playwright:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      CI: "true"
      BASE_URL: http://panel:8080
      NODE_API_URL: http://node:3000
      SETUP_TOKEN: e2e-test-setup-token
    volumes:
      - ./playwright-report:/app/playwright-report
      - ./test-results:/app/test-results
    depends_on:
      panel:
        condition: service_healthy
      node:
        condition: service_healthy
    ipc: host
    networks:
      - e2e-net

volumes:
  panel-data:
  node-data:

networks:
  e2e-net:
    driver: bridge
```

### Playwright 测试运行器 Dockerfile

```dockerfile
# e2e/Dockerfile
FROM mcr.microsoft.com/playwright:v1.50.1-noble

WORKDIR /app

# 安装 bun
RUN npm install -g bun

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY playwright.config.ts ./
COPY tests/ ./tests/

# 默认运行全部测试；可通过 command 覆盖运行特定 project
CMD ["bunx", "playwright", "test"]
```

### Playwright 配置

```typescript
// e2e/playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

const isCI = !!process.env.CI;
const baseURL = process.env.BASE_URL || 'http://localhost:8080';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,          // 顺序执行，避免共享状态冲突
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
  workers: 1,                    // 单 worker，测试间有数据依赖
  timeout: 30_000,
  expect: { timeout: 10_000 },

  reporter: isCI
    ? [['dot'], ['html', { open: 'never', outputFolder: 'playwright-report' }]]
    : [['line'], ['html', { open: 'on-failure', outputFolder: 'playwright-report' }]],

  use: {
    baseURL,
    headless: true,
    trace: isCI ? 'on-first-retry' : 'off',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    actionTimeout: 10_000,
    navigationTimeout: 30_000,
    locale: 'zh-CN',
  },

  outputDir: 'test-results',

  projects: [
    {
      name: 'smoke',
      testMatch: /.*\.smoke\.spec\.ts/,
      retries: 0,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'e2e',
      testMatch: /tests\/e2e\/.*\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
```

### 核心 Fixtures 设计

**Auth Fixture**：封装 Bootstrap → Login → 保存 storageState

```typescript
// e2e/tests/fixtures/auth.fixture.ts
import { test as base, expect } from '@playwright/test';

type AuthFixtures = {
  authenticatedPage: Page;
  apiContext: APIRequestContext;
};

export const test = base.extend<AuthFixtures>({
  authenticatedPage: async ({ page, request }, use) => {
    const baseURL = process.env.BASE_URL || 'http://localhost:8080';
    const setupToken = process.env.SETUP_TOKEN || 'e2e-test-setup-token';

    // 检查是否需要 Bootstrap
    const status = await request.post(`${baseURL}/rpc/sboard.panel.v1.AuthService/GetBootstrapStatus`, { data: {} });
    const { data } = await status.json();

    if (data.needs_setup) {
      // 执行 Bootstrap
      await request.post(`${baseURL}/rpc/sboard.panel.v1.AuthService/Bootstrap`, {
        data: {
          username: 'admin',
          password: 'admin1234',
          confirm_password: 'admin1234',
        },
        headers: { 'X-Setup-Token': setupToken },
      });
    }

    // 登录获取 token
    const loginResp = await request.post(`${baseURL}/rpc/sboard.panel.v1.AuthService/Login`, {
      data: { username: 'admin', password: 'admin1234' },
    });
    const loginData = await loginResp.json();
    const token = loginData.data.token;

    // 注入 token 到 localStorage
    await page.goto(`${baseURL}/login`);
    await page.evaluate((t) => {
      localStorage.setItem('sboard_token', t);
    }, token);

    await page.goto(`${baseURL}/`);
    await use(page);
  },

  apiContext: async ({ playwright }, use) => {
    const baseURL = process.env.BASE_URL || 'http://localhost:8080';
    // ... 创建带认证的 API context
    await use(apiContext);
  },
});
```

**API Fixture**：调用 Panel RPC + Node REST API

```typescript
// e2e/tests/fixtures/api.fixture.ts
// 封装 Panel API 和 Node API 的调用
// - 创建节点（包括传入 node 容器的实际地址和 secret）
// - 触发同步
// - 查询 Node 健康状态和配置
// - 创建用户、分组、入站等
```

### 测试用例概要

#### Smoke 测试

| 文件 | 测试点 |
|------|--------|
| `health.smoke.spec.ts` | Panel `HealthService/GetHealth` 返回 200；Node `/api/health` 返回 200 |
| `bootstrap.smoke.spec.ts` | Bootstrap 流程 → 创建 admin → 登录成功 → 重定向到 Dashboard |
| `navigation.smoke.spec.ts` | 已登录状态下，所有核心路由可正常加载（/, /users, /nodes, /subscriptions 等） |

#### E2E 测试

| 文件 | 测试点 |
|------|--------|
| `auth.spec.ts` | 登录成功/失败、Token 过期重定向、登出 |
| `users.spec.ts` | 创建用户、编辑用户信息、删除用户、搜索/筛选 |
| `groups.spec.ts` | 创建分组、分配用户到分组、删除分组 |
| `nodes.spec.ts` | 创建节点（指向 Docker 中的 node 容器）、编辑、删除、查看状态 |
| `node-sync.spec.ts` | 创建入站 → 触发同步 → 验证 Node API 返回正确配置 → 验证 sing-box 状态 |
| `inbounds.spec.ts` | 入站配置 CRUD、关联节点 |
| `subscriptions.spec.ts` | 创建订阅、生成订阅链接、验证订阅内容格式 |
| `sync-jobs.spec.ts` | 同步任务列表、手动触发同步 |

### 运行方式

```bash
# 在 e2e/ 目录下

# 一键运行全部测试
docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright

# 仅运行 smoke 测试
docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=smoke

# 仅运行 e2e 测试
docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=e2e

# 清理环境
docker compose -f docker-compose.e2e.yml down -v

# 查看报告
bunx playwright show-report ./playwright-report
```

### 根目录 Makefile 集成

```makefile
# 添加到根 Makefile
.PHONY: e2e e2e-smoke e2e-down

e2e:
	cd e2e && docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright

e2e-smoke:
	cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
	docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=smoke

e2e-down:
	cd e2e && docker compose -f docker-compose.e2e.yml down -v
```

## 设计文档

- [BDD 规格说明](./bdd-specs.md) - 行为场景和测试策略
- [架构设计](./architecture.md) - 系统架构和组件详情
- [最佳实践](./best-practices.md) - 安全、性能和代码质量指南
