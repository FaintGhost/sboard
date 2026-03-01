# Architecture

## Overview

本次优化分为两层变更：

1. 目录层：将 `panel/web/` 提升为顶层 `web/`，取消 `automation` 间接项目
2. 配置层：更新 Moon 任务定义、Docker 构建链、buf 生成路径、CI 工作流

## Directory Layer

### Before

```text
panel/
  web/            # 前端嵌套在 Go 后端下
  proto/
  buf.gen.yaml    # TS 插件引用 web/node_modules/...
  Dockerfile      # 构建上下文 = panel/
scripts/
  moon.yml        # automation 项目（alias 层）
  verify-*.sh     # 迁移残留
  moon-cli.sh     # 版本硬编码
```

### After

```text
panel/            # 纯 Go 后端
  proto/
  buf.gen.yaml    # TS 插件引用 ../web/node_modules/...
  Dockerfile      # 构建上下文 = workspace root
web/              # 独立前端项目
  moon.yml
node/             # Node 服务（补充 test 任务）
e2e/              # E2E（保持不变）
scripts/          # 工具脚本（不再是 moon 项目）
```

## Configuration Layer

### Moon Workspace

项目从 4 个变 4 个，但组成不同：

| Before | After |
|--------|-------|
| automation → scripts/ | （删除） |
| panel → panel/ | panel → panel/ |
| node → node/ | node → node/ |
| e2e → e2e/ | e2e → e2e/ |
| — | web → web/ |

### Docker Build Chain

Panel Dockerfile 构建上下文从子目录切换到工作区根：

```text
Before: context = panel/
  COPY web/...     → panel/web/...
  COPY go.mod      → panel/go.mod
  COPY . .         → panel/*

After: context = workspace root
  COPY web/...     → web/...
  COPY panel/go.mod → panel/go.mod
  COPY panel/ .     → panel/*
```

Node Dockerfile 不受影响（context 仍为 `node/`）。

### buf Generation Chain

`go generate` 在 `panel/` 目录执行 `buf generate`。buf 从 `panel/buf.gen.yaml` 读取配置。

路径变更：

```text
Before: web/node_modules/.bin/protoc-gen-es → panel/web/node_modules/...
After:  ../web/node_modules/.bin/protoc-gen-es → web/node_modules/...
```

生成器仍从 `panel/` 目录执行，所以 `../web/` 是正确的相对路径。

### CI Integration

新增独立门禁工作流，与发布工作流分离：

- `ci.yml`：在 PR 和 push 时运行 check-generate + lint + typecheck + test
- `docker-publish.yml`：仅在 main/master push 和 tag 时发布镜像

## Risks

### Path Correctness

目录移动后存在路径遗漏风险。

缓解：
- 在实施中逐步验证每条链路
- 用 `grep -r 'panel/web' .` 搜索残留引用
- 验证命令：`moon run panel:generate && moon run panel:check-generate`

### Docker Context Size

构建上下文从 `panel/` 扩大到工作区根，初始传输变大。

缓解：
- 不影响最终镜像大小（多阶段构建只 COPY 需要的）
- 可添加 `.dockerignore` 排除 `.git`、`docs/`、`e2e/` 等

### Team Adaptation

任务入口名称全面变更，团队需要重新建立 muscle memory。

缓解：
- 新入口更直觉（`panel:generate` 而非 `automation:generate`）
- 更新所有文档中的命令示例
