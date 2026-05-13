// Package cloud_account implements data.nc2_cloud_account — a
// single NC2 cloud account fetched by id.
package cloud_account //nolint:revive,staticcheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/datasources/dsshared"
)

// NewDataSource is the framework-facing constructor.
func NewDataSource() datasource.DataSource { return &cloudAccountDS{} }

type cloudAccountDS struct{ c *client.Client }

type model struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	State          types.String `tfsdk:"state"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (d *cloudAccountDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_cloud_account"
}

func (d *cloudAccountDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single NC2 cloud account by id. Credentials are NOT returned.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, Description: "Cloud account UUID."},
			"organization_id": schema.StringAttribute{Computed: true},
			"cloud_provider":  schema.StringAttribute{Computed: true},
			"name":            schema.StringAttribute{Computed: true},
			"description":     schema.StringAttribute{Computed: true},
			"state":           schema.StringAttribute{Computed: true},
			"created_at":      schema.StringAttribute{Computed: true},
			"updated_at":      schema.StringAttribute{Computed: true},
		},
	}
}

func (d *cloudAccountDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_cloud_account Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read GETs /cloud-accounts/{id}.
func (d *cloudAccountDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Path:        "/cloud-accounts/" + cfg.ID.ValueString(),
		TerraformOp: "data.nc2_cloud_account.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_cloud_account Read", err.Error())
		return
	}
	data, _ := rsp.Body["data"].(map[string]any)
	if data == nil {
		resp.Diagnostics.AddError("nc2_cloud_account Read", "response missing data envelope")
		return
	}
	cfg.OrganizationID = types.StringValue(dsshared.StringOrEmpty(data["organization_id"]))
	cfg.CloudProvider = types.StringValue(dsshared.StringOrEmpty(data["cloud_provider"]))
	cfg.Name = types.StringValue(dsshared.StringOrEmpty(data["name"]))
	cfg.Description = types.StringValue(dsshared.StringOrEmpty(data["description"]))
	cfg.State = types.StringValue(dsshared.StringOrEmpty(data["state"]))
	cfg.CreatedAt = types.StringValue(dsshared.StringOrEmpty(data["created_at"]))
	cfg.UpdatedAt = types.StringValue(dsshared.StringOrEmpty(data["updated_at"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

var (
	_ datasource.DataSource              = &cloudAccountDS{}
	_ datasource.DataSourceWithConfigure = &cloudAccountDS{}
)
