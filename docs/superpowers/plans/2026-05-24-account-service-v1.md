# Account Service V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first version of the account management system with a Fiber v3 backend in `service/` and a React management UI in `web/`, deployed independently.

**Architecture:** The backend owns all persistence, credentials, authentication, auditing, and lease logic. The frontend is a static Vite app that talks only to the backend JSON API and never stores secrets beyond in-memory reveal state.

**Tech Stack:** Go, Fiber v3, zerolog, PostgreSQL, React, Vite, shadcn-ui, Zustand, TypeScript.

---

## Source Design

Implement against [docs/account-service-design.md](/Users/jay/Documents/work/dev/ai/account-service/docs/account-service-design.md).

## Delivery Order

1. Build and test the backend API first, including schema, security, account CRUD, leases, admin auth, audit logs, health checks, and CORS.
2. Build and test the frontend management UI against the backend API contract.
3. Add local deployment files and release checks for independent `service` and `web` deployment.

## Development Rules

Apply these rules to every task in this plan:

- Use TDD for behavior changes: write the failing test first, run it and confirm the expected failure, implement the smallest passing code, then run the test again.
- Do not write production code for a task until its first relevant failing test has been observed.
- Keep each task independently reviewable and commit only after the task's verification command passes.
- Do not skip a verification command. If a command fails because dependencies or services are missing, stop and record the blocker before continuing.
- Keep backend and frontend deployable independently. The backend must not serve frontend static assets, and the frontend must not access PostgreSQL directly.
- Keep secrets out of logs, tests, frontend storage, and committed files. Tests may use dummy secrets only.
- Use uniform JSON errors from the backend and preserve `request_id` through logs, responses, audit entries, and frontend error displays.
- Use API Key authentication for internal account APIs and HttpOnly Cookie admin sessions for management APIs.
- Keep `service` and `web` changes separated by task unless a task explicitly wires their contract together.
- Prefer small focused files. If a file starts collecting unrelated responsibilities, split before adding more behavior.
- Run formatting before each commit: `gofmt` for Go files and the configured frontend formatter once `web` exists.
- Do not change the design document scope while implementing unless a requirement is impossible or unsafe; stop and ask for approval instead.

## File Structure

Backend files to create:

- `service/go.mod`: Go module and backend dependencies.
- `service/cmd/account-service/main.go`: Process entrypoint.
- `service/internal/app/app.go`: Fiber app wiring.
- `service/internal/config/config.go`: Environment-backed configuration.
- `service/internal/logging/logger.go`: zerolog setup.
- `service/internal/httpx/errors.go`: Uniform API error responses.
- `service/internal/httpx/middleware.go`: Request ID, logging, recovery, CORS, auth middleware.
- `service/internal/health/health.go`: Liveness and database readiness handlers.
- `service/internal/db/db.go`: PostgreSQL connection pool.
- `service/internal/db/migrations/*.sql`: Database schema.
- `service/internal/security/crypto.go`: Field encryption and decryption.
- `service/internal/security/apikey.go`: API Key generation, hashing, verification.
- `service/internal/security/password.go`: Admin password hashing and verification.
- `service/internal/audit/audit.go`: Audit event writer.
- `service/internal/admin/admin.go`: Admin login, current user, logout.
- `service/internal/accounts/accounts.go`: Account model, repository, service, handlers.
- `service/internal/leases/leases.go`: Lease acquisition, release, expiry cleanup.
- `service/internal/callers/callers.go`: API caller and API Key management.
- `service/internal/testutil/testdb.go`: Test database helpers.

Frontend files to create:

