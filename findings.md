# Findings & Decisions

## Requirements
- 用户要求：继续推进 M5，并使用文件化计划以便回溯。
- 当前可确认基线：
  - M1：同步任务落库 + 重试 + 节点串行。
  - M2：`/api/sync-jobs` 列表/详情/重试 API。
  - M3：同步任务前端独立页面与详情弹窗。

## Research Findings
- 当前 `sync-jobs` 页面已有：时间范围、节点、状态筛选 + 详情 + 手动重试。
- 仍缺可用性增强点：
  - 触发来源筛选（快速定位自动/手动触发任务）；
  - 分页浏览（大量任务可控）；
  - 与节点管理的上下文联动入口。

## Technical Decisions
| Decision | Rationale |
|----------|-----------|
| M5 增加 trigger source 筛选 | 降低定位成本，尤其排查 auto/user/group 触发问题 |
| M5 增加分页（limit/offset） | 避免全量请求，提升性能与可扩展性 |
| M5 增加节点页跳转入口并携带 query | 从“节点问题”直达“该节点同步历史” |

## Issues Encountered
| Issue | Resolution |
|-------|------------|
| M5 在仓库中无显式文档定义 | 采用 M1/M2/M3 顺延策略，聚焦可用性增强 |

## Resources
- `panel/web/src/pages/sync-jobs-page.tsx`
- `panel/web/src/lib/api/sync-jobs.ts`
- `panel/internal/api/sync_jobs.go`
- `panel/web/src/pages/nodes-page.tsx`

## Visual/Browser Findings
- 当前同步任务页无“触发来源”筛选与分页。
- 当前节点页无直达同步任务页并自动过滤节点的入口。

---
*Update this file after every 2 view/browser/search operations*
*This prevents visual information from being lost*

## M5 Implementation Findings
- 新增 `sync-jobs` URL 过滤解析器：
  - `parseSyncJobsSearchParams`：从 query 解析 node/status/source/range/page；
  - `buildSyncJobsSearchParams`：最小化输出 query（默认值不写入）。
- 同步任务页已接入：
  - 触发来源筛选（`trigger_source`）；
  - 分页（`limit=20 + offset=(page-1)*20`）；
  - query 同步（刷新后保留筛选与页码）。
- 节点页新增“查看同步任务”入口，跳转 `/sync-jobs?node_id=<id>`。

## Hotfix Findings (2026-02-09)
- 问题触发点：`sync-jobs` 页面切换 `status` 筛选时，如果该筛选命中过往缓存，会先短暂渲染旧缓存行，再被新请求结果替换。
- 用户观感：表格“闪回旧内容”+“出现很硬”。
- 根因归纳：筛选切换期间直接消费 `jobsQuery.data`，会把缓存命中的旧快照提前渲染出来。
- 修复策略：
  - 引入筛选键变更检测（`filtersKey`）；
  - 进入 `holdRowsDuringFilterSwitch` 保护态；
  - 保护态且请求中时仅显示骨架，不渲染缓存行；
  - 请求稳定后退出保护态，渲染新结果。
- 验证方式：新增 `sync-jobs-page.test.tsx`，构造“第二次切换到 failed 时旧缓存不可见”的回归测试。

## Hotfix Follow-up (2026-02-09)
- 用户复述后确认：真正痛点是“来源筛选从空结果切到空结果时，表格会闪出整块内容”。
- 进一步验证：该“整块内容”主要来自切换保护态中的骨架行，不是旧业务数据。
- 修复细化：
  - 新增 `lastSettledRowCount` 记录上一次稳定结果行数；
  - 当筛选切换中且上一次为 0 行时，继续显示“暂无数据”；
  - 仅在“上一次有数据”且正在切换时显示骨架，避免空态闪烁。
- 新增回归断言：来源筛选“空 -> 空”切换过程中，`暂无数据` 持续可见。

## Hotfix Follow-up 2 (2026-02-09)
- 复核后发现仍闪的关键漏点：`showSkeletonLoading` 仍直接依赖 `jobsQuery.isLoading`。
- 当来源从空结果切到另一个空结果且目标 query 无缓存时，`isLoading=true` 会再次触发骨架，造成用户可见闪烁。
- 修复：
  - 仅在“初次加载（非筛选切换态）”显示 `isLoading` 骨架；
  - 筛选切换态下，是否显示骨架完全由 `lastSettledRowCount` 决定（>0 才显示，=0 保持 no-data）。
