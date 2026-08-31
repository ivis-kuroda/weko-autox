# AGENTS.md

## Overview

`weko-autox` (Go module `github.com/ivis-kuroda/weko-autox`, binary `autox`) is a CLI that
discovers Weko/Invenio modules under `modules/*`, runs their `tox`/`pytest` suites inside the
Docker Compose `web` service, and writes logs under `log/`. It replaces the legacy `autox.sh`
at the repo root, which is kept only as historical reference (see README "Migration Notes");
new behavior goes into the Go CLI, not the shell script.

## Documentation map

Detailed, role-based context lives under `docs/agents/` instead of being scattered next to
the code. Read the relevant doc before making a change instead of loading everything up
front:

| Doc | Covers |
|---|---|
| [docs/agents/repository-structure.md](docs/agents/repository-structure.md) | Package layout, responsibilities, control flow, build/test commands, repo-wide conventions. |
| [docs/agents/cli-specification.md](docs/agents/cli-specification.md) | User-facing CLI usage (scopes, options, run modes) and which `internal/cli` file implements each rule. |
| [docs/agents/feature-specification.md](docs/agents/feature-specification.md) | Runtime behavior once a `Config` is parsed: module selection, Docker exec, run orchestration, logging, signal handling, progress reporting. |

## Quick reference

- Build: `go build ./...` or `make build` (binary at `./dist/autox`)
- Test: `go test ./...` (no live Docker daemon required; see repository-structure.md)
- Keep `cmd/autox/main.go` and `internal/app` thin; business logic belongs in `cli`,
  `module`, `dockerx`, or `runner` — see repository-structure.md for the full layout.
- `autox.sh` is legacy/reference only; do not extend it instead of the Go CLI.

## Development environment

The Go toolchain is provided via the devcontainer under `.devcontainer/` (based on
`mcr.microsoft.com/devcontainers/go`), so there is no need to install Go locally — open the
repo in the devcontainer and it is ready to build/test.

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) for the subject line:
`<type>(<scope>): <short description>` (scope is optional; e.g. `feat: ...`, `fix(cli): ...`,
`docs: ...`).

- A commit with a single, focused change needs only the subject line — a short, clear
  description is enough, no body required.
- Only when a commit bundles multiple changes together, add a body with short bullet
  points stating what was done and why, e.g.:

```
fix: prevent duplicate log directories

- reuse the existing log directory instead of creating a new one, because
  reruns were losing previous output
- guard directory creation with an existence check, to avoid a race when
  multiple runs start concurrently
```

- Bullet points do not start with a capital letter and do not end with a period.
- Keep each bullet to roughly one line (about 12-15 words); avoid stacking multiple
  subordinate clauses ("because... so...") that wrap across three or more lines.

- If Claude authored or materially contributed to the change, add a trailer (blank line,
  then the trailer) so the commit carries a co-author. Replace `Claude` with the specific
  model name in use (e.g. `Claude Sonnet 5`), not the generic assistant name:

```
Co-authored-by: Claude Sonnet 5 <noreply@anthropic.com>
```
