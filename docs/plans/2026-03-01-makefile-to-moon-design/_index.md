# Makefile to Moon Design

## Context

当前仓库根级任务仍由 [Makefile](/root/workspace/sboard/Makefile) 管理，覆盖：

- 代码生成：`generate`、`check-generate`
- E2E：`e2e`、`e2e-smoke`、`e2e-down`、`e2e-report`

而子项目已经各自拥有独立执行入口：

- `panel/web` 通过 `package.json` 承载前端脚本
- `panel/internal/rpc/generate.go` 通过 `go generate` 驱动 RPC 生成
- `e2e` 通过 `docker compose` + Playwright 承载全链路门禁

用户已确认本次方向：

- 完全弃用 `Makefile`
- 迁移到 Moon
- 由 Moon 全量接管 `Go + Node + Bun` 工具链
- 工具版本严格锁定

## Requirements

### Goals

- 建立统一入口：开发、生成、校验、测试、E2E 全部改为 `moon run ...`
- 用严格锁版本消除环境漂移
- 迁移仅替换执行编排，不改变业务行为与协议边界
- 保持当前门禁语义、退出码、工作目录与关键副作用一致

### Non-Goals

- 不改 Panel / Node 业务逻辑
- 不改 RPC / REST 契约
- 不推进订阅 REST 的去兼容化
- 不顺带重构数据库、部署模型或 CI 以外的产品功能

### Protected Boundaries

- Panel 管理面仍以 RPC 为主
- 订阅 REST `GET /api/sub/:user_uuid` 继续保留
- Node 对外 REST 接口现状保持不变

## Decision

采用「Moon 作为唯一任务编排入口 + Proto 作为严格版本源」的迁移方案：

- 在仓库根新增 `.moon/` 工作区配置，建立 `automation`、`panel`、`node`、`e2e` 四个项目
- 在仓库根新增 `.prototools`，精确锁定：
  - `moon = "2.0.0"`
  - `go = "1.26.0"`
  - `node = "22.22.0"`
  - `bun = "1.3.9"`
- 由 `automation` 项目暴露与旧 `Makefile` 同名的顶层任务别名，统一映射到真实项目任务
- 删除根 `Makefile`，同步更新文档与后续 CI 调用入口

选择 `moon 2.0.0` 的原因：

- Moon v2 已在 2026-02-18 正式发布
- 当前官方配置文档已明确面向 Moon v2
- Moon v1 文档已冻结，不适合作为新迁移基线

## Detailed Design

### Workspace Layout

新增以下配置结构：

```text
.prototools
.moon/
  workspace.yml
  toolchains.yml
  tasks/
    all.yml
scripts/
  moon.yml
panel/
  moon.yml
node/
  moon.yml
e2e/
  moon.yml
```

### Project Responsibilities

- `automation`
  - 仓库根默认项目
  - 暴露顶层命令别名：`generate`、`check-generate`、`e2e`、`e2e-smoke`、`e2e-down`、`e2e-report`
- `panel`
  - 承载 RPC 代码生成与后续 Go/前端复合任务
- `node`
  - 先建立占位项目，后续接入 Go build/test 任务
- `e2e`
  - 承载 Docker Compose + Playwright 生命周期任务

### Toolchain Strategy

- 版本锁定由 `.prototools` 负责，作为仓库唯一版本源
- `.moon/toolchains.yml` 只负责启用语言工具链与执行语义
- `go.work` 中的 `go 1.25` 继续表示语言级最低要求，不替代工具链精确版本锁定

### Task Mapping

| 旧入口 | 新入口 | 实际执行方 |
|---|---|---|
| `make generate` | `moon run automation:generate` | `panel:generate-rpc` |
| `make check-generate` | `moon run automation:check-generate` | `automation`（依赖 `generate`） |
| `make e2e` | `moon run automation:e2e` | `e2e:run` |
| `make e2e-smoke` | `moon run automation:e2e-smoke` | `e2e:smoke` |
| `make e2e-down` | `moon run automation:e2e-down` | `e2e:down` |
| `make e2e-report` | `moon run automation:e2e-report` | `e2e:report` |

### Migration Phases

1. 新增 `.prototools` 与 `.moon/*`，但暂不删 `Makefile`
2. 让 Moon 与旧命令并行对照，确认行为等价
3. 更新所有文档、脚本、门禁说明为 `moon run ...`
4. 删除根 `Makefile`
5. 后续再把更多子项目任务（如 `panel/web lint/test`、`go test`）显式纳入 Moon 图谱

## Trade-offs

### Pros

- 单一入口，减少根级 Bash 拼接
- 统一版本控制，降低“我本地可以跑”的差异
- 任务图、依赖和后续 CI 可视化更清晰

### Cons

- Moon v2 的团队迁移成本高于仅保留 Makefile，但可一次完成长期收敛
- 需要同步更新大量文档中的 `make` 示例
- `check-generate` 与 E2E 都不是纯函数任务，Moon 缓存必须谨慎配置

### Rejected Alternatives

- 继续保留 `Makefile` 仅做转发
  - 不符合“完全弃用 Makefile”
- 只迁移任务、不接管工具链
  - 不符合“全量接管 Go + Node + Bun”
- 基于 Moon v1
  - 当前官方配置文档已转向 v2，继续用 v1 会形成二次迁移成本

## Success Criteria

- 仓库内无开发文档再要求执行 `make ...`
- 根级任务均可通过 `moon run automation:<task>` 完成
- `.prototools` 成为唯一版本源，团队环境版本一致
- `check-generate`、`e2e` 的行为与迁移前等价
- 订阅 REST 与当前 RPC 边界在回归中保持不变

## Sources

- Moon config docs（v2）：https://moonrepo.dev/docs/config
- Moon workspace docs：https://moonrepo.dev/docs/config/workspace
- Moon tasks docs：https://moonrepo.dev/docs/config/tasks
- Moon project docs：https://moonrepo.dev/docs/config/project
- Moon toolchain setup：https://moonrepo.dev/docs/setup-toolchain
- Moon install docs（含 `.prototools` pin 示例）：https://moonrepo.dev/docs/install
- Moon v2 发布公告：https://moonrepo.dev/blog/moon-v2
- Proto version detection：https://moonrepo.dev/docs/proto/detection

## Design Documents

- [BDD Specifications](./bdd-specs.md) - Behavior scenarios and testing strategy
- [Architecture](./architecture.md) - System architecture and component details
- [Best Practices](./best-practices.md) - Security, performance, and code quality guidelines
