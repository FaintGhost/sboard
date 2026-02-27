#!/bin/bash
set -e

PANEL_URL="${BASE_URL:-http://panel:8080}"
NODE_URL="${NODE_API_URL:-http://node:3000}"
MAX_WAIT="${MAX_WAIT_SECONDS:-120}"

echo "Waiting for Panel at $PANEL_URL/rpc/sboard.panel.v1.HealthService/GetHealth ..."
elapsed=0
until curl -sf -X POST "$PANEL_URL/rpc/sboard.panel.v1.HealthService/GetHealth" -H "Content-Type: application/json" -d '{}' > /dev/null 2>&1; do
  if [ "$elapsed" -ge "$MAX_WAIT" ]; then
    echo "ERROR: Panel did not become healthy within ${MAX_WAIT}s"
    exit 1
  fi
  sleep 2
  elapsed=$((elapsed + 2))
done
echo "Panel is healthy (${elapsed}s)"

echo "Waiting for Node at $NODE_URL/api/health ..."
elapsed=0
until curl -sf "$NODE_URL/api/health" > /dev/null 2>&1; do
  if [ "$elapsed" -ge "$MAX_WAIT" ]; then
    echo "ERROR: Node did not become healthy within ${MAX_WAIT}s"
    exit 1
  fi
  sleep 2
  elapsed=$((elapsed + 2))
done
echo "Node is healthy (${elapsed}s)"

echo "All services ready. Running: $@"
exec "$@"
