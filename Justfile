set shell := ["bash", "-uc"]
set dotenv-filename := ".env.dev"
set dotenv-load

_default:
  @just --list

# [group('Setup')]

backend_dir := "./packages/apps/backend"
documents_server_dir := "./packages/apps/documents-server"
scripts_dir := "./scripts"

@setup:
    just install-dependencies
    just codegen
    just setup-db

@echo-env VAR:
    echo "value: '${{VAR}}'"

@install-dependencies:
    go -C {{backend_dir}} mod tidy
    bun install

@format:
    go -C {{backend_dir}}  fmt ./...
    bun run format

@reload-localias:
    localias stop && localias start
    mkdir -p "{{scripts_dir}}/certs" && cat "$(localias debug cert)" > "{{scripts_dir}}/certs/localias-ca.crt"

@run-docker IMAGE *ARGS:
    docker run \
      -v "{{scripts_dir}}/certs/localias-ca.crt:/usr/local/share/ca-certificates/localias-ca.crt:ro" \
      -e "SSL_CERT_DIR=/usr/local/share/ca-certificates" \
      --network host \
      --env-file ./.env \
      --env-file ./.env.dev \
      {{IMAGE}} {{ARGS}}

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

backend_local_docker_image := "localhost/rezible-backend:latest"

@build-backend-docker:
    docker build -t {{backend_local_docker_image}} {{backend_dir}}

@run-backend-docker *ARGS:
    just run-docker {{backend_local_docker_image}} {{ARGS}}

local_dev_api_url := "http://localhost:7002/api/v1"
@run-documents-server *ARGS:
    API_URL="{{local_dev_api_url}}" \
        bun run --filter="@rezible/documents-server" \
        {{ARGS}}

docs_local_docker_image := "localhost/rezible-backend:latest"

@build-documents-server-docker:
    docker build \
      -t {{docs_local_docker_image}} \
      -f "{{documents_server_dir}}/Dockerfile" \
      .

@run-documents-server-docker *ARGS:
    docker run {{docs_local_docker_image}} {{ARGS}}

@run-frontend *ARGS:
    bun run --filter="@rezible/frontend" {{ARGS}}

# [group('Testing')]

@test-backend: run-dev-services
    go -C {{backend_dir}} test \
        $(go -C {{backend_dir}} list ./... | grep -v /ent/)

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

setup-db: recreate-db bootstrap-db run-migrations

@run-psql *ARGS:
    just run-docker-compose exec -it postgres psql {{ARGS}}

@recreate-db:
    just run-docker-compose down postgres -v && just run-docker-compose up postgres --wait

@bootstrap-db:
    just run-backend bootstrap-db --database-url="$POSTGRES_ADMIN_URL"

@run-migrations:
    just run-backend migrate up

documents_role_grant_migration_file := backend_dir / "migrations/0002_documents_role_grant"
documents_role_grant_up_sql := 'GRANT SELECT, INSERT, UPDATE ON TABLE "documents" TO rez_documents;'
documents_role_grant_down_sql := 'REVOKE SELECT, INSERT, UPDATE ON TABLE "documents" FROM rez_documents;'

@create-initial-migrations: recreate-db
    rm -f {{backend_dir}}/migrations/*.{sql,sum}
    just run-backend generate-migration init
    echo "{{documents_role_grant_up_sql}}" > "{{documents_role_grant_migration_file}}.up.sql"
    echo "{{documents_role_grant_down_sql}}" > "{{documents_role_grant_migration_file}}.down.sql"
    just run-backend generate-migration --update-checksum
