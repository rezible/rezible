set shell := ["bash", "-uc"]

mod dev 'devenv'
mod test 'devenv/tests.Justfile'
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

[doc("Install dependencies and provision the workspace")]
@setup:
    just backend::install
    bun install --frozen-lockfile
    just dev::setup-workspace

[doc("Generate all code")]
@codegen:
    just backend::codegen
    just packages::generate-api-client

[doc("Regenerate Ent and the initial migration, then recreate this workspace's database")]
@regenerate-and-apply-db-schema:
    just backend::gen-schema
    just dev::setup-database --force --no-migrate
    just backend::create-initial-migration
    just backend::apply-migrations
