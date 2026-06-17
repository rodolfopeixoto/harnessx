SHELL := /usr/bin/env bash
GO    ?= go
BIN   ?= bin/harness
PKGS  := ./...

# gvm leaves a stale GOROOT in some shells; unset it so brew/system Go works.
export GOROOT :=

# Version stamping for `make release` (override via env: VERSION=v1.2.3 make release).
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
  -X github.com/ropeixoto/harnessx/internal/version.Version=$(VERSION) \
  -X github.com/ropeixoto/harnessx/internal/version.Commit=$(COMMIT) \
  -X github.com/ropeixoto/harnessx/internal/version.Date=$(DATE)

# Local release matrix (no GitHub Actions). Override on the command line
# to slice down: `PLATFORMS="darwin/arm64" make release`.
PLATFORMS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: all build test test-short vet lint fmt tidy check ci cd release \
        e2e e2e-all bench coverage coverage-gate security licenses sbom \
        profile-mem profile-cpu \
        clean install-hooks uninstall-hooks \
        dashboard-install dashboard-dev dashboard-build dashboard-test \
        help

audit-solid:
	$(GO) run ./cmd/harness audit-solid --root .

profile-mem:
	@mkdir -p dist/profiles
	$(GO) test -run '^$$' -bench=. -benchmem -memprofile=dist/profiles/mem.pprof ./internal/profile/... 2>/dev/null || true
	$(GO) test -bench=. -benchmem -memprofile=dist/profiles/router-mem.pprof -run '^$$' ./internal/router/... 2>/dev/null || true
	@echo "heap profiles -> dist/profiles/"

profile-cpu:
	@mkdir -p dist/profiles
	$(GO) test -run '^$$' -bench=. -cpuprofile=dist/profiles/cpu.pprof ./internal/profile/... 2>/dev/null || true
	@echo "cpu profile -> dist/profiles/cpu.pprof"

all: check

help:
	@echo "HarnessX — local-only CI/CD (GitHub Actions is not used)."
	@echo
	@echo "Most common targets:"
	@echo "  make check           vet + race tests + build"
	@echo "  make e2e-all         run every scripts/e2e-phase*.sh in order"
	@echo "  make ci              full local CI gate (lint + coverage-gate + tests + e2e)"
	@echo "  make cd              local CD: ci + dashboard build + security + licenses + sbom + release"
	@echo "  make release         multi-arch cross-build into dist/ + SHA-256 sums"
	@echo "  make bench           run benchmarks (./internal/...)"
	@echo "  make coverage        write coverage.out + per-pkg report"
	@echo "  make coverage-gate   enforce GLOBAL_MIN/CORE_MIN thresholds"
	@echo "  make security        govulncheck + harness security-audit"
	@echo "  make licenses        regen THIRD_PARTY_LICENSES.md + NOTICE"
	@echo "  make sbom            regen dist/sbom.cyclonedx.json"
	@echo "  make smoke           run cross-stack CLI smoke matrix"
	@echo "  make install-hooks   wire scripts/git-hooks/* into .git/hooks/"
	@echo

build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/harness

test:
	$(GO) test -race -cover $(PKGS)

test-short:
	$(GO) test -short $(PKGS)

vet:
	$(GO) vet $(PKGS)

lint:
	@if command -v golangci-lint >/dev/null; then \
	  golangci-lint run; \
	else \
	  echo "golangci-lint not installed; running go vet only"; \
	  $(GO) vet $(PKGS); \
	fi

fmt:
	$(GO) fmt $(PKGS)

tidy:
	$(GO) mod tidy

check: vet test build

# smoke: exercise core CLI surface against a fresh project for every
# bundled scaffold. Catches regressions where dev-repo commands break
# for downstream users.
smoke: build
	$(BIN) smoke matrix --bin $(BIN) --step-timeout 180s

# tutorial-replay: deterministic walk of the docs/tutorial-python-demo.md
# cheat-sheet (LLM-free; dry-run on `harness ship`). Catches drift
# between documented cmds and actual binary surface.
tutorial-replay: build
	HARNESS_BIN=$(BIN) bash scripts/tutorial-replay.sh

