# Task Plan: Code Quality Improvements (Code Review Fixes)

## Goal
根据深度代码审查结果，逐一修复高优先级问题，提升代码质量、可维护性和符合 Go/React best practices。

## Current Phase
Complete

## Phases

### Phase 1: 清理未使用的模板代码
- [x] 删除或重构未使用的 `data-table.tsx`（805行模板代码）
- **Status:** complete

### Phase 2: 拆分过大的 React 组件
- [x] 将 `users-page.tsx`（835行）拆分为多个子组件
  - [x] 提取 `EditUserDialog` 组件
  - [x] 提取 `DisableUserDialog` 组件
  - [x] 提取 `DeleteUserDialog` 组件
  - [x] 创建 `useUserMutations` hook
  - [x] 创建公共类型文件 `types.ts`
- **Status:** complete
- **Notes:** 主页面从 835 行减少到 410 行，组件拆分到 `pages/users/` 目录

### Phase 3: Go 后端改进
- [x] 修复 N+1 查询问题（UsersList 批量获取 group_ids）
  - 新增 `ListUserGroupIDsBatch` 批量查询方法
  - `UsersList` 改用单次批量查询代替循环查询
- [x] 为 Node 添加优雅关闭（signal handling）
  - 添加 `SIGINT/SIGTERM` 信号处理
  - 实现 HTTP server 优雅关闭
  - 添加 Core.Close() 方法释放 sing-box 资源
- **Status:** complete

### Phase 4: 添加 Error Boundary
- [x] 添加全局 Error Boundary 组件 (`components/error-boundary.tsx`)
- [x] 在 AppProviders 中集成
- [x] 添加中英文错误页面文案
- **Status:** complete

### Phase 5: 验证与测试
- [x] 运行所有测试确保无回归
  - Go 测试：全部通过
  - 前端测试：24 个测试全部通过
- [x] 执行前端构建验证：成功
- [x] 更新 planning 文件
- **Status:** complete

## Key Issues from Code Review

### High Priority (已修复)
1. ~~**未使用的模板代码** - `data-table.tsx` 包含示例数据和未定制的列定义~~
2. ~~**组件过于庞大** - `users-page.tsx` 混合了 6 个 mutation、多个 Dialog、表格渲染~~
3. ~~**N+1 查询** - `UsersList` 循环中逐个查询 user groups~~

### Medium Priority (已修复)
4. ~~**缺少 Error Boundary** - 无全局错误边界捕获渲染错误~~
5. ~~**Node 缺少优雅关闭** - 无 signal 处理和 graceful shutdown~~

### Low Priority (暂不处理)
- Service 层抽象
- 结构化日志
- 配置项硬编码问题

## Decisions Made
| Decision | Rationale |
|----------|-----------|
| 优先处理前端问题 | 改动风险低，收益明显 |
| N+1 修复采用批量查询 | 最小改动，保持 API 兼容 |
| 保留现有架构 | 避免大规模重构 |

## Files Changed

### Frontend
- **Deleted:** `panel/web/src/components/data-table.tsx`
- **Modified:** `panel/web/src/pages/users-page.tsx` (835 -> 410 行)
- **Created:** `panel/web/src/pages/users/` 目录
  - `types.ts`
  - `use-user-mutations.ts`
  - `edit-user-dialog.tsx`
  - `disable-user-dialog.tsx`
  - `delete-user-dialog.tsx`
  - `index.ts`
- **Created:** `panel/web/src/components/error-boundary.tsx`
- **Modified:** `panel/web/src/providers/app-providers.tsx`
- **Modified:** `panel/web/src/i18n/locales/zh.json`, `en.json`

### Backend
- **Modified:** `panel/internal/db/user_groups.go` (添加 ListUserGroupIDsBatch)
- **Modified:** `panel/internal/api/users.go` (UsersList 使用批量查询)
- **Modified:** `node/cmd/node/main.go` (graceful shutdown)
- **Modified:** `node/internal/core/core.go` (添加 Close 方法)

## Notes
- 遵循"问题修复协作规则"：不只修复单一页面，应主动排查同类实现
- 优先抽象为可复用模块
