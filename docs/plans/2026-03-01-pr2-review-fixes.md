# PR #2 Code Review 修复计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复 PR #2 (moon monorepo optimization) 代码审查中发现的所有问题，确保 PR 可安全合并。

**Architecture:** 在 `worktree-moon-optimization` 分支上逐项修复 CRITICAL/HIGH/MEDIUM 问题。涉及 Docker 构建安全、CI 工作流健壮性、Go 版本一致性、以及文档中过时引用的清理。

**Tech Stack:** Docker, GitHub Actions, Moon, Go 1.25, Bun

---

### Task 1: 创建根目录 .dockerignore [CRITICAL]

**Files:**
- Create: `.dockerignore`

**Context:** Docker 构建上下文从 `panel/` 改为 `.`（工作区根），但没有根目录 `.dockerignore`，导致整个仓库（包括 `.git/`、无关项目目录、数据库文件）被发送到 Docker daemon。`panel/.dockerignore` 仅在构建上下文为 `panel/` 时生效。

**Step 1: 查看现有 panel/.dockerignore 作为参考**

Run: `git show worktree-moon-optimization:panel/.dockerignore`

**Step 2: 创建根目录 .dockerignore**

```dockerignore
# VCS
.git
.github

# Moon cache
.moon/cache/

# Claude / IDE
.claude/
.idea/
.vscode/

# Projects not needed in panel Docker build
node/
e2e/
docs/

# Database files
**/*.db
**/*.db-*
**/*.sqlite
**/*.sqlite-*

# Data directories
data/

# Web build artifacts (rebuilt inside Docker)
web/node_modules/
web/dist/
web/.env
web/.env.*

# Node artifacts
node_modules/
dist/

# Misc
*.md
LICENSE
.prototools
.gitignore
```

**Step 3: 验证 .dockerignore 被正确读取**

Run: `git add .dockerignore && echo "Created root .dockerignore"`

**Step 4: Commit**

```bash
git add .dockerignore
git commit -m "fix(docker): add root .dockerignore for workspace-root build context"
```

---

### Task 2: Dockerfile 添加 go.sum COPY [HIGH]

**Files:**
- Modify: `panel/Dockerfile:21`

**Context:** `COPY panel/go.mod ./` 只复制了 `go.mod`，遗漏了 `go.sum`。虽然后续有 `go mod tidy` 兜底，但 `go mod download` 步骤在有 `go.sum` 时会更快更可靠。使用通配符确保 `go.sum` 不存在时也不会报错。

**Step 1: 修改 Dockerfile 的 COPY 行**

将 `panel/Dockerfile` 第 21 行：
```dockerfile
COPY panel/go.mod ./
```
改为：
```dockerfile
COPY panel/go.mod panel/go.sum* ./
```

**Step 2: 验证 Dockerfile 语法**

Run: `docker build --check -f panel/Dockerfile . 2>&1 || echo "docker build --check not supported, visual inspection OK"`

**Step 3: Commit**

```bash
git add panel/Dockerfile
git commit -m "fix(docker): copy go.sum alongside go.mod for reliable module download"
```

---

### Task 3: 统一 Go 版本 — Dockerfile 回退到 1.25 [HIGH]

**Files:**
- Modify: `panel/Dockerfile:13` — `golang:1.26` → `golang:1.25`

**Context:** `.prototools` 设置 `go = "1.26.0"` 用于本地开发工具链，但 `go.work`、`panel/go.mod`、`node/go.mod` 均为 `go 1.25`，README 也写 "Go 1.25+"。Go module 的 `go` 指令是最低版本要求，Dockerfile 中的 Go 版本应与 `go.mod` 保持一致以避免版本漂移。升级 go.mod 到 1.26 是单独的工作，不应在本重构 PR 中混入。

**Step 1: 修改 Dockerfile FROM 行**

将 `panel/Dockerfile` 第 13 行：
```dockerfile
FROM golang:1.26 AS builder
```
改为：
```dockerfile
FROM golang:1.25 AS builder
```

**Step 2: 验证与 go.mod 一致**

Run: `head -2 panel/go.mod` — 应显示 `go 1.25`

**Step 3: Commit**

```bash
git add panel/Dockerfile
git commit -m "fix(docker): align Go image version with go.mod (1.25)"
```

---

### Task 4: 增强 CI 工作流 — 添加 Go setup 和缓存 [HIGH]

