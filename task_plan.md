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

---

# Session Plan: Theme 配色升级（2026-02-09）

## Goal
为 shadcn/ui 面板建立更有辨识度的主题色系统，摆脱黑白单调感，同时保持可读性与一致性。

## Current Phase
Complete

## Phases

### Phase 1: 主题定义
- [x] 确定方向：Cobalt Operations
- [x] 定义 light/dark 主轴与语义色
- **Status:** complete

### Phase 2: Token 替换
- [x] 更新全局 color tokens（含 chart/sidebar）
- [x] 保持组件逻辑不变，仅升级视觉系统
- **Status:** complete

### Phase 3: 验证与记录
- [x] 前端测试全绿
- [x] 前端构建通过
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

---

# Session Plan: 节点删除功能回归修复（2026-02-09）

## Goal
恢复“节点管理”页删除节点能力，并补充回归测试，防止再次遗漏。

## Current Phase
Complete

## Phases

### Phase 1: 根因定位
- [x] 定位节点删除入口缺失位置
- [x] 确认历史提交中误删 delete 逻辑
- **Status:** complete

### Phase 2: TDD 修复
- [x] 新增节点删除失败用例
- [x] 恢复删除 mutation 与菜单入口
- [x] 补齐错误反馈
- **Status:** complete

### Phase 3: 同类排查与验证
- [x] 排查 users/groups/inbounds/nodes 删除能力一致性
- [x] 前端测试全绿（13 files, 30 tests）
- [x] 前端构建通过
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

---

# Session Plan: Settings 增加超级管理员凭证管理（2026-02-10）

## Goal
在设置页支持“修改超级管理员账号与密码”，并强制要求输入旧密码校验。

## Current Phase
Complete

## Phases

### Phase 1: 后端能力补齐
- [x] 新增管理员资料查询接口
- [x] 新增管理员凭证更新接口（旧密码必填）
- [x] 增加后端测试覆盖
- **Status:** complete

### Phase 2: 前端设置页接入
- [x] 新增管理员凭证卡片与表单
- [x] 接入管理员 profile 查询与更新 API
- [x] 增加前端测试覆盖
- **Status:** complete

### Phase 3: 严格验证与记录
- [x] `panel` 后端全量测试通过
- [x] `panel/web` 前端全量测试通过
- [x] `panel/web` 构建通过
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

---

# Session Plan: 仪表盘流量趋势图空数据回退修复（2026-02-10）

## Goal
修复流量趋势图在接口返回空数据时仍展示旧折线的问题，确保图形与 hover 数值一致。

## Current Phase
Complete

## Phases

### Phase 1: 复现与根因
- [x] 复现“无流量但折线仍上扬”
- [x] 定位到 chart 数据回退逻辑导致的陈旧数据复用
- **Status:** complete

### Phase 2: TDD 修复
- [x] 新增图表数据选择逻辑测试
- [x] 修复空数组时回退旧数据的逻辑
- [x] 接入图表组件
- **Status:** complete

### Phase 3: 验证与记录
- [x] 前端全量测试通过
- [x] 前端构建通过
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

---

# Session Plan: 仪表盘流量趋势“上扬但 0GB”修复（2026-02-10）

## Goal
修复趋势图视觉与 tooltip 数值不一致问题，避免小流量被错误感知为逻辑异常。

## Current Phase
Complete

## Phases

### Phase 1: 根因定位
- [x] 确认空数据回退逻辑已修复
- [x] 识别固定 GB 单位导致的“0GB 错觉”
- [x] 识别平滑曲线造成的视觉放大
- **Status:** complete

### Phase 2: 实施修复
- [x] 图表改为自适应单位显示
- [x] 线型由 monotone 改为 linear
- [x] 补充单位工具与测试
- **Status:** complete

### Phase 3: 验证与记录
- [x] 前端全量测试通过
- [x] 前端构建通过
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

---

# Session Plan: 仪表盘趋势图平滑曲线优化（2026-02-10）

## Goal
在不牺牲数据准确性的前提下，将趋势图恢复为更自然的平滑曲线。

## Current Phase
Complete

## Phases

### Phase 1: 影响面确认
- [x] 全局检索图表线型配置
- [x] 确认仅 dashboard 趋势图受影响
- **Status:** complete

