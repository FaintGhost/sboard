#!/usr/bin/env bash
set -euo pipefail

generated_paths=(
  panel/internal/rpc/gen
  web/src/lib/rpc/gen
  node/internal/rpc/gen
)

if ! git diff --exit-code -- "${generated_paths[@]}"; then
  echo ""
  echo "ERROR: Generated files are out of date."
  echo "Run 'moon run panel:generate' and commit the changes."
  exit 1
fi

untracked_generated="$(git ls-files --others --exclude-standard -- "${generated_paths[@]}")"
if [[ -n "${untracked_generated}" ]]; then
  echo ""
  echo "ERROR: Generated files include untracked paths."
  echo "Run 'moon run panel:generate' and commit the generated files."
  echo ""
  printf '%s\n' "${untracked_generated}"
  exit 1
fi

echo "Generated files are up to date."
