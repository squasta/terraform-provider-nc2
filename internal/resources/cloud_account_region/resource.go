// Package cloud_account_region implements the nc2_cloud_account_region
// managed resource. See data-model.md §3 and FR-007 for the
// no-DELETE caveat.
package cloud_account_region //nolint:revive,staticcheck

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/resources/orgshared"
)

// NewResource is the framework-facing constructor.
func NewResource() resource.Resource { return &regionResource{} }

type regionResource struct{ c *client.Client }

type model struct {
	ID             types.String `tfsdk:"id"`
	CloudAccountID types.String `tfsdk:"cloud_account_id"`
	Region         types.String `tfsdk:"region"`
	State          types.String `tfsdk:"state"`
}

// Metadata returns the resource type name.
func (r *regionResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "nc2_cloud_account_region"
}

// Schema returns the resource schema.
func (r *regionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema(ctx)
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (r *regionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(orgshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_cloud_account_region Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	r.c = rt.NC2Client()
}

// Create POSTs /cloud-accounts/{cloud_account_id}/regions.
func (r *regionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.c == nil {
		resp.Diagnostics.AddError("nc2_cloud_account_region", "no NC2 client; provider not configured")
		return
	}
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rsp, err := r.c.Do(ctx, client.Request{
		Method:      "POST",
		Path:        "/cloud-accounts/" + plan.CloudAccountID.ValueString() + "/regions",
		Body:        map[string]any{"data": map[string]any{"region": plan.Region.ValueString()}},
		TerraformOp: "nc2_cloud_account_region.Create",
	})
	if err != nil {
		resp.Diagnostics.AddError("region create failed", err.Error())
		return
	}
	resp.Diagnostics.Append(decodeInto(rsp.Body, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read finds the region inside the list returned by
// GET /cloud-accounts/{cloud_account_id}/regions. NC2 has no
// /regions/{id} endpoint; we filter on the client side. 404 →
// state removal (FR-009).
func (r *regionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.c == nil {
		return
	}
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rsp, err := r.c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/cloud-accounts/" + state.CloudAccountID.ValueString() + "/regions",
		TerraformOp: "nc2_cloud_account_region.Read",
	})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("region read failed", err.Error())
		return
	}
	row, ok := findRegionByID(rsp.Body, state.ID.ValueString())
	if !ok {
		// region was removed out-of-band; clean up state.
		resp.State.RemoveResource(ctx)
		return
	}
	if v, ok := row["state"].(string); ok {
		state.State = types.StringValue(v)
	}
	if v, ok := row["region"].(string); ok {
		state.Region = types.StringValue(v)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op: every mutable attribute is `RequiresReplace`,
// so the framework dispatches Delete+Create instead of calling
// Update.
func (r *regionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is state-only removal per FR-007 (NC2 has no DELETE for
// regions). A plan-time warning is emitted via ModifyPlan.
func (r *regionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"nc2_cloud_account_region: state-only destroy",
		"NC2 does not expose a DELETE endpoint for cloud account regions (FR-007). Terraform removes the resource from state only; the region remains enabled in NC2.",
	)
}

// ModifyPlan surfaces the destroy-without-DELETE caveat at plan time.
func (r *regionResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() && !req.State.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"nc2_cloud_account_region: plan-time destroy notice",
			"This destroy will only remove the resource from Terraform state; the underlying NC2 region will remain enabled (FR-007).",
		)
	}
}

// ImportState format: `<cloud_account_id>/<region_id>`. Required
// because the NC2 API for regions is keyed by both cloud account
// and region id.
func (r *regionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := splitImportID(req.ID)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("nc2_cloud_account_region import",
			"expected import ID in the form <cloud_account_id>/<region_id>")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cloud_account_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// findRegionByID walks the canonical NC2 list envelope and returns
// the matching region row (or false when not found).
func findRegionByID(body map[string]any, id string) (map[string]any, bool) {
	raw, ok := body["data"].([]any)
	if !ok {
		// Some create endpoints return data as a single object.
		if obj, ok := body["data"].(map[string]any); ok {
			if got, _ := obj["id"].(string); got == id {
				return obj, true
			}
		}
		return nil, false
	}
	for _, item := range raw {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if got, _ := row["id"].(string); got == id {
			return row, true
		}
	}
	return nil, false
}

// decodeInto handles both the create response (single object) and
// the list response.
func decodeInto(body map[string]any, m *model) diag.Diagnostics {
	var diags diag.Diagnostics
	data, ok := body["data"].(map[string]any)
	if !ok {
		// Some POSTs return {data: [{...}]}; pick the first row.
		if list, ok := body["data"].([]any); ok && len(list) > 0 {
			if row, ok := list[0].(map[string]any); ok {
				data = row
			}
		}
	}
	if data == nil {
		diags.AddError("region decode", "response missing data envelope")
		return diags
	}
	if v, ok := data["id"].(string); ok {
		m.ID = types.StringValue(v)
	}
	if v, ok := data["region"].(string); ok {
		m.Region = types.StringValue(v)
	}
	if v, ok := data["state"].(string); ok {
		m.State = types.StringValue(v)
	}
	return diags
}

// splitImportID is a tiny helper kept package-private for testability.
func splitImportID(id string) []string {
	out := []string{}
	cur := ""
	for _, ch := range id {
		if ch == '/' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(ch)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

var (
	_ resource.Resource                = &regionResource{}
	_ resource.ResourceWithConfigure   = &regionResource{}
	_ resource.ResourceWithImportState = &regionResource{}
	_ resource.ResourceWithModifyPlan  = &regionResource{}
)
