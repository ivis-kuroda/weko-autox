package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ivis-kuroda/weko-autox/internal/cli"
	"github.com/ivis-kuroda/weko-autox/internal/dockerx"
	"github.com/ivis-kuroda/weko-autox/internal/module"
	"github.com/ivis-kuroda/weko-autox/internal/report"
)

type Runner struct {
	Workspace string
	Exec      dockerx.Executor
	Progress  report.Progress
}

var ErrTestsFailed = errors.New("some tests failed")

func (r Runner) Run(ctx context.Context, cfg cli.Config) error {
	if cfg.Kill {
		return r.Exec.StopTests(ctx)
	}

	modules, err := module.Detect(filepath.Join(r.Workspace, "modules"))
	if err != nil {
		return err
	}

	targets, err := module.Select(modules, cfg)
	if err != nil {
		return err
	}

	if err := r.installTox(ctx); err != nil {
		return err
	}

	hadTestFailures := false

	for i, m := range targets {
		r.Progress.Setup(m.Name, i+1, len(targets))
		if err := r.prepareModuleLogDir(m.Name, cfg.OutputDirName); err != nil {
			return err
		}
		if err := cleanModuleBuildArtifacts(m.Path, m.Name); err != nil {
			return err
		}
		if cfg.Clean {
			if err := cleanResetArtifacts(m.Path); err != nil {
				return err
			}
		}

		_, _ = r.Exec.Exec(ctx, fmt.Sprintf("cd /code/modules/%s; .tox/c1/bin/coverage erase", m.Name))
		stopSpinner := r.Progress.StartSpinner(fmt.Sprintf("%s progressing.", m.Name))

		if len(cfg.PartialSelectors) > 0 {
			failed, err := r.runPartial(ctx, m.Name, cfg.OutputDirName, cfg.PartialSelectors, cfg.RunMode)
			hadTestFailures = hadTestFailures || failed
			if err != nil {
				stopSpinner()
				return err
			}
		} else {
			switch cfg.RunMode {
			case cli.RunModeAllAtOnce:
				failed, err := r.runAllAtOnce(ctx, m.Name, cfg.OutputDirName)
				hadTestFailures = hadTestFailures || failed
				if err != nil {
					stopSpinner()
					return err
				}
			case cli.RunModePerFile:
				failed, err := r.runPerFile(ctx, m, cfg.OutputDirName)
				hadTestFailures = hadTestFailures || failed
				if err != nil {
					stopSpinner()
					return err
				}
			default:
				stopSpinner()
				return fmt.Errorf("unsupported run mode: %s", cfg.RunMode)
			}
		}

		coverage, err := r.fetchCoverage(ctx, m.Name, cfg.OutputDirName)
		stopSpinner()
		if err != nil {
			return err
		}
		r.Progress.ModuleDone(m.Name, coverage, i+1, len(targets))
	}

	if hadTestFailures {
		r.Progress.DoneWithFailures()
		return ErrTestsFailed
	}

	r.Progress.Done()
	return nil
}

func (r Runner) installTox(ctx context.Context) error {
	_, err := r.Exec.Exec(ctx, "pip3 install tox==3.28 tox-setuptools-version")
	return err
}

