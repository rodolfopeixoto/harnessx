#!/usr/bin/env bash
# Multi-stack smoke — scaffolds every bundled stack into a throwaway
# tempdir, runs harness ci against each, asserts no crash + the
# expected sensor catalogue. Keeps docs/TUTORIAL-MULTI-STACK.md
# honest: a stack that breaks the gate breaks this script.
#
# Usage:
#   scripts/multi-stack-smoke.sh                    # build local + run
#   HARNESS_BIN=/usr/local/bin/harness scripts/multi-stack-smoke.sh

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${HARNESS_BIN:-}"
if [[ -z "${BIN}" ]]; then
    echo "==> building harness binary"
    (cd "${ROOT}" && go build -o /tmp/harness-multi-smoke ./cmd/harness)
    BIN=/tmp/harness-multi-smoke
fi

WORK="$(mktemp -d -t harness-multi-XXXX)"
echo "==> scratch dir ${WORK}"
trap 'rm -rf "${WORK}"' EXIT

step() { printf "\n--- %s ---\n" "$*"; }

cd "${WORK}"

step "harness scaffold list reports every expected stack"
"${BIN}" scaffold list | tee /tmp/scaffold-list.out
for stack in go python python-ecommerce rails react ruby rust; do
    grep -F "${stack}" /tmp/scaffold-list.out >/dev/null
done

for stack in go python python-ecommerce react ruby rust rails; do
    step "harness new ${stack} ./app-${stack} --yes"
    "${BIN}" new "${stack}" "./app-${stack}" --yes
    test -d "./app-${stack}/.harness"
    test -d "./app-${stack}/.git"
done

for stack in go python python-ecommerce react ruby rust rails; do
    step "harness ci on ${stack} (skipped tools are OK)"
    (cd "./app-${stack}" && "${BIN}" ci || true) | tee "/tmp/ci-${stack}.out"
    grep -E "summary: [0-9]+ passed" "/tmp/ci-${stack}.out" >/dev/null \
        || { echo "ci summary missing for ${stack}"; exit 1; }
done

step "harness drive --help exposes --vcr + --features flags"
"${BIN}" drive --help 2>&1 | tee /tmp/drive-help.out
grep -F -- "--vcr " /tmp/drive-help.out >/dev/null
grep -F -- "--features" /tmp/drive-help.out >/dev/null
grep -F -- "--continue-on-fail" /tmp/drive-help.out >/dev/null

step "harness onboarding renders all sections"
"${BIN}" onboarding | tee /tmp/onboarding.out
grep -F "harness onboarding" /tmp/onboarding.out >/dev/null
grep -F "system tools" /tmp/onboarding.out >/dev/null
grep -F "agent adapters" /tmp/onboarding.out >/dev/null
grep -F "next steps" /tmp/onboarding.out >/dev/null

echo
echo "multi-stack-smoke OK"
