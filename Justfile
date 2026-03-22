set shell := ["bash", "-uc"]
set dotenv-filename := ".env.dev"
set dotenv-load

set export

DB_URL_BASE := "postgresql://rezible:foobar1@localhost:7010/"
DEV_DB_DATABASE := "rezible"
DB_CONN_QUERYOPTS := "?sslmode=disable"
DB_URL := DB_URL_BASE + DEV_DB_DATABASE + DB_CONN_QUERYOPTS
TEST_DB_URL := DB_URL_BASE + DB_CONN_QUERYOPTS

_default:
  @just --list

# [group('Setup')]

@setup:
    just install-dependencies
    just codegen
    just setup-db

@install-dependencies:
    cd backend && go mod tidy
    bun install

@upgrade-dependencies:
    devbox update
    cd backend && go get -u ./... && go mod tidy
    bun update

@format:
    cd backend && go fmt ./...
    cd frontend && bun run format

@run-backend *ARGS:
    cd backend && go run ./cmd/rezible {{ARGS}}

default_dockerfile := "Dockerfile"
@build-backend-docker dockerfile=default_dockerfile:
    mkdir -p ./scripts/certs && cat $(localias debug cert) > ./scripts/certs/localias-ca.crt
    docker build -t rezible-backend -f backend/{{dockerfile}} .

@run-backend-docker:
    docker run \
      --network host \
      --env-file ./.env \
      --env-file ./.env.dev \
      -e DB_URL \
      localhost/rezible-backend:latest

@run-frontend *ARGS:
    cd frontend && \
        PUBLIC_APP_URL="${APP_URL}" \
        PUBLIC_API_URL="/api/v1" \
        PUBLIC_AUTH_ISSUER_URL="${AUTH__OIDC__ISSUER_URL}" \
        PUBLIC_AUTH_CLIENT_ID="${AUTH__OIDC__CLIENT_ID}" \
        bun run {{ARGS}}

@run-documents-server *ARGS:
    cd documents-server && \
        API_URL="http://localhost:7002/api/v1" \
        bun run {{ARGS}}

@run-docker-compose *CMD:
    docker compose \
      --env-file .env \
      --env-file .env.dev \
      -f ./scripts/docker-compose.yaml \
      {{CMD}}

# [group('Testing')]

@test-backend: run-dev-services
    cd backend && \
      DB_URL="${TEST_DB_URL}" \
      go test $(go list ./... | grep -v /ent/)

@run-backend-datasync:
    just run-backend sync-integrations

# [group('Code Generation')]

@codegen: codegen-backend && codegen-api

@codegen-backend:
    cd backend && go generate ./...

@codegen-ent:
    cd backend && go generate ./ent

@codegen-mocks:
    cd backend && go generate ./testkit/mocks

@codegen-api:
    just run-backend spec > /tmp/rezible-spec.yaml
    bun run codegen:api

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
    cd backend && reflex -s -d none -r '\.go$' -- just run-backend serve

@dev-frontend:
    just run-frontend dev

@dev-documents-server:
    just run-documents-server dev

# [group('Database')]

migrations_dir := "backend/migrations"

@setup-db:
    just run-docker-compose down postgres -v && just run-docker-compose up postgres --wait
    just create-initial-migrations
    just run-migrations

@create-initial-migrations:
    rm -f ./{{migrations_dir}}/*.{sql,sum}
    just run-backend generate-migration ent_init
    sleep 1
    migrate create -ext sql -dir "{{migrations_dir}}" river_init
    cd backend && go tool river migrate-get --all --exclude-version 1 --up > "migrations/$(ls migrations | grep 'river_init.up')"
    cd backend && go tool river migrate-get --all --exclude-version 1 --down > "migrations/$(ls migrations | grep 'river_init.down')"

@run-migrations:
    migrate -source "file://{{migrations_dir}}" -database "${DB_URL}" up

@run-psql *ARGS:
    just run-docker-compose exec -it postgres psql -U rezible {{ARGS}}