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
