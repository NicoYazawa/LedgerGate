# LedgerGate MVP — Spec

Status: ready-for-agent
Feature: ledgergate-mvp
Date: 2026-11-04

## Problem Statement

AI platform operators who resell access to multiple LLM providers cannot answer, per request, what the customer was charged, what the provider cost, and why. Existing open-source gateways track quota points or spend counters, but not a money-grade ledger: they cannot prove a balance, reproduce a historical bill after a price change, or reconcile provider invoices against customer usage.

The failures cluster at boundaries. A client disconnects mid-stream; a provider returns no usage; a request is retried after a 429; a failover attempt hits a second provider; two concurrent requests both pass a balance check. In those moments, a simple quota counter over-charges, under-charges, or loses money. Finance cannot tell whether a discrepancy is a pricing bug, a provider bill, a payment fee, or an FX difference.

Customers also lack trust. They see a balance but not the ledger behind it, a usage number but not the price snapshot used, or a charge for a stream that was cut off. Disputes become manual investigations.

LedgerGate is a commercial-grade showcase built by one developer in ten weeks. It uses sandbox payments and is not a live business, but it is designed with the invariants of one: every call can be traced to a request, model, usage, price, cost, settlement, and reconciliation outcome.

## Solution

LedgerGate is a six-service platform: a thin OpenAI-compatible gateway, a control plane for accounts and API keys, a money-grade ledger, a billing service for price catalogs and subscriptions, a payment service for sandbox PSPs and FX, and a reconciliation service.

Customers authenticate with workspace-scoped API keys and call an OpenAI-compatible chat completions endpoint. The gateway reserves funds before calling an upstream, streams the response, extracts usage, and records a durable settlement intent. The ledger rates the usage against an immutable price snapshot, writes double-entry postings, and releases the reservation. A failed or empty request costs the customer nothing; a partial stream costs only what was generated; a failover charges the customer once while recording every provider attempt.

Customers fund a USD-denominated wallet with Stripe test mode (USD) or Alipay sandbox (CNY), optionally subscribe to a monthly plan with included quota, and see per-request breakdowns, ledger entries, and reproducible historical bills. Administrators manage the model catalog, price lists, plans, refunds, and reconciliation adjustments through a React admin surface. Finance runs daily checks against provider cost APIs and PSP settlements, reviews discrepancies, and posts adjustments with an audit trail.

The system is deployed locally with Docker Compose and accepted on a local kind cluster. It uses PostgreSQL, Redis, and NATS JetStream, exposes Prometheus metrics and OpenTelemetry traces, and is tested through its public HTTP seams with real infrastructure and mock upstream providers.

## User Stories

