set shell := ["bash", "-euc"]

[doc("Run every suite in parallel. All run to completion, then a summary is printed.")]
[default]
[working-directory("..")]
all:
    #!/usr/bin/env bash
    set -uo pipefail
    suites=(backend frontend documents)
    results=$(mktemp -d)
    trap 'rm -rf "$results"' EXIT
    for suite in "${suites[@]}"; do
      (
        just test "$suite" 2>&1 \
          | sed -u -e '/\[no test files\]$/d' -e '/^ Container .* \(Running\|Waiting\|Healthy\) *$/d' -e "s/^/[$suite] /"
        echo "${PIPESTATUS[0]}" > "$results/$suite"
      ) &
    done
    wait
    failed=0
    echo
    for suite in "${suites[@]}"; do
      code=$(cat "$results/$suite")
      if [[ "$code" == 0 ]]; then
        echo "$suite: ok"
      else
        echo "$suite: FAILED (exit $code)"
        failed=1
      fi
    done
    exit "$failed"

[doc("Run backend tests, optionally selecting packages or test arguments.")]
[working-directory("..")]
[positional-arguments]
backend *ARGS:
    just dev::infra-up postgres-test
    just backend::test-selection "$@"

[doc("Test the frontend.")]
[working-directory("..")]
frontend:
    just frontend::run test

[doc("Run documents server tests.")]
[working-directory("..")]
documents:
    just documents-server::test
