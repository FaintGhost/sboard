#!/usr/bin/env bash
set -euo pipefail

required_files=("e2e/moon.yml" "scripts/moon.yml" "e2e/run.sh" "e2e/smoke.sh")
for file in "${required_files[@]}"; do
  [[ -f "$file" ]] || { echo "ERROR: missing ${file}" >&2; exit 1; }
done

assert_contains() {
  local file="$1"
  local pattern="$2"
  if ! grep -Fq -- "$pattern" "$file"; then
    echo "ERROR: ${file} missing pattern: ${pattern}" >&2
    exit 1
  fi
}

assert_contains e2e/moon.yml "run:"
assert_contains e2e/moon.yml "smoke:"
assert_contains e2e/moon.yml "down:"
assert_contains e2e/moon.yml "report:"
assert_contains e2e/moon.yml "bunx"
assert_contains e2e/moon.yml "playwright-report"
assert_contains e2e/moon.yml "runInCI: false"
assert_contains e2e/moon.yml "run.sh"
assert_contains e2e/moon.yml "smoke.sh"

assert_contains e2e/run.sh "docker compose -f docker-compose.e2e.yml down -v"
assert_contains e2e/run.sh "--abort-on-container-exit --exit-code-from playwright"
assert_contains e2e/run.sh 'ret=$?'
assert_contains e2e/smoke.sh "docker compose -f docker-compose.e2e.yml up --build -d panel node"
assert_contains e2e/smoke.sh "bunx playwright test --project=smoke"
assert_contains e2e/smoke.sh 'ret=$?'

assert_contains scripts/moon.yml "e2e:"
assert_contains scripts/moon.yml "e2e-smoke:"
assert_contains scripts/moon.yml "e2e-down:"
assert_contains scripts/moon.yml "e2e-report:"
assert_contains scripts/moon.yml "- e2e:run"
assert_contains scripts/moon.yml "- e2e:smoke"
assert_contains scripts/moon.yml "- e2e:down"
assert_contains scripts/moon.yml "- e2e:report"

echo "Moon e2e task mapping check: ok"
