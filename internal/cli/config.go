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
	RunModePerFunc   RunMode = "per-func"
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
	case RunModeAllAtOnce, RunModePerFile, RunModePerFunc, RunModePartial:
	default:
		return fmt.Errorf("invalid run mode: %s", c.RunMode)
	}

	if c.Scope != "" {
		switch c.Scope {
		case ScopeAll, ScopeWeko, ScopeInvenio:
		default:
			return fmt.Errorf("invalid scope: %s", c.Scope)
		}
	}

	if len(c.PartialSelectors) > 0 || c.RunMode == RunModePartial {
		if strings.TrimSpace(c.PartialModule) == "" {
			return errors.New(ErrPartialRequiresModule)
		}
		if len(c.PartialSelectors) == 0 {
			return errors.New(ErrPartialRequiresOneSelector)
		}
	}

	return nil
}
