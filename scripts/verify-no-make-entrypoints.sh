#!/usr/bin/env bash
set -euo pipefail

if [[ -f Makefile ]]; then
  echo "ERROR: Makefile must be removed after Moon cutover" >&2
  exit 1
fi

if rg -n "make (generate|check-generate|e2e|e2e-smoke|e2e-down|e2e-report)" README.zh.md README.en.md AGENTS.md; then
  echo "ERROR: found legacy make entrypoints in developer-facing docs" >&2
  exit 1
fi

echo "Doc entrypoint check: ok"
