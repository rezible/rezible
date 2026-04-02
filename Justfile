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

@install-dependencies:
    go -C {{backend_dir}} mod tidy
    bun install

@format:
    go -C {{backend_dir}}  fmt ./...
    bun run format

@reload-localias:
    localias -c "{{scripts_dir}}/localias.yaml" stop && localias -c "{{scripts_dir}}/localias.yaml" start
    mkdir -p "{{scripts_dir}}/certs" && cat "$(localias debug cert)" > "{{scripts_dir}}/certs/localias-ca.crt"

@run-docker IMAGE *ARGS:
    docker run \
      -v "{{scripts_dir}}/certs/localias-ca.crt:/usr/local/share/ca-certificates/localias-ca.crt:ro" \
      -e "SSL_CERT_DIR=/usr/local/share/ca-certificates" \
      --network host \
      --env-file ./.env \
      --env-file ./.env.dev \
      {{IMAGE}} {{ARGS}}

@run-backend *ARGS:
    go -C {{backend_dir}} \
        run ./cmd/rezible \
        {{ARGS}}

@run-frontend *ARGS:
    bun run --filter="@rezible/frontend" {{ARGS}}

@run-documents-server *ARGS:
    bun run --filter="@rezible/documents-server" {{ARGS}}

@build-all-docker:
    just build-app-docker backend
    just build-app-docker documents-server
    just build-app-docker frontend

@build-app-docker COMPONENT:
    docker build -t "localhost/rez-{{COMPONENT}}:latest" -f "./packages/apps/{{COMPONENT}}/Dockerfile" .

@run-app-docker COMPONENT *ARGS:
    just run-docker "localhost/rez-{{COMPONENT}}:latest" {{ARGS}}

@run-all-docker-compose:
    just run-docker-compose "--profile rezible" up

@stop-all-docker-compose:
    just run-docker-compose "--profile rezible" down

@run-docker-compose *ARGS:
    docker compose --env-file=".env.dev" {{ARGS}}

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
    just run-docker-compose up -d --wait --build

@stop-dev-services:
    just run-docker-compose down

@dev: run-dev-services && stop-dev-services
    process-compose --ordered-shutdown

@dev-backend:
    cd "{{backend_dir}}" && \
      reflex -s -d none -r '\.go$' -- \
        just run-backend serve

# [group('Database')]

@run-psql *ARGS:
    just run-docker-compose exec -it postgres psql -U postgres {{ARGS}}

setup-db: recreate-db bootstrap-db run-migrations

@recreate-db:
    just run-docker-compose down postgres -v && just run-docker-compose up postgres --wait

@bootstrap-db:
    just run-backend bootstrap-db --database-url="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/postgres?sslmode=${POSTGRES_SSLMODE}"

@run-migrations:
    just run-backend migrate up

[working-directory("packages/apps/backend/migrations")]
@create-initial-migrations: recreate-db
    rm -f ./0001_init*.sql
    rm -f ./atlas.sum
    just run-backend generate-migration init
    just run-backend generate-migration --update-checksum
