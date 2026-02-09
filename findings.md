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
