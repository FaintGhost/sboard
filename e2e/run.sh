#!/usr/bin/env bash
set -uo pipefail

docker compose -f docker-compose.e2e.yml down -v
docker compose -f docker-compose.e2e.yml up --build --abort-on-container-exit --exit-code-from playwright
ret=$?
docker compose -f docker-compose.e2e.yml down -v
exit "$ret"
