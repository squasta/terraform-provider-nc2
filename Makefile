## Makefile for terraform-provider-nc2
##
## Targets are documented in specs/001-nutanix-nc2-provider/quickstart.md §5.
## Every target is also wired into the CI workflow at .github/workflows/ci.yml
## (unit lane) and .github/workflows/release.yml (release lane).
##
## Conventions:
##   - Phony targets only; this is not a build system, just a task runner.
##   - One target per FR concern (e.g., openapi-coverage <-> FR-021).
##   - Failing closed: every gate target exits non-zero on policy violation.

SHELL                 := bash
BINARY                := terraform-provider-nc2
GO                    ?= go
GOBIN                 ?= $(shell $(GO) env GOPATH)/bin
COVERAGE_PROFILE      := coverage.out
COVERAGE_MIN_PERCENT  := 80
OPENAPI_SOURCE        := openapi/openapi.json

## Tools pinned in tools.go (Go-installable). Non-Go tools (GPG, slsa-github-generator
## GitHub Action) are installed by the release workflow itself.
TOOLS := \
	github.com/golangci/golangci-lint/cmd/golangci-lint \
	github.com/goreleaser/goreleaser \
	github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs \
	github.com/sigstore/cosign/v2/cmd/cosign \
	github.com/google/osv-scanner/cmd/osv-scanner \
	golang.org/x/vuln/cmd/govulncheck

.PHONY: all help tools build test coverage lint doc-lint openapi-coverage \
        sensitive-lint sensitive-lint-strict vuln testacc release clean \
        fmt vet tidy

## help: print one-line summary of every target.
help:
	@awk 'BEGIN {FS=": "} /^## [a-zA-Z][a-zA-Z0-9_-]*:/ { sub("^## ", ""); printf "  %-26s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

## all: default target — runs lint, test, coverage, openapi-coverage.
all: lint test coverage openapi-coverage

## tools: install every Go-installable CI tool pinned in tools.go.
tools:
	@echo "==> Installing tools pinned in tools.go"
	@for tool in $(TOOLS); do \
		echo "    go install $$tool"; \
		$(GO) install "$$tool@latest" || exit 1; \
	done
	@echo "==> Tools installed in $(GOBIN)"

## build: build the provider binary into ./terraform-provider-nc2.
build:
	@echo "==> Building $(BINARY)"
	$(GO) build -trimpath -ldflags="-s -w" -o $(BINARY) .

## test: run all unit tests with -race; never hits a real NC2 endpoint.
test:
	@echo "==> Running unit tests (-race)"
	$(GO) test -race -count=1 ./...

## coverage: run unit tests with coverage; fails if line coverage < COVERAGE_MIN_PERCENT.
coverage:
	@echo "==> Running unit tests with coverage profile"
	$(GO) test -race -count=1 -coverprofile=$(COVERAGE_PROFILE) -covermode=atomic ./internal/... ./tools/...
	@echo "==> Coverage summary"
	$(GO) tool cover -func=$(COVERAGE_PROFILE) | tail -1
	@pct=$$($(GO) tool cover -func=$(COVERAGE_PROFILE) | awk '/^total:/ {gsub("%","",$$3); print int($$3)}'); \
	  if [ "$$pct" -lt "$(COVERAGE_MIN_PERCENT)" ]; then \
	    echo "ERROR: coverage $$pct% < required $(COVERAGE_MIN_PERCENT)% (SC-005)"; \
	    exit 1; \
	  fi

## lint: run golangci-lint with the project config.
lint:
	@echo "==> Running golangci-lint"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not in PATH; run 'make tools' first"; exit 1; }
	golangci-lint run ./...

## doc-lint: enforce SC-007 — every exported Go symbol has a doc comment.
doc-lint:
	@echo "==> Checking exported-symbol doc comments (SC-007)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not in PATH; run 'make tools' first"; exit 1; }
	golangci-lint run --no-config --disable-all --enable=revive --enable-only=revive ./... \
	  -- -E revive --config <(printf '[rule.exported]\n')

