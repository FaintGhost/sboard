# Findings & Decisions

## Requirements
- 用户要求：继续推进 M4，并使用文件化计划以便回溯。
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
| M4 增加 trigger source 筛选 | 降低定位成本，尤其排查 auto/user/group 触发问题 |
| M4 增加分页（limit/offset） | 避免全量请求，提升性能与可扩展性 |
| M4 增加节点页跳转入口并携带 query | 从“节点问题”直达“该节点同步历史” |

## Issues Encountered
| Issue | Resolution |
|-------|------------|
| M4 在仓库中无显式文档定义 | 采用 M1/M2/M3 顺延策略，聚焦可用性增强 |

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

## M4 Implementation Findings
- 新增 `sync-jobs` URL 过滤解析器：
  - `parseSyncJobsSearchParams`：从 query 解析 node/status/source/range/page；
  - `buildSyncJobsSearchParams`：最小化输出 query（默认值不写入）。
- 同步任务页已接入：
  - 触发来源筛选（`trigger_source`）；
  - 分页（`limit=20 + offset=(page-1)*20`）；
  - query 同步（刷新后保留筛选与页码）。
- 节点页新增“查看同步任务”入口，跳转 `/sync-jobs?node_id=<id>`。
