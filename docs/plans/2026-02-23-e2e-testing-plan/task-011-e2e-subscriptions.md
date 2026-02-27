# Task 011: E2E - 订阅管理测试

**depends-on**: task-002

## Description

实现订阅管理的 E2E 测试：生成订阅链接、验证订阅内容格式。

## Execution Context

**Task Number**: 011 of 012
**Phase**: Core Features (E2E)
**Prerequisites**: Task 002 完成（fixtures 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "订阅管理" — 全部 2 个场景

## Files to Create

- `e2e/tests/e2e/subscriptions.spec.ts`

## Steps

### Step 1: 研究订阅功能

查看 Panel 的订阅相关实现：
- 查阅 `panel/proto/sboard/panel/v1/panel.proto` 与 `panel/internal/api/subscription.go` 中订阅相关 RPC/REST 边界
- 前端订阅页面的 UI 结构（订阅列表、用户关联、链接展示方式）
- 订阅链接的格式和生成方式
- 订阅内容格式：sing-box JSON 和 v2ray base64

### Step 2: 实现订阅管理测试

创建 `subscriptions.spec.ts`，使用 `authenticatedPage` + `panelAPI` fixtures。

**前置 setup**（`beforeAll`）：通过 API 快速创建：
- 一个分组
- 一个用户（关联到分组）
- 一个节点（指向 Docker 内的 node 容器）
- 一个入站配置（关联到节点）
- 触发同步确保配置生效

测试场景：

1. **生成订阅链接**：在 UI 中查看用户的订阅信息，验证能看到订阅链接

2. **验证订阅内容**：获取订阅链接后，通过 Playwright `request` API 直接访问订阅链接，验证：
   - 返回有效的代理配置内容
   - 配置格式正确（sing-box JSON 能被解析，或 v2ray base64 能被正确解码）
   - 配置内容包含预期的入站信息（协议类型、服务器地址等）

### Step 3: 在 Docker 环境中验证

运行 e2e project 中的 subscriptions 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
  docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=e2e tests/e2e/subscriptions.spec.ts
```

## Success Criteria

- 2 个订阅管理场景测试全部通过
- 订阅链接可正常生成
- 订阅内容格式正确且包含预期的配置信息
