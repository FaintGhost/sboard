# Progress Log

## Session: 2026-02-09

### Phase 1: M5 范围确认与设计落盘
- **Status:** in_progress
- **Started:** 2026-02-09
- Actions taken:
  - 切换到 M5 任务并重建文件化计划。
  - 基于现有提交识别 M1/M2/M3 基线能力。
  - 定义 M5 为“同步任务可用性增强”。
- Files created/modified:
  - `task_plan.md` (updated)
  - `findings.md` (updated)
  - `progress.md` (updated)

### Phase 2: TDD 先行（核心逻辑）
- **Status:** pending
- Actions taken:
  -
- Files created/modified:
  -

## Test Results
| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| 暂无 | - | - | - | - |

## Error Log
| Timestamp | Error | Attempt | Resolution |
|-----------|-------|---------|------------|
| 2026-02-09 | 暂无 | 1 | - |

## 5-Question Reboot Check
| Question | Answer |
|----------|--------|
| Where am I? | Phase 1（M5 范围确认） |
| Where am I going? | TDD -> 实现 M5 -> 验证交付 |
| What's the goal? | 完成同步任务页可用性增强与节点联动 |
| What have I learned? | M4 未显式定义，需按已完成链路自然推进 |
| What have I done? | 已完成 M4 计划落盘与范围界定 |

---
*Update after completing each phase or encountering errors*

### Phase 2: TDD 先行（核心逻辑）
- **Status:** complete
- Actions taken:
  - 新增 `sync-jobs-filters.test.ts` 测试，先执行失败（缺文件）。
  - 实现 `sync-jobs-filters.ts` 后再执行测试变绿。
- Files created/modified:
  - `panel/web/src/lib/sync-jobs-filters.test.ts` (created)
  - `panel/web/src/lib/sync-jobs-filters.ts` (created)

### Phase 3: 实现 M5 功能
- **Status:** complete
- Actions taken:
  - `sync-jobs-page` 接入 query filter + source 筛选 + 分页控件。
  - `nodes-page` 增加“查看同步任务”入口并带 `node_id` query。
  - 增补中英文文案。
- Files created/modified:
  - `panel/web/src/pages/sync-jobs-page.tsx` (modified)
  - `panel/web/src/pages/nodes-page.tsx` (modified)
  - `panel/web/src/i18n/locales/zh.json` (modified)
  - `panel/web/src/i18n/locales/en.json` (modified)

## Test Results
| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| Filters TDD 红灯 | `npm test -- src/lib/sync-jobs-filters.test.ts`（首次） | 失败（缺实现） | 失败（符合预期） | ✓ |
| Filters TDD 绿灯 | `npm test -- src/lib/sync-jobs-filters.test.ts` | 通过 | 通过 | ✓ |
| 前端构建 | `npm run build` | 构建成功 | 通过（仅 chunk size 警告） | ✓ |

## 5-Question Reboot Check
| Question | Answer |
|----------|--------|
| Where am I? | Phase 4（验证与交付） |
| Where am I going? | 提交 M5 增强并继续后续路线 |
| What's the goal? | 完成同步任务页可用性增强与节点联动 |
| What have I learned? | query 同步 + 后端分页参数组合可最小改动实现高收益 |
| What have I done? | 已完成 TDD、实现、构建验证 |

### Phase 4: 验证与交付
- **Status:** complete
- Actions taken:
  - 完成代码提交：`2424c32 feat(panel-web): enhance sync jobs filters, pagination and node linking`。
  - 更新 `task_plan.md`，标记 M5 全阶段完成。
- Files created/modified:
  - `task_plan.md` (updated)
  - `progress.md` (updated)

## Session: 2026-02-09 (Hotfix - 筛选切换闪回)

### Phase 1: 复现与根因确认
- **Status:** complete
- Actions taken:
  - 根据用户补充确认问题在筛选切换后的表格刷新体验。
  - 新建针对 `sync-jobs` 的交互复现路径，定位到缓存行短暂回显。
- Files created/modified:
  - `panel/web/src/pages/sync-jobs-page.test.tsx` (created)

### Phase 2: TDD 失败用例
- **Status:** complete
- Actions taken:
  - 编写“失败 -> 全部 -> 失败”切换场景测试。
  - 验证在未修复前，测试确实失败（能看到旧行）。
- Files created/modified:
  - `panel/web/src/pages/sync-jobs-page.test.tsx` (modified)

### Phase 3: 最小修复实现
- **Status:** complete
- Actions taken:
  - `sync-jobs-page` 增加筛选键变更检测与切换保护态。
  - 保护态期间不渲染 `jobsQuery` 缓存行，仅展示 loading 骨架。
  - 请求稳定后恢复新数据渲染。
