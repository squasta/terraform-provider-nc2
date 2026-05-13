package gcp_cluster //nolint:revive,staticcheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
	"github.com/nutanix/terraform-provider-nc2/internal/resources/orgshared"
)

// NewResource is the framework-facing constructor.
func NewResource() resource.Resource { return &gcpClusterResource{} }

type gcpClusterResource struct{ c *client.Client }

var crudSpec = clustershared.CRUD{
	CreatePath:           "/clusters/gcp",
	CreateTerraformOp:    "nc2_gcp_cluster.Create",
	ReadTerraformOp:      "nc2_gcp_cluster.Read",
	UpdateTerraformOp:    "nc2_gcp_cluster.Update",
	DeleteTerraformOp:    "nc2_gcp_cluster.Delete",
	SupportsAccessPolicy: false,
}

// Metadata returns the resource type name.
func (r *gcpClusterResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "nc2_gcp_cluster"
}

// Schema returns the GCP-cluster schema.
func (r *gcpClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema(ctx)
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (r *gcpClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(orgshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_gcp_cluster Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	r.c = rt.NC2Client()
}

func (r *gcpClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.c == nil {
		resp.Diagnostics.AddError("nc2_gcp_cluster", "no NC2 client; provider not configured")
		return
	}
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := buildCreateBody(&plan)
	id, diags := clustershared.CreateCluster(ctx, r.c, crudSpec, body)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, diags := clustershared.ReadCluster(ctx, r.c, crudSpec, id, &plan.Model); len(diags) > 0 {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gcpClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.c == nil {
		return
	}
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	notFound, diags := clustershared.ReadCluster(ctx, r.c, crudSpec, state.ID.ValueString(), &state.Model)
	resp.Diagnostics.Append(diags...)
	if notFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gcpClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.c == nil {
		return
	}
	var plan, state model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	diff := computeDiff(&state, &plan)
	ops, err := clustershared.RouteUpdate(diff)
	if err != nil {
		resp.Diagnostics.AddError("nc2_gcp_cluster Update", err.Error())
		return
	}
	bodies := buildUpdateBodies(&state, &plan)
	resp.Diagnostics.Append(clustershared.ApplyUpdates(ctx, r.c, crudSpec, state.ID.ValueString(), ops, bodies)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, diags := clustershared.ReadCluster(ctx, r.c, crudSpec, state.ID.ValueString(), &plan.Model); len(diags) > 0 {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gcpClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.c == nil {
		return
	}
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(clustershared.DeleteCluster(ctx, r.c, crudSpec, state.ID.ValueString())...)
}

func (r *gcpClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// ValidateConfig enforces NC2 cluster sizing rules at plan time:
// the sum of `capacity[*].number_of_hosts` MUST be 1 or 3..28
// inclusive (a 2-host cluster has no quorum). The per-element and
// aggregate checks live in `clustershared.ValidateCapacityHostCount`
// and are shared by every cloud's cluster resource.
func (r *gcpClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(clustershared.ValidateCapacityHostCount(cfg.Capacity)...)
}

var (
	_ resource.Resource                   = &gcpClusterResource{}
	_ resource.ResourceWithConfigure      = &gcpClusterResource{}
	_ resource.ResourceWithImportState    = &gcpClusterResource{}
	_ resource.ResourceWithValidateConfig = &gcpClusterResource{}
)
