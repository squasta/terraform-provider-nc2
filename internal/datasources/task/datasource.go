// Package task implements data.nc2_task — single-task lookup by id.
package task

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
func NewDataSource() datasource.DataSource { return &taskDS{} }

type taskDS struct{ c *client.Client }

type model struct {
	ID           types.String `tfsdk:"id"`
	Status       types.String `tfsdk:"status"`
	ClusterID    types.String `tfsdk:"cluster_id"`
	CreatedAt    types.String `tfsdk:"created_at"`
	EndedAt      types.String `tfsdk:"ended_at"`
	ErrorCode    types.String `tfsdk:"error_code"`
	ErrorMessage types.String `tfsdk:"error_message"`
}

// Metadata returns the data source type name.
func (d *taskDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_task"
}

// Schema is the framework schema.
func (d *taskDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up a single NC2 task by id.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, Description: "Task id."},
			"status":        schema.StringAttribute{Computed: true, Description: "Lifecycle status."},
			"cluster_id":    schema.StringAttribute{Computed: true, Description: "Cluster id (if applicable)."},
			"created_at":    schema.StringAttribute{Computed: true, Description: "RFC 3339 created_at."},
			"ended_at":      schema.StringAttribute{Computed: true, Description: "RFC 3339 ended_at."},
			"error_code":    schema.StringAttribute{Computed: true, Description: "Structured error code (empty on success)."},
			"error_message": schema.StringAttribute{Computed: true, Description: "Structured error message (empty on success)."},
		},
	}
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (d *taskDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_task Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read calls GET /tasks/{id}.
func (d *taskDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Path:        "/tasks/" + cfg.ID.ValueString(),
		TerraformOp: "data.nc2_task.Read",
	})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			resp.Diagnostics.AddError("nc2_task not found", fmt.Sprintf("task %s does not exist", cfg.ID.ValueString()))
			return
		}
		resp.Diagnostics.AddError("nc2_task Read", err.Error())
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
	m.ID = types.StringValue(dsshared.StringOrEmpty(data["id"]))
	m.Status = types.StringValue(dsshared.StringOrEmpty(data["status"]))
	m.ClusterID = types.StringValue(dsshared.StringOrEmpty(data["cluster_id"]))
	m.CreatedAt = types.StringValue(dsshared.StringOrEmpty(data["created_at"]))
	m.EndedAt = types.StringValue(dsshared.StringOrEmpty(data["ended_at"]))
	if errObj, ok := data["error"].(map[string]any); ok {
		m.ErrorCode = types.StringValue(dsshared.StringOrEmpty(errObj["code"]))
		m.ErrorMessage = types.StringValue(dsshared.StringOrEmpty(errObj["message"]))
	} else {
		m.ErrorCode = types.StringValue("")
		m.ErrorMessage = types.StringValue("")
	}
}

var (
	_ datasource.DataSource              = &taskDS{}
	_ datasource.DataSourceWithConfigure = &taskDS{}
)
