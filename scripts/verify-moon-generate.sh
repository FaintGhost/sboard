#!/usr/bin/env bash
set -euo pipefail

automation_json="$(mktemp)"
panel_json="$(mktemp)"
trap 'rm -f "$automation_json" "$panel_json"' EXIT

scripts/moon-cli.sh query projects --id automation >"$automation_json"
scripts/moon-cli.sh query projects --id panel >"$panel_json"

if ! grep -q '"target": "panel:generate-rpc"' "$automation_json"; then
  echo "ERROR: automation:generate must depend on panel:generate-rpc" >&2
  exit 1
fi

if ! grep -q '"target": "panel:generate-rpc"' "$panel_json"; then
  echo "ERROR: panel:generate-rpc task is missing" >&2
  exit 1
fi

scripts/moon-cli.sh run automation:generate

required_dirs=(
  "panel/internal/rpc/gen"
  "panel/web/src/lib/rpc/gen"
  "node/internal/rpc/gen"
)

for dir in "${required_dirs[@]}"; do
  if [[ ! -d "$dir" ]]; then
    echo "ERROR: missing generated directory ${dir}" >&2
    exit 1
  fi
done

echo "Moon generate task check: ok"
