# LedgerGate File Plan

Status: forecast, not a contract.
Last updated: 2026-11-04.

This document answers "which files are we likely to touch each week". The authoritative sequence of work is still the spec plus `to-tickets`; this is a capacity and layout forecast. File names may move as the design firms up.

Delivery follows `docs/pr-workflow.md`. The weekly table is capacity planning, not a PR boundary: W1 is split into small vertical-slice PRs under the risk budget, not merged as one weekly batch. Each PR must pass CI, independent review, and the rollback checks before it lands.

## Repository layout

```text
LedgerGate/
├── AGENTS.md
├── go.mod
├── go.sum
├── Makefile
├── .env.example
├── .gitignore
├── docker-compose.yml
├── .github/
│   └── workflows/ci.yml
├── cmd/
│   ├── gateway/main.go
│   ├── control/main.go
│   ├── ledger/main.go
│   ├── billing/main.go
│   ├── payment/main.go
│   └── reconcile/main.go
├── api/
│   └── openapi/
│       ├── control.yaml
│       ├── ledger.yaml
│       ├── billing.yaml
│       ├── payment.yaml
│       └── reconcile.yaml
├── internal/
│   ├── platform/
│   │   ├── config/
│   │   ├── log/
│   │   ├── observability/
│   │   ├── db/
│   │   ├── nats/
│   │   ├── redis/
│   │   ├── auth/
│   │   ├── crypt/
│   │   ├── ids/
│   │   ├── money/
│   │   ├── clock/
│   │   ├── outbox/
│   │   ├── flags/
│   │   └── httpx/
│   ├── gateway/
│   │   ├── ingress/
│   │   ├── router/
│   │   ├── ratelimit/
│   │   ├── provider/openai/
│   │   ├── provider/anthropic/
│   │   ├── provider/compat/
│   │   ├── usage/
│   │   ├── client/
│   │   ├── outbox/
│   │   └── auth/
│   ├── control/
│   │   ├── identity/
│   │   ├── apikey/
│   │   ├── session/
│   │   ├── audit/
│   │   ├── admin/
│   │   ├── httpapi/
│   │   └── web/
│   ├── ledger/
│   │   ├── account/
│   │   ├── wallet/
│   │   ├── entry/
│   │   ├── reservation/
│   │   ├── settlement/
│   │   ├── usage/
│   │   ├── pricing/
│   │   ├── consumer/
│   │   └── httpapi/
│   ├── billing/
│   │   ├── catalog/
│   │   ├── rating/
│   │   ├── subscription/
│   │   ├── grant/
│   │   ├── snapshot/
│   │   └── httpapi/
│   ├── payment/
│   │   ├── provider/stripe/
│   │   ├── provider/alipay/
│   │   ├── checkout/
│   │   ├── webhook/
│   │   ├── refund/
│   │   ├── fx/
│   │   └── httpapi/
│   ├── reconcile/
│   │   ├── provider/
│   │   ├── psp/
│   │   ├── ledger/
│   │   ├── daemon/
│   │   └── httpapi/
│   └── testkit/
│       ├── postgres.go
│       ├── nats.go
│       ├── redis.go
│       ├── mockprovider/
│       └── scenarios/
├── migrations/
│   ├── gateway/
│   ├── control/
│   ├── ledger/
│   ├── billing/
│   ├── payment/
│   └── reconcile/
├── web/
│   ├── package.json
│   ├── vite.config.ts
│   ├── index.html
│   └── src/
│       ├── main.tsx
│       ├── app.tsx
│       ├── api/
│       ├── auth/
│       ├── components/
│       ├── lib/
│       └── pages/
├── test/
│   ├── load/
│   └── fault/
├── observability/
│   ├── grafana/dashboards/
│   └── prometheus/
├── docs/
│   ├── architecture.md
│   ├── delivery-plan.md
│   ├── file-plan.md
│   ├── diagrams/
│   ├── adr/
│   └── reports/
└── .scratch/
    └── ledgergate-mvp/
        ├── spec.md
        └── issues/
```

## Weekly file forecast

