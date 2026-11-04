# Local Development and Docker Validation

Status: normative for local development and agent validation.
Last updated: 2026-11-04.

## Per-worktree `.env.local`

Every worktree gets its own gitignored `.env.local`. It is never committed. It sets:

- `COMPOSE_PROJECT_NAME` — isolates containers, networks and volumes.
- every host port used by Compose.

Generate it with:

```powershell
./scripts/worktree-env.ps1 -Name ledgergate-w1-alpha -Offset 10000
./scripts/worktree-env.ps1 -Name ledgergate-w1-beta  -Offset 20000
```

## Port map

| Variable | Container port | Base | Alpha (+10000) | Beta (+20000) |
|---|---|---|---|---|
| `GATEWAY_PORT` | 8080 | 8080 | 18080 | 28080 |
| `CONTROL_PORT` | 8081 | 8081 | 18081 | 28081 |
| `LEDGER_PORT` | 8082 | 8082 | 18082 | 28082 |
| `BILLING_PORT` | 8083 | 8083 | 18083 | 28083 |
| `PAYMENT_PORT` | 8084 | 8084 | 18084 | 28084 |
| `RECONCILE_PORT` | 8085 | 8085 | 18085 | 28085 |
| `POSTGRES_PORT` | 5432 | 5432 | 15432 | 25432 |
| `REDIS_PORT` | 6379 | 6379 | 16379 | 26379 |
| `NATS_PORT` | 4222 | 4222 | 14222 | 24222 |
| `NATS_MONITOR_PORT` | 8222 | 8222 | 18222 | 28222 |
| `JAEGER_UI_PORT` | 16686 | 16686 | 26686 | 36686 |
| `OTEL_GRPC_PORT` | 4317 | 4317 | 14317 | 24317 |
| `OTEL_HTTP_PORT` | 4318 | 4318 | 14318 | 24318 |
| `PROMETHEUS_PORT` | 9090 | 9090 | 19090 | 29090 |
| `GRAFANA_PORT` | 3000 | 3000 | 13000 | 23000 |

Container-internal communication always uses service names and container ports (`postgres:5432`, `ledger:8082`, `jaeger:4317`). Host ports are only for smoke tests and human access.

## Compose rules

```yaml
services:
  gateway:
    ports:
      - "${GATEWAY_PORT:-8080}:8080"
  postgres:
    ports:
      - "${POSTGRES_PORT:-5432}:5432"
```

Make targets must always scope the project:

```makefile
COMPOSE = docker compose -p $(COMPOSE_PROJECT_NAME) --env-file .env.local

up:
	$(COMPOSE) up -d --wait

down:
	$(COMPOSE) down -v --remove-orphans
```

Rules:

- Never run bare `docker compose down -v`; it can hit another worktree's project.
- Never run `docker system prune` or `docker volume prune`.
- Never kill containers you did not start.
- If a host port is already in use, stop and report; do not change ports silently.
- Use `--wait` and healthchecks, not `sleep`.

## Tests versus Compose

- `make test` and `make test-integration` must not start Compose.
- Integration tests use `testcontainers-go`, which allocates random host ports and cleans up after itself. Two worktrees can run tests in parallel.
- Compose is for acceptance and demos only.

## Smoke scripts

`scripts/smoke.ps1` and `scripts/smoke.sh` read the same `*_PORT` variables:

```powershell
$gateway = "http://localhost:$env:GATEWAY_PORT/healthz"
```

No hardcoded 8080.

## Cleanup

```powershell
docker compose -p $env:COMPOSE_PROJECT_NAME down -v --remove-orphans
```

Each worktree removes only its own project. Check with:

```powershell
docker compose ls
docker ps --filter "label=com.docker.compose.project"
```

## Troubleshooting

| Symptom | Action |
|---|---|
| `port is already allocated` | Stop; check `docker ps`; do not kill other projects. Report the conflict. |
| Compose sees another worktree's containers | You omitted `-p $COMPOSE_PROJECT_NAME`. Fix the command. |
| Volumes look stale | `docker compose -p $COMPOSE_PROJECT_NAME down -v --remove-orphans` for your project only. |
| Two full stacks are slow | Run one at a time, or use a Compose profile that skips Jaeger/Prometheus/Grafana for intermediate checks. |
