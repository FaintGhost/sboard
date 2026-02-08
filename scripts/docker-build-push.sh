#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

NAMESPACE=""
TAG="latest"

usage() {
  cat <<'HELP'
用法:
  ./scripts/docker-build-push.sh --namespace <dockerhub_namespace> [--tag <tag>]

示例:
  ./scripts/docker-build-push.sh --namespace faintghost --tag latest
  ./scripts/docker-build-push.sh --namespace faintghost --tag "$(git rev-parse --short HEAD)"

说明:
  - 会构建并推送两个镜像:
    - <namespace>/sboard-node:<tag>
    - <namespace>/sboard-panel:<tag>
  - 需要你已完成 docker login。
HELP
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    -n|--namespace)
      NAMESPACE="${2:-}"
      shift 2
      ;;
    -t|--tag)
      TAG="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "[错误] 未知参数: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$NAMESPACE" ]]; then
  echo "[错误] 必须提供 --namespace" >&2
  usage
  exit 1
fi

if [[ -z "$TAG" ]]; then
  echo "[错误] --tag 不能为空" >&2
  exit 1
fi

export DOCKER_BUILDKIT="${DOCKER_BUILDKIT:-1}"

NODE_IMAGE="${NAMESPACE}/sboard-node:${TAG}"
PANEL_IMAGE="${NAMESPACE}/sboard-panel:${TAG}"
PANEL_COMMIT_ID="$(git -C "${ROOT_DIR}" rev-parse --short HEAD)"
SING_BOX_VERSION="$(awk '$1 == "github.com/sagernet/sing-box" {print $2}' "${ROOT_DIR}/node/go.mod")"

if [[ -z "${SING_BOX_VERSION}" ]]; then
  SING_BOX_VERSION="unknown"
fi

echo "[1/2] 构建并推送 Node: ${NODE_IMAGE}"
(
  cd "${ROOT_DIR}/node"
  export SBOARD_NODE_IMAGE="${NODE_IMAGE}"
  docker compose -f docker-compose.yml -f docker-compose.build.yml build
  docker push "${SBOARD_NODE_IMAGE}"
)

echo "[2/2] 构建并推送 Panel: ${PANEL_IMAGE}"
(
  cd "${ROOT_DIR}/panel"
  export SBOARD_PANEL_IMAGE="${PANEL_IMAGE}"
  docker compose -f docker-compose.yml -f docker-compose.build.yml build \
    --build-arg PANEL_VERSION="${TAG}" \
    --build-arg PANEL_COMMIT_ID="${PANEL_COMMIT_ID}" \
    --build-arg SING_BOX_VERSION="${SING_BOX_VERSION}"
  docker push "${SBOARD_PANEL_IMAGE}"
)

echo "完成:"
echo "  - ${NODE_IMAGE}"
echo "  - ${PANEL_IMAGE}"