- `web/package.json`: Frontend scripts and dependencies.
- `web/vite.config.ts`: Vite config.
- `web/tsconfig.json`: TypeScript config.
- `web/index.html`: Vite entry document.
- `web/src/main.tsx`: React entrypoint.
- `web/src/App.tsx`: App routing and layout.
- `web/src/lib/api.ts`: Fetch client with credentials and error handling.
- `web/src/lib/types.ts`: Shared API response types.
- `web/src/store/auth.ts`: Zustand auth store.
- `web/src/store/accounts.ts`: Zustand account list state.
- `web/src/components/ui/*`: shadcn-ui generated components.
- `web/src/pages/LoginPage.tsx`: Admin login.
- `web/src/pages/AccountsPage.tsx`: Account list and filters.
- `web/src/pages/AccountDetailPage.tsx`: Account detail, reveal, edit.
- `web/src/pages/LeasesPage.tsx`: Lease list.
- `web/src/pages/ApiKeysPage.tsx`: API Key management.
- `web/src/pages/AuditLogsPage.tsx`: Audit log list.
- `web/src/test/*`: Frontend test setup and tests.

Repo-level files to create:

- `.gitignore`: Ignore build outputs, local env files, test artifacts.
- `docker-compose.yml`: Local PostgreSQL and service/web development dependencies.
- `docs/deployment-v1.md`: Independent deployment notes for `service` and `web`.

---

## Task 1: Backend Project Skeleton

**Files:**
- Create: `service/go.mod`
- Create: `service/cmd/account-service/main.go`
- Create: `service/internal/config/config.go`
- Create: `service/internal/logging/logger.go`
- Create: `service/internal/app/app.go`
- Create: `service/internal/health/health.go`
- Test: `service/internal/config/config_test.go`
- Test: `service/internal/health/health_test.go`

- [ ] **Step 1: Write config tests**

Create `service/internal/config/config_test.go` with tests for default values, required `DATABASE_URL`, TTL validation, and comma-separated `CORS_ALLOWED_ORIGINS`.

- [ ] **Step 2: Run config tests and verify failure**

Run: `cd service && go test ./internal/config`

Expected: FAIL because the config package does not exist yet.

- [ ] **Step 3: Implement config package**

Create `service/internal/config/config.go` with a `Config` struct containing `DatabaseURL`, `ServiceBaseURL`, `SecretEncryptionKey`, `DefaultLeaseTTLSeconds`, `MaxLeaseTTLSeconds`, `LeaseCleanupIntervalSeconds`, `AdminSessionSecret`, `CORSAllowedOrigins`, `LogLevel`, `HealthCheckDatabaseTimeoutSeconds`, `HTTPHost`, and `HTTPPort`.

- [ ] **Step 4: Write health tests**

Create `service/internal/health/health_test.go` to verify `GET /health/live` returns `200` and `GET /health/ready` returns `200` when the database checker succeeds and `503` when it fails.

- [ ] **Step 5: Implement app, logger, and health endpoints**

Create Fiber app wiring with request ID propagation, JSON error responses, and the two health routes.

- [ ] **Step 6: Run skeleton tests**

Run: `cd service && go test ./...`

Expected: PASS.

- [ ] **Step 7: Commit**

Run:

```bash
git add service
git commit -m "feat: scaffold backend service"
```

---

## Task 2: Database Schema and Test Harness

**Files:**
- Create: `service/internal/db/db.go`
- Create: `service/internal/db/migrations/000001_init.sql`
- Create: `service/internal/testutil/testdb.go`
- Test: `service/internal/db/migrations_test.go`

- [ ] **Step 1: Write migration test**

Create a test that applies `000001_init.sql` to a PostgreSQL test database and asserts these tables exist: `accounts`, `account_leases`, `api_callers`, `admin_users`, `admin_sessions`, and `audit_logs`.

- [ ] **Step 2: Run migration test and verify failure**

Run: `cd service && go test ./internal/db`

Expected: FAIL because migrations and database helpers do not exist.

- [ ] **Step 3: Add schema migration**

Create `000001_init.sql` with:

- `accounts` fields from the design document.
- `account_leases` with `active`, `released`, `expired` status constraint.
- `api_callers` with hashed API Key storage.
- `admin_users` for first-version admin login.
- `admin_sessions` for HttpOnly Cookie login state.
- `audit_logs` with actor, action, resource, request, IP, user agent, metadata, and timestamp.
- Indexes for account filtering, quota selection, lease expiry, caller status, session lookup, and audit event time.

- [ ] **Step 4: Add database connection and test helper**