1. As a customer, I want to create an account and a workspace, so that I can isolate usage and billing per project.
2. As a workspace owner, I want to invite members and assign roles, so that my team can manage keys and billing safely.
3. As a workspace owner, I want to create, rotate, and revoke API keys, so that I can control access without sharing credentials.
4. As a workspace owner, I want to restrict an API key to specific models, so that I can limit risk and spend.
5. As a workspace member, I want to see the workspace balance and usage, so that I know when to top up.
6. As a billing viewer, I want to see ledger entries and bills, so that I can reconcile spend.
7. As a customer, I want an OpenAI-compatible chat completions endpoint, so that existing SDKs and tools work without code changes.
8. As a customer, I want streaming responses, so that I receive tokens as they are generated.
9. As a customer, I want tool calling to work across supported providers, so that agent workflows work.
10. As a customer, I want vision input to work on OpenAI-routed models, so that I can send images in prompts.
11. As a customer, I want a model list endpoint and model aliases, so that I can discover and select models.
12. As a customer, I want automatic failover to another provider before the first token, so that transient upstream failures do not break my request.
13. As a customer, I want clear error responses with a request ID, so that I can debug or report failures.
14. As a customer, I want rate-limit headers and 429 responses, so that my client can back off correctly.
15. As a customer, I want a per-workspace concurrent stream limit, so that one workload cannot exhaust my capacity.
16. As a customer, I want to top up in USD with a card in Stripe test mode, so that I can pay internationally.
17. As a customer, I want to top up in CNY through Alipay sandbox, so that I can pay domestically.
18. As a customer, I want to see the CNY price at checkout and the USD credited to my wallet, so that I understand the exchange rate and fee.
19. As a customer, I want to subscribe to a monthly plan with included usage, so that I can budget predictably.
20. As a customer, I want to see my renewal date and remaining included quota, so that I can manage the subscription.
21. As a customer, I want usage beyond my plan quota to draw from my wallet, so that service continues without a hard stop.
22. As a customer, I want unused subscription quota not to roll over, so that pricing is predictable.
23. As a customer, I want a per-request breakdown of input tokens, output tokens, price snapshot, and charge, so that I can verify my bill.
24. As a customer, I want historical bills to be reproducible, so that later price changes do not alter past charges.
25. As a customer, I want retries with the same idempotency key not to double-charge me, so that I can safely retry after a network error.
26. As a customer, I want a partial stream or a mid-stream failure to charge only for generated tokens, so that I do not pay for output I did not receive.
27. As a customer, I want a zero-output or failed-before-first-token request to cost me nothing, so that provider failures are not billed to me.
28. As a customer, I want to request a refund, so that I can recover funds after a duplicate charge or service failure.
29. As a customer, I want refunds to go back to the original payment method when possible, so that I receive the money in the expected place.
30. As a customer, I want a chargeback to create an auditable reversal and a frozen account if my balance goes negative, so that the platform can recover funds.
31. As a customer, I want my account to block new requests while my balance is negative, so that I do not accumulate unpayable debt.
32. As an administrator, I want to manage model metadata and price lists with effective dates, so that pricing changes are controlled and auditable.
33. As an administrator, I want to configure plans and included grants, so that I can launch subscription offerings.
34. As an administrator, I want to disable a provider or model, so that I can mitigate an upstream incident.
35. As an administrator, I want to issue refunds with a reason code, so that disputes are handled consistently.
36. As an administrator, I want to freeze an account, so that fraud or unpaid balances are contained.
37. As an administrator, I want an audit log of sensitive actions, so that every change is attributable.
38. As a finance operator, I want daily reconciliation between provider costs, PSP settlements, and the ledger, so that discrepancies are found quickly.
39. As a finance operator, I want a discrepancy item and adjustment workflow, so that differences are resolved with an audit trail.
40. As a finance operator, I want FX gain and loss entries, so that USD/CNY differences are accounted for.
41. As a finance operator, I want to see margin by model and workspace, so that unprofitable pricing is visible.
42. As a platform operator, I want health and readiness endpoints, so that I can run and monitor the services.
43. As a platform operator, I want traces, metrics, and logs correlated by request ID, so that I can diagnose a single call end to end.
44. As a platform operator, I want the system to keep accepting requests when Redis or NATS is degraded, within documented behavior, so that outages are contained.
45. As a platform operator, I want a one-command local environment, so that I can demo and test the system reliably.
46. As a security reviewer, I want API keys stored as keyed hashes and provider keys encrypted at rest, so that a database leak does not expose credentials.
47. As a security reviewer, I want admin sessions protected with HttpOnly cookies and CSRF tokens, so that the admin surface is not trivially hijackable.
48. As a workspace owner, I want to opt in to short-lived encrypted request/response sampling for debugging, so that I can troubleshoot without making content retention the default.

## Implementation Decisions

### Domain vocabulary

