package cli

type OptionSpec[T any] struct {
	Name        string
	Short       string
	Description string
	Default     T
}

func (s OptionSpec[T]) spec() (string, string, string) {
	return s.Name, s.Short, s.Description
}

type BoolOptionSpec = OptionSpec[bool]
type StringOptionSpec = OptionSpec[string]
type StringArrayOptionSpec = OptionSpec[[]string]

type ScopeSpec struct {
	Keyword     string
	Description string
}

const (
	CommandUse   = "autox [all|weko|invenio]|[module ...] [options]"
	CommandShort = "Run tox for Weko modules"

	ScopeAll     = "all"
	ScopeWeko    = "weko"
	ScopeInvenio = "invenio"

	ErrPartialRequiresModule      = "partial mode requires module as first argument"
	ErrPartialSingleModuleOnly    = "partial mode accepts only one module argument"
	ErrPartialRequiresOneSelector = "partial mode requires at least one -p selector"
	ErrScopeAndModuleExclusive    = "scope and module arguments are mutually exclusive unless --exclude (-n) is set"
	ErrMissingTargetOrStandalone  = "specify a scope, module argument, or a standalone option (-k, -v, -h)"

	RunModeHelp = "run strategy: all-at-once | per-file | per-func (legacy: partial, default: all-at-once)"
)

var (
	OptionExclude = BoolOptionSpec{
		Name:        "exclude",
		Short:       "n",
		Description: "exclude mode for positional module names",
		Default:     false,
	}
	OptionReset = BoolOptionSpec{
		Name:        "reset",
		Short:       "r",
		Description: "remove egg-info, .tox, htmlcov and coverage.xml before run",
		Default:     false,
	}
	OptionRunMode = StringOptionSpec{
		Name:        "run-mode",
		Short:       "",
		Description: RunModeHelp,
		Default:     string(RunModeAllAtOnce),
	}
	OptionPartial = StringArrayOptionSpec{
		Name:        "partial",
		Short:       "p",
		Description: "partial test selector (repeatable), requires module positional argument",
		Default:     nil,
	}
	OptionOutput = StringOptionSpec{
		Name:        "output",
		Short:       "o",
		Description: "output subdirectory under log",
		Default:     "",
	}
	OptionKill = BoolOptionSpec{
		Name:        "kill",
		Short:       "k",
		Description: "stop running tox/pytest processes",
		Default:     false,
	}
	OptionVersion = BoolOptionSpec{
		Name:        "version",
		Short:       "v",
		Description: "show version",
		Default:     false,
	}
	OptionHelp = BoolOptionSpec{
		Name:        "help",
		Short:       "h",
		Description: "show help",
		Default:     false,
	}

	ScopeSpecs = []ScopeSpec{
		{Keyword: ScopeAll, Description: "run tests for all modules"},
		{Keyword: ScopeWeko, Description: "run tests for modules containing \"weko\""},
		{Keyword: ScopeInvenio, Description: "run tests for modules containing \"invenio\""},
	}
)

func IsScopeKeyword(scope string) bool {
	switch scope {
	case ScopeAll, ScopeWeko, ScopeInvenio:
		return true
	default:
		return false
	}
}
