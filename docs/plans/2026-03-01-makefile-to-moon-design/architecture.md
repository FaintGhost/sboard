# Architecture

## Overview

本次迁移不替换业务代码，只替换仓库根级任务编排与开发机工具链管理。

核心架构分为两层：

1. 版本层：由 `.prototools` 严格锁定工具版本
2. 编排层：由 `.moon/` 和各项目 `moon.yml` 建立任务图

## Version Layer

### Single Source of Truth

新增 `.prototools` 作为版本单一来源，建议内容：

```toml
moon = "2.0.0"
go = "1.26.0"
node = "22.22.0"
bun = "1.3.9"
```

说明：

- 这些版本来自当前工作环境的已验证版本，能减少首轮迁移摩擦
- `go.work` 的 `go 1.25` 继续保留，表示语言兼容下限，不与工具链精确版本冲突

### Why `.prototools`

- Moon 自身通过 proto 集成安装与版本解析
- proto 会优先从本地 `.prototools` 检测版本
- 这比“团队约定手工安装某版本”更可审计

## Orchestration Layer

### Workspace Root

新增 `.moon/workspace.yml`：

- 显式注册四个项目：
  - `automation` -> `scripts`
  - `panel` -> `panel`
  - `node` -> `node`
  - `e2e` -> `e2e`
- `defaultProject` 设为 `automation`
- 保持仓库根任务入口最短路径

### Shared Task Defaults

新增 `.moon/tasks/all.yml`：

- 定义共享 `fileGroups`
- 为后续的 affected/caching 提供输入边界
- 不在这里直接放复杂脚本，避免全局继承误伤

建议 file groups：

```yaml
fileGroups:
  rpcSpec:
    - /panel/proto/**/*.proto
    - /panel/buf*.yaml
  rpcGenerated:
    - /panel/internal/rpc/gen/**
    - /panel/web/src/lib/rpc/gen/**
    - /node/internal/rpc/gen/**
  e2eInfra:
    - /e2e/docker-compose.e2e.yml
    - /e2e/Dockerfile
    - /e2e/sb-client.Dockerfile
    - /e2e/entrypoint.sh
```

### Toolchain Enablement

新增 `.moon/toolchains.yml`：

```yaml
javascript:
  packageManager: bun

bun: {}
go: {}
node: {}
```

设计要点：

- 让 Moon 按语言上下文执行任务
- 不把版本写进 `.moon/toolchains.yml`
- 版本统一继续由 `.prototools` 管理，避免双重来源

## Project Design

### `automation` Project

职责：

- 提供根级别别名任务
- 统一承接旧 `Makefile` 的用户心智

建议任务：

- `generate`
- `check-generate`
- `e2e`
- `e2e-smoke`
- `e2e-down`
- `e2e-report`

其中：

- `generate` 只依赖 `panel:generate-rpc`
- `check-generate` 依赖 `generate` 后执行 `git diff --exit-code`
- `e2e*` 只做别名，不复制 Compose 逻辑

### `panel` Project

首期只承载：

- `generate-rpc`

后续可扩展：

- `web-lint`
- `web-format`
- `web-typecheck`
- `web-test`
- `test-go`

但不要求在第一阶段全部完成，否则迁移面会过大。

### `node` Project

第一阶段仅建立项目边界，不强制接入任务。

原因：

- 当前根 `Makefile` 并未管理 node 的测试或构建
- 先完成“替换现有根入口”更符合最小迁移原则

### `e2e` Project

承载 Docker Compose 与 Playwright 逻辑：

- `run`
- `smoke`
- `down`
- `report`

这里必须保留当前 shell 语义：

- 先清理
- 再启动
- 不论成功失败都收尾

这是当前 `Makefile` 中最容易被错误“抽象化”后失真的部分。

## Cache Policy

### Non-cacheable Tasks

以下任务必须默认禁用缓存：

- `automation:generate`
- `automation:check-generate`
- `e2e:run`
- `e2e:smoke`
- `e2e:down`
- `e2e:report`

原因：

- 会写工作区
- 依赖 git 状态
- 依赖 Docker 外部环境
- 输出受运行时状态影响，不是纯函数

### Future Cacheable Tasks

以下任务在第二阶段可考虑缓存：

- `panel:web-lint`
- `panel:web-typecheck`
- `panel:web-test`

前提是明确输入、输出和工作目录，并确认不写入工作区。

## Migration Sequence

### Phase 1: Introduce

- 新增 `.prototools`
- 新增 `.moon/*`
- 新增各 project 的 `moon.yml`
- 暂不删除 `Makefile`

### Phase 2: Prove Equivalence

- 逐个对照验证 6 个根任务
- 确认退出码、副作用、日志主语义一致

### Phase 3: Cut Over

- 更新 README、AGENTS、交付门禁说明
- 删除根 `Makefile`
- 将所有根入口切为 `moon run automation:<task>`

### Phase 4: Expand

- 再把前端门禁和 Go test 显式纳入 Moon
- 视需要补 CI 工作流

## Risks

### Moon v2 Baseline Risk

Moon v2 已于 2026-02-18 发布正式版本，当前配置文档也面向 v2。

缓解：

- 在 `.prototools` 严格 pin 版本
- 文档中明确“按仓库锁定版本执行”
- 不使用尚未在仓库需要的高级特性

### Documentation Drift

当前仓库大量文档仍引用 `make ...`。

缓解：

- 迁移时一并更新 README / AGENTS / docs/plans 中仍面向开发者的操作说明
- 历史设计文档可不全量回写，只在新文档里说明“旧示例已过时”
