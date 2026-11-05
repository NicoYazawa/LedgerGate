# 04: Database and migration runner with gateway and control schemas

**What to build:** Gateway and control services own their databases. A shared migration runner applies the exact schemas, is idempotent, and is exposed through a migrate subcommand. Readiness flips to 503 when the database is unreachable. Integration tests run against a real PostgreSQL 18 container.

**Blocked by:** 03

**Status:** ready-for-agent

- [ ] Gateway and control schemas apply to a fresh PostgreSQL 18 database.
- [ ] Running migrations twice is a no-op.
- [ ] `/readyz` returns 503 when the database is unreachable.
- [ ] Migration output and errors are observable.
- [ ] Testcontainers integration tests cover apply and idempotency.
