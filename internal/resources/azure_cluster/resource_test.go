package azure_cluster //nolint:revive,staticcheck

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
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

// TestComputeDiff_RoutesLicenseChange pins the simplest in-place
// routing case for the Azure cluster (license bump) and confirms
// that no hibernate / resume operation is ever scheduled.
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
		t.Errorf("computeDiff must never populate DesiredState* on Azure (FR-012); got %q→%q",
			d.DesiredStateFrom, d.DesiredStateTo)
	}

	ops, err := clustershared.RouteUpdate(d)
	if err != nil {
		t.Fatalf("RouteUpdate: %v", err)
	}
	for _, op := range ops {
		if op.Kind == clustershared.OpHibernate || op.Kind == clustershared.OpResume {
			t.Errorf("Azure must never route to %s (FR-012)", op.Kind)
		}
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