W1 is detailed because it starts immediately. W2–W10 are coarse forecasts; they will be re-cut after `to-tickets` and again at the W5 gate.

W1 is delivered as several small PRs, not one weekly batch. The day-by-day table is a capacity forecast; use the risk tiers in `docs/pr-workflow.md` to decide where each PR boundary lands.

### W1 — foundations (detailed)

**Outcome:** six services boot in compose, connect to their infrastructure, expose health/readiness, and the platform packages, OpenAPI contracts and initial migrations are in place.

**Definition of done**

- `make up` starts Postgres, Redis, NATS JetStream, Jaeger, Prometheus, Grafana and all six services.
- Every service exposes `GET /healthz` (process alive) and `GET /readyz` (its dependencies reachable).
- `make migrate` applies all six migrations to their own databases.
- `make generate` is reproducible: running it twice leaves no diff.
- `make test` and `make lint` pass.
- `make build` produces all six binaries and one image with six entrypoints.
- Any request produces structured JSON logs with `service`, `request_id` and `trace_id`.
- Traces appear in Jaeger; metrics appear in Prometheus; Grafana has a datasource for both.

**Day-by-day**

| Day | Work | Files |
|---|---|---|
| 1 | Module, build, config, IDs, money, clock, logging, flags | `go.mod`, `go.sum`, `Makefile`, `.env.example`, `.gitignore`, `.golangci.yml`, `tools.go`, `internal/platform/{config,log,ids,money,clock,httpx,flags}/*.go` + tests |
| 2 | Service bootstrap, six entrypoints, health/readiness, service auth, crypto | `internal/platform/service/{service.go,health.go,routes.go}`, `internal/platform/{db,nats,redis,auth,crypt}/*.go`, `cmd/{gateway,control,ledger,billing,payment,reconcile}/main.go` |
| 3 | OpenAPI specs, codegen config, initial migrations, migrate command | `api/openapi/*.yaml`, `api/oapi-codegen.yaml`, `api/generated/**`, `migrations/*/0001_init.sql`, `internal/platform/db/migrate.go`, `cmd/*/main.go` migrate subcommand |
| 4 | Compose, Dockerfile, observability stack, CI | `docker-compose.yml`, `deploy/Dockerfile`, `deploy/compose/{prometheus.yml,jaeger.yml}`, `deploy/compose/grafana/provisioning/**`, `.github/workflows/ci.yml`, `scripts/smoke.ps1`, `scripts/smoke.sh` |
| 5 | Buffer: compose smoke, CI fixes, README, W1 gate | `README.md`, `docs/adr/0001-service-decomposition.md`, fixes across W1 files |

**File-by-file**