Implement a PostgreSQL connection helper using context timeouts and a test helper that requires `TEST_DATABASE_URL`.

- [ ] **Step 5: Run migration test**

Run: `cd service && TEST_DATABASE_URL=postgres://account:account@localhost:5432/account_test?sslmode=disable go test ./internal/db`

Expected: PASS when local PostgreSQL is running.

- [ ] **Step 6: Commit**

Run:

```bash
git add service/internal/db service/internal/testutil
git commit -m "feat: add database schema"
```

---

## Task 3: Security, API Key, Admin Session, and Audit Foundation

**Files:**
- Create: `service/internal/security/crypto.go`
- Create: `service/internal/security/apikey.go`
- Create: `service/internal/security/password.go`
- Create: `service/internal/admin/admin.go`
- Create: `service/internal/audit/audit.go`
- Test: `service/internal/security/security_test.go`
- Test: `service/internal/admin/admin_test.go`
- Test: `service/internal/audit/audit_test.go`

- [ ] **Step 1: Write security tests**

Cover encryption round trip, wrong encryption key failure, API Key hash verification, generated API Key format, password hash verification, and password mismatch.

- [ ] **Step 2: Run security tests and verify failure**

Run: `cd service && go test ./internal/security`

Expected: FAIL because the package does not exist.

- [ ] **Step 3: Implement encryption and hashing**

Use authenticated encryption for credential fields, one-way hashing for API Keys, and password hashing for admin users.

- [ ] **Step 4: Write admin session tests**

Cover login success, login failure, `GET /api/v1/admin/me`, logout, expired session rejection, and Cookie attributes.

- [ ] **Step 5: Implement admin auth**

Add login, current user, and logout handlers using `admin_users` and `admin_sessions`.

- [ ] **Step 6: Write audit tests**

Cover audit creation and sensitive metadata redaction for `password`, `access_token`, `refresh_token`, and API Key plaintext.

- [ ] **Step 7: Implement audit writer**

Add a small audit service that accepts actor, action, resource, request ID, IP, user agent, and metadata.

- [ ] **Step 8: Run security and admin tests**

Run: `cd service && go test ./internal/security ./internal/admin ./internal/audit`

Expected: PASS.

- [ ] **Step 9: Commit**

Run:

```bash
git add service/internal/security service/internal/admin service/internal/audit
git commit -m "feat: add security and admin auth"
```

---

## Task 4: Account CRUD and Query API

**Files:**
- Create: `service/internal/accounts/accounts.go`
- Modify: `service/internal/app/app.go`
- Test: `service/internal/accounts/accounts_test.go`

- [ ] **Step 1: Write account repository tests**

Cover create, update, status update, get by ID, list filters by region/type/status/tags/min quota, and encrypted-at-rest credential fields.

- [ ] **Step 2: Write account handler tests**

Cover `POST /api/v1/accounts/query`, `GET /api/v1/accounts/{id}`, `POST /api/v1/accounts`, `PATCH /api/v1/accounts/{id}`, and `POST /api/v1/accounts/{id}/status`.

- [ ] **Step 3: Run tests and verify failure**

Run: `cd service && go test ./internal/accounts`

Expected: FAIL because account implementation does not exist.

- [ ] **Step 4: Implement account model and repository**

Use typed account status constants matching the design document and store sensitive fields only as encrypted database columns.

- [ ] **Step 5: Implement account service**

Validate status values, normalize tags, compute quota fields from request input, decrypt credentials only when preparing authorized API responses.

- [ ] **Step 6: Implement account handlers**

Return uniform JSON errors and emit audit events for query, get, create, update, and status update.

- [ ] **Step 7: Wire routes**

Register account routes under `/api/v1` in `service/internal/app/app.go`.

- [ ] **Step 8: Run account tests**

Run: `cd service && go test ./internal/accounts`

Expected: PASS.

- [ ] **Step 9: Commit**

Run:

```bash
git add service/internal/accounts service/internal/app
git commit -m "feat: add account API"
```

---

## Task 5: Lease Acquire, Release, and TTL Expiry

