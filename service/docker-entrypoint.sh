#!/usr/bin/env bash
set -euo pipefail

export HOME=/app
export XDG_CACHE_HOME=/app/.cache
export XDG_CONFIG_HOME=/app/.config
export XDG_DATA_HOME=/app/.local/share

runtime_dirs=(
  /app/logs
  /app/.aws/sso/cache
  /app/.cache
  /app/.config
  /app/.local/share
)

mkdir -p "${runtime_dirs[@]}"
chown -R app:app "${runtime_dirs[@]}"

exec gosu app "$@"
