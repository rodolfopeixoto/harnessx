#!/usr/bin/env bash
# Tiny POSIX-friendly assertion helpers for shell tests.
# Source this file: . scripts/lib/assert.sh
# Each assert prints a line and increments TESTS_PASSED/TESTS_FAILED.
TESTS_PASSED=${TESTS_PASSED:-0}
TESTS_FAILED=${TESTS_FAILED:-0}
TESTS_NAME=${TESTS_NAME:-?}

_pass() { TESTS_PASSED=$((TESTS_PASSED + 1)); printf "  ✓ %s\n" "$1"; }
_fail() { TESTS_FAILED=$((TESTS_FAILED + 1)); printf "  ✗ %s\n    %s\n" "$1" "$2" >&2; }

assert_eq() {
  if [ "$2" = "$3" ]; then _pass "$1"; else _fail "$1" "want=$2 got=$3"; fi
}

assert_contains() {
  case "$3" in
    *"$2"*) _pass "$1" ;;
    *) _fail "$1" "want substring=$2 got=$3" ;;
  esac
}

assert_exit_zero() { if "$@" >/tmp/_out 2>&1; then _pass "exit 0: $*"; else _fail "exit 0: $*" "$(cat /tmp/_out)"; fi; }
assert_exit_nonzero() { if "$@" >/tmp/_out 2>&1; then _fail "exit nonzero: $*" "got 0"; else _pass "exit nonzero: $*"; fi; }

report() {
  printf "\n%s: %d passed, %d failed\n" "$TESTS_NAME" "$TESTS_PASSED" "$TESTS_FAILED"
  [ "$TESTS_FAILED" -eq 0 ] || exit 1
}
