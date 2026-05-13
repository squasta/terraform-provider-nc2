package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-log/tflogtest"
)

// TestRecord_EmitsExpectedFields covers FR-020a: the rendered log
// line carries every required field with the documented names. We
// route the framework log channel through a buffer and parse the
// resulting JSON.
func TestRecord_EmitsExpectedFields(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	ctx := tflogtest.RootLogger(context.Background(), buf)

	Record(ctx, AuditRecord{
		CorrelationID: "corr-1",
		TerraformOp:   "nc2_organization.Create",
		HTTPMethod:    "POST",
		Path:          "/organizations",
		Status:        201,
		LatencyMs:     42,
		NC2TaskID:     "task-1",
	})

	out := buf.String()
	if out == "" {
		t.Fatalf("no log output captured")
	}
	if !strings.Contains(out, "nc2_api_call") {
		t.Errorf("output missing message; got %s", out)
	}
	for _, want := range []string{
		`"correlation_id":"corr-1"`,
		`"terraform_op":"nc2_organization.Create"`,
		`"http_method":"POST"`,
		`"path":"/organizations"`,
		`"nc2_task_id":"task-1"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %s; got %s", want, out)
		}
	}
}

// TestRecord_OmitsAbsentOptionalFields verifies the FR-020a
// "absent fields not emitted" rule.
func TestRecord_OmitsAbsentOptionalFields(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	ctx := tflogtest.RootLogger(context.Background(), buf)

	Record(ctx, AuditRecord{
		CorrelationID: "corr-2",
		TerraformOp:   "data.nc2_clusters.Read",
		HTTPMethod:    "GET",
		Path:          "/clusters",
		Status:        200,
		LatencyMs:     7,
	})

	if strings.Contains(buf.String(), "nc2_task_id") {
		t.Errorf("nc2_task_id should be absent when empty; got %s", buf.String())
	}
	if strings.Contains(buf.String(), "nc2_error_code") {
		t.Errorf("nc2_error_code should be absent when empty; got %s", buf.String())
	}
}

// TestRecord_EmitsValidJSON confirms each line is JSON-parseable.
func TestRecord_EmitsValidJSON(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	ctx := tflogtest.RootLogger(context.Background(), buf)
	Record(ctx, AuditRecord{TerraformOp: "x", HTTPMethod: "GET", Path: "/", Status: 200, LatencyMs: 1})

	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("non-JSON line %q: %v", line, err)
		}
	}
}
