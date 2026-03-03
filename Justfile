set shell := ["bash", "-uc"]
set dotenv-load

dev_db_user := "rezible"
dev_db_name := "rezible"
dev_db_url := "postgresql://"+dev_db_user+"@localhost"+"/"+dev_db_name+"?sslmode=disable"
dev_db_url_docker := "postgresql://"+dev_db_user+"@host.docker.internal"+"/"+dev_db_name+"?sslmode=disable"
test_db_url := "postgresql://"+dev_db_user+"@localhost"+"?sslmode=disable"

_default:
  @just --list

# [group('Setup')]

frontend_dist_dir := "./backend/internal/http/frontend-dist"
saml_cert_dir := "./backend/internal/http/saml/testdata"

@setup:
    mkdir -p "{{ frontend_dist_dir }}" && echo "<p>this will be replaced by the frontend build</p>" > "{{ frontend_dist_dir }}/index.html"
    mkdir -p "{{ saml_cert_dir }}"
    openssl req -x509 -newkey rsa:2048 -keyout "{{ saml_cert_dir }}/test.key" -out "{{ saml_cert_dir }}/test.cert" -days 365 -nodes -subj "/CN=test.rezible.com"
    just install-dependencies
    just codegen
    just localias-reload

@localias-reload:
    localias reload -c scripts/localias.yaml
    mkdir -p scripts/certs && cat /etc/ssl/cert.pem "$(localias debug cert)" > ./scripts/certs/ca-bundle.crt

@install-dependencies:
    cd backend && go mod tidy
    bun install

@upgrade-dependencies:
    devbox update
    cd backend && go get -u ./... && go mod tidy
    bun update

@run-docker-compose *CMD:
    docker compose \
      --env-file .env \
      -f ./scripts/docker-compose.yaml \
      {{CMD}}

@setup-dev-zitadel:
    cd scripts/zitadel/setup-dev/ && \
      ADMIN_SA_PAT="$(just run-docker-compose exec zitadel-login cat /bootstrap/admin-sa.pat)" \
      go run .

@run-backend *ARGS:
    cd backend && \
        DEBUG_MODE=true \
        DB_URL='{{ dev_db_url }}' \
        go run ./cmd/rezible {{ARGS}}

@run-frontend *ARGS:
    cd frontend && \
        PUBLIC_API_SERVER_URL="https://app.dev.rezible.com" \
        PUBLIC_API_BASE_PATH="/api/v1" \
        bun run {{ARGS}}

@run-documents-server *ARGS:
    cd documents-server && \
        DB_URL='{{ dev_db_url }}' \
        bun run {{ARGS}}

@run-backend-datasync:
    DATASYNC_MODE="true" just run-backend integrations sync

# [group('Testing')]

@test-backend:
    cd backend && \
        DB_URL='{{ test_db_url }}' \
        go test $(go list ./... | grep -v /ent/)


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

@codegen-migration NAME:
    just run-backend db-migrations generate {{NAME}}

# [group('Development Servers')]

@dev:
    process-compose --ordered-shutdown -f ./scripts/process-compose.yaml

@format:
    cd backend && go fmt ./...
    cd frontend && bun run format

@dev-backend:
    cd backend && reflex -s -d none -r '\.go$' -- just run-backend

@dev-frontend:
    just run-frontend dev

@dev-documents-server:
    just run-documents-server dev

# [group('Database')]

migrations_dir := "backend/migrations"

@create-initial-migrations:
    rm -f ./{{migrations_dir}}/*.{sql,sum}
    just run-backend db-migrations generate ent_init
    sleep 1
    migrate create -ext sql -dir "{{migrations_dir}}" river_init
    cd backend && go tool river migrate-get --all --exclude-version 1 --up > "migrations/$(ls migrations | grep 'river_init.up')"
    cd backend && go tool river migrate-get --all --exclude-version 1 --down > "migrations/$(ls migrations | grep 'river_init.down')"

@run-migrations:
    migrate -source "file://{{migrations_dir}}" -database "{{dev_db_url}}" up

@run-psql:
    psql -d {{dev_db_url}}
