package cli

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestParseArgs_PerFileWithExclude(t *testing.T) {
	cfg, err := ParseArgs([]string{"weko-admin", "--run-mode", "per-file", "-n"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RunMode != RunModePerFile {
		t.Fatalf("run mode mismatch: %s", cfg.RunMode)
	}
	if cfg.Scope != "" {
		t.Fatalf("scope mismatch: %s", cfg.Scope)
	}
	if len(cfg.ExcludeModules) != 1 || cfg.ExcludeModules[0] != "weko-admin" {
		t.Fatalf("exclude modules mismatch: %#v", cfg.ExcludeModules)
	}
}

func TestParseArgs_PartialFromShortOption(t *testing.T) {
	cfg, err := ParseArgs([]string{"weko-admin", "-p", "test_api.py::test_is_restricted_user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RunMode != RunModeAllAtOnce {
		t.Fatalf("run mode mismatch: %s", cfg.RunMode)
	}
	if cfg.PartialModule != "weko-admin" {
		t.Fatalf("partial module mismatch: %s", cfg.PartialModule)
	}
	if len(cfg.PartialSelectors) != 1 {
		t.Fatalf("selectors mismatch: %#v", cfg.PartialSelectors)
	}
}

func TestParseArgs_PartialWithMultipleSelectors(t *testing.T) {
	cfg, err := ParseArgs([]string{"weko-admin", "-p", "test_api.py::test_one", "-p", "test_tasks.py::test_two"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RunMode != RunModeAllAtOnce {
		t.Fatalf("run mode mismatch: %s", cfg.RunMode)
	}
	if cfg.PartialModule != "weko-admin" {
		t.Fatalf("partial module mismatch: %s", cfg.PartialModule)
	}
	if len(cfg.PartialSelectors) != 2 {
		t.Fatalf("selectors mismatch: %#v", cfg.PartialSelectors)
	}
}

func TestParseArgs_PartialRequiresSingleModuleArgument(t *testing.T) {
	_, err := ParseArgs([]string{"weko-admin", "extra", "-p", "test_api.py::test_one"})
	if err == nil {
		t.Fatal("expected error for extra module arguments in partial mode")
	}
}

func TestParseArgs_PartialPerFuncMode(t *testing.T) {
	cfg, err := ParseArgs([]string{"weko-admin", "--run-mode", "per-func", "-p", "test_api.py::test_one"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RunMode != RunModePerFunc {
		t.Fatalf("run mode mismatch: %s", cfg.RunMode)
	}
	if cfg.PartialModule != "weko-admin" {
		t.Fatalf("partial module mismatch: %s", cfg.PartialModule)
	}
	if len(cfg.PartialSelectors) != 1 {
		t.Fatalf("selectors mismatch: %#v", cfg.PartialSelectors)
	}
}

func TestParseArgs_InvalidRunMode(t *testing.T) {
	_, err := ParseArgs([]string{"--run-mode", "unknown"})
	if err == nil {
		t.Fatal("expected error for invalid run mode")
	}
}

func TestParseArgs_RejectsNoArguments(t *testing.T) {
	_, err := ParseArgs([]string{})
	if err == nil {
		t.Fatal("expected error when no scope/module/standalone option is provided")
	}
}

func TestParseArgs_RejectsNonStandaloneOptionsOnly(t *testing.T) {
	_, err := ParseArgs([]string{"-r"})
	if err == nil {
		t.Fatal("expected error for non-standalone options-only invocation")
	}
}

func TestParseArgs_AcceptsPositionalThenOptions(t *testing.T) {
	_, err := ParseArgs([]string{"all", "-r", "-o", "daily"})
	if err != nil {
		t.Fatalf("expected positional-then-options to be valid, got: %v", err)
	}
}

func TestParseArgs_RejectsScopeWithModuleArguments(t *testing.T) {
	_, err := ParseArgs([]string{"all", "weko-admin"})
	if err == nil {
		t.Fatal("expected error when both scope and module arguments are specified")
	}
}

func TestParseArgs_AllowsScopeWithExcludeArguments(t *testing.T) {
	cfg, err := ParseArgs([]string{"weko", "weko-admin", "-n"})
	if err != nil {
		t.Fatalf("expected scope with exclude arguments to be valid, got: %v", err)
	}
	if cfg.Scope != "weko" {
		t.Fatalf("scope mismatch: %s", cfg.Scope)
	}
	if len(cfg.ExcludeModules) != 1 || cfg.ExcludeModules[0] != "weko-admin" {
		t.Fatalf("exclude modules mismatch: %#v", cfg.ExcludeModules)
	}
}

func TestParseArgs_AcceptsOptionsOnly(t *testing.T) {
	_, err := ParseArgs([]string{"-k"})
	if err != nil {
		t.Fatalf("expected options-only form to be valid, got: %v", err)
	}
}

func TestParseArgs_RejectsOptionsThenPositional(t *testing.T) {
	_, err := ParseArgs([]string{"-r", "all"})
	if err == nil {
		t.Fatal("expected error when positional argument appears after starting with options")
	}
}

func TestParseArgs_AllowsModuleArgsWithExcludeMode(t *testing.T) {
	cfg, err := ParseArgs([]string{"weko-admin", "-n", "invenio-records"})
	if err != nil {
		t.Fatalf("expected module arguments with exclude mode to be valid, got: %v", err)
	}
	if len(cfg.ExcludeModules) != 2 || cfg.ExcludeModules[0] != "weko-admin" || cfg.ExcludeModules[1] != "invenio-records" {
		t.Fatalf("exclude modules mismatch: %#v", cfg.ExcludeModules)
	}
}

func TestExecute_VersionSkipsRunner(t *testing.T) {
	var out bytes.Buffer
	runCalled := false

	err := Execute(context.Background(), &out, []string{"-v"}, func(context.Context, Config) error {
		runCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if runCalled {
		t.Fatal("runner should not be called when -v is specified")
	}
	if out.String() == "" {
		t.Fatal("version output should not be empty")
	}
}

func TestExecute_VersionIncludesCommit(t *testing.T) {
	var out bytes.Buffer

	err := Execute(context.Background(), &out, []string{"-v"}, func(context.Context, Config) error {
		return nil
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// versionString/commitString are injected via -ldflags at build time
	// (see Makefile/.goreleaser.yaml); in tests they retain their defaults.
	// OS/arch come from runtime.GOOS/GOARCH, so compute them the same way
	// rather than hardcoding a platform.
	want := fmt.Sprintf("autox - dev, none %s/%s\n", runtime.GOOS, runtime.GOARCH)
	if out.String() != want {
		t.Fatalf("version output = %q, want %q", out.String(), want)
	}
}

func TestExecute_CallsRunnerWithParsedConfig(t *testing.T) {
	var out bytes.Buffer
	called := false

	err := Execute(context.Background(), &out, []string{"weko-admin", "--run-mode", "per-file"}, func(_ context.Context, cfg Config) error {
		called = true
		if cfg.RunMode != RunModePerFile {
			t.Fatalf("run mode mismatch: got=%s", cfg.RunMode)
		}
		if cfg.Scope != "" {
			t.Fatalf("scope mismatch: got=%s", cfg.Scope)
		}
		if len(cfg.IncludeModules) != 1 || cfg.IncludeModules[0] != "weko-admin" {
			t.Fatalf("include modules mismatch: %#v", cfg.IncludeModules)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("runner should be called")
	}
}

func TestExecute_InvalidArgsPrintsHelp(t *testing.T) {
	var out bytes.Buffer

	err := Execute(context.Background(), &out, []string{"-r"}, func(context.Context, Config) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for invalid args")
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected help output to include Usage, got: %q", out.String())
	}
}