- Files created/modified:
  - `panel/web/src/pages/sync-jobs-page.tsx` (modified)

### Phase 4: 回归与交付
- **Status:** complete
- Actions taken:
  - 运行定向测试与相关测试集合。
  - 执行 `npm run build` 构建通过。
  - 更新 planning 文件落盘。
- Files created/modified:
  - `task_plan.md` (updated)
  - `findings.md` (updated)
  - `progress.md` (updated)

## Test Results (Hotfix)
| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| 新增回归测试（红灯） | `npm test -- src/pages/sync-jobs-page.test.tsx`（修复前） | 失败（出现旧行） | 失败（符合预期） | ✓ |
| 新增回归测试（绿灯） | `npm test -- src/pages/sync-jobs-page.test.tsx` | 通过 | 通过 | ✓ |
| 相关测试集合 | `npm test -- src/lib/sync-jobs-filters.test.ts src/pages/sync-jobs-page.test.tsx` | 通过 | 通过 | ✓ |
| 前端构建 | `npm run build` | 通过 | 通过（含 chunk size 提示） | ✓ |

### Hotfix Follow-up: 来源筛选空态闪烁
- **Status:** complete
- Actions taken:
  - 按用户复现路径新增测试：全部来源(有数据) -> 手动节点同步(空) -> 手动重试(空)。
  - 先让测试失败，确认“切换中暂无数据被骨架替换”问题。
  - 引入 `lastSettledRowCount`，实现空态切换时稳定展示“暂无数据”。
  - 测试与构建全部通过。
- Files created/modified:
  - `panel/web/src/pages/sync-jobs-page.tsx` (modified)
  - `panel/web/src/pages/sync-jobs-page.test.tsx` (modified)
  - `findings.md` (updated)
  - `progress.md` (updated)

## Test Results (Hotfix Follow-up)
| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| 来源空态切换（红灯） | `npm test -- src/pages/sync-jobs-page.test.tsx`（修复前） | 失败（暂无数据被替换） | 失败（符合预期） | ✓ |
| 来源空态切换（绿灯） | `npm test -- src/pages/sync-jobs-page.test.tsx` | 通过 | 通过 | ✓ |
| 相关测试集合 | `npm test -- src/lib/sync-jobs-filters.test.ts src/pages/sync-jobs-page.test.tsx` | 通过 | 通过 | ✓ |
| 前端构建 | `npm run build` | 通过 | 通过（含 chunk size 提示） | ✓ |

### Hotfix Follow-up 2: 修复 isLoading 导致的空态闪烁
- **Status:** complete
- Actions taken:
  - 调整 `showSkeletonLoading` 逻辑，排除筛选切换态下的裸 `isLoading` 骨架。
  - 补充测试断言：空态切换后不出现 `加载中...`。
  - 重新运行测试与构建验证。
- Files created/modified:
  - `panel/web/src/pages/sync-jobs-page.tsx` (modified)
  - `panel/web/src/pages/sync-jobs-page.test.tsx` (modified)
  - `findings.md` (updated)
  - `progress.md` (updated)

### Hotfix Follow-up 3: 全筛选表格统一防闪
- **Status:** complete
- Actions taken:
  - 新增通用 hook：`useTableQueryTransition`。
  - 替换四个页面的筛选切换渲染逻辑，去除单页分叉实现。
  - 新增 hook 单测并通过；原有 `users/sync-jobs` 回归测试通过。
  - 前端构建通过。
- Files created/modified:
  - `panel/web/src/lib/table-query-transition.ts` (created)
  - `panel/web/src/lib/table-query-transition.test.ts` (created)
  - `panel/web/src/pages/sync-jobs-page.tsx` (modified)
  - `panel/web/src/pages/users-page.tsx` (modified)
  - `panel/web/src/pages/inbounds-page.tsx` (modified)
  - `panel/web/src/pages/subscriptions-page.tsx` (modified)
  - `panel/web/src/pages/sync-jobs-page.test.tsx` (modified)
  - `findings.md` (updated)
  - `progress.md` (updated)

## Test Results (Hotfix Follow-up 3)
| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| 通用 hook 单测 | `npm test -- src/lib/table-query-transition.test.ts` | 通过 | 通过 | ✓ |
| 同步任务回归 | `npm test -- src/pages/sync-jobs-page.test.tsx` | 通过 | 通过 | ✓ |
| 用户页回归 | `npm test -- src/pages/users-page.test.tsx` | 通过 | 通过 | ✓ |
| 前端构建 | `npm run build` | 通过 | 通过（含 chunk size 提示） | ✓ |

