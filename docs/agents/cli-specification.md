# CLI Specification

User-facing behavior for the `autox` command, plus a map to the code that implements each
rule (for making changes). Implemented under `internal/cli`.

## Usage forms

```
autox [all|weko|invenio] [options]
autox [module ...] [options]
autox <module> -p <selector> [-p <selector> ...] [options]
autox [options]
```

`options` means arguments starting with `-` or `--`.

## Argument order rules (`internal/cli/arg_order.go`)

- If the first argument is a scope or module name, options may follow it.
- If the first argument is an option, the command runs in options-only mode: no
  scope/module positional arguments are allowed anywhere in the args.
- Scope and module positional arguments are mutually exclusive, except when `-n`/`--exclude`
  is set (then the positionals after the scope are treated as the exclude list).
- If no scope, module, or standalone option (`-k`, `-v`, `-h`) is provided, the command is
  invalid and help is printed.
- `validateArgOrder` decides "options-only mode" before cobra parses anything; it also owns
  the `longOptionArity` / `shortOptionArity` maps that say whether a flag consumes the next
  token (needed to correctly skip flag values while scanning for stray positionals).

## Scope keyword (first positional argument, `internal/cli/specs.go` `ScopeSpecs`)

| Keyword | Meaning |
|---|---|
| `all` | all detected modules |
| `weko` | modules whose name contains `weko` |
| `invenio` | modules whose name contains `invenio` |

If scope is omitted, it is **not** treated as `all` — provide a module argument or a
standalone option instead.

## Options (`internal/cli/specs.go` `OptionSpec` values)

| Flag | Short | Default | Meaning |
|---|---|---|---|
| `--exclude` | `-n` | `false` | treat positional module names as an exclude list |
| `--reset` | `-r` | `false` | remove `*.egg-info`, `.tox`, `htmlcov`, `coverage.xml` before running |
| `--run-mode <mode>` | — | `all-at-once` | test run strategy (see below) |
| `--partial <selector>` | `-p` | `nil` | partial test selector, repeatable, requires exactly one positional module argument |
| `--output <name>` | `-o` | `""` | output subdirectory under `log/` |
| `--kill` | `-k` | `false` | stop running tox/pytest processes in the target container |
| `--version` | `-v` | `false` | print version |
| `--help` | `-h` | `false` | print help |

## Run modes (`internal/cli/config.go` `RunMode`, dispatched in `internal/runner`)

- `all-at-once` (default) — runs `tox` once per selected module.
- `per-file` — prepares the tox env once (`tox -e c1 --notest`), then runs pytest once per
  `tests/test_*.py` file.
- `per-func` — with `-p/--partial`, runs pytest once per selector.
- `partial` (legacy alias) — normalized to `per-func` during parsing
  (`internal/cli/command.go`).

### With `-p/--partial`

- `--run-mode all-at-once`: one pytest command with all provided selectors joined together.
- `--run-mode per-func`: one pytest command per selector.
- `--run-mode per-file`: behaves like `per-func` for backward compatibility.

## Examples

```bash
autox all                                    # run all modules
autox weko weko-admin -n                     # run weko modules except weko-admin
autox invenio --run-mode per-file            # per-file strategy for invenio modules
autox weko-admin -p test_api.py::test_one -p test_api.py::test_two   # partial, per-func default
autox weko-admin --run-mode all-at-once -p test_api.py::test_one -p test_api.py::test_two
autox all -o example                         # logs under log/example/
autox -k                                      # stop running tox/pytest processes
```

Invalid combinations or missing targets print help and exit non-zero.

## Adding or changing a flag — touch all of

1. `internal/cli/specs.go` — add/update the `OptionSpec` and any new error message
   constants (`Err...`).
2. `internal/cli/command.go` — bind the flag in `ParseArgs`, and use it in `RunE` if it
   affects positional-argument resolution (like `--exclude` or `--partial` do).
3. `internal/cli/arg_order.go` — add arity entries to both `longOptionArity` and (if it has a
   short form) `shortOptionArity`.
4. `internal/cli/help.go` — add it to the printed options list (help output is hand-formatted,
   not cobra's default, so it does not update itself).
5. `internal/cli/command_test.go` — add parse-success/parse-error cases.
6. `README.md` — update the user-facing CLI usage section to match.

## Gotchas

- Partial mode forces exactly one positional module argument; `RunMode` defaults to
  `all-at-once` unless already set.
- `versionString` (`internal/cli/command.go`) is overwritten at build time via
  `-ldflags -X github.com/ivis-kuroda/weko-autox/internal/cli.versionString=...` (see
  `Makefile`); don't rename that variable without updating the Makefile/goreleaser config.
