# Repository Structure

`weko-autox` (Go module `github.com/ivis-kuroda/weko-autox`, binary `autox`) discovers
Weko/Invenio modules under `modules/*`, runs their `tox`/`pytest` suites inside the Docker
Compose `web` service, and writes logs under `log/`. It replaces the legacy `autox.sh` at the
repo root, which is kept only as historical reference (see README "Migration Notes"); new
behavior goes into the Go CLI, not the shell script.

## Layout

```
cmd/autox/main.go     thin entrypoint: build App, run it, map errors to exit codes
internal/app          wiring: constructs the dockerx client, runner, signal-aware context
internal/cli          flag/arg parsing (cobra), Config type, validation, help text
internal/module       module discovery (modules/*) and scope/include/exclude/partial selection
internal/dockerx      compose file discovery + Docker SDK exec into the "web" container
internal/runner       orchestrates a full run: setup, per-module exec, logs, coverage, progress
internal/signalx      SIGINT/SIGTERM -> context cancellation + stop callback
internal/report       stdout progress reporting (spinner, per-module status lines)
```

## Package responsibilities

| Package | Responsibility |
|---|---|
| `internal/app` | Wires `cli` + `dockerx` + `runner` + `signalx` + `report` together. No business logic. |
| `internal/cli` | Parses `os.Args` into a `Config`, validates argument order/combinations, prints help. See [cli-specification.md](cli-specification.md). |
| `internal/module` | Lists modules under `modules/*` and selects the target subset from `Config`. See [feature-specification.md](feature-specification.md#module-discovery--selection). |
| `internal/dockerx` | Finds the Compose `web` container and execs shell commands into it. See [feature-specification.md](feature-specification.md#docker-execution). |
| `internal/runner` | Drives one full run: install tox, iterate modules, dispatch by run mode, write logs, fetch coverage. See [feature-specification.md](feature-specification.md#run-orchestration). |
| `internal/signalx` | Wraps a context so SIGINT/SIGTERM trigger a stop callback then cancellation. See [feature-specification.md](feature-specification.md#signal-handling). |
| `internal/report` | Prints progress lines and a spinner to stdout. See [feature-specification.md](feature-specification.md#progress-reporting). |

## Control flow

```
cmd/autox/main.go
  -> app.New(stdout, stderr).Execute(ctx, os.Args[1:])
       -> cli.Execute
            -> cli.ParseArgs        (validates arg order, builds Config)
            -> run callback:
                 dockerx.New        (detect compose file, find "web" container)
                 runner.Runner{Workspace, Exec, Progress}.Run(ctx, cfg)
                   -> module.Detect + module.Select
                   -> per module: dockerx exec (tox/pytest), report.Progress, log files under log/
       -> signalx.WithShutdownSignal wraps ctx so Ctrl-C triggers Exec.StopTests, then cancels
```

## Build & test

- `go build ./...` or `make build` (binary at `./dist/autox`)
- `go test ./...` — all packages; no live Docker daemon required (dockerx/runner tests use
  fakes and pure helper functions instead of a real container)
- `go test ./internal/<pkg>` — single package
- `go test ./internal/runner -run TestName` — single test
- `go vet ./...` before finishing a change

## Conventions

- Keep `cmd/autox/main.go` and `internal/app` thin; put actual logic in the package it
  belongs to (`cli` for parsing, `module` for selection, `dockerx` for container exec,
  `runner` for orchestration).
- Wrap errors with `fmt.Errorf("...: %w", err)`. Sentinel errors (e.g. `runner.ErrTestsFailed`)
  are compared with `errors.Is`, never by string matching.
- Tests are table-driven, standard-library `testing` only (no assertion libraries). External
  effects (Docker exec, filesystem) are faked via interfaces (`dockerx.Executor`) rather than
  requiring a real Docker daemon in unit tests.
- Don't weaken the `-k`/Ctrl-C stop path (`signalx` + `dockerx.StopTests`) silently — it's the
  only way to stop long-running tox/pytest processes inside the container.

## Non-goals

- `autox.sh` is legacy/reference only; do not extend it instead of the Go CLI.
- No network calls beyond the Docker daemon; don't add telemetry/analytics.
