package dockerx

import (
	"context"
	"testing"

	composetypes "github.com/compose-spec/compose-go/types"
	containertypes "github.com/docker/docker/api/types/container"
)

func TestNew_MissingComposeFile(t *testing.T) {
	_, err := New(context.Background(), t.TempDir(), "web")
	if err == nil {
		t.Fatal("New() expected error when compose file is missing")
	}
}

func TestHasService(t *testing.T) {
	project := &composetypes.Project{
		Services: composetypes.Services{
			{Name: "web"},
			{Name: "worker"},
		},
	}

	if !hasService(project, "web") {
		t.Fatal("hasService should return true for existing service")
	}
	if hasService(project, "db") {
		t.Fatal("hasService should return false for missing service")
	}
}

func TestChooseContainerID(t *testing.T) {
	t.Run("prefer running container", func(t *testing.T) {
		containers := []containertypes.Summary{
			{ID: "stopped-id", State: "exited"},
			{ID: "running-id", State: "running"},
		}

		got, err := chooseContainerID(containers)
		if err != nil {
			t.Fatalf("chooseContainerID() error = %v", err)
		}
		if got != "running-id" {
			t.Fatalf("chooseContainerID() = %q, want %q", got, "running-id")
		}
	})

	t.Run("fallback to first container", func(t *testing.T) {
		containers := []containertypes.Summary{
			{ID: "first-id", State: "created"},
			{ID: "second-id", State: "paused"},
		}

		got, err := chooseContainerID(containers)
		if err != nil {
			t.Fatalf("chooseContainerID() error = %v", err)
		}
		if got != "first-id" {
			t.Fatalf("chooseContainerID() = %q, want %q", got, "first-id")
		}
	})

	t.Run("empty list returns error", func(t *testing.T) {
		_, err := chooseContainerID(nil)
		if err == nil {
			t.Fatal("chooseContainerID(nil) expected error")
		}
	})
}
