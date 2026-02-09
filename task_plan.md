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

---

# Session Plan: 节点地址联动 + Last Seen + 统计文案（2026-02-09）

## Goal
围绕节点管理与流量展示做一轮 UX 收敛：
- 节点编辑支持 API 地址与公网地址联动（默认同值，可手动解耦）
- 节点列表增加/强化 Last Seen 展示
- 流量展示文案从“最近采样”统一为“最近更新”，去掉采样语义

## Current Phase
Complete

## Phases

### Phase 1: 现状核对与方案落盘
- [x] 核对 `nodes-page`、`traffic` 类型与 i18n 文案
- [x] 明确不改后端模型，仅做前端交互与展示优化
- **Status:** complete

### Phase 2: 节点地址联动 + Last Seen 列
- [x] 增加“公网地址与 API 地址一致”联动开关
- [x] 创建节点默认联动并自动回填 `public_address`
- [x] 节点列表新增 `Last Seen` 列并保持表格间距一致
- **Status:** complete

### Phase 3: 流量文案统一
- [x] 将“最近采样时间/Last Sample”统一为“最近更新/Last Updated”
- [x] 排查同类文案，避免局部修复
- **Status:** complete

### Phase 4: 验证与记录
- [x] 运行前端相关测试与构建
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

---

# Session Plan: 订阅外部地址可配置（2026-02-09）

## Goal
在设置页提供可持久化的“订阅基础地址”配置（支持域名/IP），并让订阅管理页面生成正确的公网订阅链接，适配反向代理与公网部署。

## Current Phase
Complete

## Phases

### Phase 1: 后端配置能力
- [x] 新增 `system_settings` 数据表迁移
- [x] 新增 Store 读写接口（get/upsert/delete）
- [x] 新增 API：`GET/PUT /api/system/settings`
- [x] 增加 URL 校验与标准化（仅允许 http/https）
- **Status:** complete

### Phase 2: 前端设置页接入
- [x] 设置页新增“订阅访问地址”配置区块
- [x] 支持读取、编辑、保存与错误提示
- [x] 提供“当前用于生成订阅链接的地址”预览
- **Status:** complete

### Phase 3: 订阅页接入配置
- [x] 订阅链接生成从固定 `window.location.origin` 改为“配置优先、本地回退”
- [x] 兼容带 path 的 base URL（如反代子路径）
- **Status:** complete

### Phase 4: 验证与文档化
- [x] Go 后端测试通过
- [x] 前端全量测试通过
- [x] 前端构建通过
- [x] 更新 planning 文件
- **Status:** complete

---

# Session Plan: 订阅访问地址改为协议下拉 + IP端口校验（2026-02-09）

## Goal
将设置页“订阅访问地址”改为：
- 协议通过下拉选择（HTTP/HTTPS）
- 地址单独输入 `IP:端口`
- 前后端统一做合法性校验

## Current Phase
Complete

## Phases

### Phase 1: 后端约束收敛
- [x] `subscription_base_url` 校验改为 `http|https + IP:端口`
- [x] 限制非法项：无端口、非法端口、域名、路径/query/fragment
- [x] 补充 API 测试覆盖异常分支
- **Status:** complete

### Phase 2: 设置页交互重构
- [x] 新增协议下拉 `HTTP/HTTPS`
- [x] 新增 `IP + 端口` 输入
- [x] 新增前端实时校验（格式/IP/端口范围）
- [x] 预览地址按协议+输入拼接
- **Status:** complete

### Phase 3: 文案与验证
- [x] 更新中英文文案
- [x] 更新设置页测试与订阅页相关测试
- [x] 运行后端测试、前端全量测试与构建
- **Status:** complete

---

# Session Plan: 交互设计优化（2026-02-09）

## Goal
围绕交互一致性与体感稳定性做一轮前端打磨：
- 建立统一动效 token 与降级策略（`prefers-reduced-motion`）
- 统一表格筛选切换过渡表现，减少突兀感
- 为后续按钮/异步反馈统一协议打基础