**Files:**
- Create: `service/internal/leases/leases.go`
- Modify: `service/internal/app/app.go`
- Test: `service/internal/leases/leases_test.go`

- [ ] **Step 1: Write lease selection tests**

Cover active-only selection, `quota_remaining > 0`, region/type/tag filters, highest remaining quota first, random selection among equal top quota accounts, and no available account error.

- [ ] **Step 2: Write concurrency tests**

Cover parallel acquire requests cannot exceed `max_concurrent_leases`.

- [ ] **Step 3: Write TTL and release tests**

Cover default TTL, max TTL validation, release success, release conflict for expired or released leases, and cleanup of expired leases.

- [ ] **Step 4: Run lease tests and verify failure**

Run: `cd service && go test ./internal/leases`

Expected: FAIL because lease implementation does not exist.

- [ ] **Step 5: Implement lease repository**

Use PostgreSQL transactions and row-level locking so active lease counts cannot exceed account limits.

- [ ] **Step 6: Implement lease service**

Select accounts by the design rules, create active lease rows, return complete decrypted credentials with `lease_id`, and mark releases and expiries correctly.

- [ ] **Step 7: Implement lease handlers and cleanup loop**

Register `POST /api/v1/accounts/acquire`, `POST /api/v1/accounts/release`, and `GET /api/v1/leases`; start a cleanup ticker inside the service process.

- [ ] **Step 8: Run lease tests**

Run: `cd service && go test ./internal/leases`

Expected: PASS.

- [ ] **Step 9: Commit**

Run:

```bash
git add service/internal/leases service/internal/app
git commit -m "feat: add account leasing"
```

---

## Task 6: API Caller Management, Auth Middleware, CORS, and Errors

**Files:**
- Create: `service/internal/callers/callers.go`
- Create: `service/internal/httpx/errors.go`
- Create: `service/internal/httpx/middleware.go`
- Modify: `service/internal/app/app.go`
- Test: `service/internal/callers/callers_test.go`
- Test: `service/internal/httpx/httpx_test.go`

- [ ] **Step 1: Write auth middleware tests**

Cover missing API Key, invalid API Key, disabled caller, valid caller, and admin session access to management routes.

- [ ] **Step 2: Write CORS tests**

Cover allowed origin, disallowed origin, credential support for admin cookies, and preflight response.

- [ ] **Step 3: Write caller tests**

Cover API Key creation, plaintext returned only once, hash stored in database, disable caller, and audit event creation.

- [ ] **Step 4: Run tests and verify failure**

Run: `cd service && go test ./internal/httpx ./internal/callers`

Expected: FAIL because middleware and callers implementation do not exist.

- [ ] **Step 5: Implement uniform error package**

Return JSON errors shaped as `{ "error": { "code": "...", "message": "...", "request_id": "..." } }`.

- [ ] **Step 6: Implement auth and CORS middleware**

Use API Key auth for internal account APIs and admin Cookie auth for management APIs.

- [ ] **Step 7: Implement caller management**

Register `POST /api/v1/api-keys` and caller disable behavior for admin users.

- [ ] **Step 8: Run full backend tests**

Run: `cd service && go test ./...`

Expected: PASS.

- [ ] **Step 9: Commit**

Run:

```bash
git add service/internal/httpx service/internal/callers service/internal/app
git commit -m "feat: add API auth and caller management"
```

---

## Task 7: Frontend Scaffold, API Client, and Auth Flow

