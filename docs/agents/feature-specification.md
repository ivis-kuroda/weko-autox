# Feature Specification

Runtime/functional behavior of `autox` once a `cli.Config` has been parsed (see
[cli-specification.md](cli-specification.md) for how `Config` is produced). This covers
`internal/module`, `internal/dockerx`, `internal/runner`, `internal/signalx`, and
`internal/report`.

## Module discovery & selection

Implemented in `internal/module`.

- `Detect(modulesDir)` lists immediate subdirectories of `modules/` that contain a `tests/`
  directory, sorted by name. A module without a `tests/` dir is invisible to the rest of the
  tool.
- `Select(modules, cfg)`:
  1. If `cfg.PartialSelectors` is set or `cfg.RunMode == cli.RunModePartial`, short-circuits
     to a single-module lookup by `cfg.PartialModule` (error if not found).
  2. Otherwise filters by `cfg.Scope` (`""`/`all` = passthrough, `weko`/`invenio` = substring
     match on module name).
  3. Then applies `IncludeModules` (allow-list) and `ExcludeModules` (deny-list) on top of
     the scope result.
  4. An empty result at any stage is an error, not a silent no-op.

## Docker execution

Implemented in `internal/dockerx`.

- `DetectComposeFile(workspace)` checks, in order, `compose.yaml`, `compose.yml`,
  `docker-compose.yml`, `docker-compose.yaml` in the workspace root.
- `LoadProject` parses the compose file via `compose-go`; `normalizeProjectName` sanitizes
  the workspace directory name into a compose-safe project name (lowercase, `[a-z0-9_-]`).
- `New(ctx, workspace, serviceName)` resolves the compose project, verifies `serviceName`
  exists, then finds a container by Docker labels `com.docker.compose.service` /
  `com.docker.compose.project`, preferring a running container over a stopped one.
- The `Executor` interface (`Exec`, `ExecStream`, `StopTests`) is what `internal/runner`
  depends on; extend this interface rather than exposing `*Client` internals to callers.
- `ExecStream` runs `sh -lc <shellCommand>` in the target container, tees stdout/stderr to
  caller-provided writers, and returns an error wrapping a non-zero exit code; ctx
  cancellation closes the exec-attach stream so output copying unblocks immediately.
- `StopTests` runs `pkill -f 'tox|pytest' || true` inside the container — this is the only
  kill switch for `-k`/Ctrl-C.
- No live Docker daemon is used in unit tests; tests exercise pure helpers and the
  "compose file missing" error path, faking `Executor` where a caller needs one (see
  `internal/runner/runner_test.go`'s `fakeExecutor`).

## Run orchestration

Implemented in `internal/runner` (`Runner{Workspace, Exec, Progress}.Run(ctx, cfg)`).

1. `cfg.Kill` short-circuits to `Exec.StopTests` and returns.
2. `module.Detect` under `<workspace>/modules`, then `module.Select(modules, cfg)`.
3. `installTox` runs once (`pip3 install tox==3.28 tox-setuptools-version`) before the module
   loop.
4. Per module: report setup progress, prepare the log dir, always clean
   `<module>.egg-info`/`.eggs`, additionally clean `.tox`/`htmlcov`/`coverage.xml` when
   `cfg.Clean` (`-r/--reset`), erase coverage, then dispatch by run mode:
   - partial selectors -> all-at-once or per-selector pytest run (see
     [cli-specification.md](cli-specification.md#run-modes))
   - `all-at-once` -> single `tox` run, logged to `test_all.log`
   - `per-file` -> `tox -e c1 --notest` once (`install.log`), then one pytest run per
     `tests/test_*.py` file
5. `fetchCoverage` parses `coverage report` output with a `TOTAL ... NN%` regex; no match
   yields `"0"`, not an error.
6. Failures are aggregated per module; the loop keeps running other modules even if one
   module's tests fail (a non-zero exit from tox/pytest is a test failure, not a hard error),
   but a real exec error (e.g. Docker/context error) aborts the whole run immediately.
7. If any module had test failures, `Run` returns the sentinel `runner.ErrTestsFailed`.
   Callers (`cmd/autox/main.go`) check this with `errors.Is` and suppress the generic error
   print, since `Progress.DoneWithFailures()` has already reported it.

### Logging convention

`log/<output>/<module>/<fileName>` — one file per command, containing interleaved
stdout+stderr. Typical files: `test_all.log`, `install.log` + `test_*.log` (per-file mode),
`partial.log` (partial + all-at-once) or `partialN.log` (partial + per-func), `coverage.log`.
`<output>` defaults to empty (i.e. `log/<module>/...`) unless `-o/--output` is set.

## Signal handling

Implemented in `internal/signalx`.

`WithShutdownSignal(ctx, onSignal)` returns a derived context cancelled on
`SIGINT`/`SIGTERM`, invoking `onSignal` first. `internal/app` passes a callback that calls
`execClient.StopTests`, so the in-container tox/pytest process is killed before the CLI
exits. The returned cancel func must be called to stop listening and release the signal
channel.

## Progress reporting

Implemented in `internal/report`.

`Progress` writes line-based status to stdout: `Setup`, `StartSpinner` (+ its stop func),
`ModuleDone` (includes coverage %), `Done`, `DoneWithFailures`. `Progress` also defines
`ModuleStart`, but no caller in the current `Run` loop invokes it. Per the README's
display requirement, progress output must fit on a single line: the spinner overwrites its
own line, so callers must not interleave other writes to the same writer while a spinner is
running, and must call the returned stop func before printing the next status line.