### Phase 2: 最小改动实施
- [x] 将 upload/download 曲线改为 `monotoneX`
- [x] 保持单位与数据回退逻辑不变
- **Status:** complete

### Phase 3: 严格验证与记录
- [x] 前端全量测试通过
- [x] 前端构建通过
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

---

# Session Plan: Tooltip 统一为 shadcn/ui + 图表图例 Badge 化（2026-02-10）

## Goal
统一 hover 提示交互为 `shadcn/ui Tooltip`，并提升图表图例视觉一致性（使用 `shadcn/ui Badge`）。

## Current Phase
Complete

## Phases

### Phase 1: 提示交互收口
- [x] 排查页面与组件中原生 `title` 提示
- [x] 优先替换可见交互点为 `Tooltip`
- [x] 清理重复/冲突的 `title` 属性
- **Status:** complete

### Phase 2: 图表图例统一
- [x] 图例文案切换到 i18n
- [x] download 曲线颜色提升对比
- [x] 图例改用 `Badge` + 色点表达
- **Status:** complete

### Phase 3: 验证与记录
- [x] `bun run format`
- [x] `bun run lint`
- [x] `bun run build`
- [x] 更新计划与进度文件
- **Status:** complete

---

# Session Plan: Frontend Design 第二轮（语义色彩收口）（2026-02-10）

## Goal
统一关键页面的颜色表达为主题语义 token，减少硬编码颜色导致的视觉不一致，提升暗色模式与后续主题演进稳定性。

## Current Phase
Complete

## Phases

### Phase 1: 设计基线确认
- [x] 识别关键页面硬编码颜色（`slate/amber/red`）
- [x] 明确替换策略：优先 `foreground/muted/destructive/accent`
- **Status:** complete

### Phase 2: 批量落地
- [x] 统一 `FieldHint` 触发器色彩语义
- [x] 覆盖 `subscriptions/groups/nodes/inbounds/users/settings`
- [x] 状态点组件移除 `red/slate` 硬编码为语义 token
- **Status:** complete

### Phase 3: 验证与记录
- [x] `bun run format`
- [x] `bun run lint`
- [x] `bun run build`
- [x] 更新 `findings.md` / `progress.md`
- **Status:** complete

### Phase 3（补充）: 严格测试闭环
- [x] 补跑前端全量测试（45 tests）
- [x] 修复测试暴露的 provider/交互断言问题
- **Status:** complete

---

# Session Plan: 后端测试覆盖率提升（P0 第一批）（2026-02-10）

## Goal
针对后端低覆盖包先做第一批高收益补测，优先覆盖 `panel/internal/traffic` 与 `node/internal/stats`，建立可持续提升基线。

## Current Phase
Complete

## Phases

### Phase 1: 覆盖率基线复测
- [x] 复跑 panel/node 包级覆盖率
- [x] 识别 P0 包与可快速落地场景
- **Status:** complete

### Phase 2: P0 补测落地
- [x] 新增 `panel/internal/traffic` provider 行为测试
- [x] 新增 `node/internal/stats` tracker 与系统采样测试
- [x] 覆盖参数校验、聚合逻辑、reset 行为、连接计数
- **Status:** complete

### Phase 3: 覆盖率回归验证
- [x] 定向测试通过（traffic/stats）
- [x] 全量 `go test ./... -cover` 通过（panel/node）
- [x] 输出覆盖率提升结果
- **Status:** complete

### Phase 4: P0 第二批扩展（inbounds/password）
- [x] 新增 `panel/internal/inbounds` 测试并覆盖协议校验分支
- [x] 新增 `panel/internal/password` 测试并覆盖 hash/verify 分支
- [x] 复跑 panel 覆盖率并确认提升
- **Status:** complete

### Phase 5: P0 第三批（config）
- [x] 新增 `panel/internal/config` 测试（默认值/环境覆盖/异常回退）
- [x] 新增 `node/internal/config` 测试（默认值/环境覆盖）
- [x] 复跑 panel/node 覆盖率并确认提升
- **Status:** complete

### Phase 6: P0 第四批（node/singboxcli）
- [x] 新增 `panel/internal/node/client` 测试并覆盖 HTTP 客户端关键分支
- [x] 新增 `panel/internal/singboxcli` 额外测试并覆盖 Check/Generate 分支
- [x] 复跑 panel/node 及后端合并覆盖率
- **Status:** complete

