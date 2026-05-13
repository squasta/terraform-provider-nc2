package cloud_account //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// resourceSchema mirrors §2 of data-model.md.
//
// The `credentials` attribute is a dynamic Map<String,String> rather
// than a strongly typed nested object so the user can provide the
// per-cloud subfields verbatim from their OpenAPI documentation.
// Marking the parent `credentials` map sensitive cascades to every
// child value at render time, satisfying FR-002 without forcing a
// per-cloud validator-explosion in the schema layer.
func resourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "An NC2 cloud account — a credentialed AWS / Azure / GCP account that NC2 uses to provision infrastructure.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Cloud account UUID assigned by NC2.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required:    true,
				Description: "Owning organization UUID. Changing this destroys and recreates the cloud account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_provider": schema.StringAttribute{
				Required:    true,
				Description: "One of `aws`, `azure`, `gcp`. RequiresReplace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable cloud account name (1–128 chars).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Cloud account description (≤ 2048 chars).",
			},
			"credentials": schema.MapAttribute{
				Required:    true,
				Sensitive:   true,
				ElementType: stringType(),
				Description: "Cloud-provider credentials (sensitive; per-cloud shape per the NC2 OpenAPI spec).",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Operational state: active | disabled.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 creation timestamp.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC 3339 last-update timestamp.",
			},
		},
	}
}