### Feature: 入站编辑改为 JSON 模板
- **Status:** complete
- Actions taken:
  - 新增 `inbound-template` 解析/构建工具与测试（先红后绿）。
  - `inbounds-page` 编辑弹窗改为“节点选择 + 模板 JSON”模式。
  - 新增中英文文案：模板说明、users 自动注入提示、模板解析错误提示。
  - 验证测试与构建通过。
- Files created/modified:
  - `panel/web/src/lib/inbound-template.ts` (created)
  - `panel/web/src/lib/inbound-template.test.ts` (created)
  - `panel/web/src/pages/inbounds-page.tsx` (modified)
  - `panel/web/src/i18n/locales/zh.json` (modified)
  - `panel/web/src/i18n/locales/en.json` (modified)

### Phase 6: Inbounds Monaco + sing-box 工具化（2026-02-09）
- **Status:** complete
- Actions taken:
  - 入站编辑器从 `Textarea` 切换为 `Monaco Editor`（JSON 高亮、自动布局、2 空格缩进配置）。
  - 入站模板支持预置协议快速切换：`vless/vmess/trojan/shadowsocks` + `custom`。
  - 新增后端 `sing-box` 工具 API：`/api/sing-box/format`、`/api/sing-box/check`、`/api/sing-box/generate`。
  - 前端入站编辑弹窗接入工具按钮：格式化、检查、生成（UUID/各类 keypair）。
  - 补充后端 API 测试与前端模板工具测试，完成构建验证。
- Files created/modified:
  - `panel/internal/singboxcli/service.go` (created)
  - `panel/internal/api/singbox_tools.go` (created)
  - `panel/internal/api/singbox_tools_test.go` (created)
  - `panel/internal/api/router.go` (modified)
  - `panel/internal/config/config.go` (modified)
  - `panel/web/src/pages/inbounds-page.tsx` (modified)
  - `panel/web/src/lib/api/singbox-tools.ts` (created)
  - `panel/web/src/lib/api/types.ts` (modified)
  - `panel/web/src/lib/inbound-template.test.ts` (modified)
  - `panel/web/src/i18n/locales/zh.json` (modified)
  - `panel/web/src/i18n/locales/en.json` (modified)

### Phase 7: Inbounds 模板语义修正 + 内嵌 sing-box 工具化（2026-02-09）
- **Status:** complete
- Actions taken:
  - 修正入站模板语义：模板不再展示/输出 `public_port`（避免与 sing-box 原生字段混淆）。
  - `sing-box format/check/generate` 改为 Panel 内嵌实现，不依赖容器内外部 `sing-box` 二进制。
  - 后端 wrapper 在 inbound 模式下会剔除 `public_port` 再进行格式化/校验。
  - 生成命令收敛为无需额外参数的常用项：`uuid/reality-keypair/wg-keypair/vapid-keypair`。
- Files created/modified:
  - `panel/internal/singboxcli/service.go` (modified)
  - `panel/internal/api/singbox_tools.go` (modified)
  - `panel/internal/api/singbox_tools_test.go` (modified)
  - `panel/internal/config/config.go` (modified)
  - `panel/internal/api/router.go` (modified)
  - `panel/web/src/lib/inbound-template.ts` (modified)
  - `panel/web/src/lib/inbound-template.test.ts` (modified)
  - `panel/web/src/lib/api/types.ts` (modified)
  - `panel/web/src/pages/inbounds-page.tsx` (modified)
  - `panel/go.mod` (modified)

### Phase 8: 严格检查 + 检查按钮交互优化（2026-02-09）
- **Status:** complete
- Actions taken:
  - `singboxcli.Check` 升级为严格运行态检查：反序列化后执行 `box.New(...)` 初始化并立即释放。
  - 入站编辑“检查配置”按钮文案固定，不再在极短请求中闪烁切换为“加载中”。
- Files created/modified:
  - `panel/internal/singboxcli/service.go` (modified)
  - `panel/web/src/pages/inbounds-page.tsx` (modified)

### Phase 9: SS2022 Base64 密钥一键生成并回填（2026-02-09）
- **Status:** complete
- Actions taken:
  - 新增生成命令：`rand-base64-16` 与 `rand-base64-32`（对应 SS2022 128/256 key length）。
  - 前端生成后自动回填当前模板 `password` 字段，减少手工复制粘贴错误。
  - 新增后端单测验证生成长度与非法参数。
