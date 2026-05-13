package gcp_cluster //nolint:revive,staticcheck

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

// TestComputeDiff_RoutesLicenseChange pins the simplest in-place
// routing case for the GCP cluster (license bump) and confirms that
// no hibernate / resume operation is ever scheduled.
func TestComputeDiff_RoutesLicenseChange(t *testing.T) {
	t.Parallel()

	state := &model{}
	state.License = types.StringValue("aos")
	state.AOSVersion = types.StringValue("6.7")
	state.SoftwareTier = types.StringValue("pro")

	plan := &model{}
	plan.License = types.StringValue("aos")
	plan.AOSVersion = types.StringValue("6.8")
	plan.SoftwareTier = types.StringValue("pro")

	d := computeDiff(state, plan)
	if !d.License {
		t.Errorf("computeDiff: License flag not set after AOS bump")
	}
	if d.DesiredStateFrom != "" || d.DesiredStateTo != "" {
		t.Errorf("computeDiff must never populate DesiredState* on GCP (FR-012); got %q→%q",
			d.DesiredStateFrom, d.DesiredStateTo)
	}

	ops, err := clustershared.RouteUpdate(d)
	if err != nil {
		t.Fatalf("RouteUpdate: %v", err)
	}
	for _, op := range ops {
		if op.Kind == clustershared.OpHibernate || op.Kind == clustershared.OpResume {
			t.Errorf("GCP must never route to %s (FR-012)", op.Kind)
		}
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
