# Task Plan: 修复分组删除后用户显示残留 #ID 的逻辑问题

## Goal
修复删除分组后用户仍保留失效分组 ID 的数据问题，确保用户分组展示与后端数据一致，并建立后续 M4 的文件化追踪。

## Current Phase
Phase 4

## Phases
### Phase 1: 问题确认与约束收集
- [x] 理解用户反馈与复现路径
- [x] 初始化文件化计划（task_plan/findings/progress）
- [x] 确认受影响的数据层与 API 层
- **Status:** complete

### Phase 2: 根因定位与修复方案
- [x] 定位分组删除流程中关联清理缺失点
- [x] 设计后端修复（删除分组时清理 user_groups / 相关引用）
- [x] 评估前端兜底显示策略
- **Status:** complete

### Phase 3: 实施与测试
- [x] 实现后端修复与必要事务保障
- [x] 补充或更新测试用例
- [x] 运行后端与前端验证
- **Status:** complete

### Phase 4: 交付与 M4 衔接
- [x] 更新 planning 文件（可回溯记录）
- [ ] 提交修复 commit
- [ ] 给出 M4 下一步建议
- **Status:** in_progress

## Key Questions
1. 删除 group 时，是否仅删除 groups 表记录，遗漏 user_groups 关联清理？（是）
2. 数据库外键是否配置了 ON DELETE CASCADE，还是需要代码层手动清理？（当前需要代码层手动清理）
3. 前端显示 #ID 是否为兜底逻辑导致的症状放大？（是）

## Decisions Made
| Decision | Rationale |
|----------|-----------|
| 采用后端根因修复为主，前端兜底为辅 | 避免只在 UI 掩盖脏数据，保证订阅/同步逻辑一致性 |
| 采用 planning-with-files 全程记录 | 满足用户要求，便于后续 M4 回溯 |
| 为 DeleteGroup 增加事务删除 user_groups | 保证删除分组后关联数据一致性 |

## Errors Encountered
| Error | Attempt | Resolution |
|-------|---------|------------|
| `sed: can't read panel/internal/api/test_helpers_test.go` | 1 | 改查 `panel/internal/api/users_test.go` 中的 helper |

## Notes
- 本次修复不需要 DB 迁移（通过代码层事务清理实现）。
- 如未来做 schema 升级，可考虑补 `ON DELETE CASCADE` 迁移来增强约束一致性。
