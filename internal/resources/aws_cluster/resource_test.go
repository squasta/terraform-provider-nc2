package aws_cluster //nolint:revive,staticcheck

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestSensitive_RegistersAccessPolicyIPs pins data-model.md §4.
func TestSensitive_RegistersAccessPolicyIPs(t *testing.T) {
	t.Parallel()

	got := Sensitive()
	want := map[string]bool{
		"network.management_services_access_policy.ip_addresses": false,
		"network.prism_element_access_policy.ip_addresses":       false,
		"access_policy.ip_addresses":                             false,
	}
	for _, p := range got {
		if _, ok := want[p]; ok {
			want[p] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("Sensitive() missing %q", k)
		}
	}
}

// TestComputeDiff_FlagsLicenseChange covers FR-010 update routing.
func TestComputeDiff_FlagsLicenseChange(t *testing.T) {
	t.Parallel()

	state := &model{}
	state.License = types.StringValue("aos")
	state.AOSVersion = types.StringValue("6.7")
	state.SoftwareTier = types.StringValue("pro")
	state.HostAccessSSHKey = types.StringValue("k1")

	plan := &model{}
	plan.License = types.StringValue("aos")
	plan.AOSVersion = types.StringValue("6.8")
	plan.SoftwareTier = types.StringValue("pro")
	plan.HostAccessSSHKey = types.StringValue("k1")

	d := computeDiff(state, plan)
	if !d.License {
		t.Errorf("expected License diff to be true")
	}
	if d.SSHKey {
		t.Errorf("expected SSHKey diff to be false")
	}
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

// TestOperationMappings_CoversFR021Set pins the AWS-cluster set.
func TestOperationMappings_CoversFR021Set(t *testing.T) {
	t.Parallel()

	want := map[string]bool{
		"CPanelWeb.Api.ClusterController.create_aws":            false,
		"CPanelWeb.Api.ClusterController.show":                  false,
		"CPanelWeb.Api.ClusterController.update":                false,
		"CPanelWeb.Api.ClusterController.update (2)":            false,
		"CPanelWeb.Api.ClusterController.update_capacity":       false,
		"CPanelWeb.Api.ClusterController.update_ssh_key":        false,
		"CPanelWeb.Api.ClusterController.update_license":        false,
		"CPanelWeb.Api.ClusterController.update_resource_tags":  false,
		"CPanelWeb.Api.ClusterController.update_access_policy":  false,
		"CPanelWeb.Api.ClusterController.terminate":             false,
		"CPanelWeb.Api.ClusterController.hibernate":             false,
		"CPanelWeb.Api.ClusterController.resume":                false,
	}
	for _, m := range OperationMappings {
		if _, ok := want[m.OperationID]; ok {
			want[m.OperationID] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("OperationMappings missing %q", k)
		}
	}
}