- Files created/modified:
  - `panel/internal/singboxcli/service.go` (modified)
  - `panel/internal/singboxcli/service_test.go` (created)
  - `panel/web/src/lib/api/types.ts` (modified)
  - `panel/web/src/pages/inbounds-page.tsx` (modified)
  - `panel/web/src/i18n/locales/zh.json` (modified)
  - `panel/web/src/i18n/locales/en.json` (modified)

### Phase 10: 全配置模板链路打通（Panel + Node，2026-02-09）
- **Status:** complete
- Actions taken:
  - 完成“完整 sing-box 配置模板”端到端改造：前端模板、Panel payload、Node 解析与应用链路统一。
  - `node` 改为解析完整 `option.Options` 并通过 `ApplyOptions` 重建 box，确保 `route/outbounds/dns` 真正生效。
  - 修复生成回填在完整配置下写入路径（优先 `inbounds[0].password`）。
  - 工具链在完整配置模式也统一剔除 `public_port` 元字段，避免误检。
  - 更新中英文入站模板提示文案为“完整配置”语义。
- Files created/modified:
  - `panel/web/src/lib/inbound-template.ts` (modified)
  - `panel/web/src/lib/inbound-template.test.ts` (modified)
  - `panel/web/src/pages/inbounds-page.tsx` (modified)
  - `panel/web/src/i18n/locales/zh.json` (modified)
  - `panel/web/src/i18n/locales/en.json` (modified)
  - `panel/internal/node/build_config.go` (modified)
  - `panel/internal/node/build_config_test.go` (modified)
  - `panel/internal/api/singbox_tools.go` (modified)
  - `panel/internal/api/singbox_tools_test.go` (modified)
  - `node/internal/sync/parse.go` (modified)
  - `node/internal/sync/parse_test.go` (modified)
  - `node/internal/core/core.go` (modified)
  - `node/cmd/node/main.go` (modified)
  - `task_plan.md` / `findings.md` / `progress.md` (updated)

## Test Results (Phase 10)
| Test | Input | Expected | Actual | Status |
|------|-------|----------|--------|--------|
| Panel + Node 定向 Go 测试 | `go test ./panel/internal/node ./panel/internal/api ./panel/internal/singboxcli ./node/internal/sync ./node/internal/core -count=1` | 全部通过 | 通过 | ✓ |
| Node 启动包编译检查 | `go test ./node/cmd/node -count=1` | 通过 | 通过（无测试文件） | ✓ |
| 前端模板单测 | `npm test -- src/lib/inbound-template.test.ts` | 通过 | 通过 | ✓ |
| 前端构建 | `npm run build` | 通过 | 通过（含 chunk size 提示） | ✓ |

## 2026-02-09 Session: Node UX 收敛（进行中）
- 已完成：
  - 读取并确认会话恢复信息与当前 git 状态。
  - 核对节点页、traffic API 类型、i18n 文案落点。
  - 在 `task_plan.md/findings.md/progress.md` 建立本轮计划与发现。
- 下一步：
  - 实现地址联动 + Last Seen 列。
  - 统一流量文案为“最近更新”。

## 2026-02-09 Session: Node UX 收敛（已实现，待回归收敛）
- 已完成改动：
  - `panel/web/src/pages/nodes-page.tsx`
    - 新增地址联动状态 `linkAddress` 与 UI 开关。
    - 创建/编辑流程支持 API/公网地址联动与解耦。
    - 节点表格新增 `Last Seen` 列，并接入时间格式化显示。
    - 节点流量相关文案切换为 `nodes.lastUpdatedAt`。
  - `panel/web/src/lib/table-spacing.ts`
    - 新增 `seven` 列间距规则，节点表按 7 列统一。
  - `panel/web/src/i18n/locales/zh.json`
  - `panel/web/src/i18n/locales/en.json`
    - 新增 `nodes.lastUpdatedAt` / `nodes.lastSeen` / `nodes.sameAsApiAddress`。
- 验证结果：
  - `npm run build`：通过。
  - `npm test -- src/lib/table-query-transition.test.ts src/pages/sync-jobs-page.test.tsx src/pages/users-page.test.tsx src/lib/inbound-template.test.ts`：失败（3 项，属于既有筛选过渡回归点）。
- 结论：
  - 本轮目标功能已落地。
  - 需单开一次回归修复，把 `table-query-transition` 与 `sync-jobs-page` 的失败用例重新拉齐到绿色基线。


## 2026-02-09 Session: 无骨架过渡修复（已完成）
- 需求变更：用户要求“不要骨架”。
- 实施：
  - 重写 `panel/web/src/lib/table-query-transition.ts`，去除筛选切换骨架展示；
  - 调整 `panel/web/src/lib/table-query-transition.test.ts`，断言改为无骨架 + 空态稳定。
