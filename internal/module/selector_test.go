package module

import (
	"testing"

	"github.com/ivis-kuroda/weko-autox/internal/cli"
)

func sampleModules() []Module {
	return []Module{
		{Name: "invenio-records"},
		{Name: "weko-admin"},
		{Name: "weko-items-ui"},
	}
}

func TestSelect_ByScope(t *testing.T) {
	cfg := cli.Config{RunMode: cli.RunModeAllAtOnce, Scope: "weko"}
	got, err := Select(sampleModules(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(got))
	}
}

func TestSelect_Exclude(t *testing.T) {
	cfg := cli.Config{RunMode: cli.RunModeAllAtOnce, Scope: "weko", ExcludeModules: []string{"weko-admin"}}
	got, err := Select(sampleModules(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "weko-items-ui" {
		t.Fatalf("unexpected modules: %#v", got)
	}
}

func TestSelect_PartialModule(t *testing.T) {
	cfg := cli.Config{RunMode: cli.RunModePartial, PartialModule: "weko-admin", PartialSelectors: []string{"test_a.py::x"}}
	got, err := Select(sampleModules(), cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "weko-admin" {
		t.Fatalf("unexpected module: %#v", got)
	}
}
