#!/usr/bin/env sh
set -eu

if command -v moon >/dev/null 2>&1; then
  exec moon "$@"
fi

if ! command -v bunx >/dev/null 2>&1; then
  echo "ERROR: moon command not found and bunx is unavailable for fallback." >&2
  exit 1
fi

MOON_VERSION="$(grep '^moon = ' .prototools | sed 's/moon = "\(.*\)"/\1/')"
if [ -z "$MOON_VERSION" ]; then
  echo "ERROR: cannot determine moon version from .prototools" >&2
  exit 1
fi

: "${BUN_TMPDIR:=/tmp/bun-tmp}"
: "${BUN_INSTALL:=/tmp/bun-install}"
: "${MOON_HOME:=/tmp/moon-home}"
: "${PROTO_HOME:=/tmp/proto-home}"
: "${XDG_CACHE_HOME:=/tmp/xdg-cache}"

mkdir -p "$BUN_TMPDIR" "$BUN_INSTALL" "$MOON_HOME" "$PROTO_HOME" "$XDG_CACHE_HOME"

exec env \
  BUN_TMPDIR="$BUN_TMPDIR" \
  BUN_INSTALL="$BUN_INSTALL" \
  MOON_HOME="$MOON_HOME" \
  PROTO_HOME="$PROTO_HOME" \
  XDG_CACHE_HOME="$XDG_CACHE_HOME" \
  bunx "@moonrepo/cli@${MOON_VERSION}" "$@"
