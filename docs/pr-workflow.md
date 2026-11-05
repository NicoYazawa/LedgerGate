# Pull Request Workflow

Status: normative for all development.
Last updated: 2026-11-04.

This document replaces the earlier two-worktree A/B workflow. Every change lands through a small pull request that is buildable, tested, independently reviewed, and rollback-ready.

## Principles

- **Trunk-based**: short-lived branches, merge to `main` frequently. `main` is always green.
- **Small vertical slices**: one PR delivers one demoable behavior, not one layer.
- **Risk-based sizing**: line budgets depend on what the PR touches. Money code gets the tightest budget.
- **Independent review**: the reviewer has a fresh context and does not edit the code.
- **Evidence over opinion**: a high-risk finding needs a scenario, a repro, or a failing test.
- **Flags for unfinished work**: incomplete user-visible behavior is behind a default-off flag.
- **Rollback is part of the change**: if it cannot be rolled back, it is not ready.
- **No duplicate implementations**: one branch, one PR. The A/B worktree experiment is retired.

## Roles

| Role | Who | May edit code? | Output |
|---|---|---|---|
| Implementer | developer or implementation agent | yes | Branch and PR |
| Reviewer | independent agent with fresh context | **no** | Severity-ranked findings and a verdict |
| Fixer | implementer or a separate fixer agent | yes | Regression test plus minimal fix |
| Human owner | project owner | merge | Final call on blockers and high-risk items |

The reviewer and the fixer must be different sessions. The reviewer does not approve its own fix.

## PR lifecycle

1. Take one ticket from `to-tickets`.
2. Create a short-lived branch: `feat/<ticket>`, `fix/<ticket>`, or `refactor/<ticket>`.
3. Implement test-first where practical. Add a failing test for the behavior or bug.
4. At stage completion, add the stage PlantUML source and rendered SVG under `docs/diagrams/stages/`.
5. Run the local gate and `scripts/pr-size.ps1`.
6. Open a PR with the template in `.github/pull_request_template.md`.
7. CI runs. Do not request review while CI is red.
8. An independent reviewer runs the review prompt in `docs/prompts/pr-reviewer.md`.
9. Blocker/high findings: add a regression test and a minimal fix, then re-review.
10. Merge with squash after CI and review pass.
11. Delete the branch. Remove feature flags in a follow-up PR once the behavior is stable.

## Risk tiers and size budgets

Count **hand-written production churn** (added + deleted lines). Exclude tests, generated code, vendor, lockfiles, build output, CodeGraph `graph.html` and rendered diagram SVGs. The script `scripts/pr-size.ps1` enforces this.

| Tier | Paths | Soft limit | Hard limit |
|---|---|---|---|
| Money / auth / data | `internal/ledger/`, `internal/billing/`, `internal/payment/`, `internal/platform/auth/`, `internal/platform/crypt/`, ledger/payment/control migrations, ledger/payment OpenAPI | 300 | 450 |
| Core | `internal/gateway/`, `internal/provider/`, `internal/control/`, `internal/reconcile/`, other `internal/platform/`, other migrations, API contracts, root build/config (`go.mod`, `Makefile`, `docker-compose.yml`, `deploy/`, `.env.example`) | 400 | 600 |
| UI | `web/` | 800 | 1200 |
| Docs / tooling | `docs/`, `scripts/`, `.github/`, `.scratch/`, `README.md`, `AGENTS.md`, `.gitignore` | 800 | 1200 |
| Default | everything else | 500 | 750 |

Rules:

- If a PR touches more than one tier, apply the strictest tier.
- Over the soft limit: warn. Over the hard limit: fail.
- `scripts/pr-size.ps1 -Enforce` treats the soft limit as a failure; use it in CI.
- A PR over the soft limit needs a `size:large` label and a short justification in the PR body.
- Do not game the limit by moving code into generated files, vendor, or unrelated PRs.
- `scripts/pr-size.ps1` compares committed changes only. Commit the branch before running it; untracked files are not counted.
- One logical change per PR. Prefactoring is a separate PR.
- Migrations follow expand–contract: schema expansion, data backfill, code switch, contract removal. Do not mix a destructive migration with a feature.

## Definition of done per PR

A PR is ready for review only when all of these are true:

