#!/usr/bin/env bash
set -euo pipefail

if ! git diff --exit-code -- 'panel/internal/rpc/gen/**' 'web/src/lib/rpc/gen/**' 'node/internal/rpc/gen/**'; then
  echo ""
  echo "ERROR: Generated files are out of date."
  echo "Run 'moon run panel:generate' and commit the changes."
  exit 1
fi

echo "Generated files are up to date."
