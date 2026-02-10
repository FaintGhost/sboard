# SBoard Panel Web

React + TypeScript + Vite 前端项目。

## 开发

```bash
npm run dev
```

## 构建

```bash
npm run build
```

## 测试

```bash
npm run test
```

## 代码质量（Oxc）

本项目已从 ESLint 迁移到 Oxc，推荐使用 `bun`：

- `bun run lint`：执行 `oxlint`
- `bun run lint:fix`：执行 `oxlint --fix`
- `bun run format`：执行 `oxfmt --write`
- `bun run format:check`：执行 `oxfmt --check`

也兼容 `npm run` 同名脚本。

配置文件：

- `.oxlintrc.json`
- `.oxfmtrc.json`
