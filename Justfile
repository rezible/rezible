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
    just setup-db
    localias reload

@install-dependencies:
    cd backend && go mod tidy
    bun install

@upgrade-everything:
    devbox update
    cd backend && go get -u ./... && go mod tidy
    bun update

@run-backend *ARGS:
    cd backend && \
        DEBUG_MODE=true \
        DB_URL='{{ dev_db_url }}' \
        go run ./cmd/rezible {{ARGS}}

@run-frontend *ARGS:
    cd frontend && \
        PUBLIC_REZ_API_BASE_URL="/api/v1" \
        bun run {{ARGS}}

@run-documents-server *ARGS:
    cd documents-server && \
        DB_URL='{{ dev_db_url }}' \
        bun run {{ARGS}}

@run-backend-datasync: start-db
    DATASYNC_MODE="true" just run-backend integrations sync

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

# [group('Development Servers')]

@dev: stop-db
    process-compose --ordered-shutdown

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

@create-db: stop-db
    rm -rf ./.devbox/virtenv/postgresql/data
    initdb -A trust > /dev/null
    just start-db
    psql -q -d postgres -c "CREATE ROLE {{dev_db_user}} WITH LOGIN SUPERUSER;"
    createdb -U {{dev_db_user}} {{dev_db_name}}

migrations_dir := "backend/migrations"

@create-initial-migrations: create-db
    rm -f ./{{migrations_dir}}/*.{sql,sum}
    just run-backend db-migrations generate ent_init
    sleep 1
    migrate create -ext sql -dir "{{migrations_dir}}" river_init
    cd backend && go tool river migrate-get --all --exclude-version 1 --up > "migrations/$(ls migrations | grep 'river_init.up')"
    cd backend && go tool river migrate-get --all --exclude-version 1 --down > "migrations/$(ls migrations | grep 'river_init.down')"

@generate-migration NAME:
    just run-backend db-migrations generate {{NAME}}

@setup-db:
    just create-db
    migrate -source "file://backend/migrations" -database "{{dev_db_url}}" up

@run-db:
    -pg_isready -q || pg_ctl -o "-k $PGHOST" -l "$PGDATA/postgres.log" start

@start-db:
   just run-db > /dev/null

@run-psql:
    psql -d {{dev_db_url}}

@stop-db:
    -pg_isready -q && pg_ctl stop > /dev/null

# [group('Other')]
@run-oidc-test-provider:
    docker run -p 6432:8080 \
        -e DevOidcToolkit__Users__0__Email="${REZ_DEBUG_DEFAULT_USER_EMAIL}" \
        -e DevOidcToolkit__Users__0__FirstName=Test \
        -e DevOidcToolkit__Users__0__LastName=User \
        -e DevOidcToolkit__Clients__0__Id=client \
        -e DevOidcToolkit__Clients__0__Secret=secret \
        -e DevOidcToolkit__Clients__0__RedirectUris__INDEX=https://app.dev.rezible.com/api/auth/oidc/test-provider/callback \
        ghcr.io/businesssimulations/dev-oidc-toolkit:0.2.0
