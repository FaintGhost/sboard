# Best Practices

## Principles

- 只替换编排层，不顺带重构业务逻辑
- 版本只有一个来源：`.prototools`
- 任务语义优先于“看起来更现代”的抽象
- 先做等价迁移，再做体验优化

## Tooling Recommendations

### Pin the CLI and Language Runtimes

严格锁定：

- `moon`
- `go`
- `node`
- `bun`

不要把版本同时写在多个地方作为事实来源。

可以存在的情况：

- `go.work` 继续声明语言级最低版本
- `package.json` 继续声明依赖

但它们不应替代 `.prototools` 的“精确版本”角色。

### Keep Supporting Generators Explicit

虽然本次主目标是 `Go + Node + Bun`，但 `check-generate` 仍依赖外部工具：

- `buf`
- `protoc-gen-go`
- `protoc-gen-connect-go`
- `@bufbuild/protoc-gen-es`
- `@connectrpc/protoc-gen-connect-query`

建议：

- JS 侧继续由 `package.json` 锁定
- `buf` 作为紧邻生成链路的关键工具，建议也纳入后续 `.prototools` pin

否则“主工具链已锁定，但生成仍会漂移”。

## Task Design Rules

### Prefer Thin Wrappers

Moon task 应尽量做薄封装：

- 复用当前已经通过验证的命令
- 不在第一阶段重写生成逻辑或 Compose 编排
- 保留原始 stdout/stderr

错误做法：

- 为了“纯 Moon 化”把现有稳定 shell 逻辑拆成多段隐式步骤

### Separate Read-only and Mutating Tasks

当前 `bun run format` 与 `check-generate` 都可能改工作区。

应区分：

- 本地修复型任务
- CI 断言型任务

否则会出现：

- 任务成功，但工作区被改脏
- 缓存命中掩盖真实文件变化

### Disable Cache for Stateful Tasks

以下任务必须禁用缓存：

- 生成相关
- 依赖 `git diff` 的任务
- 所有 Docker E2E
- 所有会写工作区的任务

缓存只应该用于真正纯函数、可重复的静态任务。

## Migration Controls

### Keep a One-to-One Mapping Table

迁移文档中必须保留旧入口到新入口的映射表。

作用：

- 降低迁移期沟通成本
- 便于用户把旧 muscle memory 快速映射到新命令
- 方便排障时做 A/B 对照

### Validate in This Order

推荐验证顺序：

1. `automation:generate`
2. `automation:check-generate`
3. `automation:e2e-smoke`
4. `automation:e2e`

理由：

- 先验证最便宜的静态任务
- 再验证最昂贵的 Docker 编排
- 能更快暴露环境/配置问题

### Protect Current Product Boundaries

Moon 迁移后，以下断言必须继续成立：

- Panel 管理面默认走 RPC
- Node 旧控制面 REST 不重新暴露
- 订阅 REST `GET /api/sub/:user_uuid` 保持可用

这些不是构建系统问题，但必须作为回归保护项存在。

## Rollback Strategy

- 如果 Moon 配置导致根任务失败，可临时回退到直接执行底层命令
- 回滚应只影响执行入口，不影响业务代码
- 不应把 Moon 迁移和其他架构改动绑在同一个提交里

## Documentation Hygiene

迁移完成后应优先更新：

- [README.zh.md](/root/workspace/sboard/README.zh.md)
- [README.en.md](/root/workspace/sboard/README.en.md)
- [AGENTS.md](/root/workspace/sboard/AGENTS.md)

历史设计文档可保留原始上下文，不要求全部回写，但应避免继续把旧 `make` 示例当成“当前推荐路径”。
