#!/usr/bin/env bash
# e2e-tutorial.sh — walks docs/TUTORIAL-BUILD-TASKHIVE.md end-to-end
# using the deterministic `fake` adapter. Zero cost, zero network.
#
# The script only exercises harness CLI surface — no real LLM calls,
# no Docker, no browser. Its job is to catch drift between the tutorial
# and the shipped binary before users hit it.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
HARNESS="${HARNESS_BIN:-$ROOT/bin/harness}"
WORK="${TUTORIAL_WORK:-/tmp/taskhive-e2e}"

if [ ! -x "$HARNESS" ]; then
  echo "e2e-tutorial: $HARNESS not found; run 'make build' first" >&2
  exit 1
fi

echo "== e2e-tutorial: workspace $WORK"
rm -rf "$WORK"
mkdir -p "$WORK/taskhive"
cd "$WORK/taskhive"

# --- Chapter 1: setup ---------------------------------------------------------
git init -q
git -c user.email=e2e@test.local -c user.name=e2e commit -q --allow-empty -m "init"

"$HARNESS" init >/dev/null
"$HARNESS" project index >/dev/null
"$HARNESS" doctor >/dev/null || true    # doctor reports adapter warnings; non-fatal
"$HARNESS" agent list >/dev/null
"$HARNESS" use fake >/dev/null
"$HARNESS" sensor list >/dev/null

mkdir -p .harness/config
cat > .harness/config/routes.yaml <<'EOF'
budget_usd: 5.00
default: fake
routes:
  - match: { intent: spec }
    agent: fake
  - match: { intent: implement }
    agent: fake
  - match: { intent: test }
    agent: fake
  - match: { intent: docs }
    agent: fake
EOF
git add -A && git -c user.email=e2e@test.local -c user.name=e2e commit -q -m "ch1 baseline"

# --- Chapters 2–5: backend feature calls --------------------------------------
"$HARNESS" feature "Go HTTP API scaffold with Chi router, sqlite, JWT auth" \
  --agent fake --yes >/dev/null
"$HARNESS" check >/dev/null
"$HARNESS" feature "add POST /auth/register and POST /auth/login with bcrypt + JWT" \
  --agent fake --yes >/dev/null
"$HARNESS" feature "add /projects CRUD scoped by user" \
  --agent fake --yes >/dev/null
"$HARNESS" feature "add /tasks CRUD nested under /projects" \
  --agent fake --yes >/dev/null
"$HARNESS" cost report --since 1h >/dev/null

# --- Chapter 6: frontend scaffold ---------------------------------------------
cd "$WORK"
"$HARNESS" new react taskhive-web --yes >/dev/null
cd taskhive-web

# --- Chapters 7–9: frontend + tests -------------------------------------------
"$HARNESS" feature "login form + protected route wrapper + axios 401 interceptor" \
  --agent fake --yes >/dev/null
"$HARNESS" feature "project list + detail + create/edit modals" \
  --agent fake --yes >/dev/null
"$HARNESS" feature "Vitest tests for useAuth + project list + axios interceptor" \
  --agent fake --yes >/dev/null

cd "$WORK/taskhive"
"$HARNESS" feature "table-driven httptest tests for every handler" \
  --agent fake --yes >/dev/null

# --- Chapter 10: docker (only exercise the harness feature call, not docker) --
"$HARNESS" feature "multi-stage Dockerfile + docker-compose with mem_limit" \
  --agent fake --yes >/dev/null

# --- Chapter 11: dashboard + perf snapshot ------------------------------------
"$HARNESS" perf-snapshot --label e2e-tutorial >/dev/null

# --- Chapter 12: cost breakdown -----------------------------------------------
"$HARNESS" cost report --breakdown >/dev/null

# --- Chapter 13: orchestrate (dry-run — flow lives under .harness/orch…) ------
mkdir -p .harness/orchestrations
cat > .harness/orchestrations/deploy-debate.yaml <<'EOF'
name: deploy-debate
topology: chain
steps:
  - role: planner
    adapter: fake
    command: [true]
    prompt: "Propose a deploy strategy."
  - role: reviewer
    adapter: fake
    command: [true]
    prompt: "Critique the proposal."
EOF
"$HARNESS" orchestrate run deploy-debate --dry-run >/dev/null

# --- Chapter 14: bugfix -------------------------------------------------------
"$HARNESS" bugfix "task list goes stale after status change" \
  --agent fake --yes >/dev/null

# --- Chapter 15: ci gate ------------------------------------------------------
"$HARNESS" ci >/dev/null

echo "== e2e-tutorial: OK"
