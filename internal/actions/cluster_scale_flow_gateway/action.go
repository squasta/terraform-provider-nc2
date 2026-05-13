// Package cluster_scale_flow_gateway implements
// nc2_cluster_scale_flow_gateway.
package cluster_scale_flow_gateway //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/actions/shared"
	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// NewAction is the framework-facing constructor.
func NewAction() action.Action { return &scaleAction{} }

type scaleAction struct{ c *client.Client }

type model struct {
	ClusterID       types.String `tfsdk:"cluster_id"`
	TargetNodeCount types.Int64  `tfsdk:"target_node_count"`
}

// Metadata returns the action type name.
func (a *scaleAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "nc2_cluster_scale_flow_gateway"
}

// Schema returns the action schema.
func (a *scaleAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Scales the cluster's Flow Gateway out (or in) to `target_node_count`. Issues POST /clusters/{id}/scale-out-fgw.",
		Attributes: map[string]schema.Attribute{
			"cluster_id":        schema.StringAttribute{Required: true, Description: "Target cluster UUID."},
			"target_node_count": schema.Int64Attribute{Required: true, Description: "Desired Flow Gateway node count."},
		},
	}
}

// Configure pulls the NC2 client off the framework runtime bundle.
func (a *scaleAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.c = shared.ClientFromConfigure(req, resp, "nc2_cluster_scale_flow_gateway")
}

// Invoke runs the action.
func (a *scaleAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	shared.Invoke(ctx, a.c, shared.Spec{
		Method:      "POST",
		Path:        "/clusters/" + cfg.ClusterID.ValueString() + "/scale-out-fgw",
		Body:        map[string]any{"target_node_count": cfg.TargetNodeCount.ValueInt64()},
		PollTask:    true,
		TerraformOp: "nc2_cluster_scale_flow_gateway.Invoke",
	}, resp)
}

var (
	_ action.Action              = &scaleAction{}
	_ action.ActionWithConfigure = &scaleAction{}
)
