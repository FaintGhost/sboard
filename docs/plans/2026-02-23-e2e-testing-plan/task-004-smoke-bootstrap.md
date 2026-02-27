# Task 004: Smoke - Bootstrap 初始化测试

**depends-on**: task-002

## Description

实现 Bootstrap 初始化流程的 smoke 测试：首次访问显示 Bootstrap 页面、成功完成初始化创建管理员、登录已初始化系统。

## Execution Context

**Task Number**: 004 of 012
**Phase**: Core Features (Smoke)
**Prerequisites**: Task 002 完成（auth fixture 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "Bootstrap 初始化" — 全部 3 个场景

## Files to Create

- `e2e/tests/smoke/bootstrap.smoke.spec.ts`

## Steps

### Step 1: 研究前端 Bootstrap 流程

在实现测试之前，需要理解 Panel 前端的 Bootstrap 流程：
- 查看 `panel/web/src/` 中的路由配置，找到 Bootstrap/Setup 相关的页面组件
- 确认 Bootstrap 表单的字段和提交方式
- 确认 Bootstrap 成功后的跳转逻辑

### Step 2: 实现 Bootstrap 测试

创建 `bootstrap.smoke.spec.ts`，包含以下测试（注意：这些测试应在全新数据库状态下运行）：

1. **首次访问显示 Bootstrap 页面**：全新状态下访问首页，验证重定向到登录页并显示初始化设置相关 UI 元素

2. **成功完成 Bootstrap**：在 Bootstrap 表单中填入管理员信息（用户名、密码、确认密码），提交后验证管理员账户创建成功并自动登录跳转到 Dashboard

3. **登录已初始化的系统**：Bootstrap 完成后，访问登录页，输入正确凭据，验证登录成功并跳转到 Dashboard

注意：由于测试间共享数据库状态，测试顺序很重要。Scenario 1 和 2 需要全新数据库，Scenario 3 需要已完成 Bootstrap 的状态。使用 `test.describe.serial` 确保顺序执行。

### Step 3: 在 Docker 环境中验证

运行 smoke project 中的 bootstrap 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
  docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=smoke tests/smoke/bootstrap.smoke.spec.ts
```

## Success Criteria

- 3 个 Bootstrap 场景测试全部通过
- Bootstrap 流程在 UI 层面可完整走通
