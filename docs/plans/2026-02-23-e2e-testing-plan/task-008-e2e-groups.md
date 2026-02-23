# Task 008: E2E - 分组管理测试

**depends-on**: task-002

## Description

实现分组管理的 E2E 测试：创建分组、将用户分配到分组。

## Execution Context

**Task Number**: 008 of 012
**Phase**: Core Features (E2E)
**Prerequisites**: Task 002 完成（fixtures 可用）

## BDD Scenario Reference

**Spec**: `../2026-02-23-e2e-testing-design/bdd-specs.md`
**Scenario**: Feature "分组管理" — 全部 2 个场景

## Files to Create

- `e2e/tests/e2e/groups.spec.ts`

## Steps

### Step 1: 研究分组管理 UI

查看 Panel 前端的分组管理页面：
- 分组列表/管理页面结构
- 创建分组的交互方式和表单字段
- 查阅 `panel/openapi.yaml` 中 Groups 相关的 API
- 用户编辑时如何选择分组（下拉选择框？）

### Step 2: 实现分组管理测试

创建 `groups.spec.ts`，使用 `authenticatedPage` + `panelAPI` fixtures。使用 `test.describe.serial`。

1. **创建分组**：导航到分组管理页面，点击创建分组按钮，填写分组名称（唯一前缀），提交后验证分组创建成功

2. **将用户分配到分组**：先通过 API 创建一个测试用户（快速 setup），然后在 UI 中编辑该用户，选择刚创建的分组，保存后验证用户的分组信息更新成功

### Step 3: 在 Docker 环境中验证

运行 e2e project 中的 groups 测试。

## Verification Commands

```bash
cd e2e && docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright \
  -- bunx playwright test --project=e2e tests/e2e/groups.spec.ts
```

## Success Criteria

- 2 个分组管理场景测试全部通过
- 分组创建成功
- 用户成功分配到分组
