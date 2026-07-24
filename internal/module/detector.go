package module

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Module struct {
	Name string
	Path string
}

func Detect(modulesDir string) ([]Module, error) {
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil, fmt.Errorf("read modules dir: %w", err)
	}

	result := make([]Module, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		modulePath := filepath.Join(modulesDir, entry.Name())
		testsPath := filepath.Join(modulePath, "tests")
		st, err := os.Stat(testsPath)
		if err != nil || !st.IsDir() {
			continue
		}
		result = append(result, Module{Name: entry.Name(), Path: modulePath})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}