## 2026-02-10 Session: 后端覆盖率提升（P1）

### Goal
围绕 `panel/internal/db` 与 `node/internal/api` 的低覆盖分支补齐测试，提升后端总覆盖率并沉淀可回归用例。

### Phases

#### Phase 1: 热点确认
- [x] 确认 `panel/internal/db` 低覆盖热点（`system_settings/traffic_aggregate/traffic_stats`）
- [x] 确认 `node/internal/api` 低覆盖热点（`ConfigSync/StatsInboundsGet/StatsTrafficGet/shouldDebugNodeSyncPayload`）
- **Status:** complete

#### Phase 2: Panel DB 补测
- [x] 新增 `system_settings` CRUD 测试
- [x] 新增 `traffic_aggregate` 聚合/过滤/timeseries 测试
- [x] 新增 `traffic_stats` 校验/插入/查询/用户流量累计测试
- **Status:** complete

#### Phase 3: Node API 补测
- [x] 新增 `ConfigSync` 全分支测试（400/500/200）
- [x] 新增 `StatsInboundsGet` 鉴权/reset/meta 分支测试
- [x] 新增 `StatsTrafficGet` 鉴权/env fallback/query override/success 分支测试
- [x] 新增 `shouldDebugNodeSyncPayload` 环境变量判定测试
- **Status:** complete

#### Phase 4: 覆盖率复测与回写
- [x] 复跑 `panel`/`node` 全量 Go 测试与覆盖率
- [x] 合并后端覆盖率并记录到计划文件
- **Status:** complete

## 2026-02-10 Session: 后端覆盖率提升（P2）

### Goal
继续补齐 `panel/internal/db` 低覆盖分支（admins/group_users/groups/nodes/sync_jobs/users）并拉高后端合并覆盖率。

### Phases

#### Phase 1: 低覆盖扫描
- [x] 基于 `panel/internal/db` 重新扫描 0% 函数热点
- **Status:** complete

#### Phase 2: 补测实施
- [x] 新增 `admins` 全链路测试
- [x] 新增 `group_users + groups` 关键路径测试
- [x] 新增 `user_groups batch + user delete` 测试
- [x] 新增 `nodes` 更新与在线状态测试
- [x] 新增 `sync_jobs` 列表过滤测试
- **Status:** complete

#### Phase 3: 全量复测与记录
- [x] 复跑 `panel`/`node` Go 全量测试与覆盖率
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

## 2026-02-10 Session: 后端覆盖率提升（P3）

### Goal
优先拉升 `panel/internal/api` 覆盖率，补齐 bootstrap/traffic/helper 与多条 handler 校验分支。

### Phases

#### Phase 1: 热点扫描
- [x] 扫描 `panel/internal/api` 低覆盖函数（重点 `inbounds/nodes/groups/traffic/bootstrap/helpers`）
- **Status:** complete

#### Phase 2: 补测实施
- [x] 新增 `helpers` 纯函数测试（window/bool/debug/sync error/token）
- [x] 新增 `bootstrap` 额外分支测试（store 缺失、header token、重复初始化）
- [x] 新增 `nodes/groups/inbounds` 校验分支测试
- [x] 新增 `traffic/timeseries` 参数校验与成功分支测试
- **Status:** complete

#### Phase 3: 全量复测与记录
- [x] 复跑 `panel`/`node` 全量 Go 测试与覆盖率
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

## 2026-02-10 Session: 后端覆盖率提升（P4）

### Goal
提升 `node/internal/core` 覆盖率，补齐 `Apply/InboundTraffic/Close` 等核心分支。

### Phases

#### Phase 1: 热点扫描
- [x] 扫描 `node/internal/core` 低覆盖函数
- **Status:** complete

#### Phase 2: 补测实施
- [x] 新增 `Core.Apply` 分支测试
- [x] 新增 `ApplyOptions` 工厂错误与 start 错误分支测试
- [x] 新增 `InboundTraffic/InboundTrafficMeta/Close` 分支测试
- **Status:** complete

#### Phase 3: 全量复测与记录
- [x] 复跑 `panel`/`node` 全量 Go 测试与覆盖率
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

## 2026-02-10 Session: 后端覆盖率提升（P5）

