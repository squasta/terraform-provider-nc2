// Package tasks implements data.nc2_tasks.
package tasks

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
func NewDataSource() datasource.DataSource { return &tasksDS{} }

type tasksDS struct{ c *client.Client }

type model struct {
	Tasks types.List `tfsdk:"tasks"`
}

// Metadata returns the data source type name.
func (d *tasksDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_tasks"
}

// Schema is the framework schema.
func (d *tasksDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every NC2 task visible to the configured credentials. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"tasks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, Description: "Task id."},
						"status":     schema.StringAttribute{Computed: true, Description: "Lifecycle status."},
						"cluster_id": schema.StringAttribute{Computed: true, Description: "Cluster id (if applicable)."},
						"created_at": schema.StringAttribute{Computed: true, Description: "RFC 3339 created_at."},
						"ended_at":   schema.StringAttribute{Computed: true, Description: "RFC 3339 ended_at."},
					},
				},
			},
		},
	}
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (d *tasksDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_tasks Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read calls GET /tasks.
func (d *tasksDS) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.c == nil {
		return
	}
	rsp, err := d.c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/tasks",
		TerraformOp: "data.nc2_tasks.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_tasks Read", err.Error())
		return
	}
	items := dsshared.ExtractList(rsp.Body)
	dsshared.SortByID(items)
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":         types.StringType,
		"status":     types.StringType,
		"cluster_id": types.StringType,
		"created_at": types.StringType,
		"ended_at":   types.StringType,
	}}
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":         types.StringValue(dsshared.StringOrEmpty(it["id"])),
			"status":     types.StringValue(dsshared.StringOrEmpty(it["status"])),
			"cluster_id": types.StringValue(dsshared.StringOrEmpty(it["cluster_id"])),
			"created_at": types.StringValue(dsshared.StringOrEmpty(it["created_at"])),
			"ended_at":   types.StringValue(dsshared.StringOrEmpty(it["ended_at"])),
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &model{Tasks: listVal})...)
}

var (
	_ datasource.DataSource              = &tasksDS{}
	_ datasource.DataSourceWithConfigure = &tasksDS{}
)
