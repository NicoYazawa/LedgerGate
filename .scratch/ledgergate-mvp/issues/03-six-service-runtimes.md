# 03: Six service runtimes with health and readiness

**What to build:** All six services build as separate entrypoints and start with their own configuration, logging, metrics and health endpoints. Each service answers `/healthz` and `/readyz`, and the build produces all six binaries. No business logic is implemented yet.

**Blocked by:** 01

**Status:** ready-for-agent

- [ ] The build produces six binaries.
- [ ] Each service starts with its contract port and configuration.
- [ ] All six expose `/healthz`, `/readyz` and `/metrics`.
- [ ] Missing required config fails fast for each service.
- [ ] SIGTERM drains and exits within the configured timeout for each service.