# ci: full local CI gate. Wired to the pre-push hook by `make install-hooks`.
ci: lint check coverage-gate coverage-shell test-sh e2e-all

# test-sh: shell test harness (scripts/tests/test-*.sh).
test-sh:
	@bash scripts/tests/run-all.sh

# cd: continuous delivery gate run before a push to main/develop. Layers
# dashboard build, security scan, license + SBOM generation, and the
# multi-arch release tarballs on top of ci.
cd: dashboard-install dashboard-build ci security licenses sbom release

bench:
	$(GO) test -run '^$$' -bench=. -benchtime=2x ./internal/...

coverage:
	$(GO) test -race -coverpkg=./... -coverprofile=coverage.out ./... > /dev/null
	$(GO) tool cover -func=coverage.out | tail -25
	@echo "→ html: go tool cover -html=coverage.out -o coverage.html"

coverage-gate:
	@bash scripts/coverage-gate.sh

coverage-web:
	@bash scripts/coverage-web.sh

coverage-shell:
	@bash scripts/coverage-shell.sh

security: build
	@bash scripts/security-gate.sh

licenses:
	@bash scripts/license-gate.sh

sbom:
	@mkdir -p dist
	@if command -v syft >/dev/null; then \
	  syft scan dir:. -o cyclonedx-json > dist/sbom.cyclonedx.json; \
	else \
	  python3 scripts/sbom-fallback.py > dist/sbom.cyclonedx.json; \
	fi
	@echo "→ dist/sbom.cyclonedx.json"

release: build
	@rm -rf dist && mkdir -p dist
	@for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; \
	  ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
	  target="harness-$${os}-$${arch}"; \
	  echo "→ $$target"; \
	  GOOS="$$os" GOARCH="$$arch" CGO_ENABLED=0 \
	    $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o "dist/$${target}$${ext}" ./cmd/harness; \
	  size=$$(stat -f %z "dist/$${target}$${ext}" 2>/dev/null || stat -c %s "dist/$${target}$${ext}"); \
	  printf "  %-30s %8d bytes\n" "$$target" "$$size"; \
	  if [ "$$size" -gt 41943040 ]; then \
	    echo "  ✗ $$target exceeds 40 MiB binary-size budget"; exit 1; \
	  fi; \
	  if [ "$$os" = "windows" ]; then \
	    (cd dist && zip -q "$${target}.zip" "$${target}$${ext}" && \
	      shasum -a 256 "$${target}.zip" > "$${target}.zip.sha256"); \
	  else \
	    (cd dist && tar -czf "$${target}.tar.gz" "$${target}" && \
	      shasum -a 256 "$${target}.tar.gz" > "$${target}.tar.gz.sha256"); \
	  fi; \
	done
	@echo
	@echo "dist/ contents:"
	@ls -la dist/
	@echo
	@echo "verify locally:"
	@echo "  (cd dist && shasum -a 256 -c *.sha256)"

e2e: build
	bash scripts/e2e-phase1.sh

e2e-all: build
	@for s in scripts/e2e-phase*.sh; do \
	  echo "=== $$s ==="; \
	  bash "$$s" || exit 1; \
	done

clean:
	rm -rf bin dist coverage.* *.out

# Git hook wiring — keeps the pre-push gate honest without manual setup.
install-hooks:
	@mkdir -p .git/hooks
	@for h in scripts/git-hooks/*; do \
	  name=$$(basename "$$h"); \
	  cp "$$h" ".git/hooks/$$name"; \
	  chmod +x ".git/hooks/$$name"; \
	  echo "installed .git/hooks/$$name"; \
	done

uninstall-hooks:
	@for h in scripts/git-hooks/*; do \
	  name=$$(basename "$$h"); \
	  rm -f ".git/hooks/$$name"; \
	  echo "removed .git/hooks/$$name"; \
	done

dashboard-install:
	cd web/dashboard && npm install

dashboard-dev:
	cd web/dashboard && npm run dev

dashboard-build:
	cd web/dashboard && npm run build

dashboard-test:
	cd web/dashboard && npm test -- --run
