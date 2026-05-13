package clustershared

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CommonAttributes returns the framework attribute map shared by all
// three cloud-specific cluster resources. Per-cloud schemas merge
// these with their cloud-specific extras (e.g. AWS adds
// `access_policy`, GCP adds `network.gcp.*`).
//
// The returned map is freshly allocated; mutations by callers do not
// affect later invocations.
func CommonAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "NC2-assigned cluster UUID.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"organization_id": schema.StringAttribute{
			Description: "Owning organization UUID. Changes force replacement.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"cloud_account_id": schema.StringAttribute{
			Description: "Owning cloud account UUID. Changes force replacement.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"name": schema.StringAttribute{
			Description: "DNS-label-safe cluster name. Changes force replacement.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"region": schema.StringAttribute{
			Description: "Cloud-provider region identifier. Changes force replacement.",
			Required:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"use_case": schema.StringAttribute{
			Description: "Workload classification: general or production.",
			Optional:    true,
			Computed:    true,
		},
		"host_access_ssh_key": schema.StringAttribute{
			Description: "Name of the SSH key registered on the cloud account. " +
				"In-place changes route to POST /clusters/{id}/update-ssh-key.",
			Required: true,
		},
		"license": schema.StringAttribute{
			Description: "NC2 license tier. In-place changes route to POST /clusters/{id}/update-license.",
			Required:    true,
		},
		"aos_version": schema.StringAttribute{
			Description: "AOS software version. In-place changes route via update-license.",
			Required:    true,
		},
		"software_tier": schema.StringAttribute{
			Description: "Software tier. In-place changes route via update-license.",
			Required:    true,
		},
		"capacity": schema.ListAttribute{
			Description: "Ordered list of host groups; each item is a map of host attributes. " +
				"In-place changes route to POST /clusters/{id}/update-capacity.",
			ElementType: types.MapType{ElemType: types.StringType},
			Required:    true,
		},
		"redundancy": schema.MapAttribute{
			Description: "Redundancy descriptor (e.g. {factor = \"1\"}). Changes force replacement.",
			ElementType: types.StringType,
			Required:    true,
		},
		"network": schema.MapAttribute{
			Description: "Network configuration encoded as a flat map<string,string>. " +
				"Changes force replacement (with field-level exceptions).",
			ElementType: types.StringType,
			Required:    true,
		},
		"resource_tags": schema.MapAttribute{
			Description: "Cloud-side resource tags. In-place changes route to update-resource-tags.",
			ElementType: types.StringType,
			Optional:    true,
		},
		"desired_state": schema.StringAttribute{
			Description: "Desired runtime state: running or hibernated. " +
				"Transitions route to POST /clusters/{id}/hibernate or /resume.",
			Optional: true,
			Computed: true,
		},
		"state": schema.StringAttribute{
			Description: "Observed cluster lifecycle state.",
			Computed:    true,
		},
		"created_at": schema.StringAttribute{
			Description: "RFC 3339 cluster creation timestamp.",
			Computed:    true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"updated_at": schema.StringAttribute{
			Description: "RFC 3339 most-recent-modification timestamp.",
			Computed:    true,
		},
	}
}
