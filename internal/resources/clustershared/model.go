package clustershared

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Model is the framework-decoded view of the common cluster state
// shared by every cluster resource. It deliberately omits
// `desired_state` (hibernate / resume) — that surface is AWS-only;
// see HibernatingModel below and the AWS-only `desired_state`
// schema attribute under `clustershared.CommonAttributesWithHibernate`.
//
// Cloud-specific resources embed this struct (or HibernatingModel)
// and add their cloud-specific fields. The tfsdk tags MUST match
// the attribute names returned by CommonAttributes() (or
// CommonAttributesWithHibernate() for AWS).
//
// Attribute groups that are deeply nested in the NC2 API (network,
// capacity, redundancy, resource_tags) are exposed here as flat
// Map<String,String> / List<Map<String,String>> for MVP. Strict
// typing of those subtrees is a future enhancement; the round-trip
// through map preserves user input verbatim and keeps plan diffs
// faithful for the common cases.
type Model struct {
	ID               types.String `tfsdk:"id"`
	OrganizationID   types.String `tfsdk:"organization_id"`
	CloudAccountID   types.String `tfsdk:"cloud_account_id"`
	Name             types.String `tfsdk:"name"`
	Region           types.String `tfsdk:"region"`
	UseCase          types.String `tfsdk:"use_case"`
	HostAccessSSHKey types.String `tfsdk:"host_access_ssh_key"`
	License          types.String `tfsdk:"license"`
	AOSVersion       types.String `tfsdk:"aos_version"`
	SoftwareTier     types.String `tfsdk:"software_tier"`
	Capacity         types.List   `tfsdk:"capacity"`
	Redundancy       types.Map    `tfsdk:"redundancy"`
	Network          types.Map    `tfsdk:"network"`
	ResourceTags     types.Map    `tfsdk:"resource_tags"`
	State            types.String `tfsdk:"state"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

// HibernatingModel extends Model with the AWS-only `desired_state`
// attribute. Only `nc2_aws_cluster` embeds this type; Azure / GCP
// clusters embed Model directly and therefore have no hibernate
// surface (FR-012, AWS-only).
//
// The `tfsdk:"desired_state"` tag matches the attribute added by
// CommonAttributesWithHibernate() — embedding HibernatingModel in a
// resource model that calls plain CommonAttributes() will fail at
// framework decode time because the schema and struct disagree.
type HibernatingModel struct {
	Model

	DesiredState types.String `tfsdk:"desired_state"`
}
