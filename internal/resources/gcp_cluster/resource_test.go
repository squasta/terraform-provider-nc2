package gcp_cluster //nolint:revive,staticcheck

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestComputeDiff_DesiredStateTransition covers hibernate routing.
func TestComputeDiff_DesiredStateTransition(t *testing.T) {
	t.Parallel()

	state := &model{}
	state.DesiredState = types.StringValue("hibernated")
	plan := &model{}
	plan.DesiredState = types.StringValue("running")

	d := computeDiff(state, plan)
	if d.DesiredStateFrom != "hibernated" || d.DesiredStateTo != "running" {
		t.Errorf("got %s -> %s; want hibernated -> running", d.DesiredStateFrom, d.DesiredStateTo)
	}
}

// TestOperationMappings covers create_gcp.
func TestOperationMappings(t *testing.T) {
	t.Parallel()

	if len(OperationMappings) < 1 {
		t.Fatalf("OperationMappings empty")
	}
	if OperationMappings[0].OperationID != "CPanelWeb.Api.ClusterController.create_gcp" {
		t.Errorf("OperationID=%q", OperationMappings[0].OperationID)
	}
}
