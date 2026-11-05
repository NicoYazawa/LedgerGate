## Agent skills

### Issue tracker

Issues and specs live as markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

The five canonical triage roles use their default label strings. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: `GLOSSARY.md` at the repo root and ADRs under `docs/adr/`. See `docs/agents/domain.md`.

## Development workflow

- Follow `docs/pr-workflow.md` for every change.
- Stage diagrams: at the end of every completed implementation stage or ticket, add `docs/diagrams/stages/NN-<slug>.puml`, render its `.svg`, and commit both with the stage.
- Impact gate: `scripts/pr-impact.ps1` combines Go reverse-dependency and exported-declaration checks with committed CodeGraph artifacts under `docs/codegraph/<stage>/`.
- Coverage gate: total Go statement coverage includes `cmd/*`; `make coverage` enforces at least 90%.
- W1 scope and interfaces: `docs/w1-foundation.md`.
- Normative contracts: `docs/w1-config.md`, `docs/w1-schema.md`, `docs/w1-openapi.md`, `docs/images.md`, `docs/local-dev.md`.
- Reviewer and fixer prompts: `docs/prompts/`.
- PR size check: `scripts/pr-size.ps1`.
- Impact gate: `scripts/pr-impact.ps1`, `cmd/pr-impact`, and `.github/workflows/impact.yml`.
