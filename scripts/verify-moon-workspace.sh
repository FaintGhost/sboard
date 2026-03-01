#!/usr/bin/env bash
set -euo pipefail

required_files=(
  ".moon/workspace.yml"
  ".moon/toolchains.yml"
  ".moon/tasks/all.yml"
  "scripts/moon.yml"
  "panel/moon.yml"
  "node/moon.yml"
  "e2e/moon.yml"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: missing ${file}" >&2
    exit 1
  fi
done

if ! grep -q '^defaultProject: automation$' .moon/workspace.yml; then
  echo "ERROR: defaultProject must be automation" >&2
  exit 1
fi

for project in automation panel node e2e; do
  if ! grep -q "^[[:space:]]*${project}:" .moon/workspace.yml; then
    echo "ERROR: project ${project} is not registered in .moon/workspace.yml" >&2
    exit 1
  fi
done

projects_json="$(mktemp)"
trap 'rm -f "$projects_json"' EXIT

scripts/moon-cli.sh query projects >"$projects_json"

for project in automation panel node e2e; do
  if ! grep -q "\"id\": \"${project}\"" "$projects_json"; then
    echo "ERROR: moon query does not include project ${project}" >&2
    exit 1
  fi
done

echo "Moon workspace check: ok"
