// Package cloud_accounts implements data.nc2_cloud_accounts —
// sorted-by-id list of every NC2 cloud account that belongs to the
// supplied organization.
package cloud_accounts //nolint:revive,staticcheck

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
func NewDataSource() datasource.DataSource { return &cloudAccountsDS{} }

type cloudAccountsDS struct{ c *client.Client }

type model struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	CloudAccounts  types.List   `tfsdk:"cloud_accounts"`
}

func (d *cloudAccountsDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_cloud_accounts"
}

func (d *cloudAccountsDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every NC2 cloud account that belongs to an organization. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{Required: true, Description: "Owning organization UUID."},
			"cloud_accounts": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All cloud accounts owned by the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Computed: true, Description: "Cloud account UUID."},
						"name":           schema.StringAttribute{Computed: true, Description: "Cloud account name."},
						"cloud_provider": schema.StringAttribute{Computed: true, Description: "aws | azure | gcp."},
						"state":          schema.StringAttribute{Computed: true, Description: "Operational state."},
						"created_at":     schema.StringAttribute{Computed: true, Description: "RFC 3339 created_at."},
					},
				},
			},
		},
	}
}

func (d *cloudAccountsDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_cloud_accounts Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read GETs /organizations/{id}/cloud-accounts.
func (d *cloudAccountsDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Path:        "/organizations/" + cfg.OrganizationID.ValueString() + "/cloud-accounts",
		TerraformOp: "data.nc2_cloud_accounts.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_cloud_accounts Read", err.Error())
		return
	}
	items := dsshared.ExtractList(rsp.Body)
	dsshared.SortByID(items)

	objType := caObjectType()
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":             types.StringValue(dsshared.StringOrEmpty(it["id"])),
			"name":           types.StringValue(dsshared.StringOrEmpty(it["name"])),
			"cloud_provider": types.StringValue(dsshared.StringOrEmpty(it["cloud_provider"])),
			"state":          types.StringValue(dsshared.StringOrEmpty(it["state"])),
			"created_at":     types.StringValue(dsshared.StringOrEmpty(it["created_at"])),
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &model{OrganizationID: cfg.OrganizationID, CloudAccounts: listVal})...)
}

func caObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":             types.StringType,
		"name":           types.StringType,
		"cloud_provider": types.StringType,
		"state":          types.StringType,
		"created_at":     types.StringType,
	}}
}

var (
	_ datasource.DataSource              = &cloudAccountsDS{}
	_ datasource.DataSourceWithConfigure = &cloudAccountsDS{}
)
