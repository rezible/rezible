## Setup and teardown

```sh
just setup
just dev verify-workspace
just dev cleanup-workspace
```

Setup installs dependencies and generates `devenv/.workspace-env.sh` with this
worktree's database name, HTTPS URLs, and Dex client credentials.
Devbox sources it on shell initialization. After first setup in an existing
shell, run `source devenv/.workspace-env.sh` or reopen `devbox shell`.
Reopen the shell after changing `devbox.json` or `.env`.

Setup starts shared PostgreSQL and Dex, provisions the worktree database and
Dex client, applies migrations, and registers the shared authentication alias.
Repeating setup refreshes defaults and computed configuration while preserving
the workspace identity, database name, and Dex credentials. Use
`just dev setup-workspace --force` to recreate this worktree's database, or
`just dev setup-workspace --no-migrate` to provision without applying migrations.

Teardown deletes only this worktree's database, Dex client, aliases, and generated
environment file. Stop its application services first. Shared containers and
the Localias daemon remain running.

## Application services

```sh
just dev                 # complete application
just dev backend
just dev frontend        # includes backend
just dev documents
```

Process Compose supervises these services. Paseo can supervise the same services
using `frontend`, `backend`, and `documents` in `paseo.json`. Paseo supplies service ports;
Process Compose uses `env_cmds` to call `scripts/get-port.ts` for each service,
using `get-port` through Bun auto-install
(`--install=fallback`), without a package declaration or lockfile change. Both use the common
`just frontend::dev`, `just backend::dev`, and `just documents-server::dev` recipes. Use one supervisor
per worktree; different worktrees can run concurrently. Services update their
Localias mappings when they start. Port discovery briefly precedes binding, so
a competing process can still cause a bind conflict; restart the stack if that
happens. Browse the HTTPS URLs in the generated environment file.

## Tests

```sh
just test
just test backend
just test backend ./internal/db -run TestName
just test frontend
just test documents
```

`just test` runs every suite in parallel, lets each finish even if another fails,
and ends with one line per suite. Backend tests start a shared tmpfs PostgreSQL
container and use the backend's isolated test databases.
The test container remains running after tests finish.

## Infrastructure and scripts

Docker Compose owns shared PostgreSQL, Dex, optional telemetry, and Alertmanager.
Infrastructure commands use the primary checkout's Compose configuration.

```sh
just dev infra-up telemetry
just dev infra-up alertmanager
just dev infra-stop telemetry
```

Stopping shared infrastructure affects every worktree.

`configs/prometheus.yaml` replaces the Prometheus configuration inside `otel-lgtm`: the image's
default file plus the alert definitions in `configs/prometheus-alerts.yaml` and the Alertmanager target.

## Simulations

`devenv/simulations` holds simulated services that send real telemetry to `otel-lgtm` and fire a real alert
through Alertmanager. Demos and launch-plan checks run against them. See
[simulations/README.md](simulations/README.md); run them with `just sim`.

`Justfile` defines lifecycle sequencing. `scripts/workspace-env.ts` generates the
workspace configuration, `scripts/database.ts` manages its database, and
`scripts/dex-client` is a standalone Go module using Dex's official API.
Service configuration files live in `configs/`.

### Sending simulation alerts to Rezible

The shared Alertmanager's `rezible-dev` receiver posts to Slack and to a Rezible webhook. It reads the webhook
URL from `devenv/.alertmanager/webhook-url` (gitignored) in the checkout that owns the shared infrastructure.
`just setup` and `just dev infra-up` create that file with a placeholder; deliveries to the placeholder fail in
Alertmanager without affecting Slack. This assumes one devenv and one worktree sending alerts.

1. Start the application (`just dev`), and in Settings → Integrations → Alertmanager, install Alertmanager.
2. Generate a webhook URL there and copy it. It is shown once.
3. Run `just dev alertmanager-target 'URL'` with the copied URL.

The settings URL uses the worktree's HTTPS alias, which the container cannot resolve. The recipe looks up the
backend's port in `localias list`, writes `http://host.docker.internal:<port>/webhooks/alertmanager/<token>`,
and reloads Alertmanager. The backend's port is assigned at launch, so run the recipe again with the same URL
after each backend restart. Generating a new URL in settings stops the old one working; run the recipe with the
new one.

### Reading simulation telemetry through Grafana

The Grafana integration reads Loki and Prometheus through the `otel-lgtm` Grafana's data source proxy. Start
it with `just dev infra-up telemetry`.

1. Run `just dev grafana-token create`. It creates a Viewer service account for this workspace in the shared
   Grafana, replacing any earlier one, and prints the URL and token. Grafana shows a token only once.
2. Install Grafana in Rezible with that URL (`http://localhost:$GRAFANA_PORT`; the backend runs on the host)
   and token.
3. Set the data source UIDs: `loki` for logs and `prometheus` for metrics. The simulation's services are named
   by the `service_name` label in both, which is the default.

`just dev grafana-token delete` removes the account and its tokens.

## Environment ownership

| Source | Values |
| --- | --- |
| `devbox.json` | `JUST_TEMPDIR` |
| Root `.env` | Local secrets and integration overrides; see `.env.example` |
| `devenv/.workspace-env.sh` | Development defaults, workspace identity, database, computed URLs (including `AUTH_URL`), Dex credentials |
| Service supervisor | `APP_PORT`, `BACKEND_PORT`, `DOCUMENTS_PORT` at launch |
| Generated workspace script | Backend mappings derived from workspace and shared values |
| Application recipes | Listener port and test-specific overrides |

Devbox loads `.env`, then sources the generated workspace configuration. 
Just and Process Compose inherit that environment. 
Setup sources the generated file explicitly so provisioning works on its first run.
Workspace configuration contains no application ports. 
Build and code-generation commands do not require a provisioned workspace.

Process Compose allocates backend ports from 20000–20999, frontend ports from
21000–21999, and documents ports from 22000–22999. Its environment commands
have a two-second timeout; if Bun’s first package download times out, run
`bun --no-env-file --install=fallback devenv/scripts/get-port.ts backend`
in Devbox once, then start again.

The `main` branch uses `app.dev.rezible.com`, `api.dev.rezible.com`, and
`documents.dev.rezible.com`. Other branches prefix those hosts with the workspace
ID. Run workspace setup after switching branches, then reload the environment
and restart application services to update their aliases.
