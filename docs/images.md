# Container Images

Status: normative for local development and W1.
Last updated: 2026-11-04.

All required images are already present in the local Docker engine. Do not use `latest`; use these exact tags.

| Purpose | Image | Local |
|---|---|---|
| PostgreSQL | `postgres:18-alpine` | yes |
| Redis | `redis:7.4-alpine` | yes |
| NATS JetStream | `nats:2.10-alpine` | yes |
| Jaeger with OTLP | `jaegertracing/all-in-one:1.60` | yes |
| Prometheus | `prom/prometheus:v3.1.0` | yes |
| Grafana | `grafana/grafana:11.3.0` | yes |
| Go builder | `golang:1.27-alpine` | yes |
| Runtime base | `alpine:3.23.5` | yes |
| Frontend builder (W6) | `node:24-alpine` | yes |
| Testcontainers Ryuk | `testcontainers/ryuk:0.14.0` | yes |

## Rules

- Pin every tag exactly as listed. No `latest` in `docker-compose.yml`, `Dockerfile`, testcontainers or CI.
- OTel collector is **not** part of the local stack. Services send OTLP to Jaeger at `jaeger:4317` (gRPC) or `jaeger:4318` (HTTP).
- Testcontainers must use the same PostgreSQL, Redis, NATS and Ryuk images. Set `RYUK_CONTAINER_IMAGE=testcontainers/ryuk:0.14.0` if the library does not pick it up.
- If a new service needs another image, add it to this table first.
- No downloads are required for W1.

## Why versions changed

- PostgreSQL was specified as 17. The local engine has 18 and 16, not 17. The schema uses only standard SQL and is compatible with 18, so the pinned version is 18 to avoid a download.
- Redis is pinned to the local 7.4 minor instead of the floating `7` tag.
- Jaeger all-in-one 1.60 is the local pinned trace backend and exposes OTLP directly, so the local stack does not need a separate OTel collector.