- 验证（严格）：
  - 定向：`npm test -- src/lib/table-query-transition.test.ts src/pages/sync-jobs-page.test.tsx src/pages/users-page.test.tsx src/lib/inbound-template.test.ts` 通过。
  - 全量：`npm test` 通过（10 files, 24 tests）。
  - 构建：`npm run build` 通过。

## 2026-02-09 Session: 订阅外部地址配置（已完成）
- 后端改动：
  - 新增迁移：
    - `panel/internal/db/migrations/0009_system_settings.up.sql`
    - `panel/internal/db/migrations/0009_system_settings.down.sql`
  - 新增 DB 接口：`panel/internal/db/system_settings.go`
  - 新增 API：`panel/internal/api/system_settings.go`
  - 路由接入：`panel/internal/api/router.go`
  - 新增测试：`panel/internal/api/system_settings_test.go`
- 前端改动：
  - API 封装：`panel/web/src/lib/api/system.ts`
  - 类型定义：`panel/web/src/lib/api/types.ts`
  - 设置页：`panel/web/src/pages/settings-page.tsx`
  - 订阅页：`panel/web/src/pages/subscriptions-page.tsx`
  - i18n：`panel/web/src/i18n/locales/zh.json`、`panel/web/src/i18n/locales/en.json`
  - 新增测试：`panel/web/src/pages/subscriptions-page.test.tsx`
  - 更新测试：`panel/web/src/pages/settings-page.test.tsx`
- 验证结果（严格）：
  - `go test ./panel/internal/db ./panel/internal/api -count=1` ✅
  - `npm test`（11 files, 26 tests）✅
  - `npm run build` ✅

## 2026-02-09 Session: 订阅访问地址输入模型重构（已完成）
- 后端：
  - `panel/internal/api/system_settings.go`
    - URL 校验改为协议 + IP:端口严格校验。
  - `panel/internal/api/system_settings_test.go`
    - 新增并扩展非法输入测试（scheme/domain/port/path）。
- 前端：
  - `panel/web/src/pages/settings-page.tsx`
    - 协议下拉 + IP:端口输入 + 本地校验 + 预览。
  - `panel/web/src/pages/settings-page.test.tsx`
    - 覆盖解析展示、保存值、非法输入不提交。
  - i18n 文案：`panel/web/src/i18n/locales/zh.json`、`panel/web/src/i18n/locales/en.json`
- 严格验证：
  - `go test ./panel/internal/db ./panel/internal/api -count=1` ✅
  - `npm test`（11 files, 27 tests）✅
  - `npm run build` ✅

## 2026-02-09 Session: 设置页布局与联动展示优化（已完成）
- 改动文件：`panel/web/src/pages/settings-page.tsx`
  - “订阅访问地址”卡片取消 `md:col-span-2`，与“订阅格式说明”同排显示。
  - “系统信息”卡移除 API 端点注释式描述。
  - API 端点展示从独立块改为单行字段，且值改为 `resolvedSubscriptionBaseURL` 实时联动。
- 验证结果：
  - 定向：`npm test -- src/pages/settings-page.test.tsx src/pages/subscriptions-page.test.tsx` ✅
  - 全量：`npm test`（11 files, 27 tests）✅
  - 构建：`npm run build` ✅

## 2026-02-09 Session: 设置页卡片顺序与排版微调（已完成）
- 修改：`panel/web/src/pages/settings-page.tsx`
  - 卡片顺序改为：订阅访问地址、系统信息、语言、订阅格式说明。
  - API 端点保持单行展示并联动 `resolvedSubscriptionBaseURL`。
- 验证：
  - `npm test -- src/pages/settings-page.test.tsx` ✅

## 2026-02-09 Session: Settings UX 交互打磨（已完成）
- 修改文件：
  - `panel/web/src/pages/settings-page.tsx`
    - 订阅协议+地址输入布局改为稳定双列栅格，减少横向空洞。
    - 保存流程增加最小加载时长（500ms）+ spinner，避免按钮文案闪变。
    - 地址输入支持 `onBlur` 校验，提交失败自动聚焦到输入框。
    - 预览地址新增复制按钮（复制成功图标态 + 状态文案反馈）。
    - 提示区收敛为 `aria-live`，错误提示增加 `role=alert`。
    - 系统信息字段命名改为“当前订阅基础地址”。
  - `panel/web/src/i18n/locales/zh.json`
  - `panel/web/src/i18n/locales/en.json`
    - 新增 `settings.subscriptionBaseUrlLabel`。
- 验证结果（严格）：
  - `npm test -- --run` ✅（11 files, 27 tests 全绿）
  - `npm run build` ✅

