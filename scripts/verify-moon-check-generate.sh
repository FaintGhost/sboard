#!/usr/bin/env bash
set -euo pipefail

automation_json="$(mktemp)"
output_log="$(mktemp)"
trap 'rm -f "$automation_json" "$output_log"' EXIT

scripts/moon-cli.sh query projects --id automation >"$automation_json"

if ! grep -q '"target": "automation:check-generate"' "$automation_json"; then
  echo "ERROR: automation:check-generate task is missing" >&2
  exit 1
fi

if ! grep -q '"target": "automation:generate"' "$automation_json"; then
  echo "ERROR: automation:check-generate must depend on automation:generate" >&2
  exit 1
fi

set +e
scripts/moon-cli.sh run automation:check-generate >"$output_log" 2>&1
status=$?
set -e

if [[ $status -eq 0 ]]; then
  if ! grep -q 'Generated files are up to date\.' "$output_log"; then
    echo "ERROR: success path missing expected message" >&2
    cat "$output_log" >&2
    exit 1
  fi
  echo "Moon check-generate success path: ok"
else
  if ! grep -q 'Generated files are out of date\.' "$output_log"; then
    echo "ERROR: failure path missing expected message" >&2
    cat "$output_log" >&2
    exit 1
  fi
  echo "Moon check-generate failure path: ok (workspace has pending generated diffs)"
fi
