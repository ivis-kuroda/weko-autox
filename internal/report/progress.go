package report

import (
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
)

const (
	ansiGreen = "\x1b[32m"
	ansiReset = "\x1b[0m"
)

type Progress struct {
	Out io.Writer
}

func NewProgress(out io.Writer) Progress {
	return Progress{Out: out}
}

func (p Progress) Setup(module string, idx, total int) {
	_, _ = fmt.Fprintf(p.Out, "Setup for %s. [%d/%d module(s)]\n", module, idx, total)
}

func (p Progress) ModuleStart(module string, idx, total int) {
	_, _ = fmt.Fprintf(p.Out, "%s progressing. [%d/%d module(s)]\n", module, idx, total)
}

func (p Progress) StartSpinner(message string) func() {
	s := spinner.New(
		spinner.CharSets[14],
		120*time.Millisecond,
		spinner.WithWriter(p.Out),
		spinner.WithSuffix(" "+message),
	)
	s.Start()

	return func() {
		s.Stop()
		_, _ = fmt.Fprint(p.Out, "\r")
	}
}

func (p Progress) ModuleDone(module string, coverage string, idx, total int) {
	_, _ = fmt.Fprintf(p.Out, "%s finished. cov: %s%s%%%s [%d/%d module(s)]\n", module, ansiGreen, coverage, ansiReset, idx, total)
}

func (p Progress) Done() {
	_, _ = fmt.Fprintln(p.Out, "All tests have been completed.")
}

func (p Progress) DoneWithFailures() {
	_, _ = fmt.Fprintln(p.Out, "Some tests have failed.")
}
