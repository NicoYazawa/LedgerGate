# W1 Canonical Configuration

Status: normative for both W1 implementations.
Last updated: 2026-11-04.

Both implementations must parse the same YAML shape and the same environment variables. The only difference between services is the values, not the field names.

## Precedence

Defaults < YAML file < environment variables.

- YAML is optional; `LG_CONFIG_FILE` points at it.
- Environment variables always win.
- Secrets (database password, PSP keys, provider keys, service tokens) come from environment or mounted files in production. Committed YAML contains placeholders only.
- Unknown YAML fields are an error (strict decoding), so typos fail fast.

## Canonical YAML

The same shape is used by all six services:

```yaml
env: dev
service: gateway

http:
  addr: ":8080"
  shutdown_timeout: 10s

database:
  url: "postgres://ledgergate:ledgergate@localhost:5432/ledgergate_gateway?sslmode=disable"

redis:
  url: "redis://localhost:6379/0"

nats:
  url: "nats://localhost:4222"

otel:
  endpoint: "localhost:4317"

log:
  level: info

auth:
  service_tokens:
    gateway: "dev-gateway-token"
    control: "dev-control-token"
    ledger: "dev-ledger-token"
    billing: "dev-billing-token"
    payment: "dev-payment-token"
    reconcile: "dev-reconcile-token"
```

Per-service YAML differs only in `service`, `http.addr`, `database.url`, and whether `redis` is present.

## Field contract

| YAML path | Type | Required | Default | Environment override | Notes |
|---|---|---|---|---|---|
| `env` | string | no | `dev` | `LG_ENV` | One of `dev`, `test`, `prod` |
| `service` | string | yes | — | `LG_SERVICE` | One of the six service names |
| `http.addr` | string | yes | — | `LG_HTTP_ADDR` | `host:port` or `:port` |
| `http.shutdown_timeout` | duration | no | `10s` | `LG_HTTP_SHUTDOWN_TIMEOUT` | Go duration string |
| `database.url` | string | yes | — | `LG_DATABASE_URL` | Postgres DSN; database name is `ledgergate_<service>` |
| `redis.url` | string | gateway only | — | `LG_REDIS_URL` | Required for gateway; optional for others |
| `nats.url` | string | yes | — | `LG_NATS_URL` | NATS URL |
| `otel.endpoint` | string | yes | — | `LG_OTLP_ENDPOINT` | OTLP gRPC endpoint |
| `log.level` | string | no | `info` | `LG_LOG_LEVEL` | `debug`, `info`, `warn`, `error` |
| `auth.service_tokens` | map[string]string | yes | — | `LG_SERVICE_TOKENS` | JSON object in the env var; overrides the whole map, no merge |

## Environment examples

```dotenv
LG_ENV=dev
LG_SERVICE=gateway
LG_HTTP_ADDR=:8080
LG_HTTP_SHUTDOWN_TIMEOUT=10s
LG_DATABASE_URL=postgres://ledgergate:ledgergate@localhost:5432/ledgergate_gateway?sslmode=disable
LG_REDIS_URL=redis://localhost:6379/0
LG_NATS_URL=nats://localhost:4222
LG_OTLP_ENDPOINT=localhost:4317
LG_LOG_LEVEL=info
LG_SERVICE_TOKENS={"gateway":"dev-gateway-token","control":"dev-control-token","ledger":"dev-ledger-token","billing":"dev-billing-token","payment":"dev-payment-token","reconcile":"dev-reconcile-token"}
```

Feature flags are environment-only: `LG_FEATURE_<UPPER_SNAKE>=true|false`, default off. They are not part of the YAML shape and are evaluated server-side.

## Validation rules

- `service` must be one of `gateway`, `control`, `ledger`, `billing`, `payment`, `reconcile`.
- `http.addr` must parse as a TCP address.
- `http.shutdown_timeout` must be positive and parse as a Go duration.
- `database.url`, `nats.url`, `otel.endpoint` must be non-empty and parseable.
- `redis.url` is required when `service == gateway`.
- `log.level` must be one of the four levels.
- `auth.service_tokens` must be a non-empty map whose values are non-empty.
- Missing required values fail process startup with a message naming the field and its environment variable.
- Strict YAML decoding: unknown keys are an error.

## Service mapping

| Service | `service` | `http.addr` | Database | Redis |
|---|---|---|---|---|
| gateway | `gateway` | `:8080` | `ledgergate_gateway` | required |
| control | `control` | `:8081` | `ledgergate_control` | not used |
| ledger | `ledger` | `:8082` | `ledgergate_ledger` | not used |
| billing | `billing` | `:8083` | `ledgergate_billing` | not used |
| payment | `payment` | `:8084` | `ledgergate_payment` | not used |
| reconcile | `reconcile` | `:8085` | `ledgergate_reconcile` | not used |

## Rules

- Do not invent new top-level YAML sections in W1. Add fields inside the existing sections.
- Do not read secrets from committed YAML.
- Do not merge environment maps; `LG_SERVICE_TOKENS` replaces the YAML map wholesale.
- The config loader is shared by all six services; service-specific behavior is driven by `service`, not by six different config shapes.
