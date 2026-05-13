// Package audit emits one structured audit record per NC2 HTTP call.
//
// Per research item R-10 and FR-020a..d, the records are emitted via
// the framework's tflog channel at INFO level so they ride the
// standard TF_LOG / TF_LOG_PATH routing — operators do not need to
// learn a new mechanism. The emitter is pure-by-construction: a tiny
// wrapper over `tflog.SubsystemInfo` with no allocations beyond the
// per-call field map.
//
// Sensitive values are redacted before emission via the FR-002
// pattern policy implemented in internal/redact (FR-020b). Audit
// records are NEVER written to a private sink: there is no
// internal-only file, no shadow channel, nothing besides the
// framework's logging surface (FR-020c).
package audit

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/nutanix/terraform-provider-nc2/internal/redact"
)

// Subsystem is the tflog subsystem name used by every audit record.
// Operators filter via `TF_LOG_PROVIDER_AUDIT=INFO`.
const Subsystem = "audit"

// AuditRecord is the structured per-call audit record per FR-020a.
// All fields are emitted; empty optional fields (NC2TaskID,
// NC2ErrorCode) are omitted from the rendered map so the output
// stays compact.
type AuditRecord struct {
	CorrelationID string
	TerraformOp   string
	HTTPMethod    string
	Path          string
	Status        int
	LatencyMs     int64
	NC2TaskID     string
	NC2ErrorCode  string
}

// defaultRegistry is the redact.Registry used to scrub the rendered
// field map. The audit emitter operates over field NAMES (not values
// from the NC2 response body), so the bare FR-002 pattern policy is
// sufficient — per-resource extras are applied where the response
// body itself is logged, not here.
//
//nolint:gochecknoglobals // documented per-package singleton.
var defaultRegistry = redact.NewRegistry(nil)

// Record emits the supplied audit record via tflog at INFO level.
// Side effect only; no return value because the framework's log
// channel cannot fail in a way callers can recover from.
//
// Field naming matches FR-020a verbatim. The function is safe to
// call from any goroutine; tflog itself is concurrency-safe.
func Record(ctx context.Context, r AuditRecord) {
	fields := map[string]any{
		"correlation_id": r.CorrelationID,
		"terraform_op":   r.TerraformOp,
		"http_method":    r.HTTPMethod,
		"path":           r.Path,
		"status":         strconv.Itoa(r.Status),
		"latency_ms":     strconv.FormatInt(r.LatencyMs, 10),
	}
	if r.NC2TaskID != "" {
		fields["nc2_task_id"] = r.NC2TaskID
	}
	if r.NC2ErrorCode != "" {
		fields["nc2_error_code"] = r.NC2ErrorCode
	}
	scrubbed := defaultRegistry.Redact(fields)

	subCtx := tflog.NewSubsystem(ctx, Subsystem)
	tflog.SubsystemInfo(subCtx, Subsystem, "nc2_api_call", scrubbed)
}
