# 13: Full six-service Compose and observability

**What to build:** All six services run together in the isolated Compose stack. Readiness checks their dependencies, a request produces a trace in Jaeger, metrics are scraped by Prometheus, Grafana has datasources, and smoke covers all six services.

**Blocked by:** 02, 03, 05, 06, 07, 09, 10, 11, 12

**Status:** ready-for-agent

- [ ] All six services start and become healthy in Compose.
- [ ] `/readyz` returns 503 when PostgreSQL, Redis or NATS is unavailable for the relevant service.
- [ ] A request produces a Jaeger trace.
- [ ] Prometheus scrapes all six services.
- [ ] Grafana has Prometheus and Jaeger datasources.
- [ ] Smoke passes for all six services.
- [ ] Logs carry `service`, `request_id` and `trace_id`.
