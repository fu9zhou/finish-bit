package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fu9zhou/finish-bit/internal/cli"
	"github.com/fu9zhou/finish-bit/pkg/app"
)

var version = "dev"

func main() {
	application, err := app.New(app.Config{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: initialize FinishBit: %v\n", err)
		os.Exit(1)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	os.Exit(cli.New(application, os.Stdout, os.Stderr, version).Run(ctx, os.Args[1:]))
}
