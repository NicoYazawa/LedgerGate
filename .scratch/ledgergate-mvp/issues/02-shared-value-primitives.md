# 02: Shared value primitives and test helpers

**What to build:** Cross-service value types for money, IDs, time and encryption are available and behave consistently. Money parsing, formatting, rounding and overflow are covered by tests; IDs are time-ordered and prefixed; the clock has a fake for tests; AES-GCM round-trips and rejects tampered ciphertext. The shared test harness can start real dependencies for later slices.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] Money uses integer micro-USD, rejects overflow, and passes parse, format and rounding tests.
- [ ] UUIDv7 generation is time-ordered and prefixed IDs are stable in tests.
- [ ] The fake clock makes time-dependent tests deterministic.
- [ ] AES-GCM round-trips and fails on the wrong key or tampered data.
- [ ] Test helpers can start and stop PostgreSQL, Redis and NATS for integration tests.
- [ ] Shared packages have no global mutable state.
