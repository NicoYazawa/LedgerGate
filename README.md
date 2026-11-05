# LedgerGate

LedgerGate is a multi-tenant LLM gateway and billing platform. It combines an OpenAI-compatible gateway with a money-grade ledger, prepaid wallets, subscriptions, sandbox payments, and provider reconciliation.

The goal is simple: every model call should be traceable to a request, a usage record, a price snapshot, a provider cost, a settlement, and a reconciliation outcome.

## Status

The repository currently contains the specification, architecture, delivery plan, and PR workflow. The W1 foundation is delivered as small, independently reviewed pull requests; it is not a single merge.

- Product spec: [`.scratch/ledgergate-mvp/spec.md`](.scratch/ledgergate-mvp/spec.md)
- Architecture: [`docs/architecture.md`](docs/architecture.md)
- W1 foundation: [`docs/w1-foundation.md`](docs/w1-foundation.md)
- PR workflow: [`docs/pr-workflow.md`](docs/pr-workflow.md)
- Delivery plan: [`docs/delivery-plan.md`](docs/delivery-plan.md)
- File forecast: [`docs/file-plan.md`](docs/file-plan.md)
- Local Docker rules: [`docs/local-dev.md`](docs/local-dev.md)
- Pinned images: [`docs/images.md`](docs/images.md)

## Architecture at a glance

Six services:

- `gateway` — OpenAI-compatible ingress, streaming, provider adapters, routing/failover, rate limiting, reservation client, settlement outbox.
- `control` — accounts, workspaces, users, API keys, sessions, audit, admin BFF and React UI.
- `ledger` — wallets, double-entry postings, reservations, settlements, usage attempts, credit lots.
- `billing` — price catalog, price snapshots, plans, subscriptions, grants.
- `payment` — Stripe test mode and Alipay sandbox, top-ups, refunds, webhooks, FX.
- `reconcile` — provider, PSP and ledger reconciliation.

Infrastructure: PostgreSQL 18, Redis 7.4, NATS JetStream, Jaeger, Prometheus, Grafana. Go 1.27. React 19 for the admin surface.

## Development

Follow [`docs/pr-workflow.md`](docs/pr-workflow.md). In short:

- One ticket, one small vertical-slice PR.
- Risk-based size budgets; run `powershell -NoProfile -ExecutionPolicy Bypass -File scripts/pr-size.ps1` before opening a PR.
- Every PR must build, test, pass independent review, and have a rollback plan.
- Every completed stage commits Mermaidiff, CodeGraph and PlantUML review artifacts; verify them with `make verify-artifacts STAGE=<slug>`.
- Total Go statement coverage, including `cmd/*`, must be at least 90%; run `make coverage`.
- Incomplete user-visible behavior is behind a default-off feature flag.
- Reviewer and fixer prompts live in [`docs/prompts/`](docs/prompts/).
- The Go impact gate and committed CodeGraph review bundle are described in [`docs/pr-workflow.md`](docs/pr-workflow.md). Run `scripts/pr-impact.ps1` before review when code changes.

Local Docker isolation, pinned images and smoke scripts are documented in [`docs/local-dev.md`](docs/local-dev.md) and [`docs/images.md`](docs/images.md).
