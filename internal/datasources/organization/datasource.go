// Package organization implements data.nc2_organization — a single
// NC2 organization fetched by id.
package organization

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
func NewDataSource() datasource.DataSource { return &orgDS{} }

type orgDS struct{ c *client.Client }

type model struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	State       types.String `tfsdk:"state"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *orgDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_organization"
}

func (d *orgDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single NC2 organization by id.",
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Required: true, Description: "Organization UUID."},
			"name":        schema.StringAttribute{Computed: true},
			"description": schema.StringAttribute{Computed: true},
			"state":       schema.StringAttribute{Computed: true},
			"created_at":  schema.StringAttribute{Computed: true},
			"updated_at":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *orgDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_organization Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read GETs /organizations/{id}.
func (d *orgDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Path:        "/organizations/" + cfg.ID.ValueString(),
		TerraformOp: "data.nc2_organization.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_organization Read", err.Error())
		return
	}
	data, _ := rsp.Body["data"].(map[string]any)
	if data == nil {
		resp.Diagnostics.AddError("nc2_organization Read", "response missing data envelope")
		return
	}
	cfg.Name = types.StringValue(dsshared.StringOrEmpty(data["name"]))
	cfg.Description = types.StringValue(dsshared.StringOrEmpty(data["description"]))
	cfg.State = types.StringValue(dsshared.StringOrEmpty(data["state"]))
	cfg.CreatedAt = types.StringValue(dsshared.StringOrEmpty(data["created_at"]))
	cfg.UpdatedAt = types.StringValue(dsshared.StringOrEmpty(data["updated_at"]))
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

var (
	_ datasource.DataSource              = &orgDS{}
	_ datasource.DataSourceWithConfigure = &orgDS{}
)
