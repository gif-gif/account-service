#!/usr/bin/env bash
set -euo pipefail

mkdir -p /app/logs
chown app:app /app/logs

exec gosu app "$@"