- Builds, lints and passes the existing tests.
- New behavior has tests; failure paths have tests; flag on/off has tests.
- The completed stage has a PlantUML source, a rendered SVG, and both are committed together.
- `go test -race ./...` passes for code that can race.
- Integration tests pass with real PostgreSQL, Redis and NATS via testcontainers.
- OpenAPI changes are additive; no frozen operation is removed or renamed.
- Event schema changes are versioned and consumers tolerate unknown fields.
- Database changes are expand–contract and backward-compatible.
- Config changes are additive with safe defaults.
- New paths emit logs, metrics and traces.
- A rollback plan exists: flag off, forward-compatible migration, or documented runbook.
- `scripts/pr-size.ps1` passes or the `size:large` label is present with justification.
- No secrets, `.env` files, database volumes, or build output are committed.

## Stage diagrams

Every completed implementation stage or ticket produces a durable PlantUML record:

- Source: `docs/diagrams/stages/NN-<ticket-slug>.puml`
- Render: `docs/diagrams/stages/NN-<ticket-slug>.svg`
- The `@startuml` identifier must match the filename without `.puml`, because the PlantUML CLI uses it as the output name.
- Commit the source and SVG in the same PR as the completed stage.
- Render all diagram sources with `make diagrams`. To render selected files, run `scripts/render-plantuml.ps1 -Source <path>`.

The diagram records the implemented boundary, runtime dependencies, planned work and known gaps at stage completion. Keep it focused on the stage; do not duplicate the full target architecture in every diagram.

Completed stages that change code also commit a CodeGraph review bundle:

- `docs/codegraph/<stage>/graph.mmd`
- `docs/codegraph/<stage>/graph.locs.json`
- `docs/codegraph/<stage>/graph.html`

The agent reads the code and supplies the semantic topology and source locations. CI validates the bundle; it does not invent topology automatically. The generated HTML is self-contained and supports click-to-source.

## Impact gate

When a PR changes code files, `scripts/pr-impact.ps1` runs the Go command at `cmd/pr-impact` and writes transient artifacts to the gitignored `artifacts/pr-impact/` directory:

| Artifact | Source | Purpose |
|---|---|---|
| `impact.mmd` | `AnalyzePackages` over `go list` | Changed packages and transitive reverse dependents |
| `check.json` | Go impact runner | Predicates, counts and CodeGraph validation result |
| `summary.md` | Go impact runner | Sticky PR comment body |

Run locally:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File scripts/pr-impact.ps1 `
  -Base origin/main `
  -Stage <stage-slug>