| Area | Files | Purpose |
|---|---|---|
| Build | `go.mod`, `go.sum`, `Makefile`, `.env.example`, `.gitignore`, `.golangci.yml`, `tools.go` | Module `github.com/NicoYazawa/LedgerGate`; targets `generate`, `lint`, `test`, `build`, `up`, `down`, `migrate`; pinned codegen tools |
| Local infra | `docker-compose.yml`, `deploy/Dockerfile` | Postgres 18, Redis 7.4, NATS 2.10 JetStream, Jaeger 1.60, Prometheus v3.1.0, Grafana 11.3.0, six services; multi-stage Go image |
| Observability config | `deploy/compose/prometheus.yml`, `deploy/compose/jaeger.yml`, `deploy/compose/grafana/provisioning/datasources.yml`, `deploy/compose/grafana/provisioning/dashboards.yml` | OTLP endpoint, Prometheus scrape config, Jaeger trace backend, Grafana datasources |
| CI | `.github/workflows/ci.yml` | `golangci-lint`, `go test ./...`, `go build ./cmd/...`, Docker build; Go 1.27 cache |
| Entrypoints | `cmd/{gateway,control,ledger,billing,payment,reconcile}/main.go` | Load config, run service, graceful shutdown |
| Shared bootstrap | `internal/platform/service/{service.go,health.go,routes.go}` | Common HTTP server, middleware, health/readiness, signal handling |
| Config | `internal/platform/config/{config.go,config_test.go}` | Env + YAML defaults; per-service config; fail fast on missing secrets |
| Logging | `internal/platform/log/log.go` | `slog` JSON; service name; request/trace correlation |
| Observability | `internal/platform/observability/{otel.go,metrics.go}` | OTLP tracer/meter providers; Prometheus registry |
| Database | `internal/platform/db/{pgx.go,migrate.go,tx.go,pgx_test.go}` | pgxpool, goose migrations per database, transaction helper |
| NATS | `internal/platform/nats/client.go` | Connection + JetStream context; health check |
| Redis | `internal/platform/redis/client.go` | go-redis client; health check |
| Service auth | `internal/platform/auth/{service_token.go,context.go}` | Per-service bearer token middleware; caller identity in context |
| Crypto | `internal/platform/crypt/{aesgcm.go,aesgcm_test.go}` | AES-256-GCM for provider keys; round-trip and tamper tests |
| IDs | `internal/platform/ids/{uuidv7.go,uuidv7_test.go}` | UUIDv7 plus typed prefixes (`req_`, `att_`, `settle_`) |
| Money | `internal/platform/money/{micro_usd.go,micro_usd_test.go}` | `int64` micro-USD; checked arithmetic; parse/format |
| Clock | `internal/platform/clock/{clock.go,clock_test.go}` | `Clock` interface; system and fake clocks |
| HTTP helpers | `internal/platform/httpx/{respond.go,middleware.go,health.go,respond_test.go}` | JSON errors (OpenAI shape for gateway, problem shape for admin/internal), request ID, recovery, logging |
| Outbox skeleton | `internal/platform/outbox/outbox.go` | Shared outbox row/interface; implementation lands with gateway in W4 |
| Feature flags | `internal/platform/flags/{flags.go,flags_test.go}` | Env-backed default-off flags; both states tested; no billing semantics |
| Process | `docs/pr-workflow.md`, `docs/prompts/pr-reviewer.md`, `docs/prompts/pr-fixer.md`, `.github/pull_request_template.md`, `scripts/pr-size.ps1` | Risk-tier PR gate, independent review and fixer prompts, size check |
| Contracts | `api/openapi/{control,ledger,billing,payment,reconcile}.yaml`, `api/oapi-codegen.yaml`, `api/generated/**` | OpenAPI-first contracts; generated server/client stubs |
| Schemas | `migrations/{gateway,control,ledger,billing,payment,reconcile}/0001_init.sql` | Initial tables per database, including `outbox` and `processed_events` where needed |
| Smoke scripts | `scripts/smoke.ps1`, `scripts/smoke.sh` | Curl every `/healthz` and `/readyz`; fail on non-200 |
| Docs | `README.md`, `docs/adr/0001-service-decomposition.md` | Quickstart, ports, compose commands; record the 6-service decision |

**W1 test files**

- `internal/platform/money/micro_usd_test.go` — parsing, formatting, overflow, rounding rules.
- `internal/platform/ids/uuidv7_test.go` — time ordering, prefix format, uniqueness.
- `internal/platform/crypt/aesgcm_test.go` — round-trip, wrong key, tamper detection.
- `internal/platform/config/config_test.go` — required values, env override, invalid config.
- `internal/platform/db/pgx_test.go` — testcontainers Postgres, ping, migrate, transaction rollback.
- `internal/platform/httpx/respond_test.go` — error shapes and status codes.
- `internal/platform/service/health_test.go` — health/ready behavior when a dependency is down.

**W1 dependency order**

1. `go.mod` + platform primitives (`money`, `ids`, `clock`, `config`, `log`, `httpx`).
2. `service` bootstrap + six `cmd/*/main.go` + health endpoints.
3. `db`, `redis`, `nats`, `auth`, `crypt` + migrations + `migrate` subcommand.
4. Compose, Dockerfile, observability, smoke scripts.
5. OpenAPI specs + codegen.
6. CI, README, ADR, W1 gate.

**W1 non-goals**

