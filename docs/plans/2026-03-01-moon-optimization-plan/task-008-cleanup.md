### Task 8: Cleanup

**目标：** 删除迁移残留的一次性验证脚本和根目录旧规划文件。

**Files:**
- Delete: `scripts/verify-moon-workspace.sh`
- Delete: `scripts/verify-moon-generate.sh`
- Delete: `scripts/verify-moon-check-generate.sh`
- Delete: `scripts/verify-moon-e2e.sh`
- Delete: `scripts/verify-moon-panel-gates.sh`
- Delete: `scripts/verify-moon-toolchain.sh`
- Delete: `scripts/verify-no-make-entrypoints.sh`
- Delete: `findings.md`
- Delete: `progress.md`
- Delete: `task_plan.md`

**Step 1: 删除 verify 脚本**

```bash
git rm scripts/verify-moon-workspace.sh \
       scripts/verify-moon-generate.sh \
       scripts/verify-moon-check-generate.sh \
       scripts/verify-moon-e2e.sh \
       scripts/verify-moon-panel-gates.sh \
       scripts/verify-moon-toolchain.sh \
       scripts/verify-no-make-entrypoints.sh
```

**Step 2: 删除根目录旧规划文件**

```bash
git rm findings.md progress.md task_plan.md
```

**Step 3: 验证 scripts/ 目录只剩工具脚本**

```bash
ls scripts/
```

Expected: 只剩 `moon-cli.sh`、`dev-panel-web.sh`、`docker-build-push.sh`。

**Step 4: 提交**

```bash
git add -A
git commit -m "chore: remove migration verification scripts and stale planning files"
```
