# Account Service Backend

The backend lives in `service/` and is deployed independently from the frontend.

## Environment

- `DATABASE_URL`: PostgreSQL connection string.
- `SERVICE_BASE_URL`: Public backend base URL.
- `SECRET_ENCRYPTION_KEY`: 32 byte credential encryption key.
- `DEFAULT_LEASE_TTL_SECONDS`: Default lease TTL, usually `900`.
- `MAX_LEASE_TTL_SECONDS`: Maximum lease TTL, usually `7200`.
- `LEASE_CLEANUP_INTERVAL_SECONDS`: Lease cleanup interval, usually `60`.
- `ADMIN_SESSION_SECRET`: Admin session signing secret.
- `CORS_ALLOWED_ORIGINS`: Comma-separated frontend origins.
- `LOG_LEVEL`: zerolog level, defaults to `info`.
- `HEALTH_CHECK_DATABASE_TIMEOUT_SECONDS`: Database readiness timeout.

## Commands

```bash
cd service
go test ./...
go run ./cmd/account-service
```

From the repository root, Docker Compose can manage PostgreSQL and the backend process:

```bash
docker compose up --build service
```

Runtime configuration is loaded from the root `.env` file by Docker Compose. The service container must use `HTTP_HOST=0.0.0.0` so the backend listens outside the container.

For migration integration tests, set:

```bash
export TEST_DATABASE_URL=postgres://account:account@localhost:5432/account?sslmode=disable
go test ./internal/db
```
