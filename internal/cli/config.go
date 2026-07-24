package cli

import (
	"errors"
	"fmt"
	"strings"
)

type RunMode string

const (
	RunModeAllAtOnce RunMode = "all-at-once"
	RunModePerFile   RunMode = "per-file"
	RunModePartial   RunMode = "partial"
)

type Config struct {
	RunMode          RunMode
	Scope            string
	IncludeModules   []string
	ExcludeModules   []string
	PartialModule    string
	PartialSelectors []string
	OutputDirName    string
	Clean            bool
	Kill             bool
	Version          bool
	Help             bool
}

func (c Config) Validate() error {
	switch c.RunMode {
	case RunModeAllAtOnce, RunModePerFile, RunModePartial:
	default:
		return fmt.Errorf("invalid run mode: %s", c.RunMode)
	}

	if c.Scope != "" {
		switch c.Scope {
		case "all", "weko", "invenio":
		default:
			return fmt.Errorf("invalid scope: %s", c.Scope)
		}
	}

	if c.RunMode == RunModePartial {
		if strings.TrimSpace(c.PartialModule) == "" {
			return errors.New("partial mode requires module by -p")
		}
		if len(c.PartialSelectors) == 0 {
			return errors.New("partial mode requires at least one test selector")
		}
	}

	return nil
}
