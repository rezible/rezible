# Running Rezible

Rezible runs as three container images. This document is the contract any deployment of them needs; it does
not assume a platform.

| Image | Default command | Port |
| --- | --- | --- |
| `ghcr.io/rezible/backend` | `rezible serve` | `HTTP__PORT` (or `PORT`), 7002 by default |
| `ghcr.io/rezible/documents-server` | `bun run index.js` | `PORT`, 7003 in the image |
| `ghcr.io/rezible/frontend` | Caddy | 80 |

Configuration is read from environment variables. For the backend, each variable is a configuration path in
upper case with `__` between sections: `http.auth.session_secret` is `HTTP__AUTH__SESSION_SECRET`.

## Routing

- **Frontend.** Browsers load the app from the frontend's public origin. Caddy serves the single-page app,
  falling back to `index.html`, and proxies `/api/*` to `PROXY_API_UPSTREAM_HOST` with the `/api` prefix
  removed. Leave the backend's `HTTP__BASE_PATH` empty behind it.
- **Backend through the frontend.** Browsers reach the backend's sign-in, callback and API routes through the
  frontend's origin, under `/api`. Whatever sits in front must preserve `Cookie` and `Set-Cookie` and must not
  cache authenticated responses.
- **Backend at its own domain.** Integration webhooks (Alertmanager, GitHub, Slack, webhook deployments) are
  issued as `https://<APP__API_DOMAIN>/webhooks/...`, so providers must reach the backend at that domain.
- **Documents-server.** Browsers connect to the documents-server directly at its own public URL, not through
  the frontend. The backend advertises that URL as `DOCUMENTS__SERVER_URL`, and whatever routes to it must allow
  the websocket upgrade.

## Configuration

### Backend

Required:

| Variable | Meaning |
| --- | --- |
| `APP__FRONTEND_DOMAIN` | The frontend's public host, such as `app.example.com`. Links and the default sign-in callback (`https://<domain>/api/auth/callback`) use it. |
| `APP__API_DOMAIN` | The backend's public host, used for webhook URLs and as the OIDC resource. |
| `HTTP__AUTH__SESSION_SECRET` | Secret for signing session cookies. |
| `HTTP__AUTH__OIDC__ISSUER`, `HTTP__AUTH__OIDC__CLIENT_ID`, `HTTP__AUTH__OIDC__CLIENT_SECRET` | The OIDC provider and client used for sign-in. `HTTP__AUTH__OIDC__REDIRECT_URL` overrides the callback. |
| `DOCUMENTS__SERVER_URL` | The documents-server's browser-accessible URL. When the frontend is served over HTTPS it must be secure: `wss://`, or `https://`, which browsers accept for websockets. |
| `DOCUMENTS__SESSION_KEY_HEX` | 64 hex characters (32 bytes). Must equal the documents-server's value. |
| `POSTGRES__HOST`, `POSTGRES__PORT`, `POSTGRES__DATABASE` | The database. Defaults are `postgres`, 5432 and `rezible`. |
| `POSTGRES__SSLMODE` | `require` by default. |
| `POSTGRES__ROLE_APP__NAME`, `POSTGRES__ROLE_APP__PASSWORD` | The role `rezible serve` connects as. |
| `POSTGRES__ROLE_ADMIN__NAME`, `POSTGRES__ROLE_ADMIN__PASSWORD` | The role `rezible migrations apply` connects as. |

Optional sections, each validated only when set: `AI__PROVIDERS__*` and `AI__MODELS__*` for investigations,
`INTEGRATIONS__SLACK__*`, `INTEGRATIONS__GITHUB__*` and `INTEGRATIONS__GOOGLE__*` for those integrations, and
`TELEMETRY__*` (below). `rezible print-config` prints the loaded configuration.

### Documents-server

| Variable | Meaning |
| --- | --- |
| `DOCUMENTS__SESSION_KEY_HEX` | Required. The same 64 hex characters as the backend. |
| `DOCUMENTS__ALLOWED_ORIGINS` | Comma-separated origins allowed to open websockets. It must include the frontend's origin, such as `https://app.example.com`. An empty list rejects every browser connection. |
| `DB_URL` | A PostgreSQL connection URL. Without it, the URL is built from `POSTGRES__USER` (default `documents`), `POSTGRES__PASSWORD` (required), `POSTGRES__HOST`, `POSTGRES__PORT`, `POSTGRES__DATABASE` (default `rezible`) and `POSTGRES__SSLMODE` (default `require`). |
| `PORT`, `HOST` | Listen address; the image sets 7003 and `0.0.0.0`. |

### Frontend

| Variable | Meaning |
| --- | --- |
| `PROXY_API_UPSTREAM_HOST` | Required. The backend's address as Caddy reaches it, such as `backend:7002`. |

## Database

PostgreSQL holds all persistent state. Before the first migration, the database needs:

- the `rezible` and `river` schemas, owned by the admin role;
- the app role, with `USAGE` on both schemas, default privileges for `SELECT`, `INSERT`, `UPDATE` and `DELETE`
  on their tables and `USAGE`, `SELECT` on their sequences, and `search_path` set to `rezible, river`;
- a role for the documents-server with `SELECT` and `UPDATE` on `documents`, `SELECT` on `retrospectives`, and
  `rezible` on its `search_path`. The app role is enough.

`devenv/scripts/database.ts` (`ensure`) is the reference for these statements.

## Migrations

`rezible migrations apply`, with no argument, applies the schema and job queue migrations as the admin role.
Run it to completion, once, before a new version serves traffic. It exits non-zero on failure.

## Replicas and state

- The backend and frontend keep no local state, so any number of each can run.
- The documents-server holds live documents and their pending writes in memory, with no synchronization
  between instances. Run exactly one replica, and fully stop the old instance before starting its replacement.
  Collaborative editing is briefly unavailable during a deploy. No persistent volumes are needed.

## Documents-server shutdown

On `SIGTERM` or `SIGINT` the documents-server:

1. reports not ready on `/ready` and rejects new websocket connections with 503, while HTTP stays up;
2. closes every connection and stores each document's pending writes;
3. closes the database and stops.

The whole sequence is limited to 30 seconds. If a write cannot be stored, or the time runs out, it logs the
failure and exits non-zero. Give the deployment a termination grace period longer than 30 seconds.

## Health checks

| Service | `/health` | `/ready` |
| --- | --- | --- |
| Backend | 200, or 500 when any started service reports itself unhealthy | 200 once the listener is up and the database answers a ping within 2 seconds, otherwise 503 |
| Documents-server | 204 while the process serves HTTP | 200 when the database answers `SELECT 1` within 2 seconds, otherwise 503; always 503 once shutdown has begun |

## Telemetry

- The backend logs to stdout as JSON. Traces and metrics go to an OTLP gRPC collector when
  `TELEMETRY__TRACING__ENABLED` and `TELEMETRY__METRICS__ENABLED` are set. `TELEMETRY__OTLP_ENDPOINT` names the
  collector, such as `http://collector:4317`; without it, the standard `OTEL_EXPORTER_OTLP_*` variables apply.
  `TELEMETRY__SERVICE_NAME` names the service.
- The documents-server logs plain text lines to stdout and exports no telemetry.