- **Account**: the billing entity. Owns the wallet, subscription, and billing history.
- **Workspace**: the usage-isolation entity inside an account. Owns API keys, rate-limit policy, model access, and usage views.
- **User**: a person with a role in an account or workspace.
- **API Key**: a workspace-scoped credential used for gateway authentication.
- **Wallet**: a USD-denominated balance held in the ledger, with `balance` and `held` amounts.
- **Ledger Account**: a named account in the double-entry chart (wallet liability, platform revenue, provider cost, provider payable, PSP clearing, FX gain/loss, promotional credit).
- **Ledger Entry**: an immutable debit or credit posting.
- **Reservation**: a hold against available funds created before an upstream call.
- **Settlement**: the final rating and posting for one logical request, based on its recorded attempts.
- **Usage Attempt**: one upstream request attempt, including provider, model, token counts, usage source, and provider cost.
- **Price Snapshot**: the immutable cost and sell rates used to rate a settlement.
- **Subscription**: a monthly plan with an included USD credit grant.
- **Grant**: a credit lot with an expiry, consumed before wallet balance.
- **Payment Intent**: a checkout attempt with a PSP.
- **FX Gain/Loss**: the accounting difference between presentment and settlement currency.
- **Reconciliation Item**: a discrepancy between two sources of truth, with a status and optional adjustment.

### Service boundaries and data ownership

The system is a modular set of six independently deployable services in one repository:

| Service | Responsibility | Data ownership |
|---|---|---|
| gateway | OpenAI-compatible ingress, SSE proxy, provider adapters, routing/failover, rate limiting, reservation client, settlement outbox | request log, settlement outbox |
| control | accounts, workspaces, users, API keys, sessions, audit, admin BFF and React UI | identity and audit tables |
| ledger | ledger accounts, wallets, entries, reservations, settlements, usage attempts, credit lots | money tables |
| billing | model price catalog, price snapshots, plans, subscriptions, grant policies | catalog and subscription tables |
| payment | PSP adapters, checkout, webhooks, refunds, chargebacks, FX rates | payment tables |
| reconcile | provider/PSP/ledger comparisons, discrepancy items, adjustments | reconciliation tables |

Rules:

- Each service owns exactly one database. There are no cross-database joins or transactions.
- Services communicate synchronously over HTTP/JSON with OpenAPI contracts and per-service bearer tokens.
- Services communicate asynchronously through NATS JetStream. A transactional outbox is the source of every published event.
- The ledger is the single source of truth for money. No other service writes ledger tables.

### Consistency model

- Reserve is synchronous and happens before the upstream call. If the ledger is unavailable, the gateway rejects the request rather than spending money it cannot hold.
- Settlement is asynchronous: the gateway writes a settlement intent to its local outbox and returns the response; a publisher relays it to NATS; the ledger consumes it and settles idempotently.
- The reservation protects the platform between response and settlement. Reservations expire after 30 minutes; a sweeper releases orphaned holds.
- Downstream events (`usage.recorded`, `payment.succeeded`, `price.snapshot`, `subscription.granted`, `apikey.*`) are at-least-once. Consumers deduplicate by event ID.
- The ledger uses `READ COMMITTED` with atomic conditional updates and a deterministic lock order. Cross-service operations use sagas and compensating actions, not distributed transactions.

### Money path and failure semantics

1. The gateway authenticates the API key from a Redis cache, falling back to the control service on a cache miss.
2. The gateway calls `Reserve` on the ledger with a logical request ID, an estimated cost, and an idempotency key.
3. If available balance is insufficient, the gateway returns 402 and does not call an upstream.
4. The gateway calls the selected upstream and streams SSE to the client, extracting usage and recording each attempt.
5. On client disconnect, the gateway keeps reading the upstream for a short grace period to capture usage; if usage is unavailable, it estimates from the received content and marks the source as estimated.
6. The gateway writes a settlement intent to its outbox and returns the response. The outbox publisher sends `usage.recorded` to NATS.
7. The ledger consumes the event, rates each attempt from the latest price snapshot, writes double-entry postings, consumes grants before wallet balance, and releases the unused hold.
8. If the upstream fails before the first token, the gateway releases the reservation. If all attempts fail, it releases the reservation.
9. Failover is allowed only before the first token. Every attempt is recorded; the user is charged once for the successful attempt; failed attempt costs are recorded as platform cost. There is no mid-stream failover.

