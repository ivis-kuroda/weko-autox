package cli

import (
	"bytes"
	"context"
	"testing"
)

func TestParseArgs_PerFileWithExclude(t *testing.T) {
	cfg, err := ParseArgs([]string{"--run-mode", "per-file", "weko", "-n", "weko-admin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RunMode != RunModePerFile {
		t.Fatalf("run mode mismatch: %s", cfg.RunMode)
	}
	if cfg.Scope != "weko" {
		t.Fatalf("scope mismatch: %s", cfg.Scope)
	}
	if len(cfg.ExcludeModules) != 1 || cfg.ExcludeModules[0] != "weko-admin" {
		t.Fatalf("exclude modules mismatch: %#v", cfg.ExcludeModules)
	}
}

func TestParseArgs_PartialFromShortOption(t *testing.T) {
	cfg, err := ParseArgs([]string{"-p", "weko-admin", "test_api.py::test_is_restricted_user"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RunMode != RunModePartial {
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

func TestExecute_CallsRunnerWithParsedConfig(t *testing.T) {
	var out bytes.Buffer
	called := false

	err := Execute(context.Background(), &out, []string{"--run-mode", "per-file", "weko", "weko-admin"}, func(_ context.Context, cfg Config) error {
		called = true
		if cfg.RunMode != RunModePerFile {
			t.Fatalf("run mode mismatch: got=%s", cfg.RunMode)
		}
		if cfg.Scope != "weko" {
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
