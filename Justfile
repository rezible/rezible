set shell := ["bash", "-uc"]
set dotenv-filename := ".env.dev"

mod backend 'packages/apps/backend'

_default:
  @just --list

@setup:
    just backend install
    bun install
    devbox run setup-db

@_docker-compose *ARGS:
    docker compose --ansi never {{ARGS}} > /dev/null 2>&1

@setup-db: stop-services
    just _docker-compose down postgres -v
    just start-db
    just backend apply-migrations

@stop-services:
    just _docker-compose down

@start-db:
    just _docker-compose up postgres --wait

@dev: stop-services
    process-compose --ordered-shutdown

@setup-localias:
    localias stop && localias start
    localias set ${APP_DOMAIN} ${API_PORT}
    localias set ${API_DOMAIN} ${BACKEND_PORT}
    localias set ${AUTH_DOMAIN} ${AUTH_PORT}
    # localias set ${POSTGRES_DOMAIN} $(just get-docker-postgres-port)

@build-app-docker APP:
    docker build -t "localhost/rez-{{APP}}:latest" -f "./packages/apps/{{APP}}/Dockerfile" .

@run-app-docker APP *ARGS:
    docker run \
      -v "./devenv/certs/localias-ca.crt:/usr/local/share/ca-certificates/localias-ca.crt:ro" \
      -e "SSL_CERT_DIR=/usr/local/share/ca-certificates" \
      --env-file ./.env \
      "localhost/rez-{{APP}}:latest" \
       {{ARGS}}