Failure rules:

| Scenario | User charge | Provider cost | Ledger action |
|---|---|---|---|
| Successful stream or non-stream | Actual usage | Actual or estimated | Settle |
| Upstream omits usage | Tokenizer estimate, marked estimated | Actual or estimated | Settle, reconcile later |
| Client disconnect | Generated tokens | Actual or estimated | Settle partial |
| Zero output or failure before first token | 0 | Recorded if incurred | Release full hold |
| Upstream failure after first token | Generated tokens | Actual or estimated | Settle partial |
| Failover success | Successful attempt only | All attempts recorded | Settle once |
| All attempts fail | 0 | All attempts recorded | Release full hold |
| Client retry with same idempotency key | No second charge | No second upstream call | Idempotent no-op |

Reservation state machine: `held -> settled`, `held -> released`, `held -> expired`. Each transition happens at most once and is enforced by a unique settlement or release ID.

### Pricing and currency

- All ledger money is integer micro-USD (`1 USD = 1,000,000 µUSD`). No floating point.
- The ledger base currency is USD. The wallet is USD-denominated.
- CNY is a presentment currency. Price lists exist in USD and CNY; checkout records the presentment amount, the PSP-settled USD amount, the FX rate, and the PSP fee.
- The price catalog stores input, output, cache-read, and cache-write rates per model, with `effective_from` and an immutable snapshot per settlement.
- Sell price is derived from provider cost with a per-model multiplier at catalog publication time. Historical settlements always use their snapshot.
- Rates are stored per one million tokens in micro-USD; cost is computed with integer arithmetic and a documented rounding rule.
- There is no per-image, per-request, or multi-currency wallet in the MVP.

### Subscription and grants

- A subscription is monthly, with an included USD credit grant.
- Unused quota does not roll over. Grants expire at the end of the period.
- When the grant is exhausted, usage draws from the wallet. If the wallet is also insufficient, the request is rejected with 402.
- There is no mid-cycle proration, upgrade, or downgrade; changes take effect at the next period.
- Renewal is manual in the MVP; there is no dunning or automatic retry.
- Grants are consumed earliest-expiring first. Settlement consumes grants before wallet balance.
- The MVP supports two to three plan tiers; exact quotas and prices are configuration, not hard-coded invariants.

### Identity and access

- Account is the billing entity; Workspace is the usage-isolation entity; a user can belong to multiple accounts.
- Roles are `owner`, `admin`, `member`, and `viewer`. Only owners and admins manage billing and keys; viewers are read-only.
- API keys are workspace-scoped, high-entropy, prefixed with `lg_`, and shown once at creation.
- API keys are stored as HMAC-SHA256 with a server-side pepper; only the prefix and last four characters are displayed.
- Keys can be rotated, revoked, given an expiry, and restricted to specific models.
- The gateway caches API key validation in Redis; the control service publishes `apikey.created` and `apikey.revoked` events to invalidate the cache. A cache miss falls back to the control service.
- There is no BYOK in the MVP; the platform owns all upstream provider keys.

### Gateway behavior

- The public ingress is OpenAI-compatible. Anthropic is an upstream adapter, not a public protocol.
- Public endpoints are `POST /v1/chat/completions` (streaming and non-streaming) and `GET /v1/models`.
- Supported upstreams are OpenAI, Anthropic, and one OpenAI-compatible provider for failover. The compatible provider is selected during implementation.
- Chat completions supports tool calling across supported providers. Vision is supported only when routing to OpenAI in the MVP; Anthropic vision is deferred.
- Failover is allowed only before the first token. The router uses an ordered candidate list per model alias, health checks, and a circuit breaker.
- Error responses follow the OpenAI error shape and include a request ID. Internal/admin errors use a problem-details shape.
- The gateway does not store prompt or response bodies by default.

