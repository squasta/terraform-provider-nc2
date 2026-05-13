// Package cloud_account implements the nc2_cloud_account managed
// resource. See data-model.md §2 for the schema, lifecycle, and the
// no-DELETE caveat (FR-006).
package cloud_account //nolint:revive,staticcheck

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/resources/orgshared"
)

// NewResource is the framework-facing constructor.
func NewResource() resource.Resource { return &cloudAccountResource{} }

type cloudAccountResource struct{ c *client.Client }

// model is the framework-decoded view of a cloud account.
type model struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Credentials    types.Map    `tfsdk:"credentials"`
	State          types.String `tfsdk:"state"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *cloudAccountResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "nc2_cloud_account"
}

// Schema returns the resource schema.
func (r *cloudAccountResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema(ctx)
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (r *cloudAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(orgshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_cloud_account Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	r.c = rt.NC2Client()
}

// Create POSTs /organizations/{id}/cloud-accounts/{cloud_provider}.
func (r *cloudAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.c == nil {
		resp.Diagnostics.AddError("nc2_cloud_account", "no NC2 client; provider not configured")
		return
	}
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := buildCreateBody(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	rsp, err := r.c.Do(ctx, client.Request{
		Method: "POST",
		Path: fmt.Sprintf("/organizations/%s/cloud-accounts/%s",
			plan.OrganizationID.ValueString(), plan.CloudProvider.ValueString()),
		Body:        map[string]any{"data": body},
		TerraformOp: "nc2_cloud_account.Create",
	})
	if err != nil {
		resp.Diagnostics.AddError("cloud account create failed", err.Error())
		return
	}
	resp.Diagnostics.Append(decodeInto(ctx, rsp.Body, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the resource from /cloud-accounts/{id}. 404 → state
// removal (FR-009).
func (r *cloudAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
		Path:        "/cloud-accounts/" + state.ID.ValueString(),
		TerraformOp: "nc2_cloud_account.Read",
	})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("cloud account read failed", err.Error())
		return
	}
	resp.Diagnostics.Append(decodeInto(ctx, rsp.Body, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update routes the diff to the right NC2 endpoint:
//   - credentials change → POST /cloud-accounts/{id}/update-credentials
//   - name / description change → PATCH /cloud-accounts/{id}
//   - both → credentials first, then PATCH (deterministic order; see
//     data-model.md §2 "Update routing").
func (r *cloudAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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

	credsChanged := !mapsEqual(state.Credentials, plan.Credentials)
	metaChanged := !plan.Name.Equal(state.Name) || !plan.Description.Equal(state.Description)

	if credsChanged {
		credsBody, diags := credentialsRequestBody(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		_, err := r.c.Do(ctx, client.Request{
			Method:      "POST",
			Path:        "/cloud-accounts/" + state.ID.ValueString() + "/update-credentials",
			Body:        map[string]any{"data": credsBody},
			TerraformOp: "nc2_cloud_account.Update.credentials",
		})
		if err != nil {
			resp.Diagnostics.AddError("cloud account update_credentials failed", err.Error())
			return
		}
	}
	if metaChanged {
		metaBody := map[string]any{}
		if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
			metaBody["name"] = plan.Name.ValueString()
		}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			metaBody["description"] = plan.Description.ValueString()
		}
		_, err := r.c.Do(ctx, client.Request{
			Method:      "PATCH",
			Path:        "/cloud-accounts/" + state.ID.ValueString(),
			Body:        map[string]any{"data": metaBody},
			TerraformOp: "nc2_cloud_account.Update",
		})
		if err != nil {
			resp.Diagnostics.AddError("cloud account update failed", err.Error())
			return
		}
	}

	rsp, err := r.c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/cloud-accounts/" + state.ID.ValueString(),
		TerraformOp: "nc2_cloud_account.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("cloud account refresh failed", err.Error())
		return
	}
	resp.Diagnostics.Append(decodeInto(ctx, rsp.Body, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is state-only removal per FR-006 (NC2 has no DELETE for
// cloud accounts). A plan-time warning is emitted via ModifyPlan to
// make the no-API-call semantics visible to operators.
func (r *cloudAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"nc2_cloud_account: state-only destroy",
		"NC2 does not expose a DELETE endpoint for cloud accounts (FR-006). Terraform removes the resource from state only; the cloud account remains in NC2 until you delete it via the UI or another channel.",
	)
}

// ModifyPlan surfaces the destroy-without-DELETE caveat at plan time
// per FR-006 / R-07. When the plan removes the resource, we attach
// a warning under the resource address.
func (r *cloudAccountResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() && !req.State.Raw.IsNull() {
		resp.Diagnostics.AddWarning(
			"nc2_cloud_account: plan-time destroy notice",
			"This destroy will only remove the resource from Terraform state; the underlying NC2 cloud account will remain (FR-006).",
		)
	}
}

// ImportState makes the cloud account id the import key.
func (r *cloudAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// buildCreateBody builds the POST body. `credentials` is dropped at
// the data envelope's top level (NC2 expects them in the request
// body alongside name/description for create; per the OpenAPI spec
// the spelling is the same as the credentials map keys).
func buildCreateBody(ctx context.Context, m *model) (map[string]any, diag.Diagnostics) {
	body := map[string]any{}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		body["name"] = m.Name.ValueString()
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		body["description"] = m.Description.ValueString()
	}
	creds, diags := credentialsRequestBody(ctx, m)
	if diags.HasError() {
		return nil, diags
	}
	for k, v := range creds {
		body[k] = v
	}
	return body, diags
}

// credentialsRequestBody flattens the typed Map into a
// map[string]any acceptable to NC2's update-credentials endpoint.
func credentialsRequestBody(ctx context.Context, m *model) (map[string]any, diag.Diagnostics) {
	out := map[string]any{}
	if m.Credentials.IsNull() || m.Credentials.IsUnknown() {
		return out, nil
	}
	raw := map[string]string{}
	diags := m.Credentials.ElementsAs(ctx, &raw, false)
	if diags.HasError() {
		return nil, diags
	}
	for k, v := range raw {
		out[k] = v
	}
	return out, diags
}

// decodeInto populates the framework model from the canonical NC2
// `{"data": {...}}` envelope. Credentials are NOT echoed by NC2
// (sensitive); the existing state value is preserved.
func decodeInto(_ context.Context, body map[string]any, m *model) diag.Diagnostics {
	var diags diag.Diagnostics
	data, ok := body["data"].(map[string]any)
	if !ok {
		diags.AddError("cloud account decode", "response missing data envelope")
		return diags
	}
	if v, ok := data["id"].(string); ok {
		m.ID = types.StringValue(v)
	}
	if v, ok := data["organization_id"].(string); ok {
		m.OrganizationID = types.StringValue(v)
	}
	if v, ok := data["cloud_provider"].(string); ok {
		m.CloudProvider = types.StringValue(v)
	}
	if v, ok := data["name"].(string); ok {
		m.Name = types.StringValue(v)
	}
	if v, ok := data["description"].(string); ok {
		m.Description = types.StringValue(v)
	} else if m.Description.IsUnknown() {
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
	if m.Credentials.IsNull() || m.Credentials.IsUnknown() {
		// NC2 never echoes credentials; provide an empty map so
		// state is not "unknown".
		empty, d := types.MapValue(stringType(), map[string]attr.Value{})
		diags.Append(d...)
		m.Credentials = empty
	}
	return diags
}

// mapsEqual compares two types.Map values element-by-element. Nil
// and empty maps compare equal.
func mapsEqual(a, b types.Map) bool {
	if a.IsNull() && b.IsNull() {
		return true
	}
	if a.IsUnknown() || b.IsUnknown() {
		return false
	}
	return reflect.DeepEqual(a.Elements(), b.Elements())
}

// stringType is the element type for the credentials map.
func stringType() attr.Type { return types.StringType }

var (
	_ resource.Resource                = &cloudAccountResource{}
	_ resource.ResourceWithConfigure   = &cloudAccountResource{}
	_ resource.ResourceWithImportState = &cloudAccountResource{}
	_ resource.ResourceWithModifyPlan  = &cloudAccountResource{}
)
