# Task 006: E2E - 认证管理测试

**depends-on**: task-002

## Description

实现完整的认证流程 E2E 测试：登录成功、登录失败（错误密码）、未认证访问受保护页面时重定向。

## Execution Context

**Task Number**: 006 of 012
**Phase**: Core Features (E2E)
**Prerequisites**: Task 002 完成（fixtures 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "认证管理" — 全部 3 个场景

## Files to Create

- `e2e/tests/e2e/auth.spec.ts`

## Steps

### Step 1: 研究前端登录流程

查看 Panel 前端的登录页面组件：
- 登录表单结构（用户名/密码输入框、登录按钮的定位方式）
- 错误提示信息的显示方式和文本内容
- 登录成功后的跳转目标
- 未认证访问受保护路由时的重定向逻辑

### Step 2: 实现认证测试

创建 `auth.spec.ts`，包含以下测试：

1. **登录成功**：使用 API 完成 Bootstrap 后，在登录页输入正确用户名密码，点击登录按钮，验证跳转到 Dashboard 且侧边栏导航可见

2. **登录失败 - 错误密码**：输入正确用户名和错误密码，点击登录按钮，验证显示错误提示信息且仍停留在登录页

3. **未认证访问受保护页面**：不登录状态下直接访问 `/users`，验证被重定向到登录页

注意：测试 1 和 2 需要已完成 Bootstrap 的系统。使用 `test.beforeAll` 通过 API 确保 Bootstrap 完成。

### Step 3: 在 Docker 环境中验证

运行 e2e project 中的 auth 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright \
  -- bunx playwright test --project=e2e tests/e2e/auth.spec.ts
```

## Success Criteria

- 3 个认证场景测试全部通过
- 登录成功场景正确验证跳转和导航可见性
- 登录失败场景正确验证错误提示
- 未认证访问场景正确验证重定向
