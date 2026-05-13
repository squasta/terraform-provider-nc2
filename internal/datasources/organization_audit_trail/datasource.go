// Package organization_audit_trail implements
// data.nc2_organization_audit_trail — sorted-by-id list of every
// audit entry recorded for an NC2 organization.
package organization_audit_trail //nolint:revive,staticcheck

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
func NewDataSource() datasource.DataSource { return &auditTrailDS{} }

type auditTrailDS struct{ c *client.Client }

type model struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	Entries        types.List   `tfsdk:"entries"`
}

func (d *auditTrailDS) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "nc2_organization_audit_trail"
}

func (d *auditTrailDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists audit-trail entries for an NC2 organization. Sorted ascending by id (FR-015).",
		Attributes: map[string]schema.Attribute{
			"organization_id": schema.StringAttribute{Required: true, Description: "Owning organization UUID."},
			"entries": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All audit-trail entries.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true, Description: "Audit entry id."},
						"actor":      schema.StringAttribute{Computed: true, Description: "Actor identifier (email / service principal)."},
						"action":     schema.StringAttribute{Computed: true, Description: "Action verb."},
						"target":     schema.StringAttribute{Computed: true, Description: "Acted-on target (resource id or human label)."},
						"created_at": schema.StringAttribute{Computed: true, Description: "RFC 3339 timestamp."},
					},
				},
			},
		},
	}
}

func (d *auditTrailDS) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	rt, ok := req.ProviderData.(dsshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError("nc2_organization_audit_trail Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return
	}
	d.c = rt.NC2Client()
}

// Read GETs /organizations/{id}/audit-trails.
func (d *auditTrailDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
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
		Path:        "/organizations/" + cfg.OrganizationID.ValueString() + "/audit-trails",
		TerraformOp: "data.nc2_organization_audit_trail.Read",
	})
	if err != nil {
		resp.Diagnostics.AddError("nc2_organization_audit_trail Read", err.Error())
		return
	}
	items := dsshared.ExtractList(rsp.Body)
	dsshared.SortByID(items)

	objType := entryObjectType()
	values := make([]attr.Value, 0, len(items))
	for _, it := range items {
		v, diag := types.ObjectValue(objType.AttrTypes, map[string]attr.Value{
			"id":         types.StringValue(dsshared.StringOrEmpty(it["id"])),
			"actor":      types.StringValue(dsshared.StringOrEmpty(it["actor"])),
			"action":     types.StringValue(dsshared.StringOrEmpty(it["action"])),
			"target":     types.StringValue(dsshared.StringOrEmpty(it["target"])),
			"created_at": types.StringValue(dsshared.StringOrEmpty(it["created_at"])),
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &model{OrganizationID: cfg.OrganizationID, Entries: listVal})...)
}

func entryObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":         types.StringType,
		"actor":      types.StringType,
		"action":     types.StringType,
		"target":     types.StringType,
		"created_at": types.StringType,
	}}
}

var (
	_ datasource.DataSource              = &auditTrailDS{}
	_ datasource.DataSourceWithConfigure = &auditTrailDS{}
)
