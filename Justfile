set shell := ["bash", "-uc"]
set dotenv-filename := ".env.dev"
set dotenv-load

_default:
  @just --list

pg_user_auth := env("POSTGRES__USER") + ":" + env("POSTGRES__PASSWORD")
pg_addr := env("POSTGRES__HOST") + ":" + env("POSTGRES__PORT")
pg_conn := env("POSTGRES__DATABASE") + "?sslmode=" + env("POSTGRES__SSLMODE")
DB_URL := "postgresql://" + pg_user_auth + "@" + pg_addr + "/" + pg_conn

# [group('Setup')]

backend_dir := "packages/backend"

@setup:
    just install-dependencies
    just codegen
    just setup-db

@install-dependencies:
    go -C {{backend_dir}} mod tidy
    bun install

@format:
    go -C {{backend_dir}}  fmt ./...
    bun run format

@run-backend *ARGS:
    go -C {{backend_dir}} run ./cmd/rezible {{ARGS}}

@build-documents-server-docker:
    docker build -t rezible-documents-server -f packages/documents-server/Dockerfile .

@build-backend-docker:
    mkdir -p ./scripts/certs && cat "$(localias debug cert)" > ./scripts/certs/localias-ca.crt
    docker build -t rezible-backend -f packages/backend/Dockerfile .

@run-backend-docker:
    docker run \
      --network host \
      --env-file ./.env \
      --env-file ./.env.dev \
      localhost/rezible-backend:latest

@run-frontend *ARGS:
    PUBLIC_APP_URL="${APP_URL}" \
    PUBLIC_API_URL="${API_URL}" \
    PUBLIC_API_URL_BASE="/api/v1" \
    PUBLIC_AUTH_ISSUER_URL="${AUTH__OIDC__ISSUER_URL}" \
    PUBLIC_AUTH_CLIENT_ID="${AUTH__OIDC__CLIENT_ID}" \
        bun run --filter=@rezible/frontend {{ARGS}}

@run-documents-server *ARGS:
    API_URL="http://localhost:7002/api/v1" \
    DB_URL="{{DB_URL}}" \
        bun run --filter="@rezible/documents-server" {{ARGS}}

@run-docker-compose *CMD:
    docker compose \
      --env-file .env \
      --env-file .env.dev \
      -f ./scripts/docker-compose.yaml \
      {{CMD}}

# [group('Testing')]

@test-backend: run-dev-services
    go -C {{backend_dir}} test $(go -C {{backend_dir}} list ./... | grep -v /ent/)

@run-backend-datasync:
    just run-backend sync-integrations

# [group('Code Generation')]

@codegen: codegen-backend && codegen-api

@codegen-backend:
    go -C {{backend_dir}} generate ./...

@codegen-ent:
    go -C {{backend_dir}} generate ./ent

@codegen-mocks:
    go -C {{backend_dir}} generate ./testkit/mocks

@codegen-api:
    just run-backend spec > /tmp/rezible-spec.yaml
    bun run --filter="@rezible/api-client-ts" build

@codegen-migration NAME:
    just run-backend generate-migration {{NAME}}

# [group('Development Servers')]

@run-dev-services:
    just run-docker-compose up -d --wait

@stop-dev-services:
    just run-docker-compose down

@dev: run-dev-services && stop-dev-services
    just run-migrations
    process-compose --ordered-shutdown -f ./process-compose.yaml

@dev-backend:
    cd packages/backend && reflex -s -d none -r '\.go$' -- just run-backend serve

@dev-frontend:
    just run-frontend dev

@dev-documents-server:
    just run-documents-server dev

# [group('Database')]

migrations_dir := "packages/backend/migrations"

@recreate-db:
    just run-docker-compose down postgres -v && just run-docker-compose up postgres --wait

@setup-db:
    just recreate-db
    just create-initial-migrations
    just run-migrations

@run-psql *ARGS:
    just run-docker-compose exec -it postgres psql {{ARGS}}

@create-initial-migrations:
    rm -f ./{{migrations_dir}}/*.{sql,sum}
    just run-backend generate-migration ent_init
    sleep 1
    migrate create -ext sql -dir "{{migrations_dir}}" river_init
    cd packages/backend && go tool river migrate-get --all --exclude-version 1 --up > "migrations/$(ls migrations | grep 'river_init.up')"
    cd packages/backend && go tool river migrate-get --all --exclude-version 1 --down > "migrations/$(ls migrations | grep 'river_init.down')"

#DB_URL_BASE := "postgresql://rezible:foobar1@localhost:7010/"
#DEV_DB_DATABASE := "rezible"
#DB_CONN_QUERYOPTS := "?sslmode=disable"
#DB_URL := DB_URL_BASE + DEV_DB_DATABASE + DB_CONN_QUERYOPTS
#TEST_DB_URL := DB_URL_BASE + DB_CONN_QUERYOPTS

@run-migrations direction="up":
    migrate \
        -source "file://{{migrations_dir}}" \
        -database "{{DB_URL}}" \
        {{direction}}
