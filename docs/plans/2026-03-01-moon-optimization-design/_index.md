# Moon Monorepo Optimization Design

## Context

仓库已从 Makefile 迁移到 Moon v2 作为任务编排层。经审查发现以下结构与模式问题需要优化：

- `automation` 项目映射到 `scripts/` 语义混乱，alias 层增加间接性
- `panel/web` 嵌套在 `panel/` 下导致前端 Moon 任务需要 `bash -lc "cd web && ..."` 反模式
- 7 个 `verify-*.sh` 迁移验证脚本残留
- `moon-cli.sh` 硬编码版本与 `.prototools` 构成双源
- `node` 项目无任何任务
- 无 CI 门禁工作流
- `dev-panel-web.sh` 仍用 `npm`
- fileGroups 已定义但未被任务引用

## Requirements

### Goals

- 将 `panel/web` 提升为顶层 `web/` 目录，成为独立 Moon 项目
- 取消 `automation` alias 层，任务回归所属项目
- 修复工具链版本双源
- 补充 `node` 和 `panel` 的 Go 测试任务
- 新增 CI 门禁工作流
- 清理迁移残留物

### Non-Goals

- 不改变 RPC / REST 契约
- 不改变 Proto 文件组织（仍在 `panel/proto/`）
- 不改变业务逻辑
- 不改变 `go.work` 结构

### Protected Boundaries

- Panel 管理面仍以 RPC 为主
- 订阅 REST `GET /api/sub/:user_uuid` 保持不变
- Docker 发布流程功能等价
- E2E 测试行为不变

## Decision

采用「取消 alias 层 + web 提升顶层 + 任务直属项目」方案：

- 取消 `automation` 项目，删除 `scripts/moon.yml`
- 将 `panel/web/` 移动到顶层 `web/`，注册为独立 Moon 项目
- 所有任务直属其所在项目（`panel:generate`、`web:lint`、`node:test` 等）
- Docker 构建上下文改为工作区根目录以支持跨目录引用
- 新增 CI 门禁工作流 `ci.yml`

## Detailed Design

### Directory Structure

目标态：

```text
.moon/
  workspace.yml       # 4 项目：panel, web, node, e2e
  toolchains.yml
  tasks/all.yml       # fileGroups
panel/                # Go 后端
  cmd/panel/
  internal/
  proto/
  buf.yaml
  buf.gen.yaml
  buf.node.gen.yaml
  check-generate.sh   # 从 scripts/ 迁入
  moon.yml
  Dockerfile
  go.mod
web/                  # 前端（从 panel/web/ 提升）
  src/
  package.json
  bun.lock
  moon.yml            # 新增
node/                 # Node 服务
  cmd/node/
  internal/
  moon.yml            # 补充 test 任务
  Dockerfile
  go.mod
e2e/                  # E2E 测试
  moon.yml            # 保持不变
scripts/              # 工具脚本（不再是 moon 项目）
  moon-cli.sh         # 修复版本双源
  dev-panel-web.sh    # npm → bun
  docker-build-push.sh
```

### Workspace Configuration

`.moon/workspace.yml`：

```yaml
projects:
  panel: panel
  web: web
  node: node
  e2e: e2e
```

不再设 `defaultProject`，不再有 `automation` 项目。

### Task Definitions

`panel/moon.yml`：

```yaml
tasks:
  generate:
    command: go
    args: [generate, ./internal/rpc/...]
    options:
      cache: false

  check-generate:
    deps: [generate]
    command: bash
    args: [check-generate.sh]
    options:
      cache: false
      runFromWorkspaceRoot: true

  test:
    command: go
    args: [test, ./..., -count=1]
    options:
      cache: false
```

`web/moon.yml`：

```yaml
toolchain:
  default: bun

tasks:
  lint:
    command: bun
    args: [run, lint]

  format:
    command: bun
    args: [run, format]

  typecheck:
    command: bunx
    args: [tsc, -b]

  test:
    command: bun
    args: [run, test]
```

`node/moon.yml`：

```yaml
tasks:
  test:
    command: go
    args: [test, ./..., -count=1]
    options:
      cache: false
```

`e2e/moon.yml`：保持不变。

### Task Entry Mapping

| 旧入口 | 新入口 |
|--------|--------|
| `moon run automation:generate` | `moon run panel:generate` |
| `moon run automation:check-generate` | `moon run panel:check-generate` |
| `moon run automation:e2e` | `moon run e2e:run` |
| `moon run automation:e2e-smoke` | `moon run e2e:smoke` |
| `moon run automation:e2e-down` | `moon run e2e:down` |
| `moon run automation:e2e-report` | `moon run e2e:report` |
| `moon run panel:web-lint` | `moon run web:lint` |
| `moon run panel:web-typecheck` | `moon run web:typecheck` |
| `moon run panel:web-test` | `moon run web:test` |
| — | `moon run panel:test`（新增） |
| — | `moon run node:test`（新增） |

