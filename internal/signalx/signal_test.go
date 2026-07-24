package signalx

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestWithShutdownSignal_CancelWithoutSignal(t *testing.T) {
	called := false
	ctx, stop := WithShutdownSignal(context.Background(), func() {
		called = true
	})

	stop()

	select {
	case <-ctx.Done():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not canceled by stop()")
	}

	if called {
		t.Fatal("callback should not be called when stop() is invoked directly")
	}
}

func TestWithShutdownSignal_InterruptTriggersCallback(t *testing.T) {
	calledCh := make(chan struct{}, 1)
	ctx, stop := WithShutdownSignal(context.Background(), func() {
		calledCh <- struct{}{}
	})
	defer stop()

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("FindProcess() error = %v", err)
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		t.Fatalf("Signal() error = %v", err)
	}

	select {
	case <-calledCh:
	case <-time.After(2 * time.Second):
		t.Fatal("expected callback to be called on interrupt")
	}

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not canceled after interrupt")
	}
}