## 2026-02-09 Session: Web Design Guidelines 严格审查（待修复记录）
- 状态：已完成审查，未开始修复。
- 动作：
  - 基于 `web-design-guidelines` 对前端页面进行高标准审视。
  - 输出按 `file:line` 定位的问题清单，并按优先级分组。
- 输出：
  - 可访问性高优先级问题（图标按钮命名、输入 label、表格行点击可达性）。
  - UX/信息架构中优先级问题（筛选 URL 同步、Dialog 描述完整性）。
  - 性能与一致性问题（路由懒加载、Monaco 延迟加载、省略号文案规范）。
- 备注：
  - 本阶段只记录，不改动功能行为。
  - 已在 `findings.md` 增补完整待修复清单，等待你确认进入修复。

## 2026-02-09 Session: 严格审查清单第一轮修复（进行中）
- 状态：进行中（已完成第一批高优先级和共性问题）。
- 已完成：
  - `subscriptions/users/groups` 页输入与图标按钮可访问性补齐（`aria-label`/显式 label）。
  - `users/nodes/inbounds/groups` 编辑弹窗 create 模式统一补 `DialogDescription`。
  - `users/subscriptions` 筛选状态与搜索关键字写入 URL；新增公共工具 `lib/user-list-filters.ts`。
  - 对应测试适配（`MemoryRouter` 包装）并通过。
- 验证结果：
  - 定向：`npm test -- src/pages/subscriptions-page.test.tsx src/pages/users-page.test.tsx src/pages/sync-jobs-page.test.tsx src/pages/settings-page.test.tsx` ✅
  - 全量：`npm test -- --run` ✅（11 files, 27 tests）
- 未完成：
  - 无。

## 2026-02-09 Session: 严格审查清单第二轮修复（已完成）
- 完成项：
  - `sync-jobs`：详情入口改为独立操作列按钮（去除整行伪按钮语义）。
  - `routes`：页面路由切换为 `lazy + Suspense`，降低首屏初始加载压力。
  - `inbounds`：Monaco 编辑器切换为按需加载。
  - i18n：中英文省略号统一为 `…`。
- 文件：
  - `panel/web/src/pages/sync-jobs-page.tsx`
  - `panel/web/src/routes/index.tsx`
  - `panel/web/src/pages/inbounds-page.tsx`
  - `panel/web/src/i18n/locales/zh.json`
  - `panel/web/src/i18n/locales/en.json`
- 验证：
  - `npm test -- --run` ✅（11 files, 27 tests）
  - `npm run build` ✅

## 2026-02-09 Session: Interaction Design 第一阶段实现（已完成）
- 完成项：
  - 全局交互动效变量落地（时长/缓动统一）。
  - 全局 `prefers-reduced-motion` 降级规则接入。
  - 表格筛选切换过渡抽象为通用工具并统一接入 4 个页面。
- 变更文件：
  - `panel/web/src/index.css`
  - `panel/web/src/lib/table-motion.ts`
  - `panel/web/src/pages/users-page.tsx`
  - `panel/web/src/pages/inbounds-page.tsx`
  - `panel/web/src/pages/subscriptions-page.tsx`
  - `panel/web/src/pages/sync-jobs-page.tsx`
- 验证：
  - `npm test -- --run` ✅（11 files, 27 tests）
  - `npm run build` ✅

## 2026-02-09 Session: Interaction Design 第二阶段（异步按钮统一，已完成）
- 完成项：
  - 新增通用组件：`panel/web/src/components/ui/async-button.tsx`
    - 统一异步按钮反馈策略：`pendingDelayMs` + `pendingMinMs`。
    - 保留禁用态一致性，并通过 `aria-busy` 提升可访问性。
  - 新增单测：`panel/web/src/components/ui/async-button.test.tsx`
    - 覆盖短请求不闪文案、长请求平滑显示两类场景。
  - 同类页面批量接入：
    - `panel/web/src/pages/settings-page.tsx`
    - `panel/web/src/pages/sync-jobs-page.tsx`
    - `panel/web/src/pages/users/edit-user-dialog.tsx`
    - `panel/web/src/pages/users/disable-user-dialog.tsx`
    - `panel/web/src/pages/users/delete-user-dialog.tsx`
    - `panel/web/src/pages/groups-page.tsx`
    - `panel/web/src/pages/nodes-page.tsx`
    - `panel/web/src/pages/inbounds-page.tsx`
    - `panel/web/src/components/login-form.tsx`
    - `panel/web/src/components/bootstrap-form.tsx`
