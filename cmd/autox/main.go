package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ivis-kuroda/weko-autox/internal/app"
)

func main() {
	ctx := context.Background()
	application, err := app.New(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := application.Execute(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