```

The stage defaults to the current branch name with its `feat/`, `fix/`, `refactor/` or `docs/` prefix removed. Pass `-Stage` explicitly in CI. The script compares committed changes only; commit before running it.

Policy:

- CI starts **advisory**. Set the repository variable `CODEGRAPH_ENFORCE=true` to make the gate required.
- Hard gate: invalid Go package graph, blast radius above the configured threshold, exported declaration changes without a `signature-change` waiver, or a missing/invalid CodeGraph bundle.
- To waive the signature predicate, add the `signature-change` label to the PR and justify it in the PR body. The reviewer must still approve the change.
- Go compilation rejects import cycles; the package graph additionally reports reverse-dependency blast radius.
- CI pins `@longbowzz/dsh-codegraph@0.4.0` only for `validate.mjs` and `verify-html.mjs`.
- Docs-only PRs skip the gate; the risk tiers and review still apply.
- A green gate does not replace tests or review. It validates package relationships, declaration lines and graph source locations; it cannot prove money invariants, SQL semantics, OpenAPI compatibility or runtime behavior.

`mermaidiff` is an optional **reviewer aid**, not part of the gate. It produces a semantic flow brief and stores ignored output under `.mermaidiff/`.

## Review

The reviewer is independent, read-only, and adversarial. It runs the tests and reproduces findings.

Severity:

| Severity | Meaning | Examples |
|---|---|---|
| Blocker | Must be fixed before merge | Money is charged wrong, data loss, auth bypass, secret leak, irreversible migration, frozen API break |
| High | Must be fixed or explicitly waived | Race, missing idempotency, swallowed error, unbounded resource, rollback failure, missing regression test for a money path |
| Medium | Should be fixed, can be deferred with a ticket | Boundary condition, missing metric, weak test, confusing error |
| Low / Nit | Optional | Naming, duplication, readability |

A high or blocker finding must include:

- File and line.
- The scenario that triggers it.
- A repro command, a failing test, or a concrete trace/log.
- A suggested fix or a clear direction.

Review dimensions:

1. Correctness against the spec and acceptance criteria.
2. Money and data integrity: reservations, settlements, idempotency, double charge, lost usage.
3. Concurrency: races, deadlocks, lock order, atomic updates.
4. Security: auth, key handling, secrets, injection, PII.
5. Contracts: OpenAPI, events, DB migrations, config compatibility.
6. Tests: behavior-level, failure paths, determinism, no sleeps.
7. Operability: health/readiness, logs, metrics, traces, graceful shutdown.
8. Rollback and feature flags.
9. Size and scope discipline.

Verdict: `approve`, `request changes`, or `needs design/refactor`.

## High-risk handling

Default response to a blocker/high finding:

1. Write a failing regression test that reproduces the risk.
2. Apply the smallest fix that makes the test pass.
3. Re-run the full gate.
4. Re-review the diff. Do not let the fixer approve their own change.

Rewrite is a last resort, not the default. Use a clean-room rewrite only when the design itself is wrong, not when a function has a bug.

A rewrite must:

- Preserve public function signatures, interfaces, API contracts and observable external behavior.
- **Not** preserve the buggy behavior that caused the high-risk finding.
- Start from the spec, contracts, failing tests and review findings, not from the original implementation.
- Be reviewed as a new PR with the same severity bar.

If two attempts both produce blocker/high findings:

1. Stop merging.
2. Decide whether the root cause is an ambiguous spec, a missing test oracle, or a flawed design.
3. Write or update the ADR/spec and add characterization tests.
4. Refactor in a separate behavior-preserving PR.
5. Re-implement the feature as a normal PR.

Do not refactor under time pressure without characterization tests.

## Feature flags

- Incomplete user-visible behavior ships behind a default-off flag.
- The simplest implementation is an environment/config flag: `LG_FEATURE_<UPPER_SNAKE>=true|false`.
- Flags are evaluated server-side. The default is always off.
- Every flag has an owner, a reason, and an intended removal date in the PR body.
- Test both flag states.
- A flag must not change money semantics silently. Price and billing-rule changes use price-catalog versions and effective dates, not flags.
- A flag can be a kill switch for a risky path; document the off behavior.
- Remove the flag in a follow-up PR once the feature is stable. Flag age is a tracked metric.
- Do not nest flags or use a flag to avoid writing a test.

## Non-breaking and rollback

| Surface | Rule |
|---|---|
| Public API | Additive only. Do not remove or rename a frozen operation. Run an OpenAPI diff. |
| Events | Add fields as optional; bump the event version for semantic changes; consumers ignore unknown fields. |
| Database | Expand–contract. Add columns/tables, backfill, switch reads/writes, remove old shape in a later PR. |
| Config | New env vars are optional with safe defaults. Do not require a coordinated deploy just to start. |
| Auth | Key format and validation changes require a migration path for existing keys. |

Rollback checklist:

- Feature flag can turn the behavior off.
- Database migration is backward-compatible with the previous code.
- No destructive migration in the same PR as the feature.
- Rollback procedure is written in the PR.
- Data written during the rollback window is either compatible or reconciled.

## Merge and release

- Squash merge to `main`; delete the branch.
- `main` must be green after merge.
- Tag releases when deploying; keep the commit hash in the deployment record.
- Watch logs, metrics and reconciliation for one deployment cycle before removing a flag.
- A rollback is a normal deploy of the previous tag plus flag-off, not a history rewrite.

## Metrics

Track these per iteration:

- PR size distribution by risk tier.
- Percentage of PRs with blocker/high findings.
- Review cycle time.
- Open feature flags and their age.
- Rollbacks and incidents.
- Reconciliation difference rate.

## Tooling

- `scripts/pr-size.ps1` — risk-tier size check.
- `scripts/pr-impact.ps1` — Go package impact, export guard and CodeGraph bundle gate.
- `scripts/post-impact-comment.ps1` — sticky PR comment for the impact summary.
- `.github/workflows/impact.yml` — PR impact job.
- `.github/pull_request_template.md` — PR contract.
- `docs/prompts/pr-reviewer.md` — independent reviewer prompt.
- `docs/prompts/pr-fixer.md` — regression-test-first fixer prompt.
- `mattpocock-skills:to-tickets` — vertical slices.
- `mattpocock-skills:tdd` — red-green-refactor and regression tests.
- `mattpocock-skills:code-review` — Standards + Spec review.
- `mattpocock-skills:pr` — PR body.
- `mattpocock-skills:diagnosing-bugs` — hard bugs and performance regressions.

## CI checks to require once the first code PR lands

- `go build ./...`
- `go test ./...`
- `go test -race ./...`
- `golangci-lint run`
- integration tests with testcontainers
- OpenAPI diff against `main`
- migration apply on a fresh PostgreSQL 18 database
- `pwsh scripts/pr-size.ps1 -Base origin/main -Enforce`
- `scripts/pr-impact.ps1` when code files change (advisory until `CODEGRAPH_ENFORCE=true`; then required)
- secret scanning
- `docker compose config` when compose files change

A PR with a failing required check is not reviewable.
