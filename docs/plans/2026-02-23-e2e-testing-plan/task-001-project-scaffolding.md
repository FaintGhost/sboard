# Task 001: 项目基础设施搭建

**depends-on**: (none)

## Description

创建 E2E 测试项目的完整基础设施：目录结构、Docker 编排文件、Playwright 配置、包管理配置。这是所有后续任务的基础。

## Execution Context

**Task Number**: 001 of 012
**Phase**: Setup
**Prerequisites**: 无

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: 基础设施（非特定 BDD 场景，为所有场景的运行前提）

## Files to Create

- `e2e/package.json` — Playwright 和相关依赖声明
- `e2e/playwright.config.ts` — Playwright 配置（smoke/e2e projects, CI 模式, 报告等）
- `e2e/Dockerfile` — Playwright 测试运行器镜像（基于官方 Playwright 镜像 + Bun）
- `e2e/docker-compose.e2e.yml` — 三容器编排（panel + node + playwright）
- `e2e/.gitignore` — 忽略 playwright-report/, test-results/, node_modules/
- `e2e/tests/` — 测试目录占位

## Files to Modify

- `Makefile` — 添加 e2e, e2e-smoke, e2e-down targets
- `.gitignore`（如需要）

## Steps

### Step 1: 创建 e2e/ 目录结构

创建以下目录结构：
```
e2e/
├── tests/
│   ├── smoke/
│   ├── e2e/
│   └── fixtures/
├── playwright-report/    (.gitignore)
└── test-results/         (.gitignore)
```

### Step 2: 创建 package.json

声明依赖：
- `@playwright/test` — 版本需与 Docker 镜像匹配（查看 `mcr.microsoft.com/playwright` 最新 noble 标签）
- TypeScript 相关类型

包管理器使用 Bun，与项目前端一致。

### Step 3: 创建 playwright.config.ts

按照设计文档中的配置创建，关键要素：
- `testDir: './tests'`
- 两个 projects: `smoke`（匹配 `*.smoke.spec.ts`）和 `e2e`（匹配 `tests/e2e/*.spec.ts`）
- CI 模式检测（`process.env.CI`）
- baseURL 从环境变量 `BASE_URL` 读取，默认 `http://localhost:8080`
- headless: true
- 单 worker，顺序执行（测试间有数据依赖）
- 失败时截图/trace/视频配置
- HTML 报告输出

### Step 4: 创建 Dockerfile

基于 `mcr.microsoft.com/playwright:v{version}-noble`：
- 安装 Bun
- COPY package.json + bun.lock → `bun install`
- COPY 测试代码
- 默认 CMD: `bunx playwright test`

### Step 5: 创建 docker-compose.e2e.yml

三个服务：
- `panel`: 从 `../panel/Dockerfile` 构建，设置测试环境变量（JWT_SECRET, SETUP_TOKEN, DB_PATH, SERVE_WEB 等）
- `node`: 从 `../node/Dockerfile` 构建，设置 NODE_SECRET_KEY
- `playwright`: 从当前 Dockerfile 构建，`depends_on` panel+node，`ipc: host`，volume mount 报告目录；通过 `entrypoint.sh` 等待 Panel RPC 健康检查（`POST /rpc/sboard.panel.v1.HealthService/GetHealth`）和 Node 健康检查（`GET /api/health`）

注意：优先复用 Playwright 侧 `entrypoint.sh` 的等待逻辑，避免将服务就绪判断分散到多个位置。

### Step 6: 创建 .gitignore

忽略：
- `node_modules/`
- `playwright-report/`
- `test-results/`
- `bun.lock` 或保留（取决于项目约定 — 检查 panel/web/ 的 bun.lock 是否在 git 中）

### Step 7: 安装依赖并生成 lock 文件

在 e2e/ 目录运行 `bun install` 生成 bun.lock。

### Step 8: 验证 Docker 构建

运行 `docker compose -f e2e/docker-compose.e2e.yml build` 确保三个镜像都能成功构建（不运行测试，只验证构建）。

## Verification Commands

```bash
# 验证目录结构
ls -la e2e/
ls -la e2e/tests/

# 验证依赖安装
cd e2e && bun install

# 验证 Docker 构建（仅构建，不运行）
cd e2e && docker compose -f docker-compose.e2e.yml build

# 验证 Playwright 配置可解析
cd e2e && bunx playwright test --list 2>&1 | head -5
```

## Success Criteria

- e2e/ 目录结构完整
- `bun install` 成功
- `docker compose build` 三个服务均构建成功
- Playwright 配置可被正确解析（无语法错误）