- No business logic in `ledger`, `billing`, `payment` or `reconcile` yet.
- No React app; `web/` starts in W6.
- No provider adapters, reservation or settlement; those start in W2–W5.
- No production deployment; compose only.

### W2–W3 — ledger core

**Outcome:** wallet, double-entry, reservation, settlement and idempotency work with property tests.

| Area | Files |
|---|---|
| Module | `internal/ledger/app.go` |
| Domain | `internal/ledger/{account,wallet,entry,reservation,settlement,usage,pricing}/` |
| API | `internal/ledger/httpapi/`, `internal/ledger/consumer/usage_recorded.go` |
| Schema | `migrations/ledger/0002_ledger_core.sql`, `0003_reservations.sql`, `0004_settlements.sql`, `0005_credit_lots.sql` |
| Test harness | `internal/testkit/{postgres.go,fixtures.go}` |
| Tests | `internal/ledger/**/*_test.go`, `internal/ledger/ledger_property_test.go` |

### W4–W5 — gateway and money path

**Outcome:** streaming request reserves, proxies, records usage and settles end-to-end.

| Area | Files |
|---|---|
| Module | `internal/gateway/app.go` |
| Ingress/SSE | `internal/gateway/ingress/{handler.go,sse.go,errors.go}` |
| Routing | `internal/gateway/router/{router.go,failover.go,health.go}` |
| Rate limit | `internal/gateway/ratelimit/{limiter.go,redis.go}` |
| Providers | `internal/gateway/provider/{provider.go,types.go}`, `provider/openai/`, `provider/anthropic/`, `provider/compat/` |
| Usage | `internal/gateway/usage/{extract.go,estimate.go}` |
| Ledger client | `internal/gateway/client/ledger.go` |
| Outbox | `internal/gateway/outbox/{writer.go,publisher.go}` |
| API-key cache | `internal/gateway/auth/{cache.go,redis.go}` |
| Schema | `migrations/gateway/0002_request_log.sql`, `0003_settlement_outbox.sql` |
| Fixtures | `internal/testkit/mockprovider/`, `internal/testkit/scenarios/`, `testdata/sse/{openai,anthropic}/*.txt` |
| Tests | `internal/gateway/**/*_test.go` |

### W6 — control and identity

**Outcome:** accounts, workspaces, users and API keys; admin login; key revocation reaches the gateway.

| Area | Files |
|---|---|
| Module | `internal/control/app.go` |
| Identity | `internal/control/identity/{account.go,workspace.go,user.go,service.go,repo.go}` |
| API keys | `internal/control/apikey/{key.go,service.go,repo.go,hashing.go}` |
| Sessions | `internal/control/session/{session.go,store.go,password.go}` |
| Audit | `internal/control/audit/{audit.go,repo.go}` |
| Admin BFF | `internal/control/admin/{bff.go,workspaces.go,keys.go}` |
| HTTP | `internal/control/httpapi/{handler.go,auth.go,csrf.go}`, `internal/control/web/embed.go` |
| Schema | `migrations/control/0002_identity.sql`, `0003_api_keys.sql`, `0004_sessions.sql`, `0005_audit.sql` |
| Frontend | `web/{package.json,vite.config.ts,index.html}`, `web/src/{main.tsx,app.tsx,api/client.ts}`, `web/src/pages/{login.tsx,workspaces.tsx,api-keys.tsx}` |
| Tests | `internal/control/**/*_test.go`, `web/src/**/*.test.tsx` |

### W7 — billing and subscription

**Outcome:** price catalog, price snapshots, plans, subscriptions and grants.

| Area | Files |
|---|---|
| Module | `internal/billing/app.go` |
| Catalog | `internal/billing/catalog/{model.go,service.go,repo.go}` |
| Rating | `internal/billing/rating/{rating.go,rules.go}` |
| Subscription | `internal/billing/subscription/{plan.go,subscription.go,service.go,repo.go}` |
| Grants | `internal/billing/grant/{grant.go,service.go,repo.go}` |
| Snapshots | `internal/billing/snapshot/{publisher.go,repo.go}` |
| HTTP | `internal/billing/httpapi/handler.go` |
| Schema | `migrations/billing/0002_price_catalog.sql`, `0003_subscriptions.sql`, `0004_grants.sql` |
| Frontend | `web/src/pages/{subscription.tsx,catalog.tsx}`, `web/src/api/billing.ts` |
| Tests | `internal/billing/**/*_test.go` |

