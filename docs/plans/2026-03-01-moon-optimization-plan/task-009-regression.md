### Task 9: Regression Verification

**目标：** 全量回归验证，确保所有链路在目录重构后功能等价。

**Step 1: 验证 Moon 项目注册**

```bash
scripts/moon-cli.sh query projects 2>/dev/null | grep '"id"'
```

Expected: panel、web、node、e2e 四个项目，无 automation。

**Step 2: 验证代码生成**

```bash
scripts/moon-cli.sh run panel:generate
```

Expected: 无错误退出。

**Step 3: 验证生成新鲜度检查**

```bash
scripts/moon-cli.sh run panel:check-generate
```

Expected: 退出码 0，输出 "Generated files are up to date."

**Step 4: 验证前端门禁**

```bash
scripts/moon-cli.sh run web:lint
scripts/moon-cli.sh run web:typecheck
scripts/moon-cli.sh run web:test
```

Expected: 三个命令均成功退出。

**Step 5: 验证 Panel Go 测试**

```bash
scripts/moon-cli.sh run panel:test
```

Expected: 所有测试通过。

**Step 6: 验证 Node Go 测试**

```bash
scripts/moon-cli.sh run node:test
```

Expected: 所有测试通过。

**Step 7: 验证 Docker 构建（panel）**

```bash
docker compose -f panel/docker-compose.build.yml build
```

Expected: 构建成功。

注意：构建上下文现在是工作区根（`..`），Dockerfile 从顶层 `web/` 拷贝前端代码。

**Step 8: 验证 E2E smoke**

```bash
scripts/moon-cli.sh run e2e:smoke
```

Expected: 所有 smoke 用例通过。

**Step 9: 验证无残留引用**

```bash
grep -rn 'automation:' AGENTS.md README.zh.md README.en.md
grep -rn 'panel/web' AGENTS.md README.zh.md README.en.md .moon/ panel/moon.yml panel/buf.gen.yaml panel/check-generate.sh scripts/dev-panel-web.sh
```

Expected: 两条 grep 均无输出。

**Step 10: 最终确认**

如果以上全部通过，本次优化完成。更新 `tasks/todo.md` 记录完成状态。
