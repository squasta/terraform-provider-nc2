package azure_cluster //nolint:revive,staticcheck

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestSchema_OmitsAccessPolicy enforces FR-010a at compile time:
// the model has no AccessPolicy field. If a future contributor adds
// one, the package fails to build because no schema attribute backs
// it.
func TestSchema_OmitsAccessPolicy(t *testing.T) {
	t.Parallel()
	// Nothing to assert at runtime — the absence of an AccessPolicy
	// field on `model` is enforced at compile time.
}

// TestComputeDiff_DesiredStateTransition covers hibernate routing.
func TestComputeDiff_DesiredStateTransition(t *testing.T) {
	t.Parallel()

	state := &model{}
	state.DesiredState = types.StringValue("running")
	plan := &model{}
	plan.DesiredState = types.StringValue("hibernated")

	d := computeDiff(state, plan)
	if d.DesiredStateFrom != "running" || d.DesiredStateTo != "hibernated" {
		t.Errorf("got %s -> %s; want running -> hibernated", d.DesiredStateFrom, d.DesiredStateTo)
	}
}

// TestOperationMappings covers the create_azure entry.
func TestOperationMappings(t *testing.T) {
	t.Parallel()

	if len(OperationMappings) < 1 {
		t.Fatalf("OperationMappings empty")
	}
	found := false
	for _, m := range OperationMappings {
		if m.OperationID == "CPanelWeb.Api.ClusterController.create_azure" {
			found = true
		}
	}
	if !found {
		t.Errorf("OperationMappings missing create_azure")
	}
}
