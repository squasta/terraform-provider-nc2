//go:build tools
// +build tools

// Package tools tracks build-time dependencies for terraform-provider-nc2.
//
// This file imports CI-only Go tools so that `go mod tidy` records them in
// go.mod / go.sum. The `//go:build tools` constraint guarantees the file
// is excluded from any normal build, so none of these dependencies are
// linked into the provider binary. Runtime supply-chain surface stays
// minimal (FR-032a vulnerability gating is much easier on a small graph).
//
// Tools imported here:
//
//   - github.com/getkin/kin-openapi:        OpenAPI 3.0 parser used by
//                                           tools/coverage-check (FR-021)
//                                           and tools/sensitive-lint
//                                           (FR-002a).
//   - github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs:
//                                           Generates Terraform Registry
//                                           reference documentation from
//                                           the provider schema (FR-030).
//   - github.com/golangci/golangci-lint/cmd/golangci-lint:
//                                           Linter aggregator invoked by
//                                           `make lint`.
//   - golang.org/x/vuln/cmd/govulncheck:    Reachability-aware vulnerability
//                                           scanner; HIGH/CRITICAL CVEs
//                                           hard-fail the release pipeline
//                                           (FR-032a).
//   - github.com/google/osv-scanner/cmd/osv-scanner:
//                                           Module-graph vulnerability
//                                           scanner; complements
//                                           govulncheck per FR-032a.
//   - github.com/goreleaser/goreleaser:     Cross-compile, archive,
//                                           checksum, GPG-sign release
//                                           artifacts (FR-032).
//   - github.com/sigstore/cosign/v2/cmd/cosign:
//                                           Keyless artifact signatures
//                                           via the GitHub OIDC token
//                                           (FR-032 layer 2).
//
// Non-Go tools (slsa-github-generator GitHub Action, GPG, etc.) are
// installed by the release workflow itself and are intentionally absent
// from this file.
//
// To install everything at the versions pinned in go.sum, run:
//
//	make tools
package tools

import (
	_ "github.com/getkin/kin-openapi/openapi3"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/google/osv-scanner/cmd/osv-scanner"
	_ "github.com/goreleaser/goreleaser"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
	_ "github.com/sigstore/cosign/v2/cmd/cosign"
	_ "golang.org/x/vuln/cmd/govulncheck"
)