- 补强测试：在空→空切换后 `await` 一帧，断言仍显示 `暂无数据`，且不存在 `加载中...`。

## Hotfix Follow-up 3 (2026-02-09)
- 用户反馈正确：之前仅修 `sync-jobs`，其余带筛选的 table 页仍会在切换时出现瞬时闪动。
- 统一方案：抽出 `useTableQueryTransition`，在页面层统一处理“筛选切换期间的行渲染策略”。
- 已接入页面：
  - `users-page`（状态筛选）
  - `inbounds-page`（节点筛选）
  - `subscriptions-page`（状态筛选）
  - `sync-jobs-page`（时间/节点/状态/来源筛选）
- 关键改进：
  - 同步检测 `filterKey` 变化，避免切换首帧误显示骨架；
  - 空结果切空结果时保持 `noData`；
  - 有数据切换时才显示骨架过渡，避免“旧内容闪现”。

## Inbounds Template Mode (2026-02-09)
- 目标：把“编辑入站”改成纯 JSON 模板输入，减少表单字段分裂与协议特化 UI 复杂度。
- 方案：
  - 前端保留 `node_id` 选择；
  - 其余字段统一来自模板 JSON（`type/tag/listen_port/...`）；
  - `users` 字段仅作为展示占位，提交时忽略；
  - `tls`/`transport` 顶层对象映射为 `tls_settings`/`transport_settings`；
  - 其他顶层字段自动归入 `settings`。
- 与现有后端兼容：
  - API 仍接收 `protocol/tag/listen_port/public_port/settings/tls_settings/transport_settings`；
  - Node 下发时继续由 Panel 自动注入 `users`，模板中的 `users` 不参与下发。

## Monaco + sing-box Tools Findings (2026-02-09)
- 用户诉求明确：
  - 前端编辑器改为 `Monaco`。
  - 充分利用 `sing-box format/check/generate`，避免自造轮子。
- 架构落地：
  - 新增 `panel/internal/singboxcli`，将命令执行与 API 层解耦。
  - API 层仅做请求校验、模式转换（`inbound` 模板自动包裹成最小可校验完整配置）、错误映射。
- 兼容策略：
  - `format/check` 支持 `mode=inbound|config|auto`。
  - `inbound` 模式下会自动封装 `inbounds/outbounds`，并在 format 后回提取第一个 inbound。
- 运行时注意：
  - 若运行环境无 `sing-box` 可执行文件，接口返回 `503 sing-box command not available`。
  - 可通过 `PANEL_SING_BOX_BIN` 指定二进制路径（默认 `sing-box`）。

## Hotfix Findings: public_port 语义与工具依赖（2026-02-09）
- 用户反馈成立：
  - `public_port` 不是 sing-box inbound 原生字段，放在模板里会干扰“原生配置心智模型”。
  - Docker 场景下要求外部 `sing-box` 命令不可接受，工具能力应内嵌在 Panel。
- 修复决策：
  - 将 `public_port` 视为 Panel 元数据，不再作为模板字段参与编辑；wrapper 校验前会剔除该字段。
  - 把 `format/check/generate` 下沉为 Panel 内部服务实现：
    - format：`badjson.Omitempty` + JSON pretty print
    - check：`option.Options` 反序列化校验
    - generate：内置常见命令（uuid/reality/wg/vapid）

## Strict Check + UX Findings (2026-02-09)
- 严格检查定义：不仅做 JSON/Options 反序列化，还进行 sing-box 运行初始化（`box.New`）验证。
- 交互修正：快速请求场景下按钮文字闪变会引起“跳闪”感，改为固定按钮文案，仅保留禁用态。

## SS2022 Password Generation Findings (2026-02-09)
- 需求：针对 `2022-blake3-aes-128-gcm` / `2022-blake3-aes-256-gcm` 提供基于 key length 的 base64 随机密码生成。
- 落地：
  - 通过 `Generate(rand-base64-16|32)` 统一输出 base64 文本；
  - UI 侧命令执行成功后自动写回模板 `password` 字段，降低误操作成本。
- 校验：
  - 后端单测已验证 16/32 字节解码长度正确；
  - 非法长度输入返回 `ErrInvalidGenerateKind`。

## Full Config Template End-to-End Findings (2026-02-09)
- 用户目标确认：入站模板需要直接编辑“完整 sing-box 配置”而非仅 inbound 片段，并希望可通过 `route/outbounds` 实现服务端分流。
- 关键链路缺口：
  - Panel 侧虽然可把 `settings.__config` 合并进 sync payload，但 Node 仍只解析/应用 `inbounds`，导致全局字段被忽略。
  - 生成密钥回填逻辑仅写根级 `password`，对完整模板（`inbounds[0].password`）不稳定。
