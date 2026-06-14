# Skills

Skills are versioned playbooks for high-leverage workflows (design-to-product, resource-optimization, security-review). Each one is a markdown document describing **when to use**, **when not to use**, **required inputs**, **workflow**, **sensors**, **anti-patterns**, **output format**, and **metrics**.

> Phase 6 implements the spec/plan/report pipeline that skills consume; first-class skill files (with rollout evidence + benchmark gates) land in a follow-up release. The framework is in place — see `internal/spec` and `internal/plan` for the renderers a skill writer plugs into.

## Required core skills (per spec §27)

`spec-driven-development`, `prompt-refinement`, `context-engineering`, `tdd-cycle`, `design-to-product`, `react-parity`, `feature-toggle-map`, `backend-roadmap`, `resource-optimization`, `dependency-audit`, `log-audit`, `docker-image-audit`, `security-review`, `architecture-review`, `api-contract-review`, `design-system-review`, `performance-budget`, `safe-agent-handoff`, `cost-aware-execution`, `failure-analysis`.

## Promotion rule

A skill can be updated only when:

1. There is rollout evidence (run ids that prove the new version performs).
2. Failures and successes were analysed.
3. The edit is bounded.
4. A hold-out benchmark improves.
5. No safety rule is weakened.
6. Previous working behaviour is preserved.

`skill_versions` in SQLite tracks `(skill_name, version, content_hash, score, accepted, created_at)`. The promotion gate lives in `internal/memory.Promote` for now; a dedicated `internal/skills` package will move in once a real skill ships.
