# Task 003: Smoke - 健康检查测试

**depends-on**: task-002

## Description

实现 Panel 和 Node 的 API 健康检查 smoke 测试。这是最基础的冒烟测试，验证两个服务都已启动并正常响应。

## Execution Context

**Task Number**: 003 of 012
**Phase**: Core Features (Smoke)
**Prerequisites**: Task 002 完成（fixtures 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "系统健康检查" — Scenario "Panel 健康检查" + Scenario "Node 健康检查"

## Files to Create

- `e2e/tests/smoke/health.smoke.spec.ts`

## Steps

### Step 1: 实现健康检查测试

创建 `health.smoke.spec.ts`，包含两个测试：

1. **Panel 健康检查**：发送 GET 到 `${BASE_URL}/api/health`，验证返回 200 且响应体包含 `status: "ok"`
2. **Node 健康检查**：发送 GET 到 `${NODE_API_URL}/api/health`，验证返回 200 且响应体包含 `status: "ok"`

这两个测试使用 Playwright 的 `request` API 直接发送 HTTP 请求（不需要浏览器）。

### Step 2: 在 Docker 环境中验证

运行 smoke project 验证测试通过。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright \
  -- bunx playwright test --project=smoke tests/smoke/health.smoke.spec.ts
```

## Success Criteria

- 两个健康检查测试均通过
- 测试运行时间 < 5 秒
