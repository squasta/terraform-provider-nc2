// Package cluster_open_support_tunnel implements the
// nc2_cluster_open_support_tunnel action.
package cluster_open_support_tunnel //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/actions/shared"
	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// NewAction is the framework-facing constructor.
func NewAction() action.Action { return &openSupportTunnelAction{} }

type openSupportTunnelAction struct{ c *client.Client }

type model struct {
	ClusterID types.String `tfsdk:"cluster_id"`
}

// Metadata returns the action type name.
func (a *openSupportTunnelAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "nc2_cluster_open_support_tunnel"
}

// Schema returns the action schema.
func (a *openSupportTunnelAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Opens a Support Tunnel on the cluster. Issues POST /clusters/{id}/open-support-tunnel and polls the returned task.",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{Required: true, Description: "Target cluster UUID."},
		},
	}
}

// Configure pulls the NC2 client off the framework runtime bundle.
func (a *openSupportTunnelAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.c = shared.ClientFromConfigure(req, resp, "nc2_cluster_open_support_tunnel")
}

// Invoke runs the action.
func (a *openSupportTunnelAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	shared.Invoke(ctx, a.c, shared.Spec{
		Method:      "POST",
		Path:        "/clusters/" + cfg.ClusterID.ValueString() + "/open-support-tunnel",
		PollTask:    true,
		TerraformOp: "nc2_cluster_open_support_tunnel.Invoke",
	}, resp)
}

var (
	_ action.Action              = &openSupportTunnelAction{}
	_ action.ActionWithConfigure = &openSupportTunnelAction{}
)
