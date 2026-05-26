# Account Service Backend

The backend lives in `service/` and is deployed independently from the frontend.

## Environment

Backend runtime configuration is split by deployment environment:

| Environment | File | Service base URL |
| --- | --- | --- |
| Local | `../.env.local` | `http://localhost:8000` |
| Development | `../.env.development` | `https://dev-api.example.com` |
| Test | `../.env.test` | `https://account.goio.uk` |
| Production | `../.env.production` | `https://api.example.com` |

The root `.env` file is kept as a local-compatible fallback for existing commands.

- `DATABASE_URL`: PostgreSQL connection string.
- `SERVICE_BASE_URL`: Public backend base URL.
- `SECRET_ENCRYPTION_KEY`: 32 byte credential encryption key.
- `DEFAULT_LEASE_TTL_SECONDS`: Default lease TTL, usually `900`.
- `MAX_LEASE_TTL_SECONDS`: Maximum lease TTL, usually `7200`.
- `LEASE_CLEANUP_INTERVAL_SECONDS`: Lease cleanup interval, usually `60`.
- `ADMIN_SESSION_SECRET`: Admin session signing secret.
- `CORS_ALLOWED_ORIGINS`: Comma-separated frontend origins.
- `LOG_LEVEL`: zerolog level, defaults to `info`.
- `LOG_DIR`: Log file directory for non-development environments, defaults to `logs`.
- `HEALTH_CHECK_DATABASE_TIMEOUT_SECONDS`: Database readiness timeout.

When `APP_ENV=development`, logs are printed to the console. Other environments write logs to `LOG_DIR/YYYY-MM-DD.log`.

Validate all service environment files with:

```bash
service/scripts/check-env-files.sh
```

## Commands

```bash
cd service
go test ./...
set -a
source ../.env.local
set +a
go run ./cmd/account-service
```

From the repository root, Docker Compose can manage PostgreSQL and the backend process:

```bash
docker compose up --build service
```

Docker Compose uses `.env.local` by default for the service container. Select another environment with `SERVICE_ENV_FILE`:

```bash
SERVICE_ENV_FILE=.env.development docker compose up --build service
SERVICE_ENV_FILE=.env.test docker compose up --build service
SERVICE_ENV_FILE=.env.production docker compose up --build service
```

The service container must use `HTTP_HOST=0.0.0.0` so the backend listens outside the container.

For migration integration tests, set:

```bash
export TEST_DATABASE_URL=postgres://account:account@localhost:5432/account?sslmode=disable
go test ./internal/db
```
