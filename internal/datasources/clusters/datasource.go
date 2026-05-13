// Package clusters implements data.nc2_clusters — sorted-by-id list
// of every NC2 cluster visible to the configured credentials.
package clusters

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/datasources/dsshared"
)

// NewDataSource is the framework-facing constructor.
func NewDataSource() datasource.DataSource { return &clustersDS{} }

type clustersDS struct{ c *client.Client }

type model struct {
	Clusters types.List `tfsdk:"clusters"`
}

// Metadata returns the data source type name.
func (d *clustersDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_clusters"
}

// Schema is the framework schema.
func (d *clustersDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists every NC2 cluster visible to the configured credentials. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"clusters": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All visible clusters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true, Description: "UUID."},
						"name":             schema.StringAttribute{Computed: true, Description: "Cluster name."},
						"organization_id":  schema.StringAttribute{Computed: true, Description: "Owning organization UUID."},
						"cloud_account_id": schema.StringAttribute{Computed: true, Description: "Owning cloud account UUID."},
						"cloud_provider":   schema.StringAttribute{Computed: true, Description: "aws | azure | gcp."},
						"region":           schema.StringAttribute{Computed: true, Description: "Cloud-provider region."},
						"state":            schema.StringAttribute{Computed: true, Description: "Lifecycle state."},
						"created_at":       schema.StringAttribute{Computed: true, Description: "RFC 3339 created_at."},
					},
				},
			},
		},
	}
}

// Configure extracts the wired-up NC2 client from the runtime bundle.
func (d *clustersDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_clusters Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read calls GET /clusters and returns the sorted list.
func (d *clustersDS) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.c == nil {
		return
	}
	rsp, err := d.c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/clusters",
		TerraformOp: "data.nc2_clusters.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_clusters Read", err.Error())
		return
	}
	items := extractList(rsp.Body)
	sort.Slice(items, func(i, j int) bool { return stringOrEmpty(items[i]["id"]) < stringOrEmpty(items[j]["id"]) })
	objType := clusterObjectType()
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":               types.StringValue(stringOrEmpty(it["id"])),
			"name":             types.StringValue(stringOrEmpty(it["name"])),
			"organization_id":  types.StringValue(stringOrEmpty(it["organization_id"])),
			"cloud_account_id": types.StringValue(stringOrEmpty(it["cloud_account_id"])),
			"cloud_provider":   types.StringValue(stringOrEmpty(it["cloud_provider"])),
			"region":           types.StringValue(stringOrEmpty(it["region"])),
			"state":            types.StringValue(stringOrEmpty(it["state"])),
			"created_at":       types.StringValue(stringOrEmpty(it["created_at"])),
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &model{Clusters: listVal})...)
}

func clusterObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":               types.StringType,
		"name":             types.StringType,
		"organization_id":  types.StringType,
		"cloud_account_id": types.StringType,
		"cloud_provider":   types.StringType,
		"region":           types.StringType,
		"state":            types.StringType,
		"created_at":       types.StringType,
	}}
}

func extractList(body map[string]any) []map[string]any {
	raw, ok := body["data"].([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, r := range raw {
		if m, ok := r.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func stringOrEmpty(v any) string { s, _ := v.(string); return s }

var (
	_ datasource.DataSource              = &clustersDS{}
	_ datasource.DataSourceWithConfigure = &clustersDS{}
)
