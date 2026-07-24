package dockerx

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeProjectName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "lowercase", in: "Weko-Autox", want: "weko-autox"},
		{name: "strip symbols", in: "Weko@AutoX!", want: "wekoautox"},
		{name: "fallback default", in: "!!!", want: "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeProjectName(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeProjectName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestDetectComposeFilePriority(t *testing.T) {
	workspace := t.TempDir()

	if err := os.WriteFile(filepath.Join(workspace, "docker-compose.yml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatalf("write docker-compose.yml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "compose.yaml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatalf("write compose.yaml: %v", err)
	}

	got, err := DetectComposeFile(workspace)
	if err != nil {
		t.Fatalf("DetectComposeFile() error = %v", err)
	}

	want := filepath.Join(workspace, "compose.yaml")
	if got != want {
		t.Fatalf("DetectComposeFile() = %q, want %q", got, want)
	}
}

func TestDetectComposeFileNotFound(t *testing.T) {
	workspace := t.TempDir()

	_, err := DetectComposeFile(workspace)
	if err == nil {
		t.Fatal("DetectComposeFile() expected error but got nil")
	}
}
