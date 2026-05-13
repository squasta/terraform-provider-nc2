// Package notification_acknowledge implements
// nc2_notification_acknowledge — PATCH /notifications/{id} with
// acknowledged=true.
package notification_acknowledge //nolint:revive,staticcheck

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/actions/shared"
	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// NewAction is the framework-facing constructor.
func NewAction() action.Action { return &ackAction{} }

type ackAction struct{ c *client.Client }

type model struct {
	NotificationID types.String `tfsdk:"notification_id"`
}

// Metadata returns the action type name.
func (a *ackAction) Metadata(_ context.Context, _ action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = "nc2_notification_acknowledge"
}

// Schema returns the action schema.
func (a *ackAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Acknowledges a notification by setting `acknowledged=true` via PATCH /notifications/{id}.",
		Attributes: map[string]schema.Attribute{
			"notification_id": schema.StringAttribute{Required: true, Description: "Target notification id."},
		},
	}
}

// Configure pulls the NC2 client off the framework runtime bundle.
func (a *ackAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	a.c = shared.ClientFromConfigure(req, resp, "nc2_notification_acknowledge")
}

// Invoke runs the action.
func (a *ackAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	shared.Invoke(ctx, a.c, shared.Spec{
		Method:      "PATCH",
		Path:        "/notifications/" + cfg.NotificationID.ValueString(),
		Body:        map[string]any{"acknowledged": true},
		PollTask:    false,
		TerraformOp: "nc2_notification_acknowledge.Invoke",
	}, resp)
}

var (
	_ action.Action              = &ackAction{}
	_ action.ActionWithConfigure = &ackAction{}
)
