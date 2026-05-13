// Package notifications implements data.nc2_notifications.
package notifications

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
func NewDataSource() datasource.DataSource { return &nDS{} }

type nDS struct{ c *client.Client }

type model struct {
	Notifications types.List `tfsdk:"notifications"`
}

// Metadata returns the data source type name.
func (d *nDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_notifications"
}

// Schema is the framework schema.
func (d *nDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every NC2 notification visible to the configured credentials. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"notifications": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, Description: "Notification id."},
						"severity":     schema.StringAttribute{Computed: true, Description: "Severity (info, warning, critical)."},
						"message":      schema.StringAttribute{Computed: true, Description: "Notification message."},
						"acknowledged": schema.BoolAttribute{Computed: true, Description: "Whether the notification has been acknowledged."},
						"created_at":   schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp."},
					},
				},
			},
		},
	}
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (d *nDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_notifications Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read calls GET /notifications.
func (d *nDS) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.c == nil {
		return
	}
	rsp, err := d.c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/notifications",
		TerraformOp: "data.nc2_notifications.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_notifications Read", err.Error())
		return
	}
	items := dsshared.ExtractList(rsp.Body)
	dsshared.SortByID(items)
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":           types.StringType,
		"severity":     types.StringType,
		"message":      types.StringType,
		"acknowledged": types.BoolType,
		"created_at":   types.StringType,
	}}
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		ack, _ := it["acknowledged"].(bool)
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":           types.StringValue(dsshared.StringOrEmpty(it["id"])),
			"severity":     types.StringValue(dsshared.StringOrEmpty(it["severity"])),
			"message":      types.StringValue(dsshared.StringOrEmpty(it["message"])),
			"acknowledged": types.BoolValue(ack),
			"created_at":   types.StringValue(dsshared.StringOrEmpty(it["created_at"])),
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &model{Notifications: listVal})...)
}

var (
	_ datasource.DataSource              = &nDS{}
	_ datasource.DataSourceWithConfigure = &nDS{}
)
