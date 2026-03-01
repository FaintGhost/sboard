### Task 2: Moon Config

**目标：** 更新 Moon 工作区配置，重写各项目 moon.yml，取消 automation 项目。

**Files:**
- Modify: `.moon/workspace.yml`
- Modify: `.moon/tasks/all.yml`
- Modify: `panel/moon.yml`
- Create: `web/moon.yml`
- Modify: `node/moon.yml`
- Delete: `scripts/moon.yml`

**Step 1: 更新 workspace.yml**

将 `.moon/workspace.yml` 改为：

```yaml
projects:
  panel: panel
  web: web
  node: node
  e2e: e2e
```

移除 `automation: scripts`、`defaultProject: automation`。

**Step 2: 更新 .moon/tasks/all.yml**

将 `all.yml` 中 `rpcGenerated` 路径更新：

```yaml
fileGroups:
  rpcSpec:
    - /panel/proto/**/*.proto
    - /panel/buf*.yaml
  rpcGenerated:
    - /panel/internal/rpc/gen/**
    - /web/src/lib/rpc/gen/**
    - /node/internal/rpc/gen/**
  e2eInfra:
    - /e2e/docker-compose.e2e.yml
    - /e2e/Dockerfile
    - /e2e/sb-client.Dockerfile
    - /e2e/entrypoint.sh
```

**Step 3: 重写 panel/moon.yml**

```yaml
tasks:
  generate:
    command: go
    args:
      - generate
      - ./internal/rpc/...
    options:
      cache: false

  check-generate:
    deps:
      - panel:generate
    command: bash
    args:
      - check-generate.sh
    options:
      cache: false
      runFromWorkspaceRoot: true

  test:
    command: go
    args:
      - test
      - ./...
      - -count=1
    options:
      cache: false
```

**Step 4: 创建 web/moon.yml**

```yaml
toolchain:
  default: bun

tasks:
  lint:
    command: bun
    args:
      - run
      - lint

  format:
    command: bun
    args:
      - run
      - format

  typecheck:
    command: bunx
    args:
      - tsc
      - -b

  test:
    command: bun
    args:
      - run
      - test
```

**Step 5: 更新 node/moon.yml**

```yaml
tasks:
  test:
    command: go
    args:
      - test
      - ./...
      - -count=1
    options:
      cache: false
```

**Step 6: 删除 scripts/moon.yml**

```bash
git rm scripts/moon.yml
```

**Step 7: 验证 Moon 项目注册**

```bash
scripts/moon-cli.sh query projects 2>/dev/null | grep '"id"'
```

Expected: 应包含 `panel`、`web`、`node`、`e2e`，不包含 `automation`。

**Step 8: 验证任务注册**

```bash
scripts/moon-cli.sh query projects --id panel 2>/dev/null | grep '"target"'
scripts/moon-cli.sh query projects --id web 2>/dev/null | grep '"target"'
scripts/moon-cli.sh query projects --id node 2>/dev/null | grep '"target"'
```

Expected:
- panel: `panel:generate`, `panel:check-generate`, `panel:test`
- web: `web:lint`, `web:format`, `web:typecheck`, `web:test`
- node: `node:test`

**Step 9: 提交**

```bash
git add -A
git commit -m "refactor(moon): remove automation alias layer, add per-project tasks"
```
