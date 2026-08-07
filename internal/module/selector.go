package module

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ivis-kuroda/weko-autox/internal/cli"
)

func Select(modules []Module, cfg cli.Config) ([]Module, error) {
	if len(cfg.PartialSelectors) > 0 || cfg.RunMode == cli.RunModePartial {
		for _, m := range modules {
			if m.Name == cfg.PartialModule {
				return []Module{m}, nil
			}
		}
		return nil, fmt.Errorf("partial module not found: %s", cfg.PartialModule)
	}

	base := filterByScope(modules, cfg.Scope)
	if len(base) == 0 {
		return nil, fmt.Errorf("no modules matched scope: %s", cfg.Scope)
	}

	if len(cfg.IncludeModules) > 0 {
		set := map[string]struct{}{}
		for _, n := range cfg.IncludeModules {
			set[n] = struct{}{}
		}
		filtered := make([]Module, 0, len(base))
		for _, m := range base {
			if _, ok := set[m.Name]; ok {
				filtered = append(filtered, m)
			}
		}
		base = filtered
	}

	if len(cfg.ExcludeModules) > 0 {
		filtered := make([]Module, 0, len(base))
		for _, m := range base {
			if slices.Contains(cfg.ExcludeModules, m.Name) {
				continue
			}
			filtered = append(filtered, m)
		}
		base = filtered
	}

	if len(base) == 0 {
		return nil, fmt.Errorf("no modules selected")
	}

	return base, nil
}

func filterByScope(modules []Module, scope string) []Module {
	if scope == "" || scope == "all" {
		out := make([]Module, len(modules))
		copy(out, modules)
		return out
	}

	out := make([]Module, 0, len(modules))
	for _, m := range modules {
		switch scope {
		case "weko":
			if strings.Contains(m.Name, "weko") {
				out = append(out, m)
			}
		case "invenio":
			if strings.Contains(m.Name, "invenio") {
				out = append(out, m)
			}
		}
	}
	return out
}
