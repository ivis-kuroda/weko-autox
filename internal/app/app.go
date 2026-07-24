package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ivis-kuroda/weko-autox/internal/cli"
	"github.com/ivis-kuroda/weko-autox/internal/dockerx"
	"github.com/ivis-kuroda/weko-autox/internal/report"
	"github.com/ivis-kuroda/weko-autox/internal/runner"
	"github.com/ivis-kuroda/weko-autox/internal/signalx"
)

type App struct {
	stdout io.Writer
	stderr io.Writer
}

func New(stdout io.Writer, stderr io.Writer) (*App, error) {
	if stdout == nil || stderr == nil {
		return nil, fmt.Errorf("stdout/stderr must not be nil")
	}
	return &App{stdout: stdout, stderr: stderr}, nil
}

func (a *App) Execute(ctx context.Context, args []string) error {
	return cli.Execute(ctx, a.stdout, args, func(ctx context.Context, cfg cli.Config) error {
		workspace, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("detect workspace: %w", err)
		}

		execClient, err := dockerx.New(ctx, workspace, "web")
		if err != nil {
			return err
		}
		defer execClient.Close()

		r := runner.Runner{
			Workspace: workspace,
			Exec:      execClient,
			Progress:  report.NewProgress(a.stdout),
		}

		sigCtx, stop := signalx.WithShutdownSignal(ctx, func() {
			_ = execClient.StopTests(context.Background())
		})
		defer stop()

		return r.Run(sigCtx, cfg)
	})
}
