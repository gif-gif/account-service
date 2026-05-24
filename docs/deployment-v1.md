# Account Service V1 Deployment

The first version uses one repository with two independently deployed units:

- `service/`: Go backend API service.
- `web/`: Vite React static frontend.

## Backend

Build and run `service` as a backend service. It owns PostgreSQL access, credential encryption, API Key authentication, admin sessions, audit logs, and lease cleanup.

Required configuration:

- `DATABASE_URL`
- `SECRET_ENCRYPTION_KEY`
- `ADMIN_SESSION_SECRET`
- `CORS_ALLOWED_ORIGINS`

Expose the backend only through HTTPS in production. PostgreSQL should be reachable from `service` only.

## Frontend

Build `web` with:

```bash
cd web
npm install
npm run build
```

Deploy `web/dist` to a static hosting platform. Set `VITE_API_BASE_URL` at build time to the backend HTTPS origin.

## CORS

Add the frontend origin to backend `CORS_ALLOWED_ORIGINS`, for example:

```bash
CORS_ALLOWED_ORIGINS=https://accounts-admin.example.com
```

The frontend sends admin session cookies with API requests, so the backend must allow credentials for the configured origin.

## Local Development

Start PostgreSQL:

```bash
docker compose up -d postgres
```

Run backend and frontend separately:

```bash
cd service
go run ./cmd/account-service
```

```bash
cd web
npm run dev
```