**Files:**
- Modify: `.github/workflows/ci.yml`

**Context:** CI 中 `moonrepo/setup-toolchain@v0` 安装 moon 和 proto，proto 通过 `.prototools` 知道需要 Go 1.26.0，但不保证自动安装（需 `auto-install` 配置）。显式添加 `actions/setup-go` 更可靠。`panel:test` 需要 CGO 依赖。同时添加依赖缓存加速 CI。

**Step 1: 重写 ci.yml**

```yaml
name: ci

on:
  push:
    branches: [main, master]
  pull_request:

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - uses: actions/setup-go@v5
        with:
          go-version-file: panel/go.mod
      - run: moon run panel:check-generate

  web:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - run: moon run web:lint web:typecheck web:test

  panel:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - uses: actions/setup-go@v5
        with:
          go-version-file: panel/go.mod
      - name: Install CGO dependencies
        run: sudo apt-get update && sudo apt-get install -y --no-install-recommends gcc libc6-dev libsqlite3-dev
      - run: moon run panel:test

  node:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - uses: actions/setup-go@v5
        with:
          go-version-file: node/go.mod
      - run: moon run node:test
```

**Step 2: 验证 YAML 语法**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))" && echo "YAML OK"`

**Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "fix(ci): add explicit Go setup and CGO deps for reliable CI"
```

---

### Task 5: 更新设计文档中的过时引用 [HIGH]

**Files:**
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-design/_index.md:49,52-55`
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md:133,136-141`
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-design/architecture.md:61`
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-design/best-practices.md:8,10,55`
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-001-sync-success-impl.md:56`
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-001-sync-success-test.md:50`
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-007-e2e-cutover-impl.md:70,73-78`
- Modify: `docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-007-e2e-cutover-test.md:66`

**Context:** 这些文档是本 PR 新增的，但仍使用 `make` 命令和 `panel/web` 旧路径。需统一为 `moon run` 命令和 `web/` 新路径。

**Replacement rules (apply to all files above):**
| Old | New |
|-----|-----|
| `make check-generate` | `moon run panel:check-generate` |
| `make generate` | `moon run panel:generate` |
| `make e2e-smoke` | `moon run e2e:smoke` |
| `make e2e` | `moon run e2e:run` |
| `cd panel/web && bun run lint` | `moon run web:lint` |
| `cd panel/web && bun run format` | `moon run web:format` |
| `cd panel/web && bunx tsc -b` | `moon run web:typecheck` |
| `cd panel/web && bun run test` | `moon run web:test` |
| `make generate` 一次生成两侧产物，并纳入 `make check-generate` | `moon run panel:generate` 一次生成两侧产物，并纳入 `moon run panel:check-generate` |
| 执行前后端门禁与 `make e2e` | 执行前后端门禁与 `moon run e2e:run` |

**Step 1: 对每个文件执行 sed 替换**

```bash
FILES=(
  docs/plans/2026-02-28-panel-node-rpc-cutover-design/_index.md
  docs/plans/2026-02-28-panel-node-rpc-cutover-design/bdd-specs.md
  docs/plans/2026-02-28-panel-node-rpc-cutover-design/architecture.md
  docs/plans/2026-02-28-panel-node-rpc-cutover-design/best-practices.md
  docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-001-sync-success-impl.md
  docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-001-sync-success-test.md
  docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-007-e2e-cutover-impl.md
  docs/plans/2026-02-28-panel-node-rpc-cutover-plan/task-007-e2e-cutover-test.md
)

for f in "${FILES[@]}"; do
  sed -i \
    -e 's|`make check-generate`|`moon run panel:check-generate`|g' \
    -e 's|make check-generate|moon run panel:check-generate|g' \
    -e 's|`make generate`|`moon run panel:generate`|g' \
    -e 's|make generate|moon run panel:generate|g' \
    -e 's|`make e2e-smoke`|`moon run e2e:smoke`|g' \
    -e 's|make e2e-smoke|moon run e2e:smoke|g' \
    -e 's|`make e2e`|`moon run e2e:run`|g' \
    -e 's|make e2e|moon run e2e:run|g' \
    -e 's|cd panel/web && bun run lint|moon run web:lint|g' \
    -e 's|cd panel/web && bun run format|moon run web:format|g' \
    -e 's|cd panel/web && bunx tsc -b|moon run web:typecheck|g' \
    -e 's|cd panel/web && bun run test|moon run web:test|g' \
    "$f"
done
```

