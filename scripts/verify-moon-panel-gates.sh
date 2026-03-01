#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f panel/moon.yml ]]; then
  echo "ERROR: missing panel/moon.yml" >&2
  exit 1
fi

assert_contains() {
  local file="$1"
  local pattern="$2"
  if ! grep -Fq "$pattern" "$file"; then
    echo "ERROR: ${file} missing pattern: ${pattern}" >&2
    exit 1
  fi
}

assert_contains panel/moon.yml "web-lint:"
assert_contains panel/moon.yml "web-typecheck:"
assert_contains panel/moon.yml "web-test:"
assert_contains panel/moon.yml "cd web && bun run lint"
assert_contains panel/moon.yml "cd web && bunx tsc -b"
assert_contains panel/moon.yml "cd web && bun run test"

panel_json="$(mktemp)"
trap 'rm -f "$panel_json"' EXIT
scripts/moon-cli.sh query projects --id panel >"$panel_json"

for task in panel:web-lint panel:web-typecheck panel:web-test; do
  if ! grep -q "\"target\": \"${task}\"" "$panel_json"; then
    echo "ERROR: moon does not register task ${task}" >&2
    exit 1
  fi
done

echo "Moon panel gate mapping check: ok"
