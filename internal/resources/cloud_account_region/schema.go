package cloud_account_region //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// resourceSchema mirrors §3 of data-model.md.
//
// `region` and `cloud_account_id` are RequiresReplace because NC2
// has no API to move a region between cloud accounts and the
// `region` value is used in the path of the create call.
func resourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "A cloud-provider region enabled on an NC2 cloud account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Region resource id assigned by NC2.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cloud_account_id": schema.StringAttribute{
				Required:    true,
				Description: "Owning cloud account UUID. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "Cloud-provider region identifier (e.g., us-east-1, eastus, europe-west1). RequiresReplace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Region operational state: available | disabled.",
			},
		},
	}
}