- 本次改造：
  - 前端模板：
    - 支持完整 config 读写（含 `$schema/log/dns/route/outbounds/...`）。
    - 生成 `rand-base64-16/32` 后优先回填 `inbounds[0].password`。
  - Panel sync payload：
    - `SyncPayload` 增加 `$schema`。
    - 全局配置合并改为“仅首次合并”，避免多入站重复拼接 `outbounds/services/endpoints`。
  - Node 应用链路：
    - `ParseAndValidateConfig` 解析完整 `option.Options`。
    - `Core.ApplyOptions` 基于完整 options 重建并切换 box，真正让 route/outbounds/dns 等生效。
    - 启动恢复链路同步切换到完整配置解析与应用。
  - sing-box 工具 API：
    - inbound/auto 模式遇到完整配置时也会清理 `inbounds[].public_port`（Panel 元字段）。
- 验证结果：
  - Go：`panel/internal/node`、`panel/internal/api`、`panel/internal/singboxcli`、`node/internal/sync`、`node/internal/core` 全部通过。
  - Web：`src/lib/inbound-template.test.ts` 通过，`npm run build` 通过。

## Session Findings (2026-02-09, Node UX)
- 节点后端模型已包含 `last_seen_at`（`panel/internal/api/nodes.go` 与前端 `Node` 类型都已定义）。
- 当前节点表仅展示 `name/group/api/public/status`，未单独展示 last seen。
- 节点编辑里 `api_address` 与 `public_address` 是完全独立输入，缺少“同地址”快捷体验。
- 流量相关文案当前使用 `nodes.lastSampleAt`，语义偏采样实现细节，不够用户导向。
- 本轮优先只改前端交互与文案，不引入后端 schema 变更。

## Session Findings (2026-02-09, Node UX 实施结果)
- 已在 `nodes-page` 增加“公网地址与 API 地址一致”联动开关：
  - 创建节点默认开启联动，`public_address` 初始跟随 `api_address`。
  - 联动开启时，编辑 `api_address` 会自动同步更新 `public_address`。
  - 可手动关闭联动后独立填写公网地址。
- 节点列表新增 `Last Seen` 列，并扩展统一表格间距规则到 `seven` 列场景，避免局部样式漂移。
- 流量展示文案已统一为“最近更新 / Last Updated”：
  - 节点卡片统计
  - 节点流量弹窗列头
  - 中英文 i18n 节点文案
- 测试阶段发现：`table-query-transition` 与 `sync-jobs-page` 的既有回归测试仍失败（与本轮节点页改动无直接耦合，属于当前基线问题）。


## Session Findings (2026-02-09, 无骨架修复)
- 用户明确要求筛选过渡不要骨架，统一调整 `useTableQueryTransition` 为无骨架策略。
- 过渡策略更新后行为：
  - 筛选切换期间不展示旧行、不展示骨架；
  - 空 -> 空切换保持 `暂无数据` 稳定可见；
  - 首次加载可保留顶部 loading 文案，不影响表格稳定性。
- 该策略是通用 hook 级别修复，已覆盖 `sync-jobs/users/inbounds/subscriptions` 同类筛选表格。

## Session Findings (2026-02-09, Subscription Base URL)
- 根因确认：订阅页此前使用 `window.location.origin` 拼接订阅地址，导致“管理入口地址”和“用户实际订阅入口地址”不一致时（反向代理/公网域名）链接错误。
- 设计决策：
  - 采用后端持久化配置项 `subscription_base_url`（存 DB），而不是仅前端本地缓存。
  - 输入值要求完整 URL（含协议），仅允许 `http/https`，避免歧义与非法拼接。
  - 订阅页生成 URL 时使用 `new URL("api/sub/<uuid>", base)`，兼容带路径前缀的反向代理场景。
- 覆盖范围：
  - 后端：migration + DB + API + 路由 + API 测试。
  - 前端：settings 保存链路 + subscriptions 链接生成 + 页面测试。

## Session Findings (2026-02-09, Protocol Dropdown + IP:Port)
- 用户需求从“自由输入完整 URL”收敛为“协议选择 + IP:端口输入”，可显著减少配置错误。
- 后端校验策略同步收紧，确保前后端一致：
  - 必须 `http/https`
  - Host 必须是合法 IPv4/IPv6（不接受域名）
  - 必须显式端口且范围 1-65535
  - 禁止 path/query/fragment
