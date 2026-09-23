set shell := ["bash", "-euc"]

[working-directory("..")]
@_process-compose *ARGS:
    process-compose --no-server --disable-dotenv \
      -f {{ quote(source_dir() / "process-compose.tests.yaml") }} \
      {{ ARGS }}

[doc("Run the complete test suite.")]
[default]
all: (_process-compose "up" "tests-complete")

[doc("Run backend tests, optionally selecting packages or test arguments.")]
backend *ARGS: (_process-compose "run" "backend" "--" ARGS)

[doc("Check the frontend.")]
frontend: (_process-compose "run" "frontend")

[doc("Run documents server tests.")]
documents: (_process-compose "run" "documents")
