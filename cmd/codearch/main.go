package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Cyaside/codebase-explorer/internal/app"
	"github.com/Cyaside/codebase-explorer/internal/cli"
	"github.com/Cyaside/codebase-explorer/internal/config"
)

func main() {
	ctx := context.Background()

	settings, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	service := app.New(settings)

	exitCode, err := cli.Run(ctx, os.Args[1:], service, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	os.Exit(exitCode)
}
