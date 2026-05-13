// Package cluster implements data.nc2_cluster — single-cluster lookup
// by id.
package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/datasources/dsshared"
)

// NewDataSource is the framework-facing constructor.
func NewDataSource() datasource.DataSource { return &clusterDS{} }

type clusterDS struct{ c *client.Client }

type model struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	CloudAccountID  types.String `tfsdk:"cloud_account_id"`
	CloudProvider   types.String `tfsdk:"cloud_provider"`
	Region          types.String `tfsdk:"region"`
	State           types.String `tfsdk:"state"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

// Metadata returns the data source type name.
func (d *clusterDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_cluster"
}

// Schema is the framework schema.
func (d *clusterDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single NC2 cluster by id.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Required: true, Description: "Cluster UUID."},
			"name":             schema.StringAttribute{Computed: true, Description: "Cluster name."},
			"organization_id":  schema.StringAttribute{Computed: true, Description: "Owning organization UUID."},
			"cloud_account_id": schema.StringAttribute{Computed: true, Description: "Owning cloud account UUID."},
			"cloud_provider":   schema.StringAttribute{Computed: true, Description: "aws | azure | gcp."},
			"region":           schema.StringAttribute{Computed: true, Description: "Cloud-provider region."},
			"state":            schema.StringAttribute{Computed: true, Description: "Lifecycle state."},
			"created_at":       schema.StringAttribute{Computed: true, Description: "RFC 3339 created_at."},
		},
	}
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (d *clusterDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_cluster Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read calls GET /clusters/{id}.
func (d *clusterDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.c == nil {
		return
	}
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	rsp, err := d.c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/clusters/" + cfg.ID.ValueString(),
		TerraformOp: "data.nc2_cluster.Read",
	})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.Diagnostics.AddError("nc2_cluster not found", fmt.Sprintf("cluster %s does not exist", cfg.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("nc2_cluster Read", err.Error())
		return
	}
	updateModelFromResponse(&cfg, rsp.Body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

func updateModelFromResponse(m *model, body map[string]any) {
	data, ok := body["data"].(map[string]any)
	if !ok {
		return
	}
	if v, ok := data["id"].(string); ok {
		m.ID = types.StringValue(v)
	}
	if v, ok := data["name"].(string); ok {
		m.Name = types.StringValue(v)
	}
	if v, ok := data["organization_id"].(string); ok {
		m.OrganizationID = types.StringValue(v)
	}
	if v, ok := data["cloud_account_id"].(string); ok {
		m.CloudAccountID = types.StringValue(v)
	}
	if v, ok := data["cloud_provider"].(string); ok {
		m.CloudProvider = types.StringValue(v)
	}
	if v, ok := data["region"].(string); ok {
		m.Region = types.StringValue(v)
	}
	if v, ok := data["state"].(string); ok {
		m.State = types.StringValue(v)
	}
	if v, ok := data["created_at"].(string); ok {
		m.CreatedAt = types.StringValue(v)
	}
}

var (
	_ datasource.DataSource              = &clusterDS{}
	_ datasource.DataSourceWithConfigure = &clusterDS{}
)
