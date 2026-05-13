// Package organizations implements data.nc2_organizations — a
// sorted-by-id list of every NC2 organization visible to the
// configured credentials.
package organizations

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
func NewDataSource() datasource.DataSource { return &orgsDS{} }

type orgsDS struct{ c *client.Client }

type model struct {
	Organizations types.List `tfsdk:"organizations"`
}

// Metadata returns the data source type name.
func (d *orgsDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_organizations"
}

// Schema is the framework schema.
func (d *orgsDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every NC2 organization visible to the configured credentials. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"organizations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All visible organizations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Organization UUID."},
						"name":        schema.StringAttribute{Computed: true, Description: "Organization name."},
						"description": schema.StringAttribute{Computed: true, Description: "Organization description."},
						"state":       schema.StringAttribute{Computed: true, Description: "Lifecycle state."},
						"created_at":  schema.StringAttribute{Computed: true, Description: "RFC 3339 created_at."},
					},
				},
			},
		},
	}
}

// Configure extracts the wired-up NC2 client.
func (d *orgsDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_organizations Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read GETs /organizations and returns the deterministically-sorted list.
func (d *orgsDS) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.c == nil {
		return
	}
	rsp, err := d.c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/organizations",
		TerraformOp: "data.nc2_organizations.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_organizations Read", err.Error())
		return
	}
	items := dsshared.ExtractList(rsp.Body)
	dsshared.SortByID(items)

	objType := orgObjectType()
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":          types.StringValue(dsshared.StringOrEmpty(it["id"])),
			"name":        types.StringValue(dsshared.StringOrEmpty(it["name"])),
			"description": types.StringValue(dsshared.StringOrEmpty(it["description"])),
			"state":       types.StringValue(dsshared.StringOrEmpty(it["state"])),
			"created_at":  types.StringValue(dsshared.StringOrEmpty(it["created_at"])),
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &model{Organizations: listVal})...)
}

func orgObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":          types.StringType,
		"name":        types.StringType,
		"description": types.StringType,
		"state":       types.StringType,
		"created_at":  types.StringType,
	}}
}

var (
	_ datasource.DataSource              = &orgsDS{}
	_ datasource.DataSourceWithConfigure = &orgsDS{}
)