### Rate limiting

- One Redis-backed limiter handles scopes `workspace` and `api_key`, with a model-level policy as a later extension.
- Metrics are requests per minute, tokens per minute, and concurrent streams.
- Workspace limits are account-global across API keys; creating more keys cannot bypass them.
- Token limits are optimistic at admission and reconciled at settlement. Streams are never rejected mid-flight for exceeding a token bucket.
- If Redis is unavailable, rate limiting fails open with a local fallback and an alert; money checks remain fail-closed in the ledger.
- Limits are inherited from plan to workspace to key.

### Payments, refunds, chargebacks, and FX

- Two sandbox PSP adapters: Stripe test mode for USD and Alipay sandbox for CNY. There is no production payment processing.
- The payment service defines a `PaymentProvider` interface with checkout, webhook verification, refund, and query operations.
- Webhooks are verified and deduplicated by PSP event ID. A duplicated or out-of-order webhook must not double-credit or double-debit.
- Top-ups credit the wallet through a ledger posting. CNY top-ups record presentment amount, settled USD, FX rate, and fee; the FX difference is posted to a gain/loss account.
- Refunds reverse the original posting; they go to the original PSP when possible, otherwise to wallet credit. The original transaction's FX rate is used.
- Subscription periods are not prorated or refunded; cancellation takes effect at the period end.
- Negative balances are allowed after chargebacks, late settlement, or admin adjustments. A negative balance blocks new requests until it is cleared.
- Chargebacks create a reversal, a PSP fee expense, and an account freeze if the balance goes negative.

### Reconciliation

Four checks run on a schedule:

1. Internal ledger invariants: wallet balance equals the sum of entries; reservations plus settlements equal usage; run continuously with zero tolerance.
2. Provider cost versus provider usage/cost APIs: aggregate by provider, model, and day; run daily at T+1; tolerance is 0.5% or $0.01.
3. PSP settlements versus top-ups and refunds: match by payment intent or PSP trade number; run daily; exact match expected.
4. Subscription grants versus usage: verify unused, expired, and consumed grants per period.

Discrepancies create reconciliation items. Adjustments require an administrator action and produce compensating ledger entries; discrepancies are never written off automatically.

### Data retention and privacy

- The default is metadata only: request ID, workspace, key, model, token counts, usage source, price snapshot, cost, status, latency, error code, and content hash.
- Prompt and response bodies are not stored by default. A workspace can opt in to encrypted 24-hour sampling for debugging; access is audited.
- Ledger entries are append-only and retained with the system. Usage records retain one year hot with archival; audit logs retain one year.
- API keys are stored as keyed hashes. Provider keys are encrypted with AES-256-GCM; the KEK comes from environment or a cloud secrets manager.
- Logs redact authorization headers and secrets.

### API contracts and errors

- Internal and admin APIs are OpenAPI-first; servers and clients are generated.
- The gateway's public contract is OpenAI-compatible.
- Every response carries a request ID. The gateway returns OpenAI-shaped errors; internal/admin APIs return problem details.
- Service-to-service calls use per-service bearer tokens. mTLS is a later upgrade.

### Idempotency and retries

- Clients may send `Idempotency-Key`. Scope is workspace + API key + endpoint. TTL is 24 hours.
- Same key and same request body: an in-flight request returns 409; a completed non-streaming request returns the stored response; a completed streaming request returns 409 with the request ID. Billing remains idempotent.
- Same key and different body: 422.
- Upstream retries are allowed only before the first token, with a bounded attempt count and exponential backoff. Each attempt has a unique attempt number.
- Settlement and release are idempotent by unique settlement/release ID. Payment webhooks are idempotent by PSP event ID.

### Technology stack

