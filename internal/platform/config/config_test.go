package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadMissingRequiredValueNamesFieldAndEnvironmentVariable(t *testing.T) {
	setGatewayEnvironment(t)
	t.Setenv("LG_SERVICE_TOKENS", "")

	_, err := Load("gateway")

	if err == nil {
		t.Fatal("Load() error = nil, want missing auth.service_tokens error")
	}
	if !strings.Contains(err.Error(), "auth.service_tokens") {
		t.Errorf("Load() error = %q, want field auth.service_tokens", err)
	}
	if !strings.Contains(err.Error(), "LG_SERVICE_TOKENS") {
		t.Errorf("Load() error = %q, want environment variable LG_SERVICE_TOKENS", err)
	}
}

func TestLoadEnvironmentOverridesYAML(t *testing.T) {
	configPath := writeConfig(t, `
env: test
service: gateway
http:
  addr: ":9001"
  shutdown_timeout: 3s
database:
  url: "postgres://user:pass@db:5432/ledgergate_gateway?sslmode=disable"
redis:
  url: "redis://cache:6379/0"
nats:
  url: "nats://events:4222"
otel:
  endpoint: "jaeger:4317"
log:
  level: warn
auth:
  service_tokens:
    gateway: "gateway-token"
`)
	t.Setenv("LG_CONFIG_FILE", configPath)
	t.Setenv("LG_HTTP_ADDR", ":9100")
	t.Setenv("LG_LOG_LEVEL", "debug")
	t.Setenv("LG_HTTP_SHUTDOWN_TIMEOUT", "7s")

	cfg, err := Load("gateway")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.HTTPAddr != ":9100" {
		t.Errorf("HTTPAddr = %q, want :9100", cfg.HTTPAddr)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug", cfg.LogLevel)
	}
	if cfg.ShutdownTimeout != 7*time.Second {
		t.Errorf("ShutdownTimeout = %s, want 7s", cfg.ShutdownTimeout)
	}
	if cfg.Env != "test" || cfg.DatabaseURL == "" || cfg.RedisURL == "" {
		t.Errorf("YAML values were not loaded: %+v", cfg)
	}
}

func TestLoadRejectsUnknownYAMLFields(t *testing.T) {
	configPath := writeConfig(t, `
service: gateway
unknown: true
`)
	t.Setenv("LG_CONFIG_FILE", configPath)

	_, err := Load("gateway")

	if err == nil {
		t.Fatal("Load() error = nil, want strict YAML decoding error")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("Load() error = %q, want unknown field name", err)
	}
}

func TestLoadGatewayRequiresRedis(t *testing.T) {
	configPath := writeConfig(t, `
service: gateway
http:
  addr: ":8080"
database:
  url: "postgres://user:pass@db:5432/ledgergate_gateway?sslmode=disable"
nats:
  url: "nats://events:4222"
otel:
  endpoint: "jaeger:4317"
auth:
  service_tokens:
    gateway: "gateway-token"
`)
	t.Setenv("LG_CONFIG_FILE", configPath)

	_, err := Load("gateway")

	if err == nil {
		t.Fatal("Load() error = nil, want missing redis.url error")
	}
	if !strings.Contains(err.Error(), "redis.url") || !strings.Contains(err.Error(), "LG_REDIS_URL") {
		t.Errorf("Load() error = %q, want redis.url and LG_REDIS_URL", err)
	}
}

func TestLoadRejectsMismatchedService(t *testing.T) {
	setGatewayEnvironment(t)
	t.Setenv("LG_SERVICE", "control")

	_, err := Load("gateway")

	if err == nil {
		t.Fatal("Load() error = nil, want service mismatch error")
	}
	if !strings.Contains(err.Error(), "gateway") || !strings.Contains(err.Error(), "control") {
		t.Errorf("Load() error = %q, want both service names", err)
	}
}

func TestLoadUsesTenSecondShutdownDefault(t *testing.T) {
	setGatewayEnvironment(t)
	t.Setenv("LG_HTTP_SHUTDOWN_TIMEOUT", "")

	cfg, err := Load("gateway")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %s, want 10s", cfg.ShutdownTimeout)
	}
}

func TestLoadRejectsInvalidEnvironmentValues(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError string
	}{
		{name: "environment", key: "LG_ENV", value: "staging", wantError: "LG_ENV"},
		{name: "http address", key: "LG_HTTP_ADDR", value: "localhost", wantError: "LG_HTTP_ADDR"},
		{name: "shutdown timeout", key: "LG_HTTP_SHUTDOWN_TIMEOUT", value: "-1s", wantError: "LG_HTTP_SHUTDOWN_TIMEOUT"},
		{name: "database URL", key: "LG_DATABASE_URL", value: "not-a-url", wantError: "LG_DATABASE_URL"},
		{name: "redis URL", key: "LG_REDIS_URL", value: "not-a-url", wantError: "LG_REDIS_URL"},
		{name: "nats URL", key: "LG_NATS_URL", value: "not-a-url", wantError: "LG_NATS_URL"},
		{name: "otel endpoint", key: "LG_OTLP_ENDPOINT", value: "bad", wantError: "LG_OTLP_ENDPOINT"},
		{name: "log level", key: "LG_LOG_LEVEL", value: "verbose", wantError: "LG_LOG_LEVEL"},
		{name: "service tokens", key: "LG_SERVICE_TOKENS", value: `{"gateway":" "}`, wantError: "LG_SERVICE_TOKENS"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setGatewayEnvironment(t)
			t.Setenv(test.key, test.value)

			_, err := Load("gateway")

			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("Load() error = %q, want %s", err, test.wantError)
			}
		})
	}
}

func TestLoadRejectsMalformedServiceTokens(t *testing.T) {
	setGatewayEnvironment(t)
	t.Setenv("LG_SERVICE_TOKENS", "{")

	_, err := Load("gateway")

	if err == nil || !strings.Contains(err.Error(), "LG_SERVICE_TOKENS") {
		t.Errorf("Load() error = %q, want malformed tokens error", err)
	}
}

func TestLoadRejectsUnreadableConfigFile(t *testing.T) {
	setGatewayEnvironment(t)
	t.Setenv("LG_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.yaml"))

	_, err := Load("gateway")

	if err == nil || !strings.Contains(err.Error(), "missing.yaml") {
		t.Errorf("Load() error = %q, want config path", err)
	}
}

func TestLoadAcceptsURLOTLPEndpoint(t *testing.T) {
	setGatewayEnvironment(t)
	t.Setenv("LG_OTLP_ENDPOINT", "http://jaeger:4317")

	cfg, err := Load("gateway")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.OTLPEndpoint != "http://jaeger:4317" {
		t.Errorf("OTLPEndpoint = %q", cfg.OTLPEndpoint)
	}
}

func setGatewayEnvironment(t *testing.T) {
	t.Helper()
	values := map[string]string{
		"LG_ENV":                   "test",
		"LG_SERVICE":               "gateway",
		"LG_HTTP_ADDR":             "127.0.0.1:8080",
		"LG_HTTP_SHUTDOWN_TIMEOUT": "10s",
		"LG_DATABASE_URL":          "postgres://user:pass@db:5432/ledgergate_gateway?sslmode=disable",
		"LG_REDIS_URL":             "redis://cache:6379/0",
		"LG_NATS_URL":              "nats://events:4222",
		"LG_OTLP_ENDPOINT":         "jaeger:4317",
		"LG_LOG_LEVEL":             "info",
		"LG_SERVICE_TOKENS":        `{"gateway":"gateway-token"}`,
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
