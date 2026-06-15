<!-- mode: bugfix -->
<!-- description: Reproduce -> regression test -> minimal fix -> verify. -->

## Bugfix loop

1. Reproduce locally with the smallest steps you can find; capture the exact stderr / exit code.
2. Write a failing test that mirrors the reproduction. Commit it before the fix when reviewers pair on the diff.
3. Apply the minimal fix that makes the test green. Do not bundle refactors into the same commit.
4. Re-run the full package test plus `go vet ./...` and `golangci-lint run ./...`.
5. Update CHANGELOG.md under the next version with the root cause in one sentence.
