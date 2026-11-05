# 14: CI and W1 milestone gate

**What to build:** Continuous integration runs the required checks, and the full W1 milestone gate passes from a clean checkout. Documentation explains how to run and verify the system.

**Blocked by:** 13

**Status:** ready-for-agent

- [ ] CI runs lint, unit tests, build and compose config.
- [ ] `make generate`, `make lint`, `make test`, `make build`, `make up`, `make migrate` and `make smoke` pass.
- [ ] `docker compose config` passes with `.env.local`.
- [ ] The W1 acceptance checklist is green.
- [ ] README and the service decomposition ADR exist.
- [ ] No secrets, `.env` files or database volumes are committed.
- [ ] Feature flags default off and both states are tested.
