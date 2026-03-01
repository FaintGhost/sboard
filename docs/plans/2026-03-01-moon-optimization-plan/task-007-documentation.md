### Task 7: Documentation

**目标：** 更新 AGENTS.md、README.zh.md、README.en.md，反映新的项目结构和任务入口。

**Files:**
- Modify: `AGENTS.md`
- Modify: `README.zh.md`
- Modify: `README.en.md`

**Step 1: 更新 AGENTS.md**

需要修改的区域：

1. **前端代码质量工具**（第 8 行）：`panel/web` → `web`

2. **前端交付门禁**（第 15-19 行）：
   - `moon run panel:web-lint`（等价于 `cd panel/web && bun run lint`）→ `moon run web:lint`（等价于 `cd web && bun run lint`）
   - `moon run panel:web-typecheck` → `moon run web:typecheck`
   - `moon run panel:web-test` → `moon run web:test`
   - `moon run automation:check-generate` → `moon run panel:check-generate`

3. **项目结构**（第 27-41 行）：
   - 将 `Web 前端：panel/web/` 从 panel 下面提出来，作为独立顶层条目 `web/`
   - 保持其他条目不变

4. **RPC 工作流**（第 54-57 行）：
   - `panel/web/src/lib/rpc/gen/` → `web/src/lib/rpc/gen/`
   - `moon run automation:generate` → `moon run panel:generate`
   - `moon run automation:check-generate` → `moon run panel:check-generate`

**Step 2: 更新 README.zh.md**

需要修改的区域：

1. **快速开始前端**：`cd panel/web` → `cd web`

2. **目录结构**：将 `web/` 提升为与 `panel/`、`node/` 同级

3. **RPC Proto 工作流**：
   - `moon run automation:generate` → `moon run panel:generate`
   - `moon run automation:check-generate` → `moon run panel:check-generate`

4. **前端质量工具**：`cd panel/web` → `cd web`

5. **交付门禁**：更新命令引用

**Step 3: 更新 README.en.md**

与 README.zh.md 相同的区域做相同更新（英文版本）。

**Step 4: 验证无残留引用**

```bash
grep -n 'panel/web\|automation:' AGENTS.md README.zh.md README.en.md
```

Expected: 无输出（无残留引用）。

**Step 5: 提交**

```bash
git add AGENTS.md README.zh.md README.en.md
git commit -m "docs: update project structure and task entries for moon optimization"
```
