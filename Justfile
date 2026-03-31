set shell := ["bash", "-uc"]
set dotenv-filename := ".env.dev"
set dotenv-load

_default:
  @just --list

pg_user_auth := env("POSTGRES__USER") + ":" + env("POSTGRES__PASSWORD")
pg_addr := env("POSTGRES__HOST") + ":" + env("POSTGRES__PORT")
pg_conn := env("POSTGRES__DATABASE") + "?sslmode=" + env("POSTGRES__SSLMODE")
DB_URL := "postgresql://" + pg_user_auth + "@" + pg_addr + "/" + pg_conn

DOCUMENTS_DB_URL := env("DOCUMENTS_DB_URL")

# [group('Setup')]

backend_dir := "./packages/apps/backend"
documents_server_dir := "./packages/apps/documents-server"
scripts_dir := "./scripts"

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

@reload-localias:
    localias stop && localias start
    mkdir -p "{{scripts_dir}}/certs" && cat "$(localias debug cert)" > "{{scripts_dir}}/certs/localias-ca.crt"

@run-docker-compose *CMD:
    docker compose \
      --env-file .env \
      --env-file .env.dev \
      -f "{{scripts_dir}}/docker-compose.yaml" \
      {{CMD}}

@run-backend *ARGS:
    go -C {{backend_dir}} \
        run ./cmd/rezible \
        {{ARGS}}

@build-backend-docker:
    docker build \
      -t rezible-backend \
      {{backend_dir}}

@run-backend-docker:
    docker run \
      -v "{{scripts_dir}}/certs/localias-ca.crt:/usr/local/share/ca-certificates/localias-ca.crt:ro" \
      -e "SSL_CERT_DIR=/usr/local/share/ca-certificates" \
      --network host \
      --env-file ./.env \
      --env-file ./.env.dev \
      localhost/rezible-backend:latest

local_dev_api_url := "http://localhost:7002/api/v1"
@run-documents-server *ARGS:
    API_URL="{{local_dev_api_url}}" \
    DOCUMENTS_DB_URL="{{DOCUMENTS_DB_URL}}" \
        bun run --filter="@rezible/documents-server" \
        {{ARGS}}

@build-documents-server-docker:
    docker build \
      -t rezible-documents-server \
      -f "{{documents_server_dir}}/Dockerfile" \
      .

@run-documents-server-docker:
    docker run --network host \
      -e API_URL="{{local_dev_api_url}}" \
      -e DOCUMENTS_DB_URL="{{DOCUMENTS_DB_URL}}" \
      localhost/rezible-documents-server

@run-frontend *ARGS:
    PUBLIC_APP_URL="${APP_URL}" \
    PUBLIC_API_URL_BASE="/api/v1" \
    PUBLIC_AUTH_ISSUER_URL="${AUTH__OIDC__ISSUER_URL}" \
    PUBLIC_AUTH_CLIENT_ID="${AUTH__OIDC__CLIENT_ID}" \
    PUBLIC_AUTH_CLIENT_SCOPES="${AUTH__OIDC__CLIENT_SCOPES}" \
    PUBLIC_AUTH_CLIENT_REDIRECT_URI="${AUTH__OIDC__CLIENT_REDIRECT_URI}" \
        bun run --filter="@rezible/frontend" \
        {{ARGS}}

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
    process-compose --ordered-shutdown -f "{{scripts_dir}}/process-compose.yaml"

@dev-backend:
    cd "{{backend_dir}}" && \
      reflex -s -d none -r '\.go$' -- \
        just run-backend serve

# [group('Database')]

migrations_dir := backend_dir + "/migrations"

@recreate-db:
    just run-docker-compose down postgres -v && \
      just run-docker-compose up postgres --wait

@setup-db:
    just recreate-db
    just run-migrations

@run-psql *ARGS:
    just run-docker-compose \
        exec -it postgres psql {{ARGS}}

@create-initial-migrations: recreate-db
    rm -f {{migrations_dir}}/*.{sql,sum}
    just run-backend generate-migration init

migrator_pg_user_auth := f'{{env("POSTGRES__MIGRATIONS__USER")}}:{{env("POSTGRES__MIGRATIONS__PASSWORD")}}'
@run-migrations:
    just run-backend migrate up
