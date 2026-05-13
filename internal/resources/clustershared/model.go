package clustershared

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the framework-decoded view of the common cluster state.
// Cloud-specific resources embed this struct (or wrap it) and add
// their cloud-specific fields. The tfsdk tags MUST match the
// attribute names returned by CommonAttributes().
//
// Attribute groups that are deeply nested in the NC2 API (network,
// capacity, redundancy, resource_tags) are exposed here as flat
// Map<String,String> / List<Map<String,String>> for MVP. Strict
// typing of those subtrees is a future enhancement; the round-trip
// through map preserves user input verbatim and keeps plan diffs
// faithful for the common cases.
type Model struct {
	ID              types.String `tfsdk:"id"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	CloudAccountID  types.String `tfsdk:"cloud_account_id"`
	Name            types.String `tfsdk:"name"`
	Region          types.String `tfsdk:"region"`
	UseCase         types.String `tfsdk:"use_case"`
	HostAccessSSHKey types.String `tfsdk:"host_access_ssh_key"`
	License         types.String `tfsdk:"license"`
	AOSVersion      types.String `tfsdk:"aos_version"`
	SoftwareTier    types.String `tfsdk:"software_tier"`
	Capacity        types.List   `tfsdk:"capacity"`
	Redundancy      types.Map    `tfsdk:"redundancy"`
	Network         types.Map    `tfsdk:"network"`
	ResourceTags    types.Map    `tfsdk:"resource_tags"`
	DesiredState    types.String `tfsdk:"desired_state"`
	State           types.String `tfsdk:"state"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}
