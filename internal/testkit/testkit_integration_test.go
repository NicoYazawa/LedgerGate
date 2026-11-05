//go:build integration

package testkit

import (
	"context"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestPostgresStartsAndStops(t *testing.T) {
	useExplicitContainerCleanup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	postgres, err := StartPostgres(ctx)
	if err != nil {
		t.Fatalf("StartPostgres() error = %v", err)
	}
	parsed, err := url.Parse(postgres.DSN())
	if err != nil {
		t.Fatalf("parse DSN: %v", err)
	}
	if parsed.Scheme != "postgres" || parsed.Path != "/ledgergate_test" {
		t.Errorf("DSN = %q", postgres.DSN())
	}
	assertTCP(t, parsed.Host)
	if err := postgres.Close(ctx); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestRedisStartsAndStops(t *testing.T) {
	useExplicitContainerCleanup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	redis, err := StartRedis(ctx)
	if err != nil {
		t.Fatalf("StartRedis() error = %v", err)
	}
	if !strings.HasPrefix(redis.URL(), "redis://") {
		t.Errorf("URL = %q", redis.URL())
	}
	parsed, err := url.Parse(redis.URL())
	if err != nil {
		t.Fatalf("parse Redis URL: %v", err)
	}
	assertTCP(t, parsed.Host)
	if err := redis.Close(ctx); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestNATSStartsAndStops(t *testing.T) {
	useExplicitContainerCleanup(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	nats, err := StartNATS(ctx)
	if err != nil {
		t.Fatalf("StartNATS() error = %v", err)
	}
	if !strings.HasPrefix(nats.URL(), "nats://") {
		t.Errorf("URL = %q", nats.URL())
	}
	assertTCP(t, strings.TrimPrefix(nats.URL(), "nats://"))
	if err := nats.Close(ctx); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func assertTCP(t *testing.T, address string) {
	t.Helper()
	connection, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		t.Errorf("dial %s: %v", address, err)
		return
	}
	_ = connection.Close()
}

func useExplicitContainerCleanup(t *testing.T) {
	t.Helper()
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
}