- 前端采用分段输入后，订阅链接预览与实际保存值保持一致，避免“看起来正确但存库格式不合法”。

## Session Findings (2026-02-09, Settings UI 收敛)
- 用户偏好确认：设置页信息密度要更紧凑，避免长卡片造成滚动负担。
- 本次收敛策略：
  - 将“订阅访问地址”与“订阅格式说明”改为同排双列（桌面端）。
  - 系统信息卡中移除 API 端点说明性副标题，仅保留字段行。
  - API 端点展示值改为与订阅访问地址实时联动（协议下拉 + IP:端口输入），并改为与字段文字同一行。

## Session Findings (2026-02-09, Settings Card Order)
- 按用户指定顺序重排设置页 2x2 卡片：
  - 第一行：订阅访问地址 | 系统信息
  - 第二行：语言 | 订阅格式说明
- 系统信息卡中的 API 端点展示保持与订阅访问地址实时联动，并使用单行字段样式。

## Session Findings (2026-02-09, Settings UX polish)
- 针对设置页主操作链路（订阅访问地址保存）进行了交互收敛，重点解决“反馈闪动”和“输入纠错成本高”的问题。
- 统一策略：
  - 保存按钮增加最小加载时长（500ms）+ spinner，避免极短请求导致的闪烁感。
  - 地址输入增加失焦即时校验，并在提交失败时自动回焦输入框。
  - 预览地址区增加一键复制按钮（含复制成功状态反馈）。
  - 反馈区改为 `aria-live`，提升可访问性与状态感知一致性。
- 命名收敛：系统信息中的“API 端点”字段改为“当前订阅基础地址”，与实际用途保持一致。

## Session Findings (2026-02-09, Web Design Guidelines 严格审查待修复清单)

### 高优先级（可访问性/关键交互）
- `panel/web/src/pages/subscriptions-page.tsx:77`
  - 复制图标按钮缺少可访问名称（无 `aria-label`）。
- `panel/web/src/pages/subscriptions-page.tsx:304`
  - 外链图标按钮缺少可访问名称（无 `aria-label`）。
- `panel/web/src/pages/subscriptions-page.tsx:227`
  - 搜索输入仅依赖 placeholder，缺少显式 label。
- `panel/web/src/pages/users-page.tsx:245`
  - 搜索输入仅依赖 placeholder，缺少显式 label。
- `panel/web/src/pages/groups-page.tsx:466`
  - 成员移动图标按钮（左箭头）仅有 `title`，缺少可访问名称。
- `panel/web/src/pages/groups-page.tsx:476`
  - 成员移动图标按钮（右箭头）仅有 `title`，缺少可访问名称。
- `panel/web/src/pages/sync-jobs-page.tsx:307`
  - 表格行直接 `onClick` 作为主交互入口，键盘可达性不足。

### 中优先级（UX 一致性/信息架构/性能）
- `panel/web/src/pages/subscriptions-page.tsx:175`
  - 信息 tooltip 触发器缺少明确可访问名称。
- `panel/web/src/pages/subscriptions-page.tsx:237`
  - 状态筛选 SelectTrigger 缺少 `aria-label`。
- `panel/web/src/pages/subscriptions-page.tsx:111`
  - 筛选状态未同步 URL，刷新/分享无法保留筛选上下文。
- `panel/web/src/pages/users-page.tsx:80`
  - 筛选与搜索未同步 URL，状态不可恢复。
- `panel/web/src/pages/groups-page.tsx:429`
  - “当前成员”搜索框缺少显式 label。
- `panel/web/src/pages/groups-page.tsx:493`
  - “可加入用户”搜索框缺少显式 label。
- `panel/web/src/pages/groups-page.tsx:384`
  - create 模式下 Dialog 缺少描述文本。
- `panel/web/src/pages/users/edit-user-dialog.tsx:121`
  - create 模式下 Dialog 缺少描述文本。
- `panel/web/src/pages/nodes-page.tsx:441`
  - create 模式下 Dialog 缺少描述文本。
- `panel/web/src/pages/inbounds-page.tsx:402`
  - create 模式下 Dialog 缺少描述文本。
- `panel/web/src/routes/index.tsx:10`
  - 页面路由未使用懒加载，首包压力偏大。
- `panel/web/src/pages/inbounds-page.tsx:2`
  - Monaco 作为重依赖未延迟加载，影响初始加载体积。
