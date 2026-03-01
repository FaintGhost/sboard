### Task 6: CI Workflow

**目标：** 新增 GitHub Actions CI 门禁工作流，在 PR 和 push 时自动运行 check-generate、lint、typecheck、test。

**Files:**
- Create: `.github/workflows/ci.yml`

**Step 1: 创建 CI 工作流**

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
      - run: moon run panel:check-generate

  web:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - run: moon run web:lint
      - run: moon run web:typecheck
      - run: moon run web:test

  panel:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - run: moon run panel:test

  node:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: moonrepo/setup-toolchain@v0
      - run: moon run node:test
```

设计要点：
- 4 个 job 并行运行，互不阻塞
- 使用 `moonrepo/setup-toolchain@v0` 安装 moon 和 proto，版本从 `.prototools` 读取
- `concurrency` 保护：同分支只保留最新运行
- panel:test 和 node:test 需要 Go + CGO（panel 用 sqlite3），ubuntu-latest 默认包含 gcc

**Step 2: 验证 YAML 语法**

```bash
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"
```

Expected: 无错误。

**Step 3: 提交**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add gate workflow with moon-based lint/typecheck/test"
```
