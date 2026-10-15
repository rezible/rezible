set shell := ["bash", "-euc"]

[doc("Run every suite in parallel. All run to completion; the run fails if any suite fails.")]
[default]
[parallel]
all: (_suite "backend") (_suite "frontend") (_suite "documents")

# Runs one suite with its output prefixed by the suite name, dropping lines that are noise in a combined run.
[working-directory("..")]
@_suite NAME:
    set -o pipefail; just test {{ NAME }} 2>&1 \
      | sed -u -e '/\[no test files\]$/d' -e 's/^/[{{ NAME }}] /'

[doc("Run backend tests, optionally selecting packages or test arguments.")]
[working-directory("..")]
[positional-arguments]
@backend *ARGS:
    just dev::infra-up postgres-test
    just backend::test-selection "$@"

[doc("Test the frontend.")]
[working-directory("..")]
@frontend:
    just frontend::run test

[doc("Run documents server tests.")]
[working-directory("..")]
@documents:
    just documents-server::test
