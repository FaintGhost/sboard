#!/usr/bin/env bash
set -uo pipefail

docker compose -f docker-compose.e2e.yml down -v
docker compose -f docker-compose.e2e.yml up --build -d panel node
docker compose -f docker-compose.e2e.yml build playwright
docker compose -f docker-compose.e2e.yml run --rm playwright bunx playwright test --project=smoke
ret=$?
docker compose -f docker-compose.e2e.yml down -v
exit "$ret"
