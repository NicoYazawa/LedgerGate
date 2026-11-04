## Summary

<!-- What does this PR change, and why? One paragraph. -->

## Spec / ticket

<!-- Link the spec section, ticket, or issue. -->

## Scope and non-goals

- In scope:
- Explicitly not in scope:

## Risk tier and size

- [ ] Money / auth / data (soft 300, hard 450)
- [ ] Core (soft 400, hard 600)
- [ ] UI (soft 800, hard 1200)
- [ ] Docs / tooling (soft 800, hard 1200)
- [ ] Default (soft 500, hard 750)

`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/pr-size.ps1 -Base origin/main` output:

```text
<!-- paste the tier, production churn, test churn and status -->
```

- [ ] Under the soft limit
- [ ] Over the soft limit; `size:large` label applied and justified here:

## Feature flag

- Flag name:
- Default: off
- Owner:
- Reason:
- Intended removal date:
- [ ] Both states are tested

## Contract impact

- Public API / OpenAPI:
- Events:
- Database / migrations:
- Config / environment:

- [ ] OpenAPI changes are additive
- [ ] Event changes are versioned or optional
- [ ] Migration follows expand–contract
- [ ] Config changes are optional with safe defaults

## Tests and evidence

- [ ] Unit tests
- [ ] Integration tests
- [ ] Race detector
- [ ] Failure-path test
- [ ] Flag on/off test
- [ ] Manual verification (commands or screenshots)

Evidence:

```text
<!-- commands and results -->
```

## Rollback

- How to disable or roll back:
- Is the migration backward-compatible?
- What data written during the rollback window needs reconciliation?
- [ ] Rollback procedure verified or documented

## Checklist

- [ ] `go build ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `golangci-lint run`
- [ ] Integration tests pass
- [ ] OpenAPI diff is additive
- [ ] Migrations apply to a fresh PostgreSQL 18 database
- [ ] `scripts/pr-size.ps1` passes
- [ ] No secrets, `.env`, database volumes or build output committed
- [ ] Logs, metrics and traces exist for new paths
- [ ] No unrelated changes or drive-by refactors

## Reviewer notes

<!-- Reviewer fills this in. Findings are ranked Blocker / High / Medium / Low with file:line and evidence. -->
