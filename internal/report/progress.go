package report

import (
	"fmt"
	"io"
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

func (p Progress) ModuleDone(module string, coverage string, idx, total int) {
	_, _ = fmt.Fprintf(p.Out, "%s finished. cov: %s%% [%d/%d module(s)]\n", module, coverage, idx, total)
}

func (p Progress) Done() {
	_, _ = fmt.Fprintln(p.Out, "All tests have been completed.")
}
