# Best Practices

## Principles

- 每个 Moon 项目对应一个明确的语义边界
- 任务直属其所在项目，不通过 alias 层转发
- 版本只有一个来源：`.prototools`
- Docker 构建链服务于目录结构，而非反过来

## Task Design Rules

### Tasks Belong to Their Project

每个任务都应该属于其自然所属的项目：

- `generate`、`check-generate`、`test` → `panel`
- `lint`、`format`、`typecheck`、`test` → `web`
- `test` → `node`
- `run`、`smoke`、`down`、`report` → `e2e`

不要创建 alias 项目来转发任务。

### Declare Toolchain Per Project

Moon v2 支持项目级 `toolchain.default` 声明：

- `web`：`toolchain.default: bun`
- `panel`、`node`：默认使用 system（Go）

### Disable Cache for Stateful Tasks

以下任务必须禁用缓存：

- 代码生成（`generate`、`check-generate`）
- Go 测试（带 `-count=1`）
- E2E 测试（依赖 Docker）

## Docker Build Rules

### Use Workspace Root as Context for Cross-Project Builds

当 Dockerfile 需要引用多个顶层目录时（如 panel + web），构建上下文应为工作区根。

### Keep Single-Project Builds Simple

当 Dockerfile 只需要自身目录时（如 node），构建上下文保持为项目目录。

### Add .dockerignore

排除不需要进入构建上下文的目录：

```
.git
.moon
docs
e2e
scripts
tasks
*.md
.prototools
go.work*
```

## Version Management

### Single Source of Truth

`.prototools` 是唯一的工具版本源。其他引用该版本的地方应动态读取而非硬编码。

### moon-cli.sh Pattern

Fallback 脚本应从 `.prototools` 解析版本：

```bash
MOON_VERSION="$(grep '^moon = ' .prototools | sed 's/moon = "\(.*\)"/\1/')"
```

## CI Rules

### Separate Gates from Publishing

- `ci.yml`：门禁检查（快速、每次 PR）
- `docker-publish.yml`：镜像发布（仅主分支和 tag）

### Use moonrepo/setup-toolchain

CI 中使用官方 `moonrepo/setup-toolchain@v0` action 安装 moon 和 proto，确保版本与 `.prototools` 一致。

## Documentation Hygiene

迁移完成后，所有开发者文档中的命令示例必须更新为新入口。

优先更新：

- `AGENTS.md`
- `README.zh.md` / `README.en.md`

历史设计文档可保留原始上下文，不要求全量回写。
