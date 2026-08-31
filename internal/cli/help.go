package cli

import (
	"fmt"
	"io"
	"text/tabwriter"
)

//nolint:errcheck // help output is best-effort; there's nothing actionable to do if writing it fails
func printStructuredHelp(out io.Writer) {
	fmt.Fprintf(out, "autox - %s\n", versionString)
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  autox [all|weko|invenio] [options]")
	fmt.Fprintln(out, "  autox [module ...] [options]")
	fmt.Fprintln(out, "  autox <module> -p <selector> [-p <selector> ...] [options]")
	fmt.Fprintln(out, "  autox [options]")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Scope:")
	sw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, scope := range ScopeSpecs {
		fmt.Fprintf(sw, "  %s\t%s\n", scope.Keyword, scope.Description)
	}
	_ = sw.Flush()
	fmt.Fprintln(out)

	fmt.Fprintln(out, "Options:")
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	for _, opt := range []interface {
		spec() (string, string, string)
	}{
		OptionExclude,
		OptionReset,
		OptionRunMode,
		OptionPartial,
		OptionOutput,
		OptionKill,
		OptionVersion,
		OptionHelp,
	} {
		name, short, description := opt.spec()

		shortPrefix := "-" + short + ", "
		if short == "" {
			shortPrefix = "    "
		}
		fmt.Fprintf(tw, "  %s--%s\t%s\n", shortPrefix, name, description)
	}
	_ = tw.Flush()
}
