// Package organization implements the nc2_organization managed
// resource. See data-model.md §1 for the schema and lifecycle.
package organization

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
func NewResource() resource.Resource { return &orgResource{} }

type orgResource struct{ c *client.Client }

// model is the framework-decoded view of an organization.
type model struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *orgResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "nc2_organization"
}

// Schema returns the resource schema.
func (r *orgResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema(ctx)
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (r *orgResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(orgshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_organization Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	r.c = rt.NC2Client()
}

// Create POSTs /organizations and decodes the response into state.
//
// NC2 returns the new organization synchronously (no task polling
// for org create per the OpenAPI contract).
func (r *orgResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.c == nil {
		resp.Diagnostics.AddError("nc2_organization", "no NC2 client; provider not configured")
		return
	}
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := buildBody(&plan)
	rsp, err := r.c.Do(ctx, client.Request{
		Method:      "POST",
		Path:        "/organizations",
		Body:        map[string]any{"data": body},
		TerraformOp: "nc2_organization.Create",
	})
	if err != nil {
		resp.Diagnostics.AddError("organization create failed", err.Error())
		return
	}
	resp.Diagnostics.Append(decodeInto(rsp.Body, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the resource from /organizations/{id}.
//
// 404 → state removal (FR-009).
func (r *orgResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
		Path:        "/organizations/" + state.ID.ValueString(),
		TerraformOp: "nc2_organization.Read",
	})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("organization read failed", err.Error())
		return
	}
	resp.Diagnostics.Append(decodeInto(rsp.Body, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update PATCHes /organizations/{id} when name or description change.
// Other attributes are computed and never participate in updates.
func (r *orgResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.c == nil {
		return
	}
	var plan, state model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	body := buildBody(&plan)
	rsp, err := r.c.Do(ctx, client.Request{
		Method:      "PATCH",
		Path:        "/organizations/" + state.ID.ValueString(),
		Body:        map[string]any{"data": body},
		TerraformOp: "nc2_organization.Update",
	})
	if err != nil {
		resp.Diagnostics.AddError("organization update failed", err.Error())
		return
	}
	resp.Diagnostics.Append(decodeInto(rsp.Body, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete PATCHes /organizations/{id}/terminate. NC2 returns 200 once
// the terminate request is accepted; the org transitions through
// `terminating` → `terminated`. Subsequent Read returns 404 → state
// is removed (FR-009).
func (r *orgResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.c == nil {
		return
	}
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.c.Do(ctx, client.Request{
		Method:      "PATCH",
		Path:        "/organizations/" + state.ID.ValueString() + "/terminate",
		TerraformOp: "nc2_organization.Delete",
	})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return
		}
		resp.Diagnostics.AddError("organization terminate failed", err.Error())
		return
	}
}

// ImportState makes the organization id the import key.
func (r *orgResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// buildBody assembles the request payload for create / update from
// the user-provided fields. Computed fields are never echoed back.
func buildBody(m *model) map[string]any {
	body := map[string]any{}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		body["name"] = m.Name.ValueString()
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		body["description"] = m.Description.ValueString()
	}
	return body
}

// decodeInto populates the framework model from the canonical NC2
// `{"data": {...}}` envelope.
func decodeInto(body map[string]any, m *model) diag.Diagnostics {
	var diags diag.Diagnostics
	data, ok := body["data"].(map[string]any)
	if !ok {
		diags.AddError("organization decode", "response missing data envelope")
		return diags
	}
	if v, ok := data["id"].(string); ok {
		m.ID = types.StringValue(v)
	}
	if v, ok := data["name"].(string); ok {
		m.Name = types.StringValue(v)
	}
	if v, ok := data["description"].(string); ok {
		m.Description = types.StringValue(v)
	} else if m.Description.IsUnknown() {
		// Preserve a previously-set null when API omits the field.
		m.Description = types.StringNull()
	}
	if v, ok := data["state"].(string); ok {
		m.State = types.StringValue(v)
	}
	if v, ok := data["created_at"].(string); ok {
		m.CreatedAt = types.StringValue(v)
	}
	if v, ok := data["updated_at"].(string); ok {
		m.UpdatedAt = types.StringValue(v)
	}
	return diags
}

var (
	_ resource.Resource                = &orgResource{}
	_ resource.ResourceWithConfigure   = &orgResource{}
	_ resource.ResourceWithImportState = &orgResource{}
)
