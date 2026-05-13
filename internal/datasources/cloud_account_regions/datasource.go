// Package cloud_account_regions implements
// data.nc2_cloud_account_regions — sorted-by-id list of every
// region enabled on a given NC2 cloud account.
package cloud_account_regions //nolint:revive,staticcheck

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/datasources/dsshared"
)

// NewDataSource is the framework-facing constructor.
func NewDataSource() datasource.DataSource { return &regionsDS{} }

type regionsDS struct{ c *client.Client }

type model struct {
	CloudAccountID types.String `tfsdk:"cloud_account_id"`
	Regions        types.List   `tfsdk:"regions"`
}

func (d *regionsDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_cloud_account_regions"
}

func (d *regionsDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every region enabled on an NC2 cloud account. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"cloud_account_id": schema.StringAttribute{Required: true, Description: "Owning cloud account UUID."},
			"regions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All enabled regions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":     schema.StringAttribute{Computed: true, Description: "Region resource id."},
						"region": schema.StringAttribute{Computed: true, Description: "Cloud-provider region name."},
						"state":  schema.StringAttribute{Computed: true, Description: "Region operational state."},
					},
				},
			},
		},
	}
}

func (d *regionsDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_cloud_account_regions Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read GETs /cloud-accounts/{cloud_account_id}/regions.
func (d *regionsDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Path:        "/cloud-accounts/" + cfg.CloudAccountID.ValueString() + "/regions",
		TerraformOp: "data.nc2_cloud_account_regions.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_cloud_account_regions Read", err.Error())
		return
	}
	items := dsshared.ExtractList(rsp.Body)
	dsshared.SortByID(items)

	objType := regionObjectType()
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":     types.StringValue(dsshared.StringOrEmpty(it["id"])),
			"region": types.StringValue(dsshared.StringOrEmpty(it["region"])),
			"state":  types.StringValue(dsshared.StringOrEmpty(it["state"])),
		})
		resp.Diagnostics.Append(diag...)
		values = append(values, v)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	listVal, diag := types.ListValue(objType, values)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model{CloudAccountID: cfg.CloudAccountID, Regions: listVal})...)
}

func regionObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":     types.StringType,
		"region": types.StringType,
		"state":  types.StringType,
	}}
}

var (
	_ datasource.DataSource              = &regionsDS{}
	_ datasource.DataSourceWithConfigure = &regionsDS{}
)
