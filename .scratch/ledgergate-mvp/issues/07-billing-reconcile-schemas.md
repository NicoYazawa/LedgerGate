# 07: Billing and reconcile schemas

**What to build:** Billing owns catalog, snapshots, plans, subscriptions and grant policies; reconcile owns runs, items and adjustments. Both schemas apply cleanly, enforce the required uniqueness, and readiness reflects database availability.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] Billing and reconcile schemas match the frozen schema contract.
- [ ] Both apply to a fresh PostgreSQL 18 database and reapply as no-ops.
- [ ] Active subscription uniqueness and reconciliation constraints are enforced.
- [ ] `/readyz` returns 503 when either database is unreachable.
- [ ] Integration tests cover both migrations and readiness.
