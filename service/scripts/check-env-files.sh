#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

required_files=(
  ".env.local"
  ".env.development"
  ".env.test"
  ".env.production"
)

required_keys=(
  "APP_ENV"
  "POSTGRES_DB"
  "POSTGRES_USER"
  "POSTGRES_PASSWORD"
  "POSTGRES_PORT"
  "DATABASE_URL"
  "SERVICE_BASE_URL"
  "SECRET_ENCRYPTION_KEY"
  "DEFAULT_LEASE_TTL_SECONDS"
  "MAX_LEASE_TTL_SECONDS"
  "LEASE_CLEANUP_INTERVAL_SECONDS"
  "ADMIN_SESSION_SECRET"
  "CORS_ALLOWED_ORIGINS"
  "LOG_LEVEL"
  "HEALTH_CHECK_DATABASE_TIMEOUT_SECONDS"
  "HTTP_HOST"
  "HTTP_PORT"
)

for file in "${required_files[@]}"; do
  path="$ROOT_DIR/$file"
  if [[ ! -f "$path" ]]; then
    echo "error: missing $file" >&2
    exit 1
  fi

  for key in "${required_keys[@]}"; do
    if ! grep -Eq "^${key}=.+" "$path"; then
      echo "error: $file must define $key" >&2
      exit 1
    fi
  done

  secret_key="$(grep -E "^SECRET_ENCRYPTION_KEY=" "$path" | head -1 | cut -d= -f2-)"
  if [[ "${#secret_key}" -ne 32 ]]; then
    echo "error: $file SECRET_ENCRYPTION_KEY must be exactly 32 bytes" >&2
    exit 1
  fi
done

echo "service environment files are valid"