- `panel/web/src/i18n/locales/zh.json:9`
  - 文案省略号风格不统一（`...` 与 `…` 混用）。
- `panel/web/src/i18n/locales/en.json:9`
  - 文案省略号风格不统一（`...` 与 `…` 混用）。

### 备注
- 当前仅记录问题，不做代码修复。
- 后续修复建议顺序：高优先级可访问性 -> URL 状态一致性 -> 性能优化 -> 文案规范统一。

## Session Findings (2026-02-09, 严格审查清单第一轮修复进展)

### 已覆盖范围
- 可访问性修复（图标按钮可访问名称）
  - `panel/web/src/pages/subscriptions-page.tsx`
  - `panel/web/src/pages/groups-page.tsx`
- 可访问性修复（搜索输入可访问标签）
  - `panel/web/src/pages/subscriptions-page.tsx`
  - `panel/web/src/pages/users-page.tsx`
  - `panel/web/src/pages/groups-page.tsx`
- 同类弹窗语义修复（create 模式统一补 `DialogDescription`）
  - `panel/web/src/pages/users/edit-user-dialog.tsx`
  - `panel/web/src/pages/nodes-page.tsx`
  - `panel/web/src/pages/inbounds-page.tsx`
  - `panel/web/src/pages/groups-page.tsx`
- 列表筛选 URL 状态持久化（可刷新恢复）
  - `panel/web/src/pages/users-page.tsx`
  - `panel/web/src/pages/subscriptions-page.tsx`
  - 公共解析/构建：`panel/web/src/lib/user-list-filters.ts`

### 暂未覆盖范围
- 暂无

### 后续补齐计划
- 本轮清单已全部完成，后续如继续优化可聚焦：
  - `vite` manualChunks 细分（进一步降低大 chunk 警告）；
  - 在关键输入/筛选页面补充 URL 状态恢复的 E2E 用例。

## Session Findings (2026-02-09, 严格审查清单第二轮修复进展)

### 已覆盖范围
- `sync-jobs` 语义化交互收敛
  - `panel/web/src/pages/sync-jobs-page.tsx`
  - 由“整行可点击”改为“独立详情按钮列”，并保留清晰操作入口。
  - 同步文案：`panel/web/src/i18n/locales/zh.json`、`panel/web/src/i18n/locales/en.json`（`viewDetail`）。
- 性能优化
  - 路由级懒加载：`panel/web/src/routes/index.tsx`（React.lazy + Suspense）。
  - Monaco 按需加载：`panel/web/src/pages/inbounds-page.tsx`（lazy + Suspense fallback）。
- 文案规范统一
  - `panel/web/src/i18n/locales/zh.json`
  - `panel/web/src/i18n/locales/en.json`
  - 省略号统一为 `…`。

### 验证结果
- `npm test -- --run` 通过（11 files, 27 tests）。
- `npm run build` 通过。

## Session Findings (2026-02-09, Interaction Design 第一阶段落地)

### 已覆盖范围
- 全局动效 token 与缓动规范
  - `panel/web/src/index.css`
  - 新增统一变量：`--motion-fast/base/slow`、`--ease-out/in/in-out/spring`。
- 动画降级策略
  - `panel/web/src/index.css`
  - 新增 `prefers-reduced-motion` 全局降级规则。
- 表格切换过渡统一抽象
  - 新增 `panel/web/src/lib/table-motion.ts`
  - 四个页面统一接入：
    - `panel/web/src/pages/users-page.tsx`
    - `panel/web/src/pages/inbounds-page.tsx`
    - `panel/web/src/pages/subscriptions-page.tsx`
    - `panel/web/src/pages/sync-jobs-page.tsx`

### 交互收益
- 筛选切换时过渡速度与曲线统一，页面间手感一致。
- 减少“页面 A 快、页面 B 慢”的体验割裂。
- 对系统开启“减少动态效果”的用户自动降级，避免眩晕与干扰。

### 验证结果
- `npm test -- --run` 通过（11 files, 27 tests）。
- `npm run build` 通过。

## Session Findings (2026-02-09, Interaction Design 第二阶段：异步按钮统一)
- 用户反馈点：多个页面异步按钮在短请求时会出现“检查中/保存中/重试中”瞬时闪变，体感突兀。
- 统一策略：新增通用 `AsyncButton`，内置“延迟显示 pending + 最小可见时长”机制，避免 0.x 秒抖动。
- 关键设计：
  - `pending=true` 时先禁用按钮，文案与 spinner 延迟显示；
  - 若请求极短完成，不切换到 pending 文案（避免闪）；
  - 若已显示 pending，保持最小可见时长后再回到 idle（避免抖）。