- Go 1.27, `net/http` with chi; OpenAPI-first with generated server/client code.
- PostgreSQL 17 with pgx v5, sqlc, and goose migrations.
- Redis 7 for rate limiting and API key cache.
- NATS JetStream for asynchronous events and outbox relay.
- React 19, Vite, shadcn/ui, TanStack Query, and TanStack Table for the admin surface, embedded in the control service.
- Docker Compose for local development; kind for a final Kubernetes acceptance deployment.
- GitHub Actions for lint, tests, integration tests, and builds.
- OpenTelemetry, Prometheus, Grafana, and structured `slog` JSON logs.
- Testcontainers for PostgreSQL, Redis, and NATS; k6 for load; property-based testing for ledger invariants.
- UUIDv7 identifiers; UTC `timestamptz`; integer micro-USD.

### Deployment and operations

- Local topology: six services, one PostgreSQL instance with six databases, Redis, NATS JetStream, Jaeger, Prometheus, Grafana, and an OTel collector.
- Each service exposes `/healthz` and `/readyz`.
- No day-one Kubernetes; the final acceptance uses kind.
- No multi-region, no 99.99 SLA, and no production payment processing.

## Testing Decisions

### Seams

The primary seam is the public HTTP surface:

- gateway ingress: `POST /v1/chat/completions` and `GET /v1/models`
- control/admin API
- PSP webhook endpoints
- internal service HTTP APIs where an integration test must drive a saga

Tests use real PostgreSQL, Redis, and NATS via testcontainers, and a mock upstream HTTP server that can emit recorded SSE fixtures and failure modes. Tests assert external behavior (status, headers, streamed chunks, ledger via API, usage via API), not internal implementation details.

The secondary seam is the ledger HTTP API. Property-based tests drive sequences of `Reserve`, `Release`, and `Settle` to prove invariants that are hard to force through the gateway:

- every ledger transaction balances;
- wallet balance equals the sum of entries;
- available balance never goes below zero without an explicit allowed-negative path;
- a reservation settles or releases at most once;
- duplicate settlement is a no-op.

### What makes a good test

- It drives a public API and asserts observable behavior.
- It uses real infrastructure; the database is not mocked.
- The only external boundary mocked is the upstream LLM provider, and the mock is programmable.
- It is deterministic: fixed clocks, fixed IDs where needed, and explicit fixture timestamps.
- It covers failure boundaries: disconnect, zero output, missing usage, retry, failover, duplicate webhook, Redis/NATS degradation.
- It does not assert private functions, table layouts, or log lines as the primary signal.

### Test matrix

| Acceptance area | Test shape |
|---|---|
| Happy path: key → top-up → stream → usage → settlement | Full stack, mock upstream, Stripe sandbox webhook fixture |
| Concurrent requests cannot overdraw | N concurrent gateway requests against a wallet funded for K; assert exactly K succeed and balance never negative |
| Client disconnect settles generated tokens | Mock upstream stream cut by client; assert partial usage and charge |
| Zero output charges 0 and records provider cost | Mock upstream error before first token; assert release and cost posting |
| Failover charges once and records all attempts | Primary returns 429/500 before first token; secondary succeeds; assert one user charge and two attempt costs |
| Idempotency | Same key/body does not double-charge; same key/different body is 422; in-flight is 409 |
| Subscription and wallet | Grant consumed first, then wallet; expiry does not roll over; overage draws wallet |
| Refund and chargeback | Reverse postings; negative balance freeze; FX difference |
| Reconciliation | Inject a provider or PSP discrepancy; assert a reconciliation item and a balanced adjustment |
| FX | CNY top-up records presentment, settled USD, rate, and fee; refund uses original rate |
| Rate limiting | Workspace limit cannot be bypassed by adding keys; streams are not cut mid-flight |
| Observability | A request can be traced from gateway to settlement by request ID |
| Load | 200 RPS control plane, 500 concurrent streams, settlement P99 under 50 ms, error rate under 0.1% |
| Fault injection | Upstream timeout/429/5xx/mid-stream close, Redis down, NATS down |

### Modules tested

