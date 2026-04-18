set shell := ["bash", "-uc"]

set dotenv-filename := ".env.dev"
set dotenv-load := true

mod backend 'packages/apps/backend'
mod dev-services 'scripts'

_default:
  @just --list

@setup:
    just install-dependencies
    just setup-db

@install-dependencies:
    just backend install
    bun install

@run-frontend *ARGS:
    bun run --filter="@rezible/frontend" {{ARGS}}

@run-documents-server *ARGS:
    bun run --filter="@rezible/documents-server" {{ARGS}}

@run-dev-services *ARGS:
    docker compose up {{ARGS}}

@get-dev-services-healthy:
    docker compose ps | grep -q "unhealthy" && exit 1 || exit 0

@build-app-docker APP:
    docker build -t "localhost/rez-{{APP}}:latest" -f "./packages/apps/{{APP}}/Dockerfile" .

@run-app-docker APP *ARGS:
    docker run \
      -v "./scripts/certs/localias-ca.crt:/usr/local/share/ca-certificates/localias-ca.crt:ro" \
      -e "SSL_CERT_DIR=/usr/local/share/ca-certificates" \
      --env-file ./.env \
      "localhost/rez-{{APP}}:latest" \
       {{ARGS}}

@codegen-api:
    just backend print-spec > /tmp/rezible-spec.yaml
    bun run --filter="@rezible/api-client-ts" --elide-lines 0 build

@dev:
    process-compose --ordered-shutdown

setup-db: recreate-db run-migrations

@recreate-db:
    docker compose down postgres -v
    docker compose up postgres --wait

@run-migrations:
    just backend run db migrate-up