- 严格验证：
  - `npm test -- --run` ✅（12 files, 29 tests）
  - `npm run build` ✅

## 2026-02-09 Session: Frontend Design 收敛第一轮（已完成）
- 计划执行：
  - 已写入 `task_plan.md` 并按计划完成首批页面改造。
- 实施内容：
  - 导航重排：`subscriptions` 合并到主导航。
    - `panel/web/src/components/app-sidebar.tsx`
  - Dashboard 重构为“系统概览 + 快捷入口 + TOP 节点”。
    - `panel/web/src/pages/dashboard-page.tsx`
  - Settings 主卡视觉强化（不改业务逻辑）。
    - `panel/web/src/pages/settings-page.tsx`
  - 路由级 fallback 升级为结构化 skeleton。
    - `panel/web/src/routes/index.tsx`
  - 补充 i18n 文案。
    - `panel/web/src/i18n/locales/zh.json`
    - `panel/web/src/i18n/locales/en.json`
- 验证结果（严格）：
  - `npm test -- --run` ✅（12 files, 29 tests）
  - `npm run build` ✅

## 2026-02-09 Session: Frontend Design 第二轮（已完成）
- 新增复用能力：
  - `panel/web/src/components/page-header.tsx`
  - `panel/web/src/components/table-empty-state.tsx`
  - `panel/web/src/lib/table-toolbar.ts`
- 批量改造页面：
  - `panel/web/src/pages/users-page.tsx`
  - `panel/web/src/pages/groups-page.tsx`
  - `panel/web/src/pages/subscriptions-page.tsx`
  - `panel/web/src/pages/inbounds-page.tsx`
  - `panel/web/src/pages/sync-jobs-page.tsx`
  - `panel/web/src/pages/nodes-page.tsx`
- 验证结果：
  - `npm test -- --run` ✅（12 files, 29 tests）
  - `npm run build` ✅

## 2026-02-09 Session: 主题配色系统（已完成）
- 修改文件：
  - `panel/web/src/index.css`
- 完成内容：
  - 替换 light/dark 核心 token：`background/foreground/card/popover/primary/secondary/muted/accent/border/input/ring`
  - 同步调整图表色：`chart-1..5`
  - 同步调整侧栏色：`sidebar-*`
- 验证结果：
  - `npm test -- --run` ✅（12 files, 29 tests）
  - `npm run build` ✅

## 2026-02-09 Session: 节点删除功能回归修复（已完成）
- 完成项：
  - 恢复节点页删除 API 引入与 mutation：
    - `panel/web/src/pages/nodes-page.tsx`
  - 恢复节点行操作菜单的 destructive 删除项：
    - `panel/web/src/pages/nodes-page.tsx`
  - 新增节点删除回归测试：
    - `panel/web/src/pages/nodes-page.test.tsx`
  - 同类排查删除能力：
    - `panel/web/src/pages/users-page.tsx`
    - `panel/web/src/pages/groups-page.tsx`
    - `panel/web/src/pages/inbounds-page.tsx`
    - `panel/web/src/pages/nodes-page.tsx`
- 严格验证：
  - `npm test -- --run` ✅（13 files, 30 tests）
  - `npm run build` ✅

## 2026-02-10 Session: Settings 超级管理员凭证管理（已完成）
- 新增后端接口：
  - `panel/internal/api/admin_profile.go`
  - `panel/internal/api/router.go`（新增 `/api/admin/profile` GET/PUT）
- 新增/扩展 DB 能力：
  - `panel/internal/db/admins.go`
    - `AdminGetByID`
    - `AdminGetFirst`
    - `AdminUpdateCredentials`
- 测试覆盖：
  - `panel/internal/api/auth_test.go`（新增 profile 获取与更新、旧密码校验测试）
  - `panel/web/src/pages/settings-page.test.tsx`（新增管理员凭证更新测试）
- 前端接入：
  - `panel/web/src/lib/api/types.ts`
  - `panel/web/src/lib/api/system.ts`
  - `panel/web/src/pages/settings-page.tsx`
  - `panel/web/src/i18n/locales/zh.json`
  - `panel/web/src/i18n/locales/en.json`
- 严格验证：
  - `panel` 后端全量测试 ✅
  - `panel/web` 前端全量测试 ✅
  - `panel/web` 构建 ✅

## 2026-02-10 Session: 仪表盘流量趋势图空数据回退修复（已完成）
- 新增：
  - `panel/web/src/lib/traffic-chart-data.ts`
  - `panel/web/src/lib/traffic-chart-data.test.ts`
