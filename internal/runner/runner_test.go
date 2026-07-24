package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ivis-kuroda/weko-autox/internal/cli"
	"github.com/ivis-kuroda/weko-autox/internal/dockerx"
	"github.com/ivis-kuroda/weko-autox/internal/report"
)

type fakeExecutor struct {
	commands []string
	outputs  map[string]dockerx.ExecResult
	errors   map[string]error
	stopped  bool
}

func (f *fakeExecutor) Exec(_ context.Context, shellCommand string) (dockerx.ExecResult, error) {
	f.commands = append(f.commands, shellCommand)
	if res, ok := f.outputs[shellCommand]; ok {
		if err, hasErr := f.errors[shellCommand]; hasErr {
			return res, err
		}
		return res, nil
	}
	if err, ok := f.errors[shellCommand]; ok {
		return dockerx.ExecResult{}, err
	}
	return dockerx.ExecResult{}, nil
}

func (f *fakeExecutor) StopTests(_ context.Context) error {
	f.stopped = true
	return nil
}

func TestRunnerKillMode(t *testing.T) {
	exec := &fakeExecutor{}
	r := Runner{Workspace: t.TempDir(), Exec: exec, Progress: report.NewProgress(io.Discard)}

	err := r.Run(context.Background(), cli.Config{Kill: true, RunMode: cli.RunModeAllAtOnce})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !exec.stopped {
		t.Fatal("Run() should call StopTests in kill mode")
	}
}

func TestRunnerRunAllAtOnce(t *testing.T) {
	workspace := makeWorkspaceWithModule(t, "weko-admin", nil)
	exec := &fakeExecutor{
		outputs: map[string]dockerx.ExecResult{
			"cd /code/modules/weko-admin; tox":                         {Stdout: "tox-ok"},
			"cd /code/modules/weko-admin; .tox/c1/bin/coverage report": {Stdout: "TOTAL 10 1 90%"},
		},
		errors: map[string]error{},
	}
	cfg := cli.Config{RunMode: cli.RunModeAllAtOnce, Scope: "all"}
	r := Runner{Workspace: workspace, Exec: exec, Progress: report.NewProgress(io.Discard)}

	if err := r.Run(context.Background(), cfg); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !containsCommand(exec.commands, "cd /code/modules/weko-admin; tox") {
		t.Fatalf("expected tox command in executed commands: %#v", exec.commands)
	}
	assertFileExists(t, filepath.Join(workspace, "log", "weko-admin", "test_all.log"))
	assertFileExists(t, filepath.Join(workspace, "log", "weko-admin", "coverage.log"))
}

func TestRunnerRunPerFile(t *testing.T) {
	workspace := makeWorkspaceWithModule(t, "weko-items-ui", []string{"test_a.py", "test_b.py"})
	exec := &fakeExecutor{outputs: map[string]dockerx.ExecResult{}, errors: map[string]error{}}
	cfg := cli.Config{RunMode: cli.RunModePerFile, Scope: "all"}
	r := Runner{Workspace: workspace, Exec: exec, Progress: report.NewProgress(io.Discard)}

	coverageCmd := "cd /code/modules/weko-items-ui; .tox/c1/bin/coverage report"
	exec.outputs[coverageCmd] = dockerx.ExecResult{Stdout: "TOTAL 20 0 100%"}

	if err := r.Run(context.Background(), cfg); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !containsCommand(exec.commands, "cd /code/modules/weko-items-ui; tox -e c1 --notest") {
		t.Fatalf("expected install command in commands: %#v", exec.commands)
	}
	if !containsCommandSubstring(exec.commands, "tests/test_a.py") || !containsCommandSubstring(exec.commands, "tests/test_b.py") {
		t.Fatalf("expected per-file pytest commands: %#v", exec.commands)
	}
	assertFileExists(t, filepath.Join(workspace, "log", "weko-items-ui", "install.log"))
	assertFileExists(t, filepath.Join(workspace, "log", "weko-items-ui", "test_a.log"))
	assertFileExists(t, filepath.Join(workspace, "log", "weko-items-ui", "test_b.log"))
}

func TestRunnerRunPartial(t *testing.T) {
	workspace := makeWorkspaceWithModule(t, "weko-admin", nil)
	exec := &fakeExecutor{outputs: map[string]dockerx.ExecResult{}, errors: map[string]error{}}
	r := Runner{Workspace: workspace, Exec: exec, Progress: report.NewProgress(io.Discard)}

	coverageCmd := "cd /code/modules/weko-admin; .tox/c1/bin/coverage report"
	exec.outputs[coverageCmd] = dockerx.ExecResult{Stdout: "TOTAL 5 1 80%"}

	cfg := cli.Config{
		RunMode:          cli.RunModePartial,
		PartialModule:    "weko-admin",
		PartialSelectors: []string{"test_api.py::test_one", "test_api.py::test_two"},
	}

	if err := r.Run(context.Background(), cfg); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !containsCommandSubstring(exec.commands, "tests/test_api.py::test_one") {
		t.Fatalf("missing first partial selector command: %#v", exec.commands)
	}
	if !containsCommandSubstring(exec.commands, "tests/test_api.py::test_two") {
		t.Fatalf("missing second partial selector command: %#v", exec.commands)
	}
	assertFileExists(t, filepath.Join(workspace, "log", "weko-admin", "partial1.log"))
	assertFileExists(t, filepath.Join(workspace, "log", "weko-admin", "partial2.log"))
}

