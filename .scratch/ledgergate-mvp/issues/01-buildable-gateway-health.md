# 01: Buildable gateway with health and config

**What to build:** A developer can build the Go module, start the gateway, and see it report healthy. Configuration loads from environment and optional YAML, fails fast on missing required values, and the gateway exposes health, readiness and Prometheus metrics. Feature flags are available with a default-off behavior and both states are testable.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] `make build`, `make test` and `make lint` pass.
- [ ] Starting the gateway without required config fails with a clear error naming the missing field.
- [ ] `/healthz` returns 200, `/readyz` returns 200 and `/metrics` is served.
- [ ] Logs are JSON and carry `service`, `request_id` and `trace_id`.
- [ ] A feature flag defaults off, and both on and off behavior are tested.
- [ ] SIGTERM drains and exits within the configured shutdown timeout.
