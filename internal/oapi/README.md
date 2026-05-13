# `internal/oapi`

Runtime registry of NC2 OpenAPI operationId → Terraform-side mapping declarations (R-09, FR-021, FR-021a).

## Public API

| Symbol | Kind | Purpose |
|---|---|---|
| `Mapping` | struct | `{OperationID, TerraformOp}` pair. |
| `Registry` | struct | Thread-safe collector. |
| `Default` | `*Registry` | Global singleton populated by per-package `init()`. |
| `(*Registry).Register(...Mapping)` | method | Append (de-duplicating) one or more mappings. |
| `(*Registry).Snapshot()` | method | Returns a fresh sorted slice of every registered mapping. |
| `(*Registry).CoveredOperationIDs()` | method | Returns the deduplicated, sorted list of OperationIDs. |
| `(*Registry).Reset()` | method | Wipes the registry; tests only. |
| `(*Registry).Len()` | method | Diagnostic accessor. |

## Pattern of use

Every package that talks to the NC2 API declares its coverage in `openapi_mapping.go` and registers it in a side-effect-only `init`:

```go
package aws_cluster

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

var OperationMappings = []oapi.Mapping{
    {OperationID: "CPanelWeb.Api.ClusterController.create_aws", TerraformOp: "nc2_aws_cluster.Create"},
    // ...
}

//nolint:gochecknoinits // documented mechanism per R-09.
func init() { oapi.Default.Register(OperationMappings...) }
```

`tools/coverage-check` (CI-only) loads `openapi/openapi.json`, walks `oapi.Default`, and fails CI if any operationId in the spec lacks a registered mapping.

## Concurrency

`Registry` is safe for concurrent use; the lock is held for the duration of `Register` / `Snapshot`. `Default` is the single piece of shared state in the package; everything else is pure.
