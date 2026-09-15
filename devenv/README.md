## Setup and teardown

```sh
just setup
just dev verify
just dev teardown
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

Tests use their own Process Compose configuration. Backend tests start a shared
tmpfs PostgreSQL container and use the backend's isolated test databases.
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

`Justfile` defines lifecycle sequencing. `scripts/workspace-env.ts` generates the
workspace configuration, `scripts/database.ts` manages its database, and
`scripts/dex-client` is a standalone Go module using Dex's official API.
Service configuration files live in `configs/`.

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
