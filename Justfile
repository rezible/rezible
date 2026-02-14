set shell := ["bash", "-uc"]
set dotenv-load

frontend_dist_dir := "./backend/internal/http/frontend-dist"
saml_cert_dir := "./backend/internal/http/saml/testdata"

_default:
  @just --list

# [group('Setup')]

@setup:
    mkdir -p "{{ frontend_dist_dir }}" && echo "<p>this will be replaced by the frontend build</p>" > "{{ frontend_dist_dir }}/index.html"
    mkdir -p "{{ saml_cert_dir }}"
    openssl req -x509 -newkey rsa:2048 -keyout "{{ saml_cert_dir }}/test.key" -out "{{ saml_cert_dir }}/test.cert" -days 365 -nodes -subj "/CN=test.rezible.com"
    just install-dependencies
    just codegen
    just setup-db
    localias reload

@install-dependencies:
    cd backend && go mod tidy
    bun install

@upgrade-everything:
    devbox update
    cd backend && go get -u ./... && go mod tidy
    bun update

@run-backend *ARGS:
    cd backend && DB_URL="$DB_URL" DEBUG_MODE=true go run ./cmd/rezible {{ARGS}}

# [group('Code Generation')]

@codegen: codegen-backend && codegen-api

@codegen-backend:
    cd backend && go generate ./...

@codegen-ent:
    cd backend && go generate ./ent

@codegen-mocks:
    cd backend && go generate ./testkit/mocks

@codegen-api:
    just run-backend openapi > /tmp/rezible-spec.yaml
    bun run codegen:api

# [group('Development')]

@dev: stop-db
    process-compose --ordered-shutdown

@format:
    cd backend && go fmt ./...
    cd frontend && bun run format

@dev-backend:
    cd backend && reflex -s -d none -r '\.go$' -- just run-backend

@dev-frontend:
    cd frontend && PUBLIC_REZ_API_BASE_URL="/api/v1" bun run dev

@dev-documents-server:
    cd documents-server && bun run dev

@run-datasync: start-db
    DATASYNC_MODE="true" just run-backend integrations sync

@test-backend:
    cd backend && go test ./...

@test-backend-db:
    cd backend && go test ./internal/db/...

@test-backend-db-verbose:
    cd backend && go test -v ./internal/db/...

# [group('Database')]
@create-db: stop-db
    rm -rf ./.devbox/virtenv/postgresql/data
    initdb -A trust > /dev/null
    just start-db
    createdb rezible

@setup-db:
    just create-db
    # just setup-migrations
    just run-auto-migrations
    # just run-backend load-fake-config

@setup-migrations:
    cd backend/internal/postgres/migrations && \
      go tool river migrate-get --all --exclude-version 1 --up > river_all.up.sql && \
      go tool river migrate-get --all --exclude-version 1 --down > river_all.down.sql

@run-auto-migrations:
    just run-backend db migrate apply auto

@seed-db:
    just run-backend seed

@run-db: stop-db
    -pg_isready -q || pg_ctl -o "-k $PGHOST" start

@start-db: stop-db
   just run-db > /dev/null

@run-psql:
    psql -d rezible

@stop-db:
    -pg_isready -q && pg_ctl stop > /dev/null

# [group('Other')]
@run-oidc-test-provider:
    docker run -p 6432:8080 \
        -e DevOidcToolkit__Users__0__Email="${REZ_DEBUG_DEFAULT_USER_EMAIL}" \
        -e DevOidcToolkit__Users__0__FirstName=Test \
        -e DevOidcToolkit__Users__0__LastName=User \
        -e DevOidcToolkit__Clients__0__Id=client \
        -e DevOidcToolkit__Clients__0__Secret=secret \
        -e DevOidcToolkit__Clients__0__RedirectUris__INDEX=https://app.dev.rezible.com/api/auth/oidc/callback \
        ghcr.io/businesssimulations/dev-oidc-toolkit:0.2.0
