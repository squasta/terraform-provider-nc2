// Package oapi is the OpenAPI-coverage runtime used by both the
// provider's per-resource/per-data-source/per-action packages and the
// CI-only `tools/coverage-check` and `tools/sensitive-lint` binaries.
//
// Per research item R-09 (see specs/001-nutanix-nc2-provider/research.md),
// every package that talks to the NC2 API declares the operationIds it
// covers via a package-level `OperationMappings` slice and registers it
// with the global `Default` registry from a side-effect-only `init`:
//
//	var OperationMappings = []oapi.Mapping{
//	    {OperationID: "CPanelWeb.Api.OrganizationController.show", TerraformOp: "nc2_organization.Read"},
//	}
//
//	//nolint:gochecknoinits // documented mechanism per R-09.
//	func init() { oapi.Default.Register(OperationMappings...) }
//
// The CI tool walks the registry, diffs against `openapi/openapi.json`,
// and fails CI on any uncovered operation (FR-021, FR-021a). The runtime
// surface here is intentionally tiny: a struct, a method to register, and
// accessors that return defensive copies so callers cannot mutate the
// shared state.
//
// The package is pure-by-construction: `Default` is the single piece of
// shared state, mediated by a `sync.RWMutex`, and every public accessor
// returns a fresh copy so concurrent readers cannot observe partial
// updates.
package oapi

import (
	"sort"
	"sync"
)

// Mapping ties a single NC2 OpenAPI operationId to the Terraform-side
// operation that calls it.
//
// OperationID matches the `operationId` value in `openapi/openapi.json`
// verbatim (preserving the unusual but real "(2)" suffixes used to
// disambiguate same-method operations on the same path). TerraformOp
// is the dotted identifier used in audit records and progress events
// (e.g. `"nc2_aws_cluster.Update.capacity"`). Both fields are required.
type Mapping struct {
	OperationID string
	TerraformOp string
}

// Registry collects Mapping entries from every package that talks to
// the NC2 API. The runtime needs only Register; the CI coverage tool
// uses Snapshot to enumerate.
type Registry struct {
	mu       sync.RWMutex
	mappings []Mapping
}

// Default is the global Registry every per-package init populates.
// CI-only callers use Snapshot to enumerate it without taking a lock
// for the duration of the walk.
//
//nolint:gochecknoglobals // documented per-package mechanism per R-09.
var Default = &Registry{}

// Register appends the supplied mappings to the registry. Empty
// OperationID or empty TerraformOp entries are silently ignored, so
// generated `OperationMappings` slices that include placeholder rows
// during development do not crash CI.
//
// Register is safe for concurrent use; repeated calls with the same
// (OperationID, TerraformOp) pair de-duplicate so coverage-check does
// not double-count.
func (r *Registry) Register(ms ...Mapping) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, m := range ms {
		if m.OperationID == "" || m.TerraformOp == "" {
			continue
		}
		if r.containsLocked(m) {
			continue
		}
		r.mappings = append(r.mappings, m)
	}
}

// containsLocked checks for an existing identical mapping; caller
// holds the write lock.
func (r *Registry) containsLocked(m Mapping) bool {
	for _, existing := range r.mappings {
		if existing.OperationID == m.OperationID && existing.TerraformOp == m.TerraformOp {
			return true
		}
	}
	return false
}

// Snapshot returns a freshly-allocated slice of all currently
// registered mappings, sorted ascending by OperationID then
// TerraformOp. The returned slice is safe to mutate; the registry is
// not affected.
func (r *Registry) Snapshot() []Mapping {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Mapping, len(r.mappings))
	copy(out, r.mappings)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].OperationID != out[j].OperationID {
			return out[i].OperationID < out[j].OperationID
		}
		return out[i].TerraformOp < out[j].TerraformOp
	})
	return out
}

// CoveredOperationIDs returns the deduplicated, sorted list of
// OperationIDs that have at least one Terraform-side covering call.
// This is the slice `tools/coverage-check` diffs against the OpenAPI
// inventory.
func (r *Registry) CoveredOperationIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	seen := make(map[string]struct{}, len(r.mappings))
	for _, m := range r.mappings {
		seen[m.OperationID] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// Reset wipes the registry. Intended for unit tests only; production
// code MUST NOT call it.
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mappings = nil
}

// Len returns the current mapping count. Intended for diagnostics
// and unit tests.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.mappings)
}
