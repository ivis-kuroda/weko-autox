package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

var versionString = "dev"

func ParseArgs(args []string) (Config, error) {
	if err := validateArgOrder(args); err != nil {
		return Config{}, err
	}

	var cfg Config
	var excludeMode bool
	var partialSelectors []string

	cmd := &cobra.Command{
		Use:   CommandUse,
		Short: CommandShort,
		RunE: func(_ *cobra.Command, positional []string) error {
			if cfg.Version || cfg.Kill {
				return nil
			}

			if cfg.RunMode == RunModePartial || len(partialSelectors) > 0 {
				if len(positional) == 0 {
					return errors.New(ErrPartialRequiresModule)
				}
				if len(positional) > 1 {
					return errors.New(ErrPartialSingleModuleOnly)
				}
				if len(partialSelectors) == 0 {
					return errors.New(ErrPartialRequiresOneSelector)
				}
				if cfg.RunMode == "" {
					cfg.RunMode = RunModeAllAtOnce
				}
				if cfg.RunMode == RunModePartial {
					cfg.RunMode = RunModePerFunc
				}

				cfg.PartialModule = positional[0]
				cfg.PartialSelectors = partialSelectors
				return cfg.Validate()
			}

			if cfg.RunMode == "" {
				cfg.RunMode = RunModeAllAtOnce
			}

			rest := positional
			if len(rest) > 0 {
				if IsScopeKeyword(rest[0]) {
					cfg.Scope = rest[0]
					rest = rest[1:]
					if len(rest) > 0 && !excludeMode {
						return errors.New(ErrScopeAndModuleExclusive)
					}
				}
			}

			if excludeMode {
				cfg.ExcludeModules = rest
			} else {
				cfg.IncludeModules = rest
			}

			if cfg.Scope == "" && len(cfg.IncludeModules) == 0 && len(cfg.ExcludeModules) == 0 {
				return errors.New(ErrMissingTargetOrStandalone)
			}

			return cfg.Validate()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	bindOption(cmd.Flags().BoolVarP, &excludeMode, OptionExclude)
	bindOption(cmd.Flags().BoolVarP, &cfg.Clean, OptionReset)
	bindOption(cmd.Flags().StringVarP, (*string)(&cfg.RunMode), OptionRunMode)
	bindOption(cmd.Flags().StringArrayVarP, &partialSelectors, OptionPartial)
	bindOption(cmd.Flags().StringVarP, &cfg.OutputDirName, OptionOutput)
	bindOption(cmd.Flags().BoolVarP, &cfg.Kill, OptionKill)
	bindOption(cmd.Flags().BoolVarP, &cfg.Version, OptionVersion)

	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		cfg.Help = true
		printStructuredHelp(c.OutOrStdout())
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
		printStructuredHelp(out)
		return err
	}

	if cfg.Help || cfg.Version {
		if cfg.Version {
			_, _ = fmt.Fprintf(out, "autox - %s\n", versionString)
		}
		return nil
	}

	return run(ctx, cfg)
}

func bindOption[T any](bind func(*T, string, string, T, string), target *T, spec OptionSpec[T]) {
	bind(target, spec.Name, spec.Short, spec.Default, spec.Description)
}
