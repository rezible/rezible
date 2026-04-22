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
    process-compose -f process-compose.dev.yaml

@run-frontend *ARGS:
    bun run --filter=@rezible/frontend --elide-lines 0 {{ARGS}}

@build-app-docker APP:
    docker build -t "localhost/rez-{{APP}}:latest" -f "./packages/apps/{{APP}}/Dockerfile" .

@run-app-docker APP *ARGS:
    docker run \
      -v "./devenv/certs/localias-ca.crt:/usr/local/share/ca-certificates/localias-ca.crt:ro" \
      -e "SSL_CERT_DIR=/usr/local/share/ca-certificates" \
      --env-file ./.env \
      "localhost/rez-{{APP}}:latest" \
       {{ARGS}}
