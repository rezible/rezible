set shell := ["bash", "-uc"]

set dotenv-filename := ".env.workspace"
set dotenv-load

mod dev 'devenv'
mod backend 'packages/apps/backend'
mod frontend 'packages/apps/frontend'
mod documents-server 'packages/apps/documents-server'
mod packages 'packages'

@_default:
    just --list dev --unsorted --list-heading $'Development Workspace\n'
    just --list backend --unsorted --list-heading $'Backend\n'
    just --list frontend --unsorted --list-heading $'Frontend\n'
    just --list documents-server --unsorted --list-heading $'Documents Server\n'
    just --list packages --unsorted --list-heading $'Packages\n'

@codegen:
    just backend::codegen
    just packages::generate-api-client

[doc("Run all tests, or a backend/frontend/documents test target")]
@test target="all" *ARGS:
    just dev::test {{ target }} {{ ARGS }}

@regenerate-and-apply-db-schema:
    just backend::gen-schema
    just dev::setup-workspace --force
