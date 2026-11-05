# 12: Compose infrastructure and gateway smoke

**What to build:** A developer can bring up PostgreSQL, Redis, NATS JetStream, Jaeger, Prometheus, Grafana and the gateway with one project-scoped command, using the pinned images and per-checkout `.env.local` ports. The gateway becomes healthy and the smoke check passes. Cleanup is project-scoped.

**Blocked by:** 01, 04

**Status:** ready-for-agent

- [ ] `docker compose config` passes with the checkout's `.env.local`.
- [ ] `make up` starts the infrastructure and gateway and waits for health.
- [ ] The gateway `/healthz` and `/readyz` are reachable on the isolated host ports.
- [ ] Smoke scripts read ports from environment variables.
- [ ] No script uses bare `docker compose down -v` or global prune.
- [ ] Images match the pinned image contract.
