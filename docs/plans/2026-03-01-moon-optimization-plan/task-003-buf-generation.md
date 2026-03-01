### Task 3: Buf Generation Path Adaptation

**目标：** 更新 buf 生成配置中的 TS 插件路径和输出路径，适配 web 目录提升。

**Files:**
- Modify: `panel/buf.gen.yaml`
- Modify: `panel/check-generate.sh`

**Step 1: 更新 panel/buf.gen.yaml**

将 TS 相关插件的路径从 `web/` 前缀改为 `../web/`：

```yaml
version: v2
plugins:
  - local: [go, tool, protoc-gen-go]
    out: internal/rpc/gen
    opt:
      - paths=source_relative
  - local: [go, tool, protoc-gen-connect-go]
    out: internal/rpc/gen
    opt:
      - paths=source_relative
      - simple
  - local: [../web/node_modules/.bin/protoc-gen-es]
    out: ../web/src/lib/rpc/gen
    include_imports: true
    opt:
      - target=ts
      - import_extension=none
  - local: [../web/node_modules/.bin/protoc-gen-connect-query]
    out: ../web/src/lib/rpc/gen
    include_imports: true
    opt:
      - target=ts
      - import_extension=none
```

**Step 2: 更新 panel/check-generate.sh**

将文件内容改为：

```bash
#!/usr/bin/env bash
set -euo pipefail

if ! git diff --exit-code -- 'panel/internal/rpc/gen/**' 'web/src/lib/rpc/gen/**' 'node/internal/rpc/gen/**'; then
  echo ""
  echo "ERROR: Generated files are out of date."
  echo "Run 'moon run panel:generate' and commit the changes."
  exit 1
fi

echo "Generated files are up to date."
```

变更点：
- `panel/web/src/lib/rpc/gen/**` → `web/src/lib/rpc/gen/**`
- `moon run automation:generate` → `moon run panel:generate`

**Step 3: 验证 buf generate 可执行**

```bash
cd panel && buf generate && cd ..
```

Expected: 无错误，web/src/lib/rpc/gen/ 和 panel/internal/rpc/gen/ 中有生成产物。

**Step 4: 验证 node proto 生成**

```bash
cd panel && buf generate --template buf.node.gen.yaml --path proto/sboard/node/v1/node.proto && cd ..
```

Expected: node/internal/rpc/gen/ 中有生成产物。

**Step 5: 验证 go generate 链路**

```bash
cd panel && go generate ./internal/rpc/... && cd ..
```

Expected: 无错误，所有三个生成目录的产物已更新。

**Step 6: 提交**

```bash
git add panel/buf.gen.yaml panel/check-generate.sh
git commit -m "refactor(buf): adapt generation paths for top-level web directory"
```
