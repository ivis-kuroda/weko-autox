package runner

import (
	"context"
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
		r.Progress.ModuleStart(m.Name, i+1, len(targets))

		switch cfg.RunMode {
		case cli.RunModeAllAtOnce:
			if err := r.runAllAtOnce(ctx, m.Name, cfg.OutputDirName); err != nil {
				return err
			}
		case cli.RunModePerFile:
			if err := r.runPerFile(ctx, m, cfg.OutputDirName); err != nil {
				return err
			}
		case cli.RunModePartial:
			if err := r.runPartial(ctx, m.Name, cfg.OutputDirName, cfg.PartialSelectors); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported run mode: %s", cfg.RunMode)
		}

		coverage, err := r.fetchCoverage(ctx, m.Name, cfg.OutputDirName)
		if err != nil {
			return err
		}
		r.Progress.ModuleDone(m.Name, coverage, i+1, len(targets))
	}

	r.Progress.Done()
	return nil
}

func (r Runner) installTox(ctx context.Context) error {
	_, err := r.Exec.Exec(ctx, "pip3 install tox==3.28 tox-setuptools-version")
	return err
}

func (r Runner) runAllAtOnce(ctx context.Context, moduleName string, outputName string) error {
	res, err := r.Exec.Exec(ctx, fmt.Sprintf("cd /code/modules/%s; tox", moduleName))
	if err != nil {
		_ = r.writeModuleLog(outputName, moduleName, "test_all.log", res.Stdout+res.Stderr)
		return err
	}
	return r.writeModuleLog(outputName, moduleName, "test_all.log", res.Stdout+res.Stderr)
}

func (r Runner) runPerFile(ctx context.Context, m module.Module, outputName string) error {
	install, err := r.Exec.Exec(ctx, fmt.Sprintf("cd /code/modules/%s; tox -e c1 --notest", m.Name))
	if err != nil {
		_ = r.writeModuleLog(outputName, m.Name, "install.log", install.Stdout+install.Stderr)
		return err
	}
	if err := r.writeModuleLog(outputName, m.Name, "install.log", install.Stdout+install.Stderr); err != nil {
		return err
	}

	testFiles, err := filepath.Glob(filepath.Join(m.Path, "tests", "test_*.py"))
	if err != nil {
		return fmt.Errorf("glob test files: %w", err)
	}
	for _, file := range testFiles {
		base := filepath.Base(file)
		res, execErr := r.Exec.Exec(ctx, fmt.Sprintf("cd /code/modules/%s; .tox/c1/bin/pytest --cov=%s tests/%s -v --cov-append --cov-branch --cov-report=term --cov-report=html --basetemp=/code/modules/%s/.tox/c1/tmp --full-trace", m.Name, strings.ReplaceAll(m.Name, "-", "_"), base, m.Name))
		if err := r.writeModuleLog(outputName, m.Name, strings.TrimSuffix(base, ".py")+".log", res.Stdout+res.Stderr); err != nil {
			return err
		}
		if execErr != nil {
			return execErr
		}
	}
	return nil
}

func (r Runner) runPartial(ctx context.Context, moduleName string, outputName string, selectors []string) error {
	for i, selector := range selectors {
		res, err := r.Exec.Exec(ctx, fmt.Sprintf("cd /code/modules/%s; .tox/c1/bin/pytest --cov=%s tests/%s -v -vv -s --cov-append --cov-branch --cov-report=term --cov-report=html --basetemp=/code/modules/%s/.tox/c1/tmp --full-trace", moduleName, strings.ReplaceAll(moduleName, "-", "_"), selector, moduleName))
		if writeErr := r.writeModuleLog(outputName, moduleName, fmt.Sprintf("partial%d.log", i+1), res.Stdout+res.Stderr); writeErr != nil {
			return writeErr
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r Runner) fetchCoverage(ctx context.Context, moduleName string, outputName string) (string, error) {
	res, err := r.Exec.Exec(ctx, fmt.Sprintf("cd /code/modules/%s; .tox/c1/bin/coverage report", moduleName))
	if writeErr := r.writeModuleLog(outputName, moduleName, "coverage.log", res.Stdout+res.Stderr); writeErr != nil {
		return "", writeErr
	}
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`TOTAL\s+\d+\s+\d+\s+(\d+)%`)
	m := re.FindStringSubmatch(res.Stdout + "\n" + res.Stderr)
	if len(m) < 2 {
		return "0", nil
	}
	return m[1], nil
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
