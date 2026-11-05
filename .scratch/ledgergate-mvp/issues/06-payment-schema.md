# 06: Payment schema and migration

**What to build:** The payment service owns payment intents, top-ups, refunds, webhook events, PSP settlements and FX rates. The schema applies cleanly, webhook events are unique by provider event ID, and readiness reflects database availability.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] The payment schema matches the frozen schema contract.
- [ ] Applying it to a fresh PostgreSQL 18 database succeeds and reapplying is a no-op.
- [ ] Webhook event uniqueness is enforced.
- [ ] `/readyz` returns 503 when the payment database is unreachable.
- [ ] Integration tests cover migration and readiness.
