set shell := ["bash", "-uc"]

set dotenv-filename := ".env.workspace"

mod dev 'devenv'
mod backend 'packages/apps/backend'
mod frontend 'packages/apps/frontend'
mod documents-server 'packages/apps/documents-server'
mod packages 'packages'

@_default:
    just --list dev --unsorted --list-heading $'Development Workspace\n'
    just --list backend --unsorted --list-heading $'Backend\n'
    just --list backend --unsorted --list-heading $'Frontend\n'
    just --list backend --unsorted --list-heading $'Documents Server\n'
    just --list packages --unsorted --list-heading $'Packages\n'
