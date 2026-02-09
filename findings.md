# Findings & Decisions

## Requirements
- 修复：新建分组 -> 给用户加分组 -> 删除分组后，用户页显示 `#2`（失效 ID 占位）的问题。
- 目标：删除分组后，用户分组关系应被正确清理，前端不应展示失效 ID。
- 额外要求：后续工作持续写入 `task_plan.md` / `findings.md` / `progress.md`。

## Research Findings
- 根因 1（核心）：`panel/internal/db/groups.go` 的 `DeleteGroup` 只删除 `groups`，未删除 `user_groups` 关联。
- 根因 2（结构）：`0004_groups.up.sql` 的 `user_groups.group_id` 外键未声明 `ON DELETE CASCADE`，不会自动清理。
- 表现放大点：`panel/web/src/pages/users-page.tsx` 在 group 不存在时使用 `#${groupID}` 兜底，直接暴露脏关联。

## Technical Decisions
| Decision | Rationale |
|----------|-----------|
| 后端事务化清理关联（先删 user_groups，再删 groups） | 从根因上修复数据一致性，避免影响订阅/同步逻辑 |
| 增加 DB 与 API 回归测试 | 防止后续回归，确保删除行为可验证 |
| 前端改为仅渲染存在分组，未知分组不显示 #ID | 防御性 UI，避免异常数据污染体验 |

## Issues Encountered
| Issue | Resolution |
|-------|------------|
| `sed` 读取 `panel/internal/api/test_helpers_test.go` 报不存在 | 确认测试辅助函数在 `users_test.go`，改为读取正确文件 |

## Resources
- `panel/internal/db/groups.go`
- `panel/internal/db/groups_test.go`
- `panel/internal/api/groups_test.go`
- `panel/internal/db/migrations/0004_groups.up.sql`
- `panel/web/src/pages/users-page.tsx`

## Visual/Browser Findings
- 用户报告：分组删除后，用户页分组列出现 `#2`。
- 修复后行为：删除分组后用户分组为空时展示 `-`，不会出现 `#ID`。

---
*Update this file after every 2 view/browser/search operations*
*This prevents visual information from being lost*
