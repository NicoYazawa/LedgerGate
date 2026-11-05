# CodeGraph Impact Gate Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the `@optave/codegraph` CLI and replace its PR gate with a deterministic Go impact analyzer plus validation of committed `longbowzz/CodeGraph` artifacts.

**Architecture:** A thin PowerShell wrapper invokes a Go command. The Go analyzer compares changed Go files against a `go list` package graph, reports reverse dependencies and blast radius, detects exported declaration changes, and validates the stage's committed CodeGraph bundle with the versioned CodeGraph skill scripts. CI installs only the pinned CodeGraph npm distribution and Go toolchain.

**Tech Stack:** Go 1.27 standard library, PowerShell, Node.js 24, Mermaid 11.12 CodeGraph validator, GitHub Actions.

**Spec:** The approved replacement plan in the Work thread dated 2026-11-05.

**Status:** Implemented and verified across the stacked refactor branches.

## Global Constraints

- No invocation or installation of `@optave/codegraph` remains.
- CodeGraph output is committed under `docs/codegraph/<stage>/`.
- Each completed stage also has a PlantUML source and rendered SVG.
- Gate failures are advisory by default and blocking when `-Enforce` is set or `CODEGRAPH_ENFORCE=true`.
- The analyzer works from committed changes between `<base>...HEAD`.
- No generated graph database is committed.

---

### Task 1: Go Package Impact Analyzer

**Files:**

- Create: `internal/tooling/primpact/packages.go`
- Create: `internal/tooling/primpact/packages_test.go`

**Interfaces:**

- Produces: `AnalyzePackages(ctx context.Context, repo string, changed []string, maxDepth, limit int) (PackageImpact, error)`
- Produces: `PackageImpact{ChangedPackages []string, Dependents []Dependent, MaxDepth int, Failed bool}`

- [ ] Write a failing test for a changed leaf package with one direct and one transitive dependent.
- [ ] Run `go test ./internal/tooling/primpact -run TestAnalyzePackagesFindsTransitiveDependents`; expect undefined API.
- [ ] Implement `go list -f '{{json .}}' ./...` parsing and reverse dependency traversal.
- [ ] Add and fail a test for a blast-radius count above the configured limit.
- [ ] Implement threshold enforcement and Mermaid generation.
- [ ] Run package tests and commit.

### Task 2: Exported Declaration Guard

**Files:**

- Create: `internal/tooling/primpact/signatures.go`
- Create: `internal/tooling/primpact/signatures_test.go`

**Interfaces:**

- Produces: `FindBreakingExports(diff string) []string`

- [ ] Add tests for removed/changed exported functions, methods and types.
- [ ] Add a passing case for added exports and moved unchanged declarations.
- [ ] Verify tests fail before implementation.
- [ ] Parse zero-context unified diff declaration lines and compare removed against added declarations.
- [ ] Run package tests and commit.

### Task 3: CodeGraph Bundle Validation

**Files:**

- Create: `internal/tooling/primpact/codegraph.go`
- Create: `internal/tooling/primpact/codegraph_test.go`

**Interfaces:**

- Produces: `ValidateCodeGraph(ctx context.Context, opts CodeGraphOptions) (CodeGraphResult, error)`
- Produces stage slug normalization from branch names such as `feat/02-shared-value-primitives`.

- [ ] Add a failing test that a code change with no bundle fails.
- [ ] Add a passing test using a real Mermaid/location fixture and the installed CodeGraph validator.
- [ ] Add a failing test for a location outside the repository.
- [ ] Implement bundle existence checks, `validate.mjs`, and `verify-html.mjs` invocation.
- [ ] Run package tests and commit.

### Task 4: CLI Report and PowerShell Wrapper

**Files:**

- Create: `cmd/pr-impact/main.go`
- Create: `internal/tooling/primpact/run.go`
- Create: `internal/tooling/primpact/run_test.go`
- Modify: `scripts/pr-impact.ps1`

**Interfaces:**

- Produces: `Run(ctx context.Context, opts Options) (Report, error)`
- PowerShell command remains: `scripts/pr-impact.ps1 -Base origin/main`

- [ ] Add a failing integration test for docs-only PRs skipping CodeGraph and package analysis.
- [ ] Add a failing integration test that writes `summary.md`, `impact.mmd`, and `check.json`.
- [ ] Verify failure before implementation.
- [ ] Implement change discovery, predicate aggregation, report rendering and advisory/enforced exit codes.
- [ ] Replace the PowerShell implementation with a wrapper that invokes `go run ./cmd/pr-impact`.
- [ ] Run tests and commit.

### Task 5: CI Replacement

**Files:**

- Modify: `.github/workflows/impact.yml`

**Interfaces:**

- Consumes: `scripts/pr-impact.ps1 -Stage <head-ref> -CodeGraphScripts <path>`

- [ ] Remove Node-only `@optave` installation and graph database cache.
- [ ] Add `actions/setup-go@v5` using `go.mod`.
- [ ] Install `@longbowzz/dsh-codegraph@0.4.0` with `npm install --no-save --ignore-scripts`.
- [ ] Run the gate with the PR head ref and package skill script directory.
- [ ] Keep artifact upload and sticky PR comment steps.

### Task 6: Stage CodeGraph Bundle and Cleanup

**Files:**

- Create: `docs/codegraph/codegraph-gate/graph.mmd`
- Create: `docs/codegraph/codegraph-gate/graph.locs.json`
- Create: `docs/codegraph/codegraph-gate/graph.html`
- Create: `docs/diagrams/stages/02-codegraph-gate.puml`
- Create: `docs/diagrams/stages/02-codegraph-gate.svg`

- [ ] Generate a review-mode CodeGraph for the gateway startup and request flow using verified source locations.
- [ ] Validate and build the single-file HTML with the pinned CodeGraph scripts.
- [ ] Generate the stage 02 PlantUML and SVG.
- [ ] Run `scripts/pr-impact.ps1` against `origin/main` with the stage bundle.
- [ ] Uninstall `@optave/codegraph`.
- [ ] Remove legacy `.codegraph/graph.db` and `.codegraph/changes.journal`.

### Task 7: Documentation and Verification

**Files:**

- Modify: `AGENTS.md`
- Modify: `README.md`
- Modify: `docs/pr-workflow.md`
- Modify: `docs/architecture.md`
- Modify: `.github/pull_request_template.md`

- [ ] Replace `@optave/codegraph` gate documentation with Go impact plus CodeGraph validation.
- [ ] Document commands, artifact paths, enforcement, rollback and the semantic limit of CodeGraph validation.
- [ ] Run `go build ./...`, `go test ./...`, `go test -race ./...`, `golangci-lint run`, `make diagrams`, `scripts/pr-impact.ps1`, and `gh pr checks`.
- [ ] Run `rg -n "@optave|codegraph branch-compare|codegraph diff-impact|codegraph check"` and require zero matches.
- [ ] Commit the completed stage.