### Goal
提升 `node/internal/state` 与 `node/internal/sync` 覆盖率，补齐持久化与配置解析错误分支。

### Phases

#### Phase 1: 热点扫描
- [x] 扫描 `state/sync` 低覆盖函数
- **Status:** complete

#### Phase 2: 补测实施
- [x] 新增 `state` 边界与错误分支测试
- [x] 新增 `sync parse` BadRequest 分支测试
- [x] 补充 `sync parse` context canceled 兼容测试
- **Status:** complete

#### Phase 3: 全量复测与记录
- [x] 复跑 `panel`/`node` 全量 Go 测试与覆盖率
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

## 2026-02-10 Session: 后端覆盖率提升（P6）

### Goal
把后端合并覆盖率推进到 `70%+`，优先补 `panel/internal/subscription` 与 `panel/internal/monitor` 的低覆盖函数。

### Phases

#### Phase 1: 热点扫描
- [x] 扫描 `subscription/monitor` 的低覆盖函数
- **Status:** complete

#### Phase 2: 补测实施
- [x] 新增 `v2ray` 多协议构建测试（vmess/trojan/ss）
- [x] 新增 `v2ray` 校验与 unknown 协议跳过测试
- [x] 新增 `traffic monitor Run` 行为测试（0 interval、cancel 停止）
- **Status:** complete

#### Phase 3: 全量复测与记录
- [x] 复跑 `panel`/`node` 全量 Go 测试与覆盖率
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

---

# Session Plan: 后端健壮性与模块化整改（2026-02-12）

## Goal
根据本轮后端审查结果，修复影响正确性与可维护性的关键问题，并补齐回归测试，确保行为可验证。

## Current Phase
Complete

## Phases

### Phase 1: 基线与问题固化
- [x] 运行后端测试基线并确认通过
- [x] 固化问题清单与优先级
- **Status:** complete

### Phase 2: 用户删除与同步一致性
- [x] 修复 `hard delete` 用户后未触发节点同步的问题
- [x] 增加回归测试（覆盖有分组绑定的硬删除）
- **Status:** complete

### Phase 3: 用户状态筛选分页正确性
- [x] 修复“先分页后筛选”导致的结果偏差
- [x] 增加回归测试（小 limit + offset 场景）
- **Status:** complete

### Phase 4: 节点 group_id 可置空更新
- [x] 修复 `PUT /api/nodes/:id` 无法设置 `group_id = null`
- [x] 增加回归测试（置空与非法值）
- **Status:** complete

### Phase 5: 回归验证与文档更新
- [x] 运行受影响后端测试集
- [x] 更新 `findings.md` 与 `progress.md`
- **Status:** complete

### Phase 6: Node Sync 请求体防护
- [x] 为 `POST /api/config/sync` 增加 body size 上限
- [x] 增加 413 回归测试
- **Status:** complete

### Phase 7: 同步编排模块拆分
- [x] 拆分 `node_sync_helpers.go`，将运行时辅助函数迁移到独立文件
- [x] 保持行为不变并通过全量回归
- **Status:** complete

## Risks
- `users status` 由“存储状态 + 运行时状态”组合计算，分页修复需兼顾正确性与性能。
- `group_id` 字段的 JSON 绑定改动可能影响现有更新分支，必须补测试兜底。

## Decisions Made
| Decision | Rationale |
|----------|-----------|
| 先修复 correctness 再做结构重构 | 先消除行为风险，控制改动面 |
| 使用增量扫描方式修复状态筛选分页 | 保持现有接口语义，同时确保分页正确 |
| 通过字段级 presence 解析支持 group_id 置空 | 避免大规模改 DTO 绑定模型 |

### Phase 8: Users 状态筛选性能优化（增量）
- [x] 将 `status=disabled|expired` 下沉为 DB 侧有效状态过滤分页
- [x] 保留 `status=active|traffic_exceeded` 的增量扫描路径（避免流量重置语义偏差）
- [x] 增加 expired 分页回归测试
- **Status:** complete

### Phase 9: Users 状态筛选全量下沉（active/traffic_exceeded）
- [x] 新增 `ListUsersByEffectiveStatus` 对 `active/traffic_exceeded` 的 DB 分页支持
- [x] 引入查询前流量重置归一化，保障语义一致
- [x] 增加 active 与 traffic_exceeded（含重置边界）回归测试
- **Status:** complete

