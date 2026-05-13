// Package vpcs implements data.nc2_vpcs.
package vpcs

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
func NewDataSource() datasource.DataSource { return &vpcsDS{} }

type vpcsDS struct{ c *client.Client }

type model struct {
	CloudAccountID types.String `tfsdk:"cloud_account_id"`
	RegionID       types.String `tfsdk:"region_id"`
	Vpcs           types.List   `tfsdk:"vpcs"`
}

// Metadata returns the data source type name.
func (d *vpcsDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_vpcs"
}

// Schema is the framework schema.
func (d *vpcsDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists AWS VPCs discovered inside a cloud-account region. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"cloud_account_id": schema.StringAttribute{Required: true, Description: "Parent cloud account UUID."},
			"region_id":        schema.StringAttribute{Required: true, Description: "Parent region id."},
			"vpcs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true, Description: "VPC id."},
						"name": schema.StringAttribute{Computed: true, Description: "VPC name."},
						"cidr": schema.StringAttribute{Computed: true, Description: "VPC CIDR."},
					},
				},
			},
		},
	}
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (d *vpcsDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_vpcs Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read calls GET /cloud-accounts/{cloud_account_id}/regions/{region_id}/vpcs.
func (d *vpcsDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Path:        "/cloud-accounts/" + cfg.CloudAccountID.ValueString() + "/regions/" + cfg.RegionID.ValueString() + "/vpcs",
		TerraformOp: "data.nc2_vpcs.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_vpcs Read", err.Error())
		return
	}
	items := dsshared.ExtractList(rsp.Body)
	dsshared.SortByID(items)
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":   types.StringType,
		"name": types.StringType,
		"cidr": types.StringType,
	}}
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":   types.StringValue(dsshared.StringOrEmpty(it["id"])),
			"name": types.StringValue(dsshared.StringOrEmpty(it["name"])),
			"cidr": types.StringValue(dsshared.StringOrEmpty(it["cidr"])),
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
	cfg.Vpcs = listVal
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}

var (
	_ datasource.DataSource              = &vpcsDS{}
	_ datasource.DataSourceWithConfigure = &vpcsDS{}
)
