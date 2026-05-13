package aws_cluster //nolint:revive,staticcheck

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

// nullModel returns a model with every nullable attribute initialized
// to a null value. This mirrors the framework state when no field has
// been set yet and avoids two-unknowns-are-not-Equal false positives
// from zero-value `types.Map{}` / `types.List{}` defaults.
func nullModel() *model {
	emptyMap := types.MapNull(types.StringType)
	return &model{
		HibernatingModel: clustershared.HibernatingModel{
			Model: clustershared.Model{
				ID:               types.StringNull(),
				OrganizationID:   types.StringNull(),
				CloudAccountID:   types.StringNull(),
				Name:             types.StringNull(),
				Region:           types.StringNull(),
				UseCase:          types.StringNull(),
				HostAccessSSHKey: types.StringNull(),
				License:          types.StringNull(),
				AOSVersion:       types.StringNull(),
				SoftwareTier:     types.StringNull(),
				Capacity:         types.ListNull(types.MapType{ElemType: types.StringType}),
				Redundancy:       emptyMap,
				Network:          emptyMap,
				ResourceTags:     emptyMap,
				State:            types.StringNull(),
				CreatedAt:        types.StringNull(),
				UpdatedAt:        types.StringNull(),
			},
			DesiredState: types.StringNull(),
		},
		AccessPolicy: types.MapNull(types.StringType),
	}
}

// TestHibernate_RouteIsHibernate_FromRunning pins the FR-011
// behaviour for nc2_aws_cluster.
func TestHibernate_RouteIsHibernate_FromRunning(t *testing.T) {
	t.Parallel()

	state := nullModel()
	state.DesiredState = types.StringValue("running")
	plan := nullModel()
	plan.DesiredState = types.StringValue("hibernated")

	ops, err := clustershared.RouteUpdate(computeDiff(state, plan))
	if err != nil {
		t.Fatalf("RouteUpdate: %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != clustershared.OpHibernate {
		t.Errorf("ops = %+v; want single OpHibernate", ops)
	}
}

// TestHibernate_RouteIsResume_FromHibernated pins resume routing.
func TestHibernate_RouteIsResume_FromHibernated(t *testing.T) {
	t.Parallel()

	state := nullModel()
	state.DesiredState = types.StringValue("hibernated")
	plan := nullModel()
	plan.DesiredState = types.StringValue("running")

	ops, err := clustershared.RouteUpdate(computeDiff(state, plan))
	if err != nil {
		t.Fatalf("RouteUpdate: %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != clustershared.OpResume {
		t.Errorf("ops = %+v; want single OpResume", ops)
	}
}

// TestHibernate_RejectsCombinedDiff pins the
// "hibernate + capacity in same diff is not allowed" guard.
func TestHibernate_RejectsCombinedDiff(t *testing.T) {
	t.Parallel()

	state := nullModel()
	state.DesiredState = types.StringValue("running")
	state.HostAccessSSHKey = types.StringValue("k1")
	plan := nullModel()
	plan.DesiredState = types.StringValue("hibernated")
	plan.HostAccessSSHKey = types.StringValue("k2")

	if _, err := clustershared.RouteUpdate(computeDiff(state, plan)); err == nil {
		t.Errorf("expected error combining hibernate with field changes; got nil")
	}
}

// TestHibernate_NoOp_WhenUnchanged pins idempotency.
func TestHibernate_NoOp_WhenUnchanged(t *testing.T) {
	t.Parallel()

	state := nullModel()
	state.DesiredState = types.StringValue("running")
	plan := nullModel()
	plan.DesiredState = types.StringValue("running")

	ops, err := clustershared.RouteUpdate(computeDiff(state, plan))
	if err != nil {
		t.Fatalf("RouteUpdate: %v", err)
	}
	if len(ops) != 0 {
		t.Errorf("ops = %+v; want empty (no-op)", ops)
	}
}

// _ avoids unused import lint if attr isn't directly referenced.
var _ attr.Value = types.StringNull()