### W8 — payment and FX

**Outcome:** Stripe test mode and Alipay sandbox top-ups, refunds, webhooks and FX.

| Area | Files |
|---|---|
| Module | `internal/payment/app.go` |
| Provider interface | `internal/payment/provider/{provider.go,types.go}` |
| Stripe | `internal/payment/provider/stripe/{client.go,checkout.go,webhook.go,refund.go}` |
| Alipay | `internal/payment/provider/alipay/{client.go,checkout.go,notify.go,refund.go}` |
| Checkout/refund | `internal/payment/checkout/{service.go,repo.go}`, `internal/payment/refund/{service.go,repo.go}` |
| Webhooks | `internal/payment/webhook/{handler.go,verify.go,idempotency.go}` |
| FX | `internal/payment/fx/{rate.go,repo.go}` |
| HTTP | `internal/payment/httpapi/handler.go` |
| Schema | `migrations/payment/0002_payment_intents.sql`, `0003_webhook_events.sql`, `0004_refunds.sql`, `0005_fx.sql` |
| Frontend | `web/src/pages/wallet.tsx`, `web/src/api/payment.ts` |
| Tests | `internal/payment/**/*_test.go`, `testdata/webhooks/{stripe,alipay}/*.json` |

### W9 — reconciliation, rate limiting, failover

**Outcome:** discrepancies are detected; rate limits and circuit breakers are enforced; admin reconciliation views work.

| Area | Files |
|---|---|
| Module | `internal/reconcile/app.go` |
| Provider checks | `internal/reconcile/provider/{openai.go,anthropic.go}` |
| PSP checks | `internal/reconcile/psp/{stripe.go,alipay.go}` |
| Ledger checks | `internal/reconcile/ledger/check.go` |
| Daemon | `internal/reconcile/daemon/{scheduler.go,runner.go}` |
| HTTP | `internal/reconcile/httpapi/handler.go` |
| Schema | `migrations/reconcile/0002_reconciliation.sql`, `0003_adjustments.sql` |
| Gateway hardening | `internal/gateway/ratelimit/{policy.go,limiter_test.go}`, `internal/gateway/router/{circuitbreaker.go,failover_test.go}` |
| Frontend | `web/src/pages/{reconcile.tsx,audit.tsx}`, `web/src/api/reconcile.ts` |
| Tests | `internal/reconcile/**/*_test.go`, `internal/gateway/**/*_test.go` |

### W10 — admin UI, observability, load/fault, docs, kind

**Outcome:** remaining admin pages, dashboards, load and fault reports, runbook, kind acceptance.

| Area | Files |
|---|---|
| Frontend | `web/src/pages/{usage.tsx,dashboard.tsx}`, `web/src/components/charts/`, `web/src/lib/format.ts` |
| Kind | `deploy/kind/kind.yaml`, `deploy/kind/manifests/*.yaml`, `deploy/kind/README.md` |
| Load | `test/load/{gateway.js,concurrency.js}` |
| Fault | `test/fault/{inject.go,scenarios.md}` |
| Observability | `observability/grafana/dashboards/*.json`, `observability/prometheus/prometheus.yml` |
| ADRs | `docs/adr/0001-microservices.md`, `0002-reserve-settle-saga.md`, `0003-single-usd-wallet.md`, `0004-nats-outbox.md` |
| Docs | `README.md`, `docs/runbook.md`, `docs/reports/load.md`, `docs/reports/fault.md` |
| Tests | final acceptance suite under `internal/testkit/` and `test/` |

## Rules of thumb

- Do not create a file before its first vertical slice needs it.
- Prefer adding tests next to the slice, not a separate "testing week".
- Shared platform changes ripple: batch them early in the week, not on Friday.
- When the W5 gate fails, cut UI and PSP breadth before touching ledger or gateway files.
