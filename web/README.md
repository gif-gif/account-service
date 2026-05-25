# Account Service Web

The frontend lives in `web/` and is deployed as static assets independently from the backend.

## Environment

The frontend uses Vite modes to separate runtime configuration. Business code reads `VITE_API_BASE_URL`; environment selection is handled by the command.

| Environment | Vite mode | File | API base URL |
| --- | --- | --- | --- |
| Local | `localdev` | `.env.localdev` | `http://localhost:8000` |
| Development | `development` | `.env.development` | `https://dev-api.example.com` |
| Test | `test` | `.env.test` | `https://account.goio.uk` |
| Production | `production` | `.env.production` | `https://api.example.com` |

Required variables:

- `VITE_APP_ENV`: Environment name exposed to the frontend.
- `VITE_API_BASE_URL`: Backend API origin used by `src/lib/api.ts`.

## Commands

```bash
cd web
npm install
npm run check:env
npm test -- --run
npm run build
npm run dev:local
npm run build:local
npm run build:dev
npm run build:test
npm run build:prod
```