- 覆盖范围（同类一起修）：
  - 页面：`settings`、`sync-jobs`、`users`、`groups`、`nodes`、`inbounds`
  - 认证：`login-form`、`bootstrap-form`
- 测试：新增 `panel/web/src/components/ui/async-button.test.tsx`
  - 覆盖“短 pending 不闪文案”与“长 pending 平滑收敛”两条核心行为。
- 验证结果：
  - `npm test -- --run` 全绿（12 files, 29 tests）。
  - `npm run build` 通过。

## Session Findings (2026-02-09, Frontend Design 收敛第一轮)
- 方向：采用 `Network Ops Console` 风格，在不改业务逻辑前提下提升可扫描性和视觉识别。
- 导航信息架构：
  - 将 `subscriptions` 从 documents 分组并入主导航，减少路径分叉。
  - 保留 settings 在 secondary 区，形成“主业务 / 系统配置”分层。
- Dashboard 改造：
  - 用“系统概览卡片 + 快捷入口”替代低价值“后端连通性”信息块。
  - 快捷卡提供核心计数（用户、节点等）和上下文提示（活跃/在线）。
  - 保留并强化 Top Nodes 列表，增加“查看全部节点”CTA。
- Settings 视觉层级：
  - 订阅访问地址卡增加主色弱渐变强调，突出核心配置属性。
  - 预览地址改为高对比边框+底色，提升可读性。
- 路由 fallback：
  - 从简单省略号替换为结构化 skeleton，占位语义更清晰，体感更稳定。

## Session Findings (2026-02-09, Frontend Design 第二轮：页面骨架统一)
- 目标：统一页面头、空状态和表格工具栏密度，减少各页视觉割裂。
- 新增复用模块：
  - `panel/web/src/components/page-header.tsx`
    - 统一页面标题区域视觉（标题/副标题/右侧主操作）。
  - `panel/web/src/components/table-empty-state.tsx`
    - 统一表格空状态文案与 CTA 行为。
  - `panel/web/src/lib/table-toolbar.ts`
    - 统一 table header 工具栏布局密度与断点行为。
- 批量接入页面（同类一次性覆盖）：
  - `users/groups/subscriptions/inbounds/sync-jobs/nodes`
- 结果：
  - 各管理页页面头风格统一，主操作按钮位置一致；
  - 表格空态从“仅 no data 文本”升级为“文案 + 快捷 CTA”；
  - 过滤器布局统一，减少不同页面切换的认知负担。

## Session Findings (2026-02-09, Theme Color System)
- 诉求：默认黑白主题过于单调，需要一套更有识别度的 shadcn/ui 配色。
- 主题方向：`Cobalt Operations`（冷静蓝青主轴 + 清晰语义色），适配运维控制台场景。
- 改动策略：
  - 仅替换 `index.css` 的设计 token（OKLCH），不改业务组件逻辑。
  - 同时覆盖 light/dark、chart、sidebar 变量，避免“只改主色，其它割裂”。
- 结果：
  - 整体从黑白过渡为蓝青科技感，保留高可读对比；
  - 卡片、侧栏、图表统一语义；
  - 现有组件无回归。

## Session Findings (2026-02-09, 节点删除功能回归修复)
- 问题现象：节点管理列表操作菜单缺少“删除节点”，导致节点只能新增/编辑，无法删除。
- 根因定位：历史提交 `chore(ui): simplify nodes actions` 在简化菜单时误移除了 `deleteNode` API 引用、`deleteMutation` 和 destructive 菜单项。
- 修复方案（最小闭环）：
  - 恢复 `deleteNode` API 调用与 `deleteMutation`。
  - 在节点行菜单恢复 destructive 的“删除节点”入口。
  - 删除失败时统一通过 `actionMessage` 显示错误（复用 `nodes.deleteFailed` / `ApiError`）。
- TDD 回归保护：
  - 新增 `panel/web/src/pages/nodes-page.test.tsx`。
  - 覆盖“从行操作删除节点并触发 DELETE 请求”的关键路径。
- 同类排查：
  - 已核查 `users/groups/inbounds/nodes` 页面删除入口，当前均具备删除能力。
- 验证结果：
  - `npm test -- --run` ✅（13 files, 30 tests）
  - `npm run build` ✅
