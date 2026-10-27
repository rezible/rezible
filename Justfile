set shell := ["bash", "-uc"]

mod dev 'devenv'
mod sim 'devenv/simulations'
mod test 'devenv/test.Justfile'
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
    bun run --filter=@rezible/api-client-ts build