## Current Phase
Complete

## Phases

### Phase 1: 规划与基线落盘
- [x] 将交互优化目标写入计划文件
- [x] 明确首批落地范围与验证口径
- **Status:** complete

### Phase 2: 全局动效能力
- [x] 在全局样式中新增 motion token（fast/base/slow + easing）
- [x] 增加 `prefers-reduced-motion` 降级策略
- **Status:** complete

### Phase 3: 列表切换过渡统一
- [x] 抽离可复用表格过渡 class 生成逻辑
- [x] 在 users/inbounds/subscriptions/sync-jobs 页面统一接入
- **Status:** complete

### Phase 4: 验证与记录
- [x] 运行前端全量测试
- [x] 运行前端构建
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

---

# Session Plan: 交互设计优化（二阶段：异步按钮统一，2026-02-09）

## Goal
统一高频异步操作按钮反馈，消除“0.x 秒文字闪变”的生硬感，并确保同类场景一次性覆盖。

## Current Phase
Complete

## Phases

### Phase 1: 规范与范围确认
- [x] 盘点页面异步按钮位点
- [x] 定义统一交互规范（延迟显示 + 最小可见时长）
- **Status:** complete

### Phase 2: 抽象复用组件
- [x] 新增 `AsyncButton` 组件
- [x] 内置 spinner 延迟显示与最小展示时长
- [x] 补充组件单测覆盖“短请求不闪/长请求平滑显示”
- **Status:** complete

### Phase 3: 同类页面批量接入
- [x] 接入 `settings/sync-jobs/users/groups/nodes/inbounds`
- [x] 接入 `login/bootstrap` 表单主提交按钮
- [x] 保持文案与禁用态一致
- **Status:** complete

### Phase 4: 严格验证与记录
- [x] 运行前端全量测试（全绿）
- [x] 运行前端构建（通过）
- [x] 同步更新 `findings.md` 与 `progress.md`
- **Status:** complete

---

# Session Plan: Frontend Design 收敛（2026-02-09）

## Goal
提升管理面板的视觉辨识度与信息架构清晰度，重点改造导航、Dashboard 首屏层级、Settings 主卡视觉与路由加载体验。

## Current Phase
Complete

## Phases

### Phase 1: 设计方向与信息架构
- [x] 确认方向：Network Ops Console（克制工业感）
- [x] 定义首批改造范围：Sidebar / Dashboard / Settings / Route Fallback
- **Status:** complete

### Phase 2: 导航与入口重排
- [x] 将订阅入口合并到主导航
- [x] 清理冗余导航分区，提升路径直觉性
- **Status:** complete

### Phase 3: Dashboard 层级重构
- [x] 用“系统概览 + 快捷动作”替代低价值卡片
- [x] 强化关键指标信息密度与可扫描性
- **Status:** complete

### Phase 4: Settings 视觉优先级优化
- [x] 强化订阅访问配置卡片层级与识别度
- [x] 保持交互逻辑不变，仅优化视觉组织
- **Status:** complete

### Phase 5: 验证与记录
- [x] 运行前端全量测试（必须全绿）
- [x] 运行前端构建
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

---

# Session Plan: Frontend Design 收敛（二阶段骨架统一，2026-02-09）

## Goal
统一页面头、表格工具栏与空状态表现，建立跨页面一致的视觉骨架。

## Current Phase
Complete

## Phases

### Phase 1: 抽象复用组件
- [x] 新增 `PageHeader`
- [x] 新增 `TableEmptyState`
- [x] 新增 `table-toolbar` 规范
- **Status:** complete

### Phase 2: 批量页面接入
- [x] 接入 `users/groups/subscriptions/inbounds/sync-jobs/nodes`
- [x] 统一 header、toolbar、no-data CTA
- **Status:** complete

### Phase 3: 验证与沉淀
- [x] 前端测试全绿
- [x] 前端构建通过
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete
