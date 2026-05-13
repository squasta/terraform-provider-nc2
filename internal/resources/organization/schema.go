package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// resourceSchema mirrors §1 of data-model.md.
//
// Lifecycle: pending → active → terminating → terminated. Read after
// terminate eventually returns 404 → state removal (FR-009).
//
// `id` and `created_at` are immutable identifiers and must use
// `UseStateForUnknown` so a subsequent plan after Apply does not
// mark them unknown.
func resourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "An NC2 organization — root of the resource hierarchy and unit of billing/audit.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Organization UUID assigned by NC2.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable organization name (1–128 chars).",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Organization description (≤ 2048 chars).",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Lifecycle state: active | terminating | terminated.",
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