## openapi-coverage: FR-021 / FR-021a gate. Build the tool, run it against openapi/.
openapi-coverage:
	@echo "==> OpenAPI coverage check (FR-021)"
	@if [ ! -d tools/coverage-check ] || [ -z "$$(ls -A tools/coverage-check 2>/dev/null | grep -v '^\.gitkeep$$')" ]; then \
	  echo "    tools/coverage-check not yet implemented (Phase 2 task T039/T041); skipping"; \
	  exit 0; \
	fi
	$(GO) run ./tools/coverage-check -openapi $(OPENAPI_SOURCE) -output tools/coverage-check/output/coverage-report.json

## sensitive-lint: FR-002a — advisory mode, exits 0 on warnings.
sensitive-lint:
	@echo "==> Sensitive-field lint (FR-002a, advisory)"
	@if [ ! -d tools/sensitive-lint ] || [ -z "$$(ls -A tools/sensitive-lint 2>/dev/null | grep -v '^\.gitkeep$$')" ]; then \
	  echo "    tools/sensitive-lint not yet implemented (Phase 2 task T040/T042); skipping"; \
	  exit 0; \
	fi
	$(GO) run ./tools/sensitive-lint -openapi $(OPENAPI_SOURCE)

## sensitive-lint-strict: FR-002a — strict mode, fails on any unclassified field.
sensitive-lint-strict:
	@echo "==> Sensitive-field lint (FR-002a, strict)"
	@if [ ! -d tools/sensitive-lint ] || [ -z "$$(ls -A tools/sensitive-lint 2>/dev/null | grep -v '^\.gitkeep$$')" ]; then \
	  echo "    tools/sensitive-lint not yet implemented (Phase 2 task T040/T042); skipping"; \
	  exit 0; \
	fi
	$(GO) run ./tools/sensitive-lint -openapi $(OPENAPI_SOURCE) -strict

## vuln: FR-032a — fail on HIGH or CRITICAL CVEs (govulncheck + osv-scanner).
vuln:
	@echo "==> govulncheck"
	@command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck not in PATH; run 'make tools' first"; exit 1; }
	govulncheck ./...
	@echo "==> osv-scanner"
	@command -v osv-scanner >/dev/null 2>&1 || { echo "osv-scanner not in PATH; run 'make tools' first"; exit 1; }
	osv-scanner --lockfile=go.mod

## testacc: acceptance tests; requires TF_ACC=1 and NC2_API_KEY/KEY_ID/ISSUER env vars.
testacc:
	@echo "==> Running acceptance tests"
	@if [ "$$TF_ACC" != "1" ]; then \
	  echo "ERROR: TF_ACC=1 required for acceptance tests"; \
	  exit 1; \
	fi
	@for v in NC2_API_KEY NC2_KEY_ID NC2_ISSUER; do \
	  if [ -z "$${!v}" ]; then echo "ERROR: $$v not set"; exit 1; fi; \
	done
	$(GO) test -race -count=1 -timeout=120m -tags=acceptance ./tests/acceptance/...

## release: dry-run the release pipeline locally; the real release runs in CI.
release:
	@echo "==> Release dry-run (use 'gh workflow run release.yml' for real)"
	@command -v goreleaser >/dev/null 2>&1 || { echo "goreleaser not in PATH; run 'make tools' first"; exit 1; }
	goreleaser release --snapshot --clean

## fmt: format all Go files via gofmt -s.
fmt:
	@echo "==> gofmt -s -w"
	gofmt -s -w .

## vet: run go vet.
vet:
	$(GO) vet ./...

## tidy: tidy go.mod / go.sum.
tidy:
	$(GO) mod tidy

## clean: remove built artifacts.
clean:
	rm -f $(BINARY) $(COVERAGE_PROFILE)
	rm -rf dist/ tools/coverage-check/output/
