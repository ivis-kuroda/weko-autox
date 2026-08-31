# weko-autox v2

Go-based CLI tool for running Weko module tests in Docker.

## Overview

`weko-autox` discovers modules under `modules/*`, executes tox/pytest inside the Docker `web` service, and writes logs under `log/`.

The command behavior is based on the previous shell script, with the following improvements:

- `cmd` is thin and all logic is implemented under `internal/`
- run strategy is selected by CLI (`--run-mode`)
- testable package structure (`cli`, `module`, `dockerx`, `runner`, `signalx`)

## Requirements

- Docker daemon reachable from the environment
- Compose file in workspace root (`compose.yaml`, `compose.yml`, `docker-compose.yml`, or `docker-compose.yaml`)
- Compose service name `web`
- Project mounted in container as `/code` (current runner commands expect `/code/modules/...`)

## Install Binary

Download a release asset from GitHub Releases and place `autox` in your `PATH`.

Example for Linux amd64 (`v2.0.0-alpha.2`):

```bash
VERSION="v2.0.0-alpha.2"
curl -fL -o /tmp/autox.tar.gz \
  "https://github.com/ivis-kuroda/weko-autox/releases/download/${VERSION}/autox_${VERSION#v}_linux_amd64.tar.gz"
tar -xzf /tmp/autox.tar.gz -C /tmp
install -m 0755 /tmp/autox /usr/local/bin/autox
```

For macOS, replace `linux_amd64` with `darwin_amd64` or `darwin_arm64`.

If you do not have permission to write `/usr/local/bin`, install to `$HOME/.local/bin` and add it to `PATH`.

## Quick Start

Run version:

```bash
autox -v
```

Run all modules explicitly with the default strategy (`all-at-once`):

```bash
autox all
```

## CLI Usage

```bash
autox [all|weko|invenio] [options]
autox [module ...] [options]
autox <module> -p <selector> [-p <selector> ...] [options]
autox [options]
```

`options` means arguments starting with `-` or `--`.

Argument order rules:

- If the first argument is a scope or module name, options can follow.
- If the first argument is an option, run in options-only mode (no scope/module positional arguments).
- Scope and module positional arguments are mutually exclusive, except when `-n/--exclude` is set.
- If no scope, module, or standalone option (`-k`, `-v`, `-h`) is provided, the command is invalid and help is shown.

### Scope (first positional argument)

- `all`: all detected modules
- `weko`: modules containing `weko` in the module name
- `invenio`: modules containing `invenio` in the module name

If scope is omitted, it is not treated as `all`.
Provide a module argument or a standalone option instead.

### Options

- `-n, --exclude`: treat positional module names as exclude list
- `-r, --reset`: remove local artifacts (`*.egg-info`, `.tox`, `htmlcov`, `coverage.xml`) before running
- `--run-mode <mode>`: test run strategy
- `-p, --partial <selector>`: partial test selector (repeatable), requires one positional module argument
- `-o, --output <name>`: output subdirectory under `log/`
- `-k, --kill`: stop running tox/pytest processes in the target container
- `-v, --version`: print version
- `-h, --help`: print help

### Run Modes

- `all-at-once` (default)
  - Runs `tox` for each selected module
- `per-file`
  - Prepares tox env once, then runs pytest for each `tests/test_*.py`
- `per-func`
  - For `-p/--partial`, runs pytest once per selector

With `-p/--partial`:

- `--run-mode all-at-once`: runs one pytest command with all provided selectors
- `--run-mode per-func`: runs one pytest command per selector
- `--run-mode per-file` behaves like `per-func` for backward compatibility

## Examples

Run all modules:

```bash
autox all
```

Run weko modules except `weko-admin`:

```bash
autox weko weko-admin -n
```

Run per-file strategy for invenio modules:

```bash
autox invenio --run-mode per-file
```

Run partial tests for one module:

```bash
autox weko-admin -p test_api.py::test_is_restricted_user -p test_tasks.py::test_send_all_reports
```

Run partial selectors all at once in a single pytest command:

```bash
autox weko-admin --run-mode all-at-once -p test_api.py::test_one -p test_api.py::test_two
```

Run partial selectors one by one:

```bash
autox weko-admin --run-mode per-func -p test_api.py::test_one -p test_api.py::test_two
```

Write logs under `log/example/`:

```bash
autox all -o example
```

Stop running tox/pytest processes:

```bash
autox -k
```

Invalid combinations or missing targets print help and exit.

## Output

- Default: `log/<module>/...`
- With `-o name`: `log/name/<module>/...`

