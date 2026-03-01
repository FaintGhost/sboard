# BDD Specifications

## Feature: Moon Replace Root Makefile

### Scenario 1: Generate command remains behaviorally equivalent

```gherkin
Given panel RPC proto and buf config are valid
When 执行 `moon run automation:generate`
Then 它应执行与原 `make generate` 等价的生成流程
And `panel/internal/rpc/gen`
And `panel/web/src/lib/rpc/gen`
And `node/internal/rpc/gen`
Should reflect freshly generated artifacts
```

### Scenario 2: Check-generate succeeds when artifacts are fresh

```gherkin
Given 所有生成产物已经与 proto 规格同步
When 执行 `moon run automation:check-generate`
Then 任务退出码应为 0
And 输出应包含 “Generated files are up to date.”
```

### Scenario 3: Check-generate fails when artifacts are stale

```gherkin
Given proto 规格已变化但生成产物未同步
When 执行 `moon run automation:check-generate`
Then 任务退出码应为非 0
And 输出应包含 “Generated files are out of date.”
And 失败原因应是 git diff 检测到生成目录变化
```

### Scenario 4: E2E run preserves cleanup semantics

```gherkin
Given Docker daemon 和 docker compose 可用
When 执行 `moon run automation:e2e`
Then 它应先执行 compose down -v
And 再执行 compose up --build --abort-on-container-exit --exit-code-from playwright
And 最终无论成功失败都执行 compose down -v
```

### Scenario 5: E2E smoke remains cheaper than full E2E

```gherkin
Given Docker daemon 和 docker compose 可用
When 执行 `moon run automation:e2e-smoke`
Then 它只应启动 panel 和 node 基础服务
And 应执行 Playwright smoke project
And 不应替代 full e2e 的全部覆盖面
```

### Scenario 6: Report task remains local-only

```gherkin
Given 已存在 Playwright report 目录
When 执行 `moon run automation:e2e-report`
Then 它应调用 `bunx playwright show-report playwright-report`
And 在 CI 中默认不运行
```

### Scenario 7: Toolchain versions are strictly pinned

```gherkin
Given 仓库根存在 `.prototools`
When 开发者在仓库内执行 `proto use`
Then `moon`、`go`、`node`、`bun` 均应解析到仓库锁定版本
And 任务执行不得依赖系统全局版本漂移
```

### Scenario 8: Moon becomes the only documented root entrypoint

```gherkin
Given 迁移完成
When 开发者阅读 README 或项目交付门禁说明
Then 根级任务示例应全部使用 `moon run ...`
And 不再把 `make ...` 作为默认入口
```

### Scenario 9: Existing RPC/REST boundaries are not changed

```gherkin
Given 当前系统已完成 Panel 管理面 RPC 化
And 订阅 REST 仍作为兼容入口保留
When 执行本次 Moon 迁移
Then 迁移后 `GET /api/sub/:user_uuid` 行为不得改变
And Node RPC cutover 边界断言必须继续通过
```

### Scenario 10: Frontend gates can be lifted into Moon without semantic drift

```gherkin
Given `panel/web` 仍以 bun scripts 承载 lint/format/test
When 后续执行 `moon run panel:web-lint`、`panel:web-typecheck`、`panel:web-test`
Then 它们应分别等价于 `bun run lint`、`bunx tsc -b`、`bun run test`
And 不得改变各自退出码与错误输出结构
```

## Verification Matrix

- 任务等价性
  - 对比 `make <target>` 与 `moon run automation:<target>` 的退出码与副作用
- 工具链一致性
  - 对比 `go version`、`node --version`、`bun --version`
- 回归保护
  - `moon run automation:check-generate`
  - `moon run automation:e2e-smoke`
  - `moon run automation:e2e`
