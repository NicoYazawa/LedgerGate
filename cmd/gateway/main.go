package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/NicoYazawa/LedgerGate/internal/platform/config"
	"github.com/NicoYazawa/LedgerGate/internal/platform/service"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "gateway: %v\n", err)
		os.Exit(1)
	}
}
func run() error {
	cfg, err := config.Load("gateway")
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return service.Run(ctx, cfg, service.Dependencies{}, nil)
}
