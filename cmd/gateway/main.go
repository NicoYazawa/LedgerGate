package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/NicoYazawa/LedgerGate/internal/platform/config"
	"github.com/NicoYazawa/LedgerGate/internal/platform/service"
)

type runDependencies struct {
	loadConfig    func(string) (config.Config, error)
	runService    func(context.Context, config.Config) error
	notifyContext func(context.Context, ...os.Signal) (context.Context, context.CancelFunc)
}

func main() {
	os.Exit(run(defaultDependencies(), os.Stderr))
}

func defaultDependencies() runDependencies {
	return runDependencies{
		loadConfig: config.Load,
		runService: func(ctx context.Context, cfg config.Config) error {
			return service.Run(ctx, cfg, service.Dependencies{}, nil)
		},
		notifyContext: signal.NotifyContext,
	}
}

func run(deps runDependencies, stderr io.Writer) int {
	ctx, stop := deps.notifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, err := deps.loadConfig("gateway")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "gateway: %v\n", err)
		return 1
	}
	if err := deps.runService(ctx, cfg); err != nil {
		_, _ = fmt.Fprintf(stderr, "gateway: %v\n", err)
		return 1
	}
	return 0
}