### Toolchain Fix

`scripts/moon-cli.sh` 从 `.prototools` 动态读取版本，消除硬编码双源：

```bash
MOON_VERSION="$(grep '^moon = ' .prototools | sed 's/moon = "\(.*\)"/\1/')"
exec bunx "@moonrepo/cli@${MOON_VERSION}" "$@"
```

`scripts/dev-panel-web.sh` 将 `npm` 替换为 `bun`，路径从 `panel/web` 更新为 `web`。

### Docker Build Adaptation

`panel/Dockerfile` 构建上下文从 `panel/` 改为工作区根：

- web 阶段：`COPY web/package.json` — 从顶层 `web/` 拷贝
- builder 阶段：`COPY panel/go.mod ./` 和 `COPY panel/ .` — 指定 panel 子目录

受影响的构建上下文配置：

| 文件 | 旧 context | 新 context |
|------|-----------|-----------|
| `panel/docker-compose.build.yml` | `.` | `..` |
| `e2e/docker-compose.e2e.yml` (panel) | `../panel` | `..` |
| `.github/workflows/docker-publish.yml` (panel) | `./panel` | `.` |

### buf Generation Path Adaptation

`panel/buf.gen.yaml` TS 插件路径更新：

| 字段 | 旧值 | 新值 |
|------|------|------|
| protoc-gen-es local | `web/node_modules/.bin/protoc-gen-es` | `../web/node_modules/.bin/protoc-gen-es` |
| protoc-gen-es out | `web/src/lib/rpc/gen` | `../web/src/lib/rpc/gen` |
| protoc-gen-connect-query local | `web/node_modules/.bin/protoc-gen-connect-query` | `../web/node_modules/.bin/protoc-gen-connect-query` |
| protoc-gen-connect-query out | `web/src/lib/rpc/gen` | `../web/src/lib/rpc/gen` |

`panel/check-generate.sh` 路径更新：

```bash
git diff --exit-code -- 'panel/internal/rpc/gen/**' 'web/src/lib/rpc/gen/**' 'node/internal/rpc/gen/**'
```

### CI Gate Workflow

新增 `.github/workflows/ci.yml`：

```yaml
name: ci
on:
  push:
    branches: [main, master]
  pull_request:

jobs:
  gates:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - run: moon run panel:check-generate
      - run: moon run web:lint
      - run: moon run web:typecheck
      - run: moon run web:test
      - run: moon run panel:test
      - run: moon run node:test
```

### File Cleanup

删除：

- `scripts/moon.yml`（automation 项目定义）
- `scripts/verify-moon-workspace.sh`
- `scripts/verify-moon-generate.sh`
- `scripts/verify-moon-check-generate.sh`
- `scripts/verify-moon-e2e.sh`
- `scripts/verify-moon-panel-gates.sh`
- `scripts/verify-moon-toolchain.sh`
- `scripts/verify-no-make-entrypoints.sh`
- `findings.md`（根目录旧规划残留）
- `progress.md`（根目录旧规划残留）
- `task_plan.md`（根目录旧规划残留）

文档更新：

- `AGENTS.md`：更新项目结构、任务入口、前端门禁
- `README.zh.md` / `README.en.md`：更新目录结构、开发命令
- `.moon/tasks/all.yml`：更新 fileGroups 路径

## Trade-offs

### Pros

- 每个项目语义清晰，无间接层
- 前端任务原生运行在 `web/` 目录，无 bash cd 反模式
- 版本单一来源，无硬编码风险
- CI 门禁自动化
- 符合常见 monorepo 最佳实践

### Cons

- 目录移动是有破坏性的 — 涉及 Docker、buf、CI 多处路径更新
- Docker 构建上下文变大（整个工作区而非单个子目录），但不影响镜像大小
- 团队需要适应新的任务入口命名

### Rejected Alternatives

- 保留 automation alias 层但映射到根目录
  - 不符合"取消 alias 层"决定
- 仅注册 `panel-web` 为 `panel/web` 的项目（不移动目录）
  - 不符合"同级目录"结构偏好

## Success Criteria

- `moon run panel:generate` 等价于旧 `moon run automation:generate`
- `moon run panel:check-generate` 退出码与副作用一致
- `moon run web:lint`、`web:typecheck`、`web:test` 等价于 `cd panel/web && bun run ...`
- `panel:test` 和 `node:test` 可通过
- Docker 构建成功（panel + node 镜像）
- E2E smoke 回归通过
- CI 工作流在 PR 上触发并通过
- 仓库内无 `make` 或 `automation:` 残留引用

## Design Documents

- [BDD Specifications](./bdd-specs.md)
- [Architecture](./architecture.md)
- [Best Practices](./best-practices.md)
