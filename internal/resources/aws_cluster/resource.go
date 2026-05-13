package aws_cluster //nolint:revive,staticcheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
	"github.com/nutanix/terraform-provider-nc2/internal/resources/orgshared"
)

// NewResource is the framework-facing constructor.
func NewResource() resource.Resource { return &awsClusterResource{} }

type awsClusterResource struct{ c *client.Client }

var crudSpec = clustershared.CRUD{
	CreatePath:           "/clusters/aws",
	CreateTerraformOp:    "nc2_aws_cluster.Create",
	ReadTerraformOp:      "nc2_aws_cluster.Read",
	UpdateTerraformOp:    "nc2_aws_cluster.Update",
	DeleteTerraformOp:    "nc2_aws_cluster.Delete",
	SupportsAccessPolicy: true,
}

// Metadata returns the resource type name.
func (r *awsClusterResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "nc2_aws_cluster"
}

// Schema returns the AWS-cluster schema.
func (r *awsClusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceSchema(ctx)
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (r *awsClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(orgshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_aws_cluster Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	r.c = rt.NC2Client()
}

// Create POSTs /clusters/aws and waits for the task to complete.
func (r *awsClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.c == nil {
		resp.Diagnostics.AddError("nc2_aws_cluster", "no NC2 client; provider not configured")
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
	plan.ID = stringValue(id)
	if _, diags := clustershared.ReadHibernatingCluster(ctx, r.c, crudSpec, id, &plan.HibernatingModel); len(diags) > 0 {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the resource from /clusters/{id}.
func (r *awsClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.c == nil {
		return
	}
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	notFound, diags := clustershared.ReadHibernatingCluster(ctx, r.c, crudSpec, state.ID.ValueString(), &state.HibernatingModel)
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

// Update routes the diff through clustershared.RouteUpdate and
// applies each operation in order.
func (r *awsClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
		resp.Diagnostics.AddError("nc2_aws_cluster Update", err.Error())
		return
	}
	bodies := buildUpdateBodies(&state, &plan)
	resp.Diagnostics.Append(clustershared.ApplyUpdates(ctx, r.c, crudSpec, state.ID.ValueString(), ops, bodies)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if _, diags := clustershared.ReadHibernatingCluster(ctx, r.c, crudSpec, state.ID.ValueString(), &plan.HibernatingModel); len(diags) > 0 {
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete POSTs /clusters/{id}/terminate and waits for the task.
func (r *awsClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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

// ImportState makes the cluster id the import key.
func (r *awsClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// ValidateConfig enforces NC2 cluster sizing rules at plan time:
// the sum of `capacity[*].number_of_hosts` MUST be 1 or 3..28
// inclusive (a 2-host cluster has no quorum). The per-element and
// aggregate checks live in `clustershared.ValidateCapacityHostCount`
// and are shared by every cloud's cluster resource.
func (r *awsClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(clustershared.ValidateCapacityHostCount(cfg.Capacity)...)
}

var (
	_ resource.Resource                   = &awsClusterResource{}
	_ resource.ResourceWithConfigure      = &awsClusterResource{}
	_ resource.ResourceWithImportState    = &awsClusterResource{}
	_ resource.ResourceWithValidateConfig = &awsClusterResource{}
	_ rschema.Schema                      = rschema.Schema{}
)