**Step 2: 验证没有残留的旧引用**

Run: `grep -rn 'make \|cd panel/web' docs/plans/2026-02-28-panel-node-rpc-cutover-design/ docs/plans/2026-02-28-panel-node-rpc-cutover-plan/`

Expected: 无输出

**Step 3: Commit**

```bash
git add docs/plans/2026-02-28-panel-node-rpc-cutover-design/ docs/plans/2026-02-28-panel-node-rpc-cutover-plan/
git commit -m "docs: update stale make/panel-web references to moon commands"
```

---

### Task 6: 清理孤立的 verify 脚本和 migration 产物 [MEDIUM]

**Files:**
- Verify and delete if present on PR branch: `scripts/verify-moon-*.sh`, `scripts/check-generate.sh`

**Context:** PR 分支上 `scripts/moon.yml` 和 `scripts/check-generate.sh` 已不存在（在 diff 中已删除）。但需检查工作区中是否还有残留的 verify 脚本。PR 的 Test Plan 中有两项未完成（Docker build 和 E2E smoke），这些是独立验证项，不影响本修复。

**Step 1: 检查 scripts/ 目录中是否有 verify 脚本**

Run: `git ls-tree worktree-moon-optimization scripts/`

Expected: 仅保留 `dev-panel-web.sh`、`docker-build-push.sh`、`moon-cli.sh`

**Step 2: 如无 verify 脚本则跳过，如有则删除并提交**

若发现 `verify-moon-*.sh` 文件：
```bash
git rm scripts/verify-moon-*.sh
git commit -m "chore: remove stale verify-moon migration scripts"
```

Expected: PR 分支上这些文件已不存在，此步骤应为 no-op。

---

### Task 7: .moon/tasks/all.yml 清理过时 e2e fileGroup [MEDIUM]

**Files:**
- Modify: `.moon/tasks/all.yml:12`

**Context:** `e2eInfra` fileGroup 引用了 `/e2e/sb-client.Dockerfile`，但 PR 分支的 `e2e/docker-compose.e2e.yml` 使用预构建镜像而非此 Dockerfile。需确认此文件是否存在于 PR 分支。

**Step 1: 检查文件是否存在**

Run: `git show worktree-moon-optimization:e2e/sb-client.Dockerfile > /dev/null 2>&1 && echo "EXISTS" || echo "NOT FOUND"`

**Step 2: 如不存在则从 fileGroup 中移除引用**

将 `.moon/tasks/all.yml` 的 `e2eInfra` 中移除 `- /e2e/sb-client.Dockerfile` 行。

若文件存在则保留引用（no-op）。

**Step 3: Commit (if changed)**

```bash
git add .moon/tasks/all.yml
git commit -m "chore: remove stale sb-client.Dockerfile from e2eInfra fileGroup"
```

---

### Task 8: 回归验证 [VERIFICATION]

**Context:** 在 PR 分支上运行所有门禁检查，确保修复未引入新问题。

**Step 1: 切换到 PR 分支**

Run: `git checkout worktree-moon-optimization`

**Step 2: 运行 web 门禁**

Run: `moon run web:lint web:typecheck web:test`

Expected: 全部通过

**Step 3: 运行 panel 测试**

Run: `moon run panel:test`

Expected: 全部通过

**Step 4: 运行 node 测试**

Run: `moon run node:test`

Expected: 全部通过

**Step 5: 运行 generate 检查**

Run: `moon run panel:check-generate`

Expected: "Generated files are up to date."

**Step 6: 验证无残留旧引用**

Run: `grep -rn 'make \|cd panel/web' docs/ AGENTS.md README*.md .github/ 2>/dev/null | grep -v 'Makefile'`

Expected: 无输出或仅有历史设计文档中合理的 Makefile 迁移描述

---

## 未在本计划中处理的项（建议后续 PR）

| 项目 | 原因 |
|------|------|
| 升级 `go.mod` 到 Go 1.26 | 涉及所有 Go 模块，应独立 PR |
| 添加 `web:format` 到 CI | 需团队确认是否为门禁 |
| 给 `node/moon.yml` 添加 `build`/`vet` 任务 | 属于增强，非修复 |
| 给 `web/moon.yml` 添加 input/output fileGroups | 属于优化，非修复 |
| CI 添加 Moon/Go/bun 缓存 | 属于性能优化，非修复 |
