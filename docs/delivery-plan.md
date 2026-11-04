# LedgerGate Delivery Plan

Status: forecast, inputs to the spec and tickets.
Last updated: 2026-11-04.

This is a milestone plan, not a file-by-file schedule. The weekly code content is produced by `to-tickets` as vertical slices; this document only fixes the outcomes, gates and critical path.

## Capacity assumptions

- Solo developer, 10 weeks.
- Four productive development days per week plus one buffer day.
- 20–30% contingency for integration, provider quirks and rework.
- Re-plan at the W5 go/no-go.

## Milestones

| Week | Outcome | Gate |
|---|---|---|
| W1 | Six service skeletons boot; `docker compose up` brings up Postgres, Redis, NATS, Jaeger, Prometheus, Grafana; OpenAPI + codegen wired; platform packages (config, slog/OTel, pgx, outbox, bearer auth); health endpoints. | Compose is green; CI runs lint, tests and build. |
| W2–W3 | Ledger core: wallets, double-entry accounts and entries, reservations, settlements, credit lots; `Reserve`/`Release`/`Settle` APIs with idempotency; property tests for invariants. | Concurrent reserve test cannot overdraw; balance equals sum of entries; duplicate settle is a no-op. |
| W4–W5 | Gateway and money path: OpenAI non-stream and stream; Anthropic text + tool calling; OpenAI-compatible ingress; usage extraction; reserve before upstream; outbox → NATS → ledger settle; zero-output, disconnect, failover and all-fail paths. | End-to-end stream settles correctly; concurrency, disconnect and failover tests pass. **W5 go/no-go.** |
| W6 | Control: accounts, workspaces, users, API keys; hashed keys; Redis cache + NATS invalidation; admin auth (Argon2id + session); audit log; admin BFF endpoints. | Key revocation reaches the gateway within the configured TTL; RBAC tests pass. |
| W7 | Billing: USD/CNY price catalog with effective dates; price snapshots published to ledger; plans, subscriptions and grants; FIFO grant consumption; overage to wallet. | Subscription + wallet end-to-end; historical bills reproduce from snapshots. |
| W8 | Payment: Stripe test mode + Alipay sandbox; checkout, webhook, refund; FX capture; ledger credit/debit; webhook idempotency. | USD and CNY top-ups; refund and chargeback simulation; FX entries balance. |
| W9 | Reconcile, rate limiting, failover hardening, admin UI pages 1–5: provider/PSP/ledger jobs; discrepancy items; RPM/TPM/concurrency limits; circuit breaker. | Injected discrepancy is detected; rate-limit and failover tests pass. |
| W10 | Remaining admin pages; Grafana dashboards; k6 load test; fault injection; docs (README, spec, ADRs, runbook, OpenAPI); kind acceptance deployment. | Acceptance table green; load/fault reports produced; demo script works. |

## Critical path

`ledger core -> gateway reserve/settle -> money path failure semantics -> reconciliation`

The critical path is money, not UI. If the W5 gate fails:

1. Cut admin pages (8 → 5) and UI polish.
2. Cut Alipay breadth (keep Stripe test mode and a mock adapter).
3. Keep the ledger, reserve/settle, outbox and gateway paths intact.

Do not cut the ledger invariants or the high-seam tests that prove them.

## Go/no-go at W5

Go if all of the following are true:

- A streaming request reserves, proxies, records usage and settles end-to-end.
- Concurrent requests cannot overdraw a wallet.
- Client disconnect settles generated tokens.
- Zero output charges the user 0 and records provider cost.
- Failover before the first token charges the user once and records all attempts.
- Duplicate settlement is idempotent.

No-go means the milestone plan is re-cut before W6; it does not mean the project is abandoned.

## Ticket generation

After the spec exists, run `to-tickets` to produce tracer-bullet vertical slices with blocking edges. Assign those tickets to weeks by capacity; do not pre-write a file-level schedule here. Each ticket should be demoable or verifiable on its own and small enough for a fresh context window.

## Risk register

| Risk | Impact | Mitigation |
|---|---|---|
| Cross-service settle saga loses usage | Revenue loss / wrong bill | Gateway outbox + durable NATS + ledger idempotency; reservation expiry sweeper; reconciliation alerts |
| Anthropic tool-call translation | Gateway schedule | Contract tests with recorded SSE fixtures; keep translation narrow |
| Dual-currency/FX reconciliation | Ledger complexity | Single USD wallet; record presentment/settled/FX; daily PSP reconciliation |
| 10-week full-scope schedule | Quality / unfinished admin | W5 gate; cut UI and PSP breadth first |
| Microservice operational overhead | Slow delivery | Compose for development; static service discovery; shared Postgres instance with separate databases |
| Provider API changes | Gateway breakage | Thin adapters; contract tests; pinned versions |