**Files:**
- Create: `web/package.json`
- Create: `web/vite.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/index.html`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/lib/api.ts`
- Create: `web/src/lib/types.ts`
- Create: `web/src/store/auth.ts`
- Create: `web/src/pages/LoginPage.tsx`
- Test: `web/src/lib/api.test.ts`
- Test: `web/src/store/auth.test.ts`
- Test: `web/src/pages/LoginPage.test.tsx`

- [ ] **Step 1: Write API client tests**

Cover `VITE_API_BASE_URL`, `credentials: "include"`, JSON error parsing, and request ID display data.

- [ ] **Step 2: Write auth store and login tests**

Cover login success, failed login message, `/admin/me` restoration, and logout clearing state.

- [ ] **Step 3: Run tests and verify failure**

Run: `cd web && npm test -- --run`

Expected: FAIL because frontend project does not exist yet.

- [ ] **Step 4: Scaffold Vite React TypeScript app**

Add Vite, React, TypeScript, Vitest, Testing Library, Zustand, and shadcn-ui prerequisites.

- [ ] **Step 5: Implement API client**

Create a typed fetch wrapper that prefixes `VITE_API_BASE_URL`, sends credentials, and throws normalized API errors.

- [ ] **Step 6: Implement auth store and login page**

Build a simple login screen that calls `/api/v1/admin/login`, restores state from `/api/v1/admin/me`, and logs out with `/api/v1/admin/logout`.

- [ ] **Step 7: Run frontend auth tests**

Run: `cd web && npm test -- --run`

Expected: PASS.

- [ ] **Step 8: Commit**

Run:

```bash
git add web
git commit -m "feat: scaffold web admin auth"
```

---

## Task 8: Account Management UI

**Files:**
- Create: `web/src/store/accounts.ts`
- Create: `web/src/pages/AccountsPage.tsx`
- Create: `web/src/pages/AccountDetailPage.tsx`
- Create: `web/src/components/AccountForm.tsx`
- Create: `web/src/components/RevealSecret.tsx`
- Modify: `web/src/App.tsx`
- Test: `web/src/pages/AccountsPage.test.tsx`
- Test: `web/src/pages/AccountDetailPage.test.tsx`
- Test: `web/src/components/RevealSecret.test.tsx`

- [ ] **Step 1: Write account list tests**

Cover filters for region, account type, status, tags, minimum quota, pagination state, loading state, empty state, and backend error display.

- [ ] **Step 2: Write account detail and form tests**

Cover create, edit, status update, quota update, token/password update, validation errors, and successful save refresh.

- [ ] **Step 3: Write reveal tests**

Cover secrets hidden by default, revealed only after user action, and cleared on unmount.

- [ ] **Step 4: Run tests and verify failure**

Run: `cd web && npm test -- --run src/pages src/components`

Expected: FAIL because account UI does not exist.

- [ ] **Step 5: Implement accounts store and API calls**

Track filters, pagination, loading, errors, selected account, and mutation state in Zustand.

- [ ] **Step 6: Implement account pages**

Use shadcn-ui inputs, buttons, tabs, tables, dialogs, badges, and forms for account list and detail workflows.

- [ ] **Step 7: Implement reveal component**

Keep revealed `password`, `access_token`, and `refresh_token` in component memory only.

- [ ] **Step 8: Run account UI tests**

Run: `cd web && npm test -- --run src/pages src/components`

Expected: PASS.

- [ ] **Step 9: Commit**

Run:

```bash
git add web/src
git commit -m "feat: add account management UI"
```

---

## Task 9: Leases, API Keys, and Audit UI

**Files:**
- Create: `web/src/pages/LeasesPage.tsx`
- Create: `web/src/pages/ApiKeysPage.tsx`
- Create: `web/src/pages/AuditLogsPage.tsx`
- Create: `web/src/components/OneTimeSecret.tsx`
- Modify: `web/src/App.tsx`
- Test: `web/src/pages/LeasesPage.test.tsx`
- Test: `web/src/pages/ApiKeysPage.test.tsx`
- Test: `web/src/pages/AuditLogsPage.test.tsx`
- Test: `web/src/components/OneTimeSecret.test.tsx`

- [ ] **Step 1: Write leases page tests**

Cover filtering by account ID, caller ID, status, and time range.

- [ ] **Step 2: Write API Key tests**

Cover create key, display plaintext once, hide after dismissal, and disabled caller state.

- [ ] **Step 3: Write audit log tests**

Cover audit list rendering, actor/action filters, metadata display, and request ID display.

- [ ] **Step 4: Run tests and verify failure**

Run: `cd web && npm test -- --run src/pages/LeasesPage.test.tsx src/pages/ApiKeysPage.test.tsx src/pages/AuditLogsPage.test.tsx`

Expected: FAIL because these pages do not exist.

- [ ] **Step 5: Implement leases page**

Render active and historical leases with filters and status badges.

- [ ] **Step 6: Implement API Key page**

Create API Keys through the backend and keep plaintext visible only in the one-time result component.

- [ ] **Step 7: Implement audit log page**

Render audit events with request ID and redacted metadata.

- [ ] **Step 8: Run UI tests**

Run: `cd web && npm test -- --run`

Expected: PASS.

- [ ] **Step 9: Commit**

Run:

```bash
git add web/src
git commit -m "feat: add admin operations UI"
```

---

## Task 10: Local Development and Deployment Documentation

**Files:**
- Create: `.gitignore`
- Create: `docker-compose.yml`
- Create: `docs/deployment-v1.md`
- Create: `service/README.md`
- Create: `web/README.md`

- [ ] **Step 1: Write local environment docs**

Document required environment variables for `service` and `web`, including `DATABASE_URL`, `SECRET_ENCRYPTION_KEY`, `ADMIN_SESSION_SECRET`, `CORS_ALLOWED_ORIGINS`, and `VITE_API_BASE_URL`.

- [ ] **Step 2: Add docker compose**

Add PostgreSQL for local development and testing. Keep `service` and `web` as independently runnable processes rather than one bundled deployment.

- [ ] **Step 3: Add backend README**

Document commands:

```bash
cd service
go test ./...
go run ./cmd/account-service
```

- [ ] **Step 4: Add frontend README**

Document commands:

```bash
cd web
npm install
npm test -- --run
npm run build
npm run dev
```

- [ ] **Step 5: Add deployment notes**

Document independent deployment:

- Build `service` as a backend service.
- Build `web` as static assets.
- Configure backend CORS with the frontend origin.
- Configure frontend API base URL with the backend HTTPS origin.
- Keep PostgreSQL reachable only from `service`.

- [ ] **Step 6: Run final verification**

Run:

```bash
cd service && go test ./...
cd ../web && npm test -- --run && npm run build
```

Expected: all backend tests pass, frontend tests pass, and frontend build succeeds.

- [ ] **Step 7: Commit**

Run:

```bash
git add .gitignore docker-compose.yml docs/deployment-v1.md service/README.md web/README.md
git commit -m "docs: add local development and deployment guide"
```

---

## Final Acceptance Checklist

- [ ] `service` exposes health endpoints.
- [ ] `service` stores credentials encrypted at rest.
- [ ] `service` authenticates internal APIs with API Key hashes.
- [ ] `service` authenticates admin APIs with secure HttpOnly Cookie sessions.
- [ ] `service` enforces CORS allowlist for the frontend origin.
- [ ] `service` returns uniform JSON errors with `request_id`.
- [ ] `service` records redacted audit logs for sensitive actions.
- [ ] Account CRUD and query APIs work.
- [ ] Lease acquire honors status, quota, filters, concurrency limit, and TTL.
- [ ] Lease release and expiry cleanup work.
- [ ] API Key plaintext is returned only once.
- [ ] `web` logs in, restores session, and logs out.
- [ ] `web` supports account list, detail, create, edit, status, quota, and reveal flows.
- [ ] `web` supports lease list, API Key management, and audit log pages.
- [ ] `web` uses `VITE_API_BASE_URL` and sends credentials.
- [ ] Backend and frontend can be built and deployed independently.

## Self-Review

Spec coverage:

- Backend API, PostgreSQL, TTL leases, account selection, security, audit, config, health, and CORS are covered by Tasks 1 through 6.
- React, shadcn-ui, Zustand, Vite, admin login, account UI, lease UI, API Key UI, and audit UI are covered by Tasks 7 through 9.
- Independent deployment and environment configuration are covered by Task 10.

Placeholder scan:

- The plan avoids unresolved placeholder markers and vague future-work language.

Type consistency:

- API names and route paths match `docs/account-service-design.md`.
- Directory names match the requested `service` and `web` split.
