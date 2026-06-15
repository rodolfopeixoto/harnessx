#!/usr/bin/env bash
# event: pre-tool-use
# description: Refuse any run when the project tree contains an .env file with apparent secret values.
set -euo pipefail

if [[ -f .env ]] && grep -E '(KEY|TOKEN|SECRET|PASSWORD)=[A-Za-z0-9]{16,}' .env >/dev/null 2>&1; then
  echo "[pre-tool-use-secrets] refusing: .env appears to contain plain-text secrets" >&2
  echo "  fix: move to harness secret set <name> or .env.local outside the worktree" >&2
  exit 1
fi
exit 0
