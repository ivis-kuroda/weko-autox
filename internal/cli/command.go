package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const versionString = "autox - dev"

func ParseArgs(args []string) (Config, error) {
	var cfg Config
	var excludeMode bool

	cmd := &cobra.Command{
		Use:   "autox [all|weko|invenio] [target1 target2 ...]",
		Short: "Run tox for Weko modules",
		RunE: func(_ *cobra.Command, positional []string) error {
			if cfg.Version || cfg.Kill {
				return nil
			}

			if cfg.PartialModule != "" {
				cfg.RunMode = RunModePartial
				cfg.PartialSelectors = positional
				return cfg.Validate()
			}

			if cfg.RunMode == "" {
				cfg.RunMode = RunModeAllAtOnce
			}

			rest := positional
			if len(rest) > 0 {
				switch rest[0] {
				case "all", "weko", "invenio":
					cfg.Scope = rest[0]
					rest = rest[1:]
				}
			}

			if excludeMode {
				cfg.ExcludeModules = rest
			} else {
				cfg.IncludeModules = rest
			}

			return cfg.Validate()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.Flags().BoolVarP(&excludeMode, "exclude", "n", false, "exclude mode for positional module names")
	cmd.Flags().BoolVarP(&cfg.Clean, "reset", "r", false, "remove egg-info, .tox, htmlcov and coverage.xml before run")
	cmd.Flags().StringVarP((*string)(&cfg.RunMode), "run-mode", "", string(RunModeAllAtOnce), "run strategy: all-at-once | per-file | partial")
	cmd.Flags().StringVarP(&cfg.PartialModule, "partial", "p", "", "target module for partial mode")
	cmd.Flags().StringVarP(&cfg.OutputDirName, "output", "o", "", "output subdirectory under log")
	cmd.Flags().BoolVarP(&cfg.Kill, "kill", "k", false, "stop running tox/pytest processes")
	cmd.Flags().BoolVarP(&cfg.Version, "version", "v", false, "show version")

	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		cfg.Help = true
		_ = c.Usage()
	})

	cmd.SetArgs(args)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Execute(ctx context.Context, out io.Writer, args []string, run func(context.Context, Config) error) error {
	cfg, err := ParseArgs(args)
	if err != nil {
		return err
	}

	if cfg.Help || cfg.Version {
		if cfg.Version {
			_, _ = fmt.Fprintln(out, versionString)
		}
		return nil
	}

	return run(ctx, cfg)
}