- 修改：
  - `panel/web/src/components/chart-area-interactive.tsx`
    - 空数组结果不再回退旧折线；
    - 仅在 query 数据为 `undefined`（加载中）时保留旧数据。
- 严格验证：
  - `cd panel/web && npm test -- --run` ✅（14 files, 34 tests）
  - `cd panel/web && npm run build` ✅

## 2026-02-10 Session: 仪表盘流量趋势单位与曲线感知修复（已完成）
- 变更：
  - `panel/web/src/components/chart-area-interactive.tsx`
    - 图表 Y 轴与 tooltip 改为自适应单位显示。
    - 线型从 `monotone` 调整为 `linear`。
  - `panel/web/src/lib/units.ts`
    - 新增 `pickByteUnit` / `formatBytesWithUnit`。
  - `panel/web/src/lib/units.test.ts`
    - 新增单位选择与格式化测试。
- 既有回归保持：
  - `panel/web/src/lib/traffic-chart-data.ts` 与其测试继续生效，确保空数组不回退旧折线。
- 验证：
  - `npm test -- --run` ✅
  - `npm run build` ✅

## 2026-02-10 Session: 仪表盘趋势图平滑曲线优化（已完成）
- 修改文件：
  - `panel/web/src/components/chart-area-interactive.tsx`
- 完成内容：
  - 将 upload/download 曲线类型由 `linear` 改为 `monotoneX`，恢复平滑观感；
  - 保留自适应单位显示与空数据回退修复，确保视觉与数值一致。
- 严格验证：
  - `cd panel/web && npm test -- --run` ✅（15 files, 39 tests）
  - `cd panel/web && npm run build` ✅

## 2026-02-10 Session: Settings 通用设置卡 + 时区输入匹配（已完成）
- 完成内容：
  - 将时区从“订阅访问地址”卡移出，合并到“通用设置”卡（语言 + 时区）。
  - 引入成熟时区库 `@vvo/tzdb`，支持时区输入联想匹配（datalist）与手动输入。
  - 保存时区时做 IANA 校验，非法值前端即时报错并阻断提交。
  - 订阅设置与通用设置拆分为独立保存反馈，减少状态串扰。
- 修改文件：
  - panel/web/src/pages/settings-page.tsx
  - panel/web/src/i18n/locales/zh.json
  - panel/web/src/i18n/locales/en.json
  - panel/web/src/pages/settings-page.test.tsx
  - panel/web/src/pages/settings-timezone.test.tsx
  - panel/web/package.json
  - panel/web/package-lock.json
- 严格验证：
  - cd panel/web && npm test -- --run src/pages/settings-page.test.tsx src/pages/settings-timezone.test.tsx ✅
  - cd panel/web && npm test -- --run ✅（17 files, 44 tests）
  - cd panel/web && npm run build ✅

## Session: 2026-02-10 (Panel graceful shutdown + force node delete)

### Task: 节点删除策略与 Panel 停机行为落地
- **Status:** complete
- Actions taken:
  - 修复 `panel/internal/api/nodes.go` 语法损坏问题（`toNodeDTO` 函数拼接错误）。
  - 落地 `DELETE /api/nodes/:id?force=true`：
    - 默认删除保持安全策略（有入站返回冲突）。
    - force 删除时先向 node 下发空入站配置（drain），再删除该节点所有入站，最后删除节点。
  - 落地 `panel/cmd/panel/main.go` 优雅停机：接收 `SIGINT/SIGTERM`，停止 monitor，shutdown HTTP server，关闭 DB。
  - 补齐测试：
    - API：`TestNodesDelete_ForceDrainsNodeAndDeletesInbounds`
    - DB：`TestDeleteInboundsByNode`
  - 完整验证通过：`go test ./panel/... ./node/... -count=1`
- Files created/modified:
  - `panel/cmd/panel/main.go`
  - `panel/internal/api/nodes.go`
  - `panel/internal/db/inbounds.go`
  - `panel/internal/api/nodes_inbounds_test.go`
  - `panel/internal/db/nodes_inbounds_test.go`

### Hotfix: 节点删除弹窗 i18n 混排修复（2026-02-10）
- **Status:** complete
- Actions taken:
  - 节点删除错误新增本地化映射：`ApiError.status === 409` 或命中 `node is in use` 时统一展示 `nodes.deleteInUse`。
  - 删除弹窗不再拼接原始英文错误与中文提示，避免出现 `node is in use该节点...` 混排。
  - 同步补充中英文文案键：`nodes.deleteInUse`。
- Files created/modified:
  - `panel/web/src/pages/nodes-page.tsx`
  - `panel/web/src/i18n/locales/zh.json`
  - `panel/web/src/i18n/locales/en.json`
