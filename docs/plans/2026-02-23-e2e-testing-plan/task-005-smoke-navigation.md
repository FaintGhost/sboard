# Task 005: Smoke - 核心页面导航测试

**depends-on**: task-002

## Description

实现已登录状态下核心页面导航的 smoke 测试，验证所有关键路由可正常加载（无错误弹窗或白屏）。

## Execution Context

**Task Number**: 005 of 012
**Phase**: Core Features (Smoke)
**Prerequisites**: Task 002 完成（auth fixture 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "核心页面导航" — Scenario Outline "核心页面加载"（8 个路由）

## Files to Create

- `e2e/tests/smoke/navigation.smoke.spec.ts`

## Steps

### Step 1: 研究前端路由结构

查看 `panel/web/src/` 的路由配置，确认：
- 路由路径列表（/, /users, /groups, /nodes, /inbounds, /sync-jobs, /subscriptions, /settings）
- 每个页面的标题或可识别的关键元素（用于验证页面加载成功）
- 侧边栏导航结构

### Step 2: 实现导航测试

创建 `navigation.smoke.spec.ts`，使用 `authenticatedPage` fixture（自动完成 Bootstrap + Login）。

对每个核心路由实现参数化测试（Scenario Outline）：
- 导航到指定路由
- 验证页面正常加载（无白屏）：检查页面标题或关键 DOM 元素可见
- 验证无未捕获的 JavaScript 错误（监听 `page.on('pageerror')`）

测试的路由列表：`/`, `/users`, `/groups`, `/nodes`, `/inbounds`, `/sync-jobs`, `/subscriptions`, `/settings`

### Step 3: 在 Docker 环境中验证

运行 smoke project 中的 navigation 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
  docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=smoke tests/smoke/navigation.smoke.spec.ts
```

## Success Criteria

- 8 个路由页面全部加载成功
- 无 JavaScript 错误
- 测试运行时间 < 15 秒
