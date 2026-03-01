### Task 1: Directory Move

**目标：** 将 `panel/web/` 移动到顶层 `web/`，将 `scripts/check-generate.sh` 迁入 `panel/`。

**Files:**
- Move: `panel/web/` → `web/`
- Move: `scripts/check-generate.sh` → `panel/check-generate.sh`

**Step 1: 移动 web 目录**

```bash
git mv panel/web web
```

**Step 2: 验证移动成功**

```bash
ls web/package.json web/src/lib/rpc/gen/ web/bun.lock
```

Expected: 三个路径都存在。

**Step 3: 移动 check-generate.sh**

```bash
git mv scripts/check-generate.sh panel/check-generate.sh
```

**Step 4: 验证移动成功**

```bash
ls panel/check-generate.sh
```

Expected: 文件存在。

**Step 5: 确认 git 状态**

```bash
git status
```

Expected: 显示 renamed 文件，无 untracked 残留。

**Step 6: 提交**

```bash
git add -A
git commit -m "refactor: promote web/ to top level and move check-generate.sh"
```
