#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PANEL_HTTP_ADDR="${PANEL_HTTP_ADDR:-:8080}"
PANEL_DB_PATH="${PANEL_DB_PATH:-panel.db}"
PANEL_JWT_SECRET="${PANEL_JWT_SECRET:-dev-secret-change-me}"
PANEL_CORS_ALLOW_ORIGINS="${PANEL_CORS_ALLOW_ORIGINS:-http://127.0.0.1:5173,http://localhost:5173}"
PANEL_SERVE_WEB="false"
GO_BUILD_TAGS="${GO_BUILD_TAGS:-with_utls}"

VITE_PROXY_TARGET="${VITE_PROXY_TARGET:-http://127.0.0.1:8080}"
WEB_HOST="${WEB_HOST:-0.0.0.0}"
WEB_PORT="${WEB_PORT:-5173}"

if ! command -v go >/dev/null 2>&1; then
  echo "[错误] 未找到 go，请先安装 Go。" >&2
  exit 1
fi

if ! command -v npm >/dev/null 2>&1; then
  echo "[错误] 未找到 npm，请先安装 Node.js/npm。" >&2
  exit 1
fi

cleanup() {
  local code=$?
  echo
  echo "[dev] 正在停止本地 panel/web..."

  if [[ -n "${PANEL_PID:-}" ]] && kill -0 "${PANEL_PID}" >/dev/null 2>&1; then
    kill "${PANEL_PID}" >/dev/null 2>&1 || true
  fi

  if [[ -n "${WEB_PID:-}" ]] && kill -0 "${WEB_PID}" >/dev/null 2>&1; then
    kill "${WEB_PID}" >/dev/null 2>&1 || true
  fi

  wait >/dev/null 2>&1 || true
  exit "$code"
}

trap cleanup INT TERM EXIT

echo "[dev] 启动 panel API: ${PANEL_HTTP_ADDR} (tags=${GO_BUILD_TAGS})"
(
  cd "${ROOT_DIR}/panel"
  PANEL_HTTP_ADDR="${PANEL_HTTP_ADDR}" \
  PANEL_DB_PATH="${PANEL_DB_PATH}" \
  PANEL_JWT_SECRET="${PANEL_JWT_SECRET}" \
  PANEL_CORS_ALLOW_ORIGINS="${PANEL_CORS_ALLOW_ORIGINS}" \
  PANEL_SERVE_WEB="${PANEL_SERVE_WEB}" \
  GOFLAGS="-tags=${GO_BUILD_TAGS}" \
  go run ./cmd/panel
) &
PANEL_PID=$!

echo "[dev] 启动 web dev server: http://${WEB_HOST}:${WEB_PORT}"
echo "[dev] 前端 API 代理目标: ${VITE_PROXY_TARGET}"
(
  cd "${ROOT_DIR}/panel/web"
  VITE_PROXY_TARGET="${VITE_PROXY_TARGET}" npm run dev -- --host "${WEB_HOST}" --port "${WEB_PORT}"
) &
WEB_PID=$!

echo "[dev] 启动完成，按 Ctrl+C 结束。"
echo "[dev] 浏览器访问: http://${WEB_HOST}:${WEB_PORT}"

wait -n "${PANEL_PID}" "${WEB_PID}"
