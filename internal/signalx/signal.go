package signalx

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func WithShutdownSignal(ctx context.Context, onSignal func()) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-ch:
			onSignal()
			cancel()
		}
	}()

	return ctx, func() {
		signal.Stop(ch)
		close(ch)
		cancel()
	}
}