- gateway: ingress, SSE, provider adapters, usage extraction, routing/failover, rate limiting, outbox.
- ledger: reserve, release, settle, balance, entries, grant consumption, idempotency.
- billing: price catalog, snapshots, plans, subscriptions, grants.
- payment: checkout, webhook verification, refunds, chargebacks, FX.
- reconcile: provider/PSP/ledger checks, discrepancy items, adjustments.
- control: identity, API keys, roles, sessions, audit.

### Prior art

This is a greenfield repository. The test harness establishes the prior art: testcontainers helpers, a programmable mock upstream, recorded SSE golden files, a ledger invariant property suite, and k6 scripts. TDD should follow the same seam pattern as features are implemented.

## Out of Scope

- Real production payment processing, merchant onboarding, tax/VAT, invoices, PDF statements, accounting periods, multi-entity accounting, dunning, coupons, and referral programs.
- BYOK (bring your own provider key).
- SSO, SAML, SCIM, and custom roles beyond `owner`, `admin`, `member`, and `viewer`.
- Anthropic-native ingress.
- Embeddings, images, audio, Realtime, Batch, and fine-tuning.
- Anthropic vision in the MVP; text and tool calling only.
- Mid-stream failover.
- Dynamic cost- or latency-based routing, semantic caching, guardrails, and content moderation.
- Provider invoice PDF parsing.
- Stream replay. Billing is idempotent, but streamed responses are not replayed.
- Multi-region deployment, day-one Kubernetes, 99.99% SLA, on-call rotation, data residency, and compliance certifications.
- Mobile UI, full internationalization, dark mode, and a custom design system.
- Storing prompt or response bodies by default.
- Real production payment credentials; sandbox only.
- A customer-facing self-service marketing site or billing portal beyond the admin/workspace UI.

## Further Notes

### Delivery

The project is planned for ten weeks by one developer. The critical path is ledger core, then the gateway reserve/settle path, then reconciliation.

Milestones:

- W1: foundations, six service skeletons, OpenAPI, compose, initial migrations.
- W2–W3: ledger core and invariant tests.
- W4–W5: gateway, streaming, usage extraction, reserve/settle saga, failure paths. W5 is a go/no-go gate.
- W6: control, identity, API keys, sessions, audit.
- W7: billing, price catalog, subscriptions, grants.
- W8: payment, sandbox PSPs, refunds, FX.
- W9: reconciliation, rate limiting, failover hardening, admin UI start.
- W10: admin UI completion, observability, load and fault tests, docs, kind acceptance.

If the W5 gate fails, cut admin pages and PSP breadth first; do not cut the ledger or the high-seam tests that prove its invariants.

### Risks

- The cross-service settlement saga can lose usage if the gateway outbox, NATS, and ledger retries are not durably wired. The reservation expiry sweeper and reconciliation alerts are the backstop.
- Anthropic tool-call translation is the riskiest gateway work. Contract tests with recorded SSE fixtures are required.
- Dual-currency presentment and FX reconciliation add accounting complexity. The single USD wallet keeps it bounded.
- Ten weeks with six services and a full admin surface is aggressive. The W5 gate and the scope cuts above are mandatory.
- Microservice operational overhead is real. Compose for development, static service discovery, and one PostgreSQL instance with separate databases keep it manageable.

### Related documents

- Architecture, container and sequence diagrams: the architecture document in `docs/`.
- PlantUML sources: the diagrams directory in `docs/`; render them with the VS Code PlantUML extension.
- Weekly file forecast and W1 detail: the file plan in `docs/`.
- Milestones, gates and risk register: the delivery plan in `docs/`.

### Configuration values to settle during implementation

- Exact model prices and the per-model markup multiplier.
- Exact plan tiers, included quotas, and prices.
- The third OpenAI-compatible upstream provider.
- Default rate limits per plan and the reservation per-request cap.
- Reconciliation schedule and alert thresholds.
