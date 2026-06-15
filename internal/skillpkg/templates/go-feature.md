<!-- mode: feature -->
<!-- description: Go feature loop: types -> ctor -> behaviour -> test -> wire. -->

## Go feature loop

1. Sketch the public type + method signature first; keep zero-value usable when possible.
2. Add a constructor only when defaults differ from the zero value.
3. Implement the smallest behaviour that satisfies one named test case.
4. Add the test with `t.Run("scenario", func(t *testing.T) { ... })`; one assertion per case.
5. Wire from the next layer up (`cmd/` or app); update the changelog entry in the same commit.
6. Run `gofmt`, `golangci-lint run ./...`, and the package test before pushing.