---

# Session Plan: 后端分页参数模块化与上限防护（2026-02-12）

## Goal
统一后端列表接口的分页参数解析，消除跨模块散落实现并加上 `limit` 上限防护，提升健壮性与可维护性。

## Current Phase
Complete

## Phases

### Phase 1: 现状核查
- [x] 确认 `users/groups/nodes/inbounds/sync-jobs` 共用分页解析逻辑但定义在 `users.go`
- [x] 确认现有实现缺少 `limit` 上限，存在单次大查询风险
- **Status:** complete

### Phase 2: 模块化抽取
- [x] 新增独立请求参数模块 `request_params.go`
- [x] 抽取 `parseID` 与 `parseLimitOffset`
- **Status:** complete

### Phase 3: 健壮性增强
- [x] 为通用分页解析增加 `maxListLimit=500` 上限约束
- [x] 保持现有默认值与错误语义（`invalid pagination`）
- **Status:** complete

### Phase 4: 回归验证
- [x] 新增跨接口分页边界测试（超限拒绝 + 边界值通过）
- [x] 运行后端相关测试
- **Status:** complete

---

# Session Plan: Traffic Handler Store Guard 一致性修复（2026-02-12）

## Goal
统一 traffic 相关 API 的 `store` 就绪检查，避免 `store=nil` 时行为不一致（panic/recovery 或错误信息漂移）。

## Current Phase
Complete

## Phases

### Phase 1: 同类问题排查
- [x] 排查所有使用 `*db.Store` 的 traffic handler
- [x] 识别 `NodeTrafficList` 与 traffic aggregate 三个接口缺少 `ensureStore`
- **Status:** complete

### Phase 2: 统一修复
- [x] 在 `NodeTrafficList` 增加 `ensureStore`
- [x] 在 `TrafficNodesSummary/TrafficTotalSummary/TrafficTimeseries` 增加 `ensureStore`
- **Status:** complete

### Phase 3: 回归验证
- [x] 新增 nil store 场景测试覆盖 4 个 traffic 接口
- [x] 运行后端测试通过
- **Status:** complete

### Phase 4: 流量明细接口参数解析统一（增量）
- [x] `GET /api/nodes/:id/traffic` 改为复用 `parseID + parseLimitOffset`
- [x] 移除该接口独立分页解析分支，统一错误语义为 `invalid pagination`
- [x] 补分页边界测试（非法值拒绝、500 上限通过）
- **Status:** complete

---

# Session Plan: 前后端分页上限对齐（2026-02-12）

## Goal
消除前端 `limit=1000` 与后端 `maxListLimit=500` 的不兼容，确保页面功能不回退（尤其全量成员/节点场景）。

## Current Phase
Complete

## Phases

### Phase 1: 影响面核查
- [x] 定位前端仍使用 `limit=1000` 的调用点
- [x] 评估功能影响（群组成员编辑、仪表盘节点概览）
- **Status:** complete

### Phase 2: 对齐实现
- [x] 新增分页聚合工具 `listAllByPage`（按 500 分批拉取）
- [x] 新增 `listAllUsers` / `listAllNodes` API
- [x] 群组页与仪表盘改为使用全量聚合 API
- **Status:** complete

### Phase 3: 验证
- [x] 新增 API 分页聚合测试（users/nodes）
- [x] 回归群组页面测试
- [x] 前端构建通过
- **Status:** complete

---

# Session Plan: Sync Jobs 空态错误跳转修复（2026-02-12）

## Goal
修复 `sync-jobs` 页面空态按钮文案/跳转不一致导致的交互混乱（"查看同步任务" 按钮却跳回节点页）。

## Current Phase
Complete

## Phases

### Phase 1: 根因定位
- [x] 确认节点页入口为 `/sync-jobs?node_id=<id>`，逻辑正常
- [x] 确认问题在 `sync-jobs-page` 空态按钮错误复用 `nodes.viewSyncJobs`
- **Status:** complete

### Phase 2: 修复
- [x] 移除 `sync-jobs` 空态的误导跳转按钮
- [x] 保留空态文案 `暂无数据`
- **Status:** complete

### Phase 3: 验证
- [x] 更新并通过 `sync-jobs-page` 回归测试
- **Status:** complete
