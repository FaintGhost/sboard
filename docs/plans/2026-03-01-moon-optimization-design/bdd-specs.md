# BDD Specifications

## Feature: Moon Monorepo Optimization

### Scenario 1: web directory promoted to top level

```gherkin
Given panel/web/ 已移动到顶层 web/
When 执行 `moon run web:lint`
Then 任务应在 web/ 目录下执行 `bun run lint`
And 退出码应与旧 `cd panel/web && bun run lint` 等价
```

### Scenario 2: panel:generate replaces automation:generate

```gherkin
Given automation 项目已删除
When 执行 `moon run panel:generate`
Then 应调用 `go generate ./internal/rpc/...`
And panel/internal/rpc/gen、web/src/lib/rpc/gen、node/internal/rpc/gen 应反映最新生成产物
```

### Scenario 3: panel:check-generate detects stale artifacts

```gherkin
Given proto 规格已变化但生成产物未同步
When 执行 `moon run panel:check-generate`
Then 退出码应为非 0
And 输出应包含 "Generated files are out of date."
```

### Scenario 4: panel:check-generate passes with fresh artifacts

```gherkin
Given 所有生成产物已与 proto 同步
When 执行 `moon run panel:check-generate`
Then 退出码应为 0
And 输出应包含 "Generated files are up to date."
```

### Scenario 5: Docker panel build succeeds with workspace root context

```gherkin
Given panel/Dockerfile 使用工作区根作为构建上下文
When 执行 `docker compose -f panel/docker-compose.build.yml build`
Then 构建应成功
And 镜像应包含 Go 二进制 + web 静态资源
```

### Scenario 6: E2E smoke passes after restructure

```gherkin
Given 目录已重构
When 执行 `moon run e2e:smoke`
Then 应与重构前行为等价
And 所有 smoke 用例通过
```

### Scenario 7: moon-cli.sh reads version from .prototools

```gherkin
Given .prototools 中 moon 版本为 "2.0.3"
When moon-cli.sh 执行 fallback 路径
Then 应调用 `bunx @moonrepo/cli@2.0.3`（版本与 .prototools 一致）
And 不应出现硬编码版本
```

### Scenario 8: node project has working test task

```gherkin
Given node/moon.yml 定义了 test 任务
When 执行 `moon run node:test`
Then 应执行 `go test ./... -count=1`
And 退出码应反映测试结果
```

### Scenario 9: CI gate workflow runs on PR

```gherkin
Given .github/workflows/ci.yml 已配置
When 提交 PR 到 main/master
Then CI 应运行 check-generate、lint、typecheck、test 门禁
And 任一失败应阻止合并
```

### Scenario 10: No legacy references remain

```gherkin
Given 迁移完成
When 搜索 AGENTS.md、README.zh.md、README.en.md
Then 不应包含 `automation:` 任务引用
And 不应包含 `panel/web` 路径引用（应为 `web/`）
And 不应包含 `make ...` 入口
```

### Scenario 11: buf generation paths work after web promotion

```gherkin
Given panel/buf.gen.yaml 已更新为 ../web/ 路径
When 在 panel/ 目录执行 `buf generate`
Then TS 生成产物应正确输出到 web/src/lib/rpc/gen/
And 编译器插件应正确从 web/node_modules/.bin/ 加载
```

## Verification Matrix

| 验证项 | 命令 |
|--------|------|
| 前端门禁 | `moon run web:lint && moon run web:typecheck && moon run web:test` |
| Panel Go 测试 | `moon run panel:test` |
| Node Go 测试 | `moon run node:test` |
| 代码生成 | `moon run panel:generate` |
| 生成新鲜度 | `moon run panel:check-generate` |
| Docker 构建 | `docker compose -f panel/docker-compose.build.yml build` |
| E2E smoke | `moon run e2e:smoke` |
| 残留检查 | `grep -r 'automation:' AGENTS.md README.*.md` |
