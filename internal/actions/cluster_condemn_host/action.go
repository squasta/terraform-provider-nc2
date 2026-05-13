// Package cluster_condemn_host implements nc2_cluster_condemn_host —
// the action that asks NC2 to remove a specific host from a cluster.
package cluster_condemn_host //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/actions/shared"
	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// NewAction is the framework-facing constructor.
func NewAction() action.Action { return &condemnHostAction{} }

type condemnHostAction struct{ c *client.Client }

type model struct {
	ClusterID types.String `tfsdk:"cluster_id"`
	HostID    types.String `tfsdk:"host_id"`
}

// Metadata returns the action type name.
func (a *condemnHostAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "nc2_cluster_condemn_host"
}

// Schema returns the action schema.
func (a *condemnHostAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Marks a specific host inside the cluster as condemned. Issues POST /clusters/{id}/condemn-host and polls the returned task.",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{Required: true, Description: "Target cluster UUID."},
			"host_id":    schema.StringAttribute{Required: true, Description: "Target host id."},
		},
	}
}

// Configure pulls the NC2 client off the framework runtime bundle.
func (a *condemnHostAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.c = shared.ClientFromConfigure(req, resp, "nc2_cluster_condemn_host")
}

// Invoke runs the action.
func (a *condemnHostAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	shared.Invoke(ctx, a.c, shared.Spec{
		Method:      "POST",
		Path:        "/clusters/" + cfg.ClusterID.ValueString() + "/condemn-host",
		Body:        map[string]any{"host_id": cfg.HostID.ValueString()},
		PollTask:    true,
		TerraformOp: "nc2_cluster_condemn_host.Invoke",
	}, resp)
}

var (
	_ action.Action              = &condemnHostAction{}
	_ action.ActionWithConfigure = &condemnHostAction{}
)
