# 05: Ledger schema and migration

**What to build:** The ledger service owns the exact money schema: wallets, double-entry entries, reservations, settlements, usage attempts, price snapshots and credit lots. The schema applies cleanly, preserves the required constraints and indexes, and readiness reflects database availability.

**Blocked by:** 04

**Status:** ready-for-agent

- [ ] The ledger schema matches the frozen schema contract.
- [ ] Applying it to a fresh PostgreSQL 18 database succeeds.
- [ ] Reapplying it is a no-op.
- [ ] Required constraints and unique indexes exist.
- [ ] `/readyz` returns 503 when the ledger database is unreachable.
- [ ] Integration tests cover migration and readiness.
