// Package cluster_extend_support_tunnel implements
// nc2_cluster_extend_support_tunnel.
package cluster_extend_support_tunnel //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/actions/shared"
	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// NewAction is the framework-facing constructor.
func NewAction() action.Action { return &extAction{} }

type extAction struct{ c *client.Client }

type model struct {
	ClusterID     types.String `tfsdk:"cluster_id"`
	DurationHours types.Int64  `tfsdk:"duration_hours"`
}

// Metadata returns the action type name.
func (a *extAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "nc2_cluster_extend_support_tunnel"
}

// Schema returns the action schema.
func (a *extAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Extends an open Support Tunnel by `duration_hours`. Issues POST /clusters/{id}/extend-support-tunnel.",
		Attributes: map[string]schema.Attribute{
			"cluster_id":     schema.StringAttribute{Required: true, Description: "Target cluster UUID."},
			"duration_hours": schema.Int64Attribute{Required: true, Description: "Hours to extend the support tunnel for."},
		},
	}
}

// Configure pulls the NC2 client off the framework runtime bundle.
func (a *extAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.c = shared.ClientFromConfigure(req, resp, "nc2_cluster_extend_support_tunnel")
}

// Invoke runs the action.
func (a *extAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	shared.Invoke(ctx, a.c, shared.Spec{
		Method:      "POST",
		Path:        "/clusters/" + cfg.ClusterID.ValueString() + "/extend-support-tunnel",
		Body:        map[string]any{"duration_hours": cfg.DurationHours.ValueInt64()},
		PollTask:    true,
		TerraformOp: "nc2_cluster_extend_support_tunnel.Invoke",
	}, resp)
}

var (
	_ action.Action              = &extAction{}
	_ action.ActionWithConfigure = &extAction{}
)