func (r Runner) runAllAtOnce(ctx context.Context, moduleName string, outputName string) (bool, error) {
	res, err := r.execToModuleLog(ctx, moduleName, outputName, "test_all.log", fmt.Sprintf("cd /code/modules/%s; tox", moduleName))
	if err != nil {
		if isNonZeroExit(err, res) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (r Runner) runPerFile(ctx context.Context, m module.Module, outputName string) (bool, error) {
	install, err := r.execToModuleLog(ctx, m.Name, outputName, "install.log", fmt.Sprintf("cd /code/modules/%s; tox -e c1 --notest", m.Name))
	if err != nil {
		return false, err
	}
	_ = install

	hadTestFailures := false

	testFiles, err := filepath.Glob(filepath.Join(m.Path, "tests", "test_*.py"))
	if err != nil {
		return false, fmt.Errorf("glob test files: %w", err)
	}
	for _, file := range testFiles {
		base := filepath.Base(file)
		res, execErr := r.execToModuleLog(ctx, m.Name, outputName, strings.TrimSuffix(base, ".py")+".log", fmt.Sprintf("cd /code/modules/%s; .tox/c1/bin/pytest --cov=%s tests/%s -v --cov-append --cov-branch --cov-report=term --cov-report=html -W ignore --basetemp=/code/modules/%s/.tox/c1/tmp", m.Name, strings.ReplaceAll(m.Name, "-", "_"), base, m.Name))
		if execErr != nil {
			if isNonZeroExit(execErr, res) {
				hadTestFailures = true
				continue
			}
			return false, execErr
		}
	}
	return hadTestFailures, nil
}

func (r Runner) runPartial(ctx context.Context, moduleName string, outputName string, selectors []string, runMode cli.RunMode) (bool, error) {
	switch runMode {
	case cli.RunModeAllAtOnce:
		return r.runPartialAllAtOnce(ctx, moduleName, outputName, selectors)
	case cli.RunModePerFunc, cli.RunModePerFile, cli.RunModePartial:
		return r.runPartialPerSelector(ctx, moduleName, outputName, selectors)
	default:
		return false, fmt.Errorf("unsupported run mode for partial selectors: %s", runMode)
	}
}

func (r Runner) runPartialAllAtOnce(ctx context.Context, moduleName string, outputName string, selectors []string) (bool, error) {
	targets := make([]string, 0, len(selectors))
	for _, selector := range selectors {
		targets = append(targets, "tests/"+selector)
	}

	res, err := r.execToModuleLog(
		ctx,
		moduleName,
		outputName,
		"partial.log",
		fmt.Sprintf(
			"cd /code/modules/%s; .tox/c1/bin/pytest --cov=%s %s -v -vv -s --cov-append --cov-branch --cov-report=term --cov-report=html -W ignore --basetemp=/code/modules/%s/.tox/c1/tmp",
			moduleName,
			strings.ReplaceAll(moduleName, "-", "_"),
			strings.Join(targets, " "),
			moduleName,
		),
	)
	if err != nil {
		if isNonZeroExit(err, res) {
			return true, nil
		}
		return false, err
	}

	return false, nil
}

func (r Runner) runPartialPerSelector(ctx context.Context, moduleName string, outputName string, selectors []string) (bool, error) {
	hadTestFailures := false

	for i, selector := range selectors {
		res, err := r.execToModuleLog(ctx, moduleName, outputName, fmt.Sprintf("partial%d.log", i+1), fmt.Sprintf("cd /code/modules/%s; .tox/c1/bin/pytest --cov=%s tests/%s -v -vv -s --cov-append --cov-branch --cov-report=term --cov-report=html -W ignore --basetemp=/code/modules/%s/.tox/c1/tmp", moduleName, strings.ReplaceAll(moduleName, "-", "_"), selector, moduleName))
		if err != nil {
			if isNonZeroExit(err, res) {
				hadTestFailures = true
				continue
			}
			return false, err
		}
	}
	return hadTestFailures, nil
}

func (r Runner) fetchCoverage(ctx context.Context, moduleName string, outputName string) (string, error) {
	res, err := r.execToModuleLog(ctx, moduleName, outputName, "coverage.log", fmt.Sprintf("cd /code/modules/%s; .tox/c1/bin/coverage report", moduleName))
	if err != nil {
		if isNonZeroExit(err, res) {
			return "0", nil
		}
		return "", err
	}

	re := regexp.MustCompile(`TOTAL\s+\d+\s+\d+\s+(\d+)%`)
	m := re.FindStringSubmatch(res.Stdout + "\n" + res.Stderr)
	if len(m) < 2 {
		return "0", nil
	}
	return m[1], nil
}

func (r Runner) execToModuleLog(ctx context.Context, moduleName string, outputName string, fileName string, shellCommand string) (dockerx.ExecResult, error) {
	path := filepath.Join(moduleLogDir(r.Workspace, outputName, moduleName), fileName)
	logFile, err := os.Create(path)
	if err != nil {
		return dockerx.ExecResult{}, fmt.Errorf("create log file %s: %w", path, err)
	}
	defer logFile.Close()

	return r.Exec.ExecStream(ctx, shellCommand, logFile, logFile)
}

func isNonZeroExit(err error, res dockerx.ExecResult) bool {
	if err == nil {
		return false
	}
	// Context cancellation is an interruption, not a test failure.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var exitErr interface{ ExitCode() int }
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode() != 0
	}

	if strings.Contains(err.Error(), "exec failed with code") {
		return true
	}

	return res.ExitCode != 0
}

func (r Runner) prepareModuleLogDir(moduleName string, outputName string) error {
	dir := moduleLogDir(r.Workspace, outputName, moduleName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	return nil
}

func (r Runner) writeModuleLog(outputName string, moduleName string, fileName string, body string) error {
	path := filepath.Join(moduleLogDir(r.Workspace, outputName, moduleName), fileName)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write log file %s: %w", path, err)
	}
	return nil
}

func moduleLogDir(workspace string, outputName string, moduleName string) string {
	if outputName == "" {
		return filepath.Join(workspace, "log", moduleName)
	}
	return filepath.Join(workspace, "log", outputName, moduleName)
}

func cleanModuleBuildArtifacts(modulePath string, moduleName string) error {
	paths := []string{
		filepath.Join(modulePath, moduleName+".egg-info"),
		filepath.Join(modulePath, ".eggs"),
	}
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}
	return nil
}

func cleanResetArtifacts(modulePath string) error {
	paths := []string{
		filepath.Join(modulePath, ".tox"),
		filepath.Join(modulePath, "htmlcov"),
		filepath.Join(modulePath, "coverage.xml"),
	}
	for _, p := range paths {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("remove %s: %w", p, err)
		}
	}
	return nil
}
