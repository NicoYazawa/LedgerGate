package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env, Service, HTTPAddr         string
	ShutdownTimeout                time.Duration
	DatabaseURL, RedisURL, NATSURL string
	OTLPEndpoint, LogLevel         string
	ServiceTokens                  map[string]string
}
type fileConfig struct {
	Env      *string       `yaml:"env"`
	Service  *string       `yaml:"service"`
	HTTP     *fileHTTP     `yaml:"http"`
	Database *fileURL      `yaml:"database"`
	Redis    *fileURL      `yaml:"redis"`
	NATS     *fileURL      `yaml:"nats"`
	OTel     *fileEndpoint `yaml:"otel"`
	Log      *fileLog      `yaml:"log"`
	Auth     *fileAuth     `yaml:"auth"`
}
type fileHTTP struct {
	Addr            *string `yaml:"addr"`
	ShutdownTimeout *string `yaml:"shutdown_timeout"`
}
type fileURL struct {
	URL *string `yaml:"url"`
}
type fileEndpoint struct {
	Endpoint *string `yaml:"endpoint"`
}
type fileLog struct {
	Level *string `yaml:"level"`
}
type fileAuth struct {
	ServiceTokens map[string]string `yaml:"service_tokens"`
}

func Load(service string) (Config, error) {
	cfg := Config{Env: "dev", Service: service, LogLevel: "info", ShutdownTimeout: 10 * time.Second}
	if path := os.Getenv("LG_CONFIG_FILE"); path != "" {
		if err := loadFile(path, &cfg); err != nil {
			return Config{}, err
		}
	}
	if err := loadEnvironment(&cfg); err != nil {
		return Config{}, err
	}
	if err := validate(cfg, service); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %q: %w", path, err)
	}
	var raw fileConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return fmt.Errorf("decode config file %q: %w", path, err)
	}

	set(&cfg.Env, raw.Env)
	set(&cfg.Service, raw.Service)
	if raw.HTTP != nil {
		set(&cfg.HTTPAddr, raw.HTTP.Addr)
		if raw.HTTP.ShutdownTimeout != nil {
			cfg.ShutdownTimeout, err = parseDuration(*raw.HTTP.ShutdownTimeout)
			if err != nil {
				return err
			}
		}
	}
	setURL(&cfg.DatabaseURL, raw.Database)
	setURL(&cfg.RedisURL, raw.Redis)
	setURL(&cfg.NATSURL, raw.NATS)
	if raw.OTel != nil {
		set(&cfg.OTLPEndpoint, raw.OTel.Endpoint)
	}
	if raw.Log != nil {
		set(&cfg.LogLevel, raw.Log.Level)
	}
	if raw.Auth != nil && raw.Auth.ServiceTokens != nil {
		cfg.ServiceTokens = raw.Auth.ServiceTokens
	}
	return nil
}
func loadEnvironment(cfg *Config) error {
	overrides := map[*string]string{
		&cfg.Env: "LG_ENV", &cfg.Service: "LG_SERVICE", &cfg.HTTPAddr: "LG_HTTP_ADDR",
		&cfg.DatabaseURL: "LG_DATABASE_URL", &cfg.RedisURL: "LG_REDIS_URL",
		&cfg.NATSURL: "LG_NATS_URL", &cfg.OTLPEndpoint: "LG_OTLP_ENDPOINT",
		&cfg.LogLevel: "LG_LOG_LEVEL",
	}
	for destination, name := range overrides {
		if value := os.Getenv(name); value != "" {
			*destination = value
		}
	}
	if value := os.Getenv("LG_HTTP_SHUTDOWN_TIMEOUT"); value != "" {
		timeout, err := parseDuration(value)
		if err != nil {
			return err
		}
		cfg.ShutdownTimeout = timeout
	}
	if value, ok := os.LookupEnv("LG_SERVICE_TOKENS"); ok {
		if strings.TrimSpace(value) == "" {
			cfg.ServiceTokens = nil
			return nil
		}
		if err := json.Unmarshal([]byte(value), &cfg.ServiceTokens); err != nil {
			return fmt.Errorf("parse field auth.service_tokens (LG_SERVICE_TOKENS): %w", err)
		}
	}
	return nil
}
func validate(cfg Config, expectedService string) error {
	if !validService(cfg.Service) {
		return fmt.Errorf("invalid field service (LG_SERVICE): %q", cfg.Service)
	}
	if cfg.Service != expectedService {
		return fmt.Errorf("field service (LG_SERVICE) = %q, must match binary %q", cfg.Service, expectedService)
	}
	if cfg.Env != "dev" && cfg.Env != "test" && cfg.Env != "prod" {
		return fmt.Errorf("invalid field env (LG_ENV): %q", cfg.Env)
	}
	if cfg.HTTPAddr == "" {
		return fmt.Errorf("invalid field http.addr (LG_HTTP_ADDR): missing required value")
	}
	if err := validatePort(cfg.HTTPAddr); err != nil {
		return fmt.Errorf("invalid field http.addr (LG_HTTP_ADDR): %w", err)
	}
	if cfg.ShutdownTimeout <= 0 {
		return fmt.Errorf("invalid field http.shutdown_timeout (LG_HTTP_SHUTDOWN_TIMEOUT): must be positive")
	}
	if err := validateURL("database.url", "LG_DATABASE_URL", cfg.DatabaseURL); err != nil {
		return err
	}
	if cfg.Service == "gateway" && cfg.RedisURL == "" {
		return fmt.Errorf("missing required field redis.url (LG_REDIS_URL)")
	}
	if cfg.RedisURL != "" {
		if err := validateURL("redis.url", "LG_REDIS_URL", cfg.RedisURL); err != nil {
			return err
		}
	}
	if err := validateURL("nats.url", "LG_NATS_URL", cfg.NATSURL); err != nil {
		return err
	}
	if err := validateEndpoint(cfg.OTLPEndpoint); err != nil {
		return fmt.Errorf("invalid field otel.endpoint (LG_OTLP_ENDPOINT): %w", err)
	}
	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid field log.level (LG_LOG_LEVEL): %q", cfg.LogLevel)
	}
	if len(cfg.ServiceTokens) == 0 {
		return fmt.Errorf("missing required field auth.service_tokens (LG_SERVICE_TOKENS)")
	}
	for caller, token := range cfg.ServiceTokens {
		if strings.TrimSpace(caller) == "" || strings.TrimSpace(token) == "" {
			return fmt.Errorf("invalid field auth.service_tokens (LG_SERVICE_TOKENS): names and values must be non-empty")
		}
	}
	return nil
}
func validateURL(field, env, value string) error {
	if value == "" {
		return fmt.Errorf("missing required field %s (%s)", field, env)
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid field %s (%s): %q is not a URL", field, env, value)
	}
	return nil
}
func validateEndpoint(endpoint string) error {
	if endpoint == "" {
		return fmt.Errorf("missing required value")
	}
	if strings.Contains(endpoint, "://") {
		parsed, err := url.Parse(endpoint)
		if err != nil || parsed.Host == "" {
			return fmt.Errorf("invalid endpoint %q", endpoint)
		}
		return nil
	}
	return validatePort(endpoint)
}

func validatePort(address string) error {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return fmt.Errorf("invalid port %q", port)
	}
	return nil
}

func parseDuration(value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse field http.shutdown_timeout (LG_HTTP_SHUTDOWN_TIMEOUT): %w", err)
	}
	return duration, nil
}

func validService(service string) bool {
	return strings.Contains("|gateway|control|ledger|billing|payment|reconcile|", "|"+service+"|")
}

func set(destination *string, source *string) {
	if source != nil {
		*destination = *source
	}
}

func setURL(destination *string, source *fileURL) {
	if source != nil {
		set(destination, source.URL)
	}
}
