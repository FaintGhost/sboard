# Task Plan: M4 - 同步任务可用性增强

## Goal
在现有 M1/M2/M3（同步任务持久化、API、基础页面）之上，完成 M4 可用性增强：更高效筛选、分页浏览、节点联动入口，并保持可回溯记录。

## Current Phase
Phase 4

## Phases
### Phase 1: M4 范围确认与设计落盘
- [x] 识别已完成基线（M1/M2/M3）
- [x] 定义 M4 范围（可用性增强）
- [x] 记录影响文件与测试策略
- **Status:** complete

### Phase 2: TDD 先行（核心逻辑）
- [x] 为 M4 新增逻辑先写失败测试
- [x] 验证测试先红后绿
- **Status:** complete

### Phase 3: 实现 M4 功能
- [x] 同步任务页新增“触发来源”筛选
- [x] 同步任务页支持分页（limit/offset + 上下页）
- [x] 节点页增加“查看该节点同步任务”入口（联动过滤）
- **Status:** complete

### Phase 4: 验证与交付
- [x] 运行测试与构建
- [x] 更新 planning 文件（findings/progress）
- [ ] 提交 commit
- **Status:** in_progress

## Key Questions
1. M4 最小闭环如何定义？（可用性增强优先）
2. 分页是前端切片还是后端分页参数？（后端分页参数优先）
3. 节点联动是否走 query 参数？（是）

## Decisions Made
| Decision | Rationale |
|----------|-----------|
| M4 采用“同步任务可用性增强”路线 | 基于已落地 M1/M2/M3 的自然延伸，收益高且风险低 |
| 使用后端分页参数 | 保持大数据量可扩展性，避免前端全量拉取 |
| 用 URL query 做节点联动 | 支持刷新后状态保持与可分享链接 |
| 以纯函数抽离筛选参数解析 | 可测试、可复用、减少页面复杂度 |

## Errors Encountered
| Error | Attempt | Resolution |
|-------|---------|------------|
| `sync-jobs-filters` 导入失败（文件不存在） | 1 | 按 TDD 创建 `sync-jobs-filters.ts` 实现 |

## Notes
- 保持与现有 `sync-jobs` API 契约兼容。
- M4 仅前端改动，不涉及 node 服务。