func TestFetchCoverageFallback(t *testing.T) {
	r := Runner{Workspace: t.TempDir(), Progress: report.NewProgress(io.Discard)}
	exec := &fakeExecutor{outputs: map[string]dockerx.ExecResult{}, errors: map[string]error{}}
	r.Exec = exec

	moduleName := "weko-admin"
	outDir := ""
	if err := r.prepareModuleLogDir(moduleName, outDir); err != nil {
		t.Fatalf("prepareModuleLogDir() error = %v", err)
	}
	cmd := "cd /code/modules/weko-admin; .tox/c1/bin/coverage report"
	exec.outputs[cmd] = dockerx.ExecResult{Stdout: "no total line"}

	coverage, err := r.fetchCoverage(context.Background(), moduleName, outDir)
	if err != nil {
		t.Fatalf("fetchCoverage() error = %v", err)
	}
	if coverage != "0" {
		t.Fatalf("fetchCoverage() = %q, want %q", coverage, "0")
	}
}

func makeWorkspaceWithModule(t *testing.T, moduleName string, testFiles []string) string {
	t.Helper()
	workspace := t.TempDir()
	moduleTestsDir := filepath.Join(workspace, "modules", moduleName, "tests")
	if err := os.MkdirAll(moduleTestsDir, 0o755); err != nil {
		t.Fatalf("mkdir module tests dir: %v", err)
	}
	for _, file := range testFiles {
		path := filepath.Join(moduleTestsDir, file)
		if err := os.WriteFile(path, []byte("def test_x():\n    assert True\n"), 0o644); err != nil {
			t.Fatalf("write test file %s: %v", file, err)
		}
	}
	return workspace
}

func containsCommand(commands []string, want string) bool {
	for _, cmd := range commands {
		if cmd == want {
			return true
		}
	}
	return false
}

func containsCommandSubstring(commands []string, wantSubstr string) bool {
	for _, cmd := range commands {
		if strings.Contains(cmd, wantSubstr) {
			return true
		}
	}
	return false
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %s (err=%v)", path, err)
	}
}

func TestCleanLocalArtifacts(t *testing.T) {
	workspace := t.TempDir()
	moduleName := "weko-admin"
	modulePath := filepath.Join(workspace, moduleName)
	if err := os.MkdirAll(modulePath, 0o755); err != nil {
		t.Fatalf("mkdir module path: %v", err)
	}
	paths := []string{
		filepath.Join(modulePath, moduleName+".egg-info"),
		filepath.Join(modulePath, ".tox"),
		filepath.Join(modulePath, "htmlcov"),
		filepath.Join(modulePath, "coverage.xml"),
	}
	for _, p := range paths {
		if strings.HasSuffix(p, ".xml") {
			if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
				t.Fatalf("write file %s: %v", p, err)
			}
		} else {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", p, err)
			}
		}
	}

	if err := cleanLocalArtifacts(modulePath, moduleName); err != nil {
		t.Fatalf("cleanLocalArtifacts() error = %v", err)
	}
	for _, p := range paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed", p)
		}
	}
}

func TestModuleLogDir(t *testing.T) {
	workspace := "/tmp/weko"
	gotDefault := moduleLogDir(workspace, "", "weko-admin")
	wantDefault := filepath.Join(workspace, "log", "weko-admin")
	if gotDefault != wantDefault {
		t.Fatalf("moduleLogDir default = %q, want %q", gotDefault, wantDefault)
	}

	gotNamed := moduleLogDir(workspace, "daily", "weko-admin")
	wantNamed := filepath.Join(workspace, "log", "daily", "weko-admin")
	if gotNamed != wantNamed {
		t.Fatalf("moduleLogDir named = %q, want %q", gotNamed, wantNamed)
	}
}

func TestRunnerInstallToxError(t *testing.T) {
	workspace := makeWorkspaceWithModule(t, "weko-admin", nil)
	cmd := "pip3 install tox==3.28 tox-setuptools-version"
	exec := &fakeExecutor{
		outputs: map[string]dockerx.ExecResult{cmd: {Stdout: "", Stderr: "fail"}},
		errors:  map[string]error{cmd: fmt.Errorf("install failed")},
	}
	r := Runner{Workspace: workspace, Exec: exec, Progress: report.NewProgress(io.Discard)}

	err := r.Run(context.Background(), cli.Config{RunMode: cli.RunModeAllAtOnce, Scope: "all"})
	if err == nil {
		t.Fatal("Run() expected install error")
	}
}
