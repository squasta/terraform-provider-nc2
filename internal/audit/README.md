# `internal/audit`

One structured audit record per NC2 HTTP call (FR-020a..d, R-10).

## Public API

| Symbol | Kind | Purpose |
|---|---|---|
| `AuditRecord` | struct | The flat per-call shape per FR-020a. |
| `Record(ctx, AuditRecord)` | func | Emits one `tflog.SubsystemInfo` line at the `audit` subsystem. |
| `Subsystem` | const | The tflog subsystem name (`"audit"`); operators filter via `TF_LOG_PROVIDER_AUDIT=INFO`. |

## Schema (FR-020a)

```go
type AuditRecord struct {
    CorrelationID string  // 16-byte hex; same value sent as X-Correlation-Id
    TerraformOp   string  // e.g. "nc2_organization.Create"
    HTTPMethod    string
    Path          string
    Status        int
    LatencyMs     int64
    NC2TaskID     string  // omitted when empty
    NC2ErrorCode  string  // omitted when empty
}
```

## Routing (FR-020c)

Records ride the framework's standard `tflog` channel. They're routed through the operator's existing `TF_LOG` / `TF_LOG_PATH` configuration — no internal-only sink, no shadow channel (FR-020c).

## Redaction (FR-020b)

Every string-valued field is run through `internal/redact.Registry` before emission. The default registry uses the bare FR-002 pattern policy; per-resource registries are applied where the response body itself is logged (in `internal/client.Do`), not here.
