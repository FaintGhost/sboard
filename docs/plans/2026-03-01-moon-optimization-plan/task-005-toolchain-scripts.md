### Task 5: Toolchain Scripts

**目标：** 修复 moon-cli.sh 版本双源，修复 dev-panel-web.sh npm→bun 和路径。

**Files:**
- Modify: `scripts/moon-cli.sh`
- Modify: `scripts/dev-panel-web.sh`

**Step 1: 修复 scripts/moon-cli.sh**

将硬编码的 `@moonrepo/cli@2.0.3` 改为从 `.prototools` 动态读取。

完整文件内容：

```bash
#!/usr/bin/env sh
set -eu

if command -v moon >/dev/null 2>&1; then
  exec moon "$@"
fi

if ! command -v bunx >/dev/null 2>&1; then
  echo "ERROR: moon command not found and bunx is unavailable for fallback." >&2
  exit 1
fi

MOON_VERSION="$(grep '^moon = ' .prototools | sed 's/moon = "\(.*\)"/\1/')"
if [ -z "$MOON_VERSION" ]; then
  echo "ERROR: cannot determine moon version from .prototools" >&2
  exit 1
fi

: "${BUN_TMPDIR:=/tmp/bun-tmp}"
: "${BUN_INSTALL:=/tmp/bun-install}"
: "${MOON_HOME:=/tmp/moon-home}"
: "${PROTO_HOME:=/tmp/proto-home}"
: "${XDG_CACHE_HOME:=/tmp/xdg-cache}"

mkdir -p "$BUN_TMPDIR" "$BUN_INSTALL" "$MOON_HOME" "$PROTO_HOME" "$XDG_CACHE_HOME"

exec env \
  BUN_TMPDIR="$BUN_TMPDIR" \
  BUN_INSTALL="$BUN_INSTALL" \
  MOON_HOME="$MOON_HOME" \
  PROTO_HOME="$PROTO_HOME" \
  XDG_CACHE_HOME="$XDG_CACHE_HOME" \
  bunx "@moonrepo/cli@${MOON_VERSION}" "$@"
```

**Step 2: 验证 moon-cli.sh**

```bash
scripts/moon-cli.sh --version
```

Expected: 输出 moon 版本号（应与 .prototools 中 `2.0.3` 一致）。

**Step 3: 修复 scripts/dev-panel-web.sh**

三处修改：

1. 第 23-26 行：`npm` 检查改为 `bun` 检查
   - `if ! command -v npm` → `if ! command -v bun`
   - 错误信息改为：`"[错误] 未找到 bun，请先安装 Bun。"`

2. 第 63 行：web 目录路径
   - `cd "${ROOT_DIR}/panel/web"` → `cd "${ROOT_DIR}/web"`

3. 第 64 行：npm 改为 bun
   - `npm run dev -- --host` → `bun run dev --host`

**Step 4: 验证脚本语法**

```bash
bash -n scripts/dev-panel-web.sh
```

Expected: 无语法错误。

**Step 5: 提交**

```bash
git add scripts/moon-cli.sh scripts/dev-panel-web.sh
git commit -m "fix(scripts): resolve moon version dual-source, replace npm with bun"
```
