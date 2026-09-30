# Backend test helpers

Run backend tests from the repository root with `just test backend`, optionally followed by packages and Go test flags. The recipe establishes the shared PostgreSQL test server. Do not invoke Go tests directly.

## Configuration

`test.Suite` loads validated synthetic configuration without loading ambient application settings. The only environment variables read by the test loader are:

- `POSTGRES_TEST_HOST`, `POSTGRES_TEST_PORT`, `POSTGRES_TEST_DB`: required test-server coordinates. Missing/invalid coordinates fail setup; development-database variables are not a fallback.
- `POSTGRES_ADMIN_USER`, `POSTGRES_APP_USER`: default to `postgres` and `rez_app`. Passwords match these names, as configured by the repository's local runner.

`WithConfigOverrides` applies explicit per-suite values last. Required auth/documents values are synthetic; auth bypass, provider clients, Redis and telemetry exporters remain disabled by default. Configuration tests changing process environment run serially.

## Database and identity ownership

`Suite.CreateTestDatabase()` creates a fresh migrated database on every call and seeds a default tenant, organization and user. The seed tenant is always ID 1 (seeding fails if it is not), so `SeedTenantContext()` may be called before the database exists. Existing skip-seeding options retain their behavior. Cleanup is registered as resources are acquired; dependent services must stop before database cleanup.

`Suite.NewIdentity(database, label)` adds a fresh tenant, organization, user and persisted session to a supplied database. It returns `(context.Context, Identity)`; `Identity` contains the persisted session, and the context stays local to the caller. Use it for multi-tenant tests; use `SeedTenantContext()` for existing seeded service fixtures.

The two application scenarios, `TestIngestAndQuery` and `TestInvestigation`, live together in `cmd/rezible/app_test.go`. Shared application setup, bounded waits and the scripted model live in `app_helpers_test.go`.

The application suite follows the CLI evaluation path: its injector creates and owns the isolated database and dependent services. It does not also call `CreateTestDatabase`. Its harness runs `Application.RunLifecycle` and waits for lifecycle completion before shutting down injector-owned resources. It wraps the application's registered v1 API with `humatest.Wrap` and dispatches through the complete production `Server.Handler()`. Requests use plain test contexts; the outer server middleware initializes the HTTP execution context, and production cookies and Huma security middleware handle authentication. No OIDC login calls or live model/provider calls are involved.

Focused service tests stay in their existing suites. Application journeys use real workers and production registrations with periodic scheduling disabled; their model is scripted. They assert specific committed records and job states rather than waiting for an empty queue.

`Suite.Telemetry()` provides an exporter-free `rez.TelemetryService` (SDK tracer, no-op metrics, quiet logger) for injection. It never installs process globals; only the production entrypoint calls `opentelemetry.Service.Init`. Genkit falls back to its own global SDK tracer provider when none is installed.

The harness authenticates its Huma test APIs using persisted identities and production cookies. `api.Operation[Request, Response](definition)` binds an existing operation to its Go contracts. `Call(ctx, request)` builds the path, query, headers and JSON body from Huma tags and returns a fresh typed response; `ExpectStatus` checks rejected requests. Requests take contexts explicitly. These journeys exercise the full HTTP handler in process, including outer Chi routing and middleware, without a listening server. The ingestion journey also verifies that a missing cookie returns HTTP 401. River observations use a read-only client and the public job APIs against the application database, without starting another worker.

## Commands

```sh
just test backend ./test ./internal/postgres/pgtestdb -count=1
just test backend ./cmd/rezible ./test ./pkg/openapi ./internal/postgres/river -count=1
just test backend ./cmd/rezible -run TestBackendSuite -race -count=3
just test backend ./internal/postgres/river -run TestJobServiceSuite -race -count=3
just test backend
```

The provider journey begins at the ingestion service; it does not test the unfinished demo webhook. Existing focused investigation tests cover failure/retry and output-selection cases outside the application journey.