Typical files:

- `test_all.log`
- `install.log` (per-file mode)
- `test_*.log` (per-file mode)
- `partial.log` (`-p` with `--run-mode all-at-once`)
- `partialN.log` (`-p` with `--run-mode per-func` or `--run-mode per-file`)
- `coverage.log`

## Note
The log files are stored in the log directory.

> [!IMPORTANT]
> The following conditions must be satisfied in order for the progress to be displayed correctly
> - The display must fit on a single line.
> - docker does not issue a warning. Create a file in the project root as shown below.
>
>       # .env
>       ELASTICSEARCH_S3_ACCESS_KEY=
>       ELASTICSEARCH_S3_SECRET_KEY=
>       ELASTICSEARCH_S3_ENDPOINT=
>       ELASTICSEARCH_S3_BUCKET=

## Development

Go 1.26+ is required for development and building from source.

Run all tests:

```bash
go test ./...
```

Run specific package tests:

```bash
go test ./internal/dockerx
go test ./internal/runner
```

Run a specific test:

```bash
go test ./internal/runner -run TestRunnerRunPerFile
```

## Dependency Update Notes

For routine updates:

```bash
go get -u=patch ./...
go mod tidy
go test ./...
```

For larger updates (including direct dependencies like Docker SDK), update in steps and verify after each step.

## Migration Notes (Shell Script -> Go CLI) 🚚

This project has been migrated from `autox.sh` to the Go-based `autox` CLI.

### ✅ Behavior kept from `autox.sh`

- Scope-based selection (`all`, `weko`, `invenio`)
- Exclude mode (`-n`), clean mode (`-r`), partial mode (`-p`), output directory (`-o`), stop (`-k`), version (`-v`)
- Coverage report output per module (`coverage.log`)
- Log output under `log/<module>/...` or `log/<output>/<module>/...`
- Always remove `<module>.egg-info` and `.eggs` before each module run

### 🔁 Behavior changed in Go implementation

- Container execution backend changed from `docker-compose exec` to Docker SDK-based exec.
  - Compose files are auto-detected from workspace root.
  - The `web` service is validated before execution.
- Progress output is now line-based and stable (no terminal overwrite animation).
- Stop behavior (`-k`) now runs `pkill -f 'tox|pytest'` inside the target container.
- Version output is now build-time injectable (`autox - <version>`), instead of fixed script text.
- `--run-mode per-file` is now explicitly selectable for any target module set.

### ⚠️ Spec differences to be aware of

- Log directory cleanup behavior differs.
  - `autox.sh`: frequently removed module log directories before runs.
  - Go CLI: creates directories as needed and writes files; existing files may remain unless overwritten.
- Module argument validation differs.
  - `autox.sh`: unknown arguments were rejected early.
  - Go CLI: unknown module names are filtered during selection; if nothing remains, execution fails with selection error.
- Per-file install phase differs.
  - `autox.sh`: started `tox` in background, watched install log, then killed install process.
  - Go CLI: runs `tox -e c1 --notest` synchronously, then executes file-by-file pytest.

### 🎯 Why this migration

- Better testability (unit tests for CLI/module selection/docker runner/signal handling)
- Cleaner architecture (`internal/*` packages)
- More robust release automation with versioned binary distribution


## Change Log
### v2.0.0-alpha.1 / v2.0.0-alpha.2
- Note: There is no functional difference between `v2.0.0-alpha.1` and `v2.0.0-alpha.2`. `v2.0.0-alpha.2` exists because the `v2.0.0-alpha.1` release failed during release automation.
- Migrate runtime from shell script to Go CLI (`autox`).
- Add binary-first usage and release distribution flow.
- Add GitHub Actions + GoReleaser automation for release artifacts.
- Add `make build-snapshot` for local snapshot builds via GoReleaser.
- Improve version output to use build-time injected version.
- Keep shell-compatible cleanup behavior for `<module>.egg-info` and `.eggs` (always removed).

### v1.2.1
fix: create log directories only when needed to avoid failures during partial test runs.  
change: suppress excessive per-file pytest logging when running modules separately.  
docs: add installation steps.  
change: adopt semantic version notation for the version display in preparation for the Go implementation.

### ver.1.2.0
add options: `-p`; :clap: Tests can now be run on a per-function basis.  
Coverage reports are now output after tests.

### ver.1.1.1
delete options: `-i` and `-w`.  
add options: `-o`; specify the output directory for the log files. `-v`; show the version.

### ver.1.1.0
add commands: all, invenio, weko.

### ver.1.0
the first script.
