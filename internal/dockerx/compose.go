package dockerx

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	composetypes "github.com/compose-spec/compose-go/types"
)

var composeCandidates = []string{
	"compose.yaml",
	"compose.yml",
	"docker-compose.yml",
	"docker-compose.yaml",
}

func DetectComposeFile(workspace string) (string, error) {
	for _, name := range composeCandidates {
		path := filepath.Join(workspace, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("compose file not found in %s", workspace)
}

func LoadProject(composeFile string) (*composetypes.Project, error) {
	wd := filepath.Dir(composeFile)
	base := filepath.Base(composeFile)

	projectName := normalizeProjectName(filepath.Base(wd))
	details := composetypes.ConfigDetails{
		WorkingDir: wd,
		ConfigFiles: []composetypes.ConfigFile{
			{Filename: base},
		},
		Environment: map[string]string{
			"COMPOSE_PROJECT_NAME": projectName,
		},
	}

	project, err := loader.Load(details, func(options *loader.Options) {
		options.SetProjectName(projectName, true)
	})
	if err != nil {
		return nil, fmt.Errorf("load compose project: %w", err)
	}
	return project, nil
}

func normalizeProjectName(name string) string {
	lower := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9_-]`)
	lower = re.ReplaceAllString(lower, "")
	if lower == "" {
		return "default"
	}
	return lower
}
