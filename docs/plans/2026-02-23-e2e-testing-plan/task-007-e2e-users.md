# Task 007: E2E - 用户管理 CRUD 测试

**depends-on**: task-002

## Description

实现用户管理的完整 CRUD E2E 测试：创建用户、编辑用户信息、删除用户。

## Execution Context

**Task Number**: 007 of 012
**Phase**: Core Features (E2E)
**Prerequisites**: Task 002 完成（fixtures 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "用户管理" — 全部 3 个场景

## Files to Create

- `e2e/tests/e2e/users.spec.ts`

## Steps

### Step 1: 研究用户管理 UI

查看 Panel 前端的用户管理页面：
- 用户列表页面结构（表格/卡片、列名、操作按钮位置）
- "创建用户"按钮的位置和标签
- 用户创建表单的字段（用户名、密码、流量限制、过期时间、分组等）
- 编辑用户的交互方式（点击编辑按钮 → 弹窗/抽屉/跳转）
- 删除确认的交互方式（确认对话框）
- 查阅 `panel/proto/sboard/panel/v1/panel.proto` 中 Users 相关 RPC 了解字段定义

### Step 2: 实现用户管理测试

创建 `users.spec.ts`，使用 `authenticatedPage` fixture。使用 `test.describe.serial` 确保顺序执行。

1. **创建新用户**：导航到用户管理页面，点击创建用户按钮，填写用户信息表单（使用唯一名称前缀），提交后验证用户列表中出现新用户

2. **编辑用户信息**：找到刚创建的用户，点击编辑按钮，修改某个字段（如备注/流量限制），保存后验证列表中显示更新后的信息

3. **删除用户**：找到测试用户，点击删除按钮，确认删除，验证用户从列表中消失

### Step 3: 在 Docker 环境中验证

运行 e2e project 中的 users 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build -d panel node && \
  docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=e2e tests/e2e/users.spec.ts
```

## Success Criteria

- 3 个用户管理场景测试全部通过
- 创建的用户在列表中可见
- 编辑后的信息正确更新
- 删除后用户从列表消失
