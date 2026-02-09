# Task Plan: M5 - 同步任务可用性增强

## Goal
在现有 M1/M2/M3（同步任务持久化、API、基础页面）之上，完成 M5 可用性增强：更高效筛选、分页浏览、节点联动入口，并保持可回溯记录。

## Current Phase
Phase 4

## Phases
### Phase 1: M5 范围确认与设计落盘
- [x] 识别已完成基线（M1/M2/M3）
- [x] 定义 M5 范围（可用性增强）
- [x] 记录影响文件与测试策略
- **Status:** complete

### Phase 2: TDD 先行（核心逻辑）
- [x] 为 M5 新增逻辑先写失败测试
- [x] 验证测试先红后绿
- **Status:** complete

### Phase 3: 实现 M5 功能
- [x] 同步任务页新增“触发来源”筛选
- [x] 同步任务页支持分页（limit/offset + 上下页）
- [x] 节点页增加“查看该节点同步任务”入口（联动过滤）
- **Status:** complete

### Phase 4: 验证与交付
- [x] 运行测试与构建
- [x] 更新 planning 文件（findings/progress）
- [x] 提交 commit
- **Status:** complete

## Key Questions
1. M5 最小闭环如何定义？（可用性增强优先）
2. 分页是前端切片还是后端分页参数？（后端分页参数优先）
3. 节点联动是否走 query 参数？（是）

## Decisions Made
| Decision | Rationale |
|----------|-----------|
| M5 采用“同步任务可用性增强”路线 | 基于已落地 M1/M2/M3 的自然延伸，收益高且风险低 |
| 使用后端分页参数 | 保持大数据量可扩展性，避免前端全量拉取 |
| 用 URL query 做节点联动 | 支持刷新后状态保持与可分享链接 |
| 以纯函数抽离筛选参数解析 | 可测试、可复用、减少页面复杂度 |

## Errors Encountered
| Error | Attempt | Resolution |
|-------|---------|------------|
| `sync-jobs-filters` 导入失败（文件不存在） | 1 | 按 TDD 创建 `sync-jobs-filters.ts` 实现 |

## Notes
- 保持与现有 `sync-jobs` API 契约兼容。
- M5 仅前端改动，不涉及 node 服务。
- M5 交付 commit：`2424c32`

---

# Task Plan: Hotfix - 同步任务筛选切换表格闪回

## Goal
修复 `sync-jobs` 页面在切换状态筛选时，表格短暂闪回旧缓存数据、观感突兀的问题。

## Current Phase
Phase 4

## Phases
### Phase 1: 复现与根因确认
- [x] 从用户描述确认问题在“筛选后表格内容切换”，非下拉菜单本体
- [x] 用测试复现“缓存旧数据闪回”场景
- **Status:** complete

### Phase 2: TDD 失败用例
- [x] 新增页面测试，验证二次切换筛选时不应出现旧行
- [x] 先看到测试失败（红灯）
- **Status:** complete

### Phase 3: 最小修复实现
- [x] 在筛选键变化时进入“切换保护态”
- [x] 切换保护态期间隐藏缓存行，仅显示 loading 骨架
- [x] 数据请求稳定后恢复正常渲染
- **Status:** complete

### Phase 4: 回归与交付
- [x] 运行相关测试
- [x] 执行前端构建验证
- [x] 更新 planning 文件
- **Status:** complete

## Decisions Made
| Decision | Rationale |
|----------|-----------|
| 用页面内最小状态机屏蔽筛选切换期缓存行 | 直接消除“旧表格闪回”根因，改动范围可控 |
| 保留 React Query 缓存能力，仅在筛选切换窗口期屏蔽渲染 | 避免影响常规轮询刷新与整体性能 |
| 通过测试固化场景（失败 -> 全部 -> 失败） | 该路径最容易复现用户反馈的闪回体验 |

## Follow-up Decision
| Decision | Rationale |
|----------|-----------|
| 空态筛选切换时保持“暂无数据”而非骨架 | 避免视觉闪烁，交互更稳定，符合用户预期 |

## Follow-up Decision 2
| Decision | Rationale |
|----------|-----------|
| 将防闪逻辑抽成通用 hook 并全页面复用 | 防止单页补丁反复遗漏，保证交互一致性 |
