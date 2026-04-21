set shell := ["bash", "-uc"]

set dotenv-filename := ".env.workspace"

mod backend 'packages/apps/backend'
mod devenv 'devenv'

_default:
  @just --list

@setup:
    just backend install
    bun install
    just devenv setup

@dev:
    just devenv ensure-postgres-ready
    process-compose --ordered-shutdown

@build-app-docker APP:
    docker build -t "localhost/rez-{{APP}}:latest" -f "./packages/apps/{{APP}}/Dockerfile" .

@run-app-docker APP *ARGS:
    docker run \
      -v "./devenv/certs/localias-ca.crt:/usr/local/share/ca-certificates/localias-ca.crt:ro" \
      -e "SSL_CERT_DIR=/usr/local/share/ca-certificates" \
      --env-file ./.env \
      "localhost/rez-{{APP}}:latest" \
       {{ARGS}}
