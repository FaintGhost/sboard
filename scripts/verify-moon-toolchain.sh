#!/usr/bin/env bash
set -euo pipefail

PROTOTOOLS_FILE=".prototools"

if [[ ! -f "$PROTOTOOLS_FILE" ]]; then
  echo "ERROR: missing $PROTOTOOLS_FILE" >&2
  exit 1
fi

expect_pin() {
  local key="$1"
  local version="$2"
  if ! grep -Eq "^${key} = \"${version}\"$" "$PROTOTOOLS_FILE"; then
    echo "ERROR: expected ${key} to be pinned to ${version} in ${PROTOTOOLS_FILE}" >&2
    exit 1
  fi
}

expect_pin moon 2.0.0
expect_pin go 1.26.0
expect_pin node 22.22.0
expect_pin bun 1.3.9

if command -v proto >/dev/null 2>&1; then
  proto use >/dev/null
  echo "proto use check: ok"
else
  echo "proto command not found; skipped runtime activation check."
fi

echo "Moon toolchain pin check: ok"
