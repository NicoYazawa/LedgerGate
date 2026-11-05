package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/NicoYazawa/LedgerGate/internal/platform/config"
)

func TestRunStartsServiceWithLoadedConfig(t *testing.T) {
	var stderr bytes.Buffer
	stopCalled := false
	serviceCalled := false
	expected := config.Config{Service: "gateway", HTTPAddr: ":8080"}
	deps := runDependencies{
		loadConfig: func(service string) (config.Config, error) {
			if service != "gateway" {
				t.Errorf("service = %q, want gateway", service)
			}
			return expected, nil
		},
		runService: func(ctx context.Context, cfg config.Config) error {
			serviceCalled = true
			if ctx == nil {
				t.Error("context is nil")
			}
			if cfg.HTTPAddr != expected.HTTPAddr {
				t.Errorf("HTTPAddr = %q, want %q", cfg.HTTPAddr, expected.HTTPAddr)
			}
			return nil
		},
		notifyContext: func(ctx context.Context, _ ...os.Signal) (context.Context, context.CancelFunc) {
			return ctx, func() { stopCalled = true }
		},
	}

	code := run(deps, &stderr)

	if code != 0 {
		t.Fatalf("run() = %d, stderr = %s", code, stderr.String())
	}
	if !serviceCalled || !stopCalled {
		t.Errorf("serviceCalled = %v, stopCalled = %v", serviceCalled, stopCalled)
	}
}

func TestRunReportsConfigFailure(t *testing.T) {
	var stderr bytes.Buffer
	deps := runDependencies{
		loadConfig: func(string) (config.Config, error) {
			return config.Config{}, errors.New("missing LG_HTTP_ADDR")
		},
		runService: func(context.Context, config.Config) error {
			t.Fatal("runService called after config failure")
			return nil
		},
		notifyContext: func(ctx context.Context, _ ...os.Signal) (context.Context, context.CancelFunc) {
			return ctx, func() {}
		},
	}

	code := run(deps, &stderr)

	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "missing LG_HTTP_ADDR") {
		t.Errorf("stderr = %q, want config error", stderr.String())
	}
}

func TestRunReportsServiceFailure(t *testing.T) {
	var stderr bytes.Buffer
	deps := runDependencies{
		loadConfig: func(string) (config.Config, error) {
			return config.Config{Service: "gateway"}, nil
		},
		runService: func(context.Context, config.Config) error {
			return errors.New("listen failed")
		},
		notifyContext: func(ctx context.Context, _ ...os.Signal) (context.Context, context.CancelFunc) {
			return ctx, func() {}
		},
	}

	code := run(deps, &stderr)

	if code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "listen failed") {
		t.Errorf("stderr = %q, want service error", stderr.String())
	}
}

func TestDefaultDependenciesAreConfigured(t *testing.T) {
	deps := defaultDependencies()
	if deps.loadConfig == nil || deps.runService == nil || deps.notifyContext == nil {
		t.Fatalf("defaultDependencies() = %+v, want all functions", deps)
	}

	err := deps.runService(context.Background(), config.Config{
		Service: "gateway", HTTPAddr: "not-an-address", LogLevel: "info",
	})
	if err == nil {
		t.Fatal("default runService error = nil, want listen error")
	}
}
