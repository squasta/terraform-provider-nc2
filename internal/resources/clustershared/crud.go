package clustershared

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// CRUD bundles the per-cloud invariants the common lifecycle code
// needs: the create endpoint, the OpenAPI op identifier, and an
// indicator for whether AccessPolicy participates in the diff
// (AWS-only).
type CRUD struct {
	CreatePath          string
	CreateTerraformOp   string // e.g. "nc2_aws_cluster.Create"
	ReadTerraformOp     string
	UpdateTerraformOp   string
	DeleteTerraformOp   string
	SupportsAccessPolicy bool
}

// CreateCluster issues POST {CreatePath} with the supplied body, polls
// the returned task to completion, and returns the new cluster id.
func CreateCluster(ctx context.Context, c *client.Client, crud CRUD, body map[string]any) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	rsp, err := c.Do(ctx, client.Request{
		Method:      "POST",
		Path:        crud.CreatePath,
		Body:        map[string]any{"data": body},
		TerraformOp: crud.CreateTerraformOp,
	})
	if err != nil {
		diags.AddError("cluster create failed", err.Error())
		return "", diags
	}
	taskID := stringFromBody(rsp.Body, "data", "id")
	if taskID == "" {
		diags.AddError("cluster create: no task id", "NC2 response did not include data.id (task id)")
		return "", diags
	}
	res, err := c.PollTask(ctx, taskID)
	if err != nil {
		diags.AddError("cluster create task failed", err.Error())
		return "", diags
	}
	if res.ClusterID == "" {
		diags.AddError("cluster create: no cluster id", fmt.Sprintf("task %s completed without a cluster_id field", taskID))
		return "", diags
	}
	return res.ClusterID, diags
}

// ReadCluster issues GET /clusters/{id} and decodes the response into m.
// Returns (notFound=true, nil diagnostics) when NC2 reports 404.
func ReadCluster(ctx context.Context, c *client.Client, crud CRUD, id string, m *Model) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	rsp, err := c.Do(ctx, client.Request{
		Method:      "GET",
		Path:        "/clusters/" + id,
		TerraformOp: crud.ReadTerraformOp,
	})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.IsNotFound() {
			return true, diags
		}
		diags.AddError("cluster read failed", err.Error())
		return false, diags
	}
	UpdateModelFromResponse(m, rsp.Body)
	return false, diags
}

// DeleteCluster issues POST /clusters/{id}/terminate and waits for the
// task. Returns nil diagnostics on success.
func DeleteCluster(ctx context.Context, c *client.Client, crud CRUD, id string) diag.Diagnostics {
	var diags diag.Diagnostics
	rsp, err := c.Do(ctx, client.Request{
		Method:      "POST",
		Path:        "/clusters/" + id + "/terminate",
		TerraformOp: crud.DeleteTerraformOp,
	})
	if err != nil {
		diags.AddError("cluster terminate failed", err.Error())
		return diags
	}
	taskID := stringFromBody(rsp.Body, "data", "id")
	if taskID == "" {
		// terminate might be synchronous in some scenarios; treat as ok
		return diags
	}
	if _, err := c.PollTask(ctx, taskID); err != nil {
		diags.AddError("cluster terminate task failed", err.Error())
	}
	return diags
}

// ApplyUpdates routes the planned diff to the right NC2 endpoints
// and waits for each task in turn. The body suppliers let per-cloud
// callers shape the per-endpoint payloads (some take a flat map,
// others a structured object).
type UpdateBodies struct {
	License       map[string]any
	SSHKey        map[string]any
	Capacity      map[string]any
	ResourceTags  map[string]any
	AccessPolicy  map[string]any
	GenericPatch  map[string]any
	HibernateBody map[string]any
	ResumeBody    map[string]any
}

// ApplyUpdates issues each operation in `ops` against the right
// /clusters/{id}/* endpoint and polls each returned task. Returns
// the first error encountered; subsequent ops are not attempted.
func ApplyUpdates(ctx context.Context, c *client.Client, crud CRUD, id string, ops []Operation, bodies UpdateBodies) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, op := range ops {
		var (
			method string
			path   string
			body   any
		)
		switch op.Kind {
		case OpUpdateLicense:
			method, path, body = "POST", "/clusters/"+id+"/update-license", wrap(bodies.License)
		case OpUpdateSSHKey:
			method, path, body = "POST", "/clusters/"+id+"/update-ssh-key", wrap(bodies.SSHKey)
		case OpUpdateCapacity:
			method, path, body = "POST", "/clusters/"+id+"/update-capacity", wrap(bodies.Capacity)
		case OpUpdateResourceTags:
			method, path, body = "POST", "/clusters/"+id+"/update-resource-tags", wrap(bodies.ResourceTags)
		case OpUpdateAccessPolicy:
			if !crud.SupportsAccessPolicy {
				diags.AddError("access_policy not supported", "this cluster type does not support access_policy (FR-010a)")
				return diags
			}
			method, path, body = "POST", "/clusters/"+id+"/update-access-policy", wrap(bodies.AccessPolicy)
		case OpGenericPatch:
			method, path, body = "PATCH", "/clusters/"+id, wrap(bodies.GenericPatch)
		case OpHibernate:
			method, path, body = "POST", "/clusters/"+id+"/hibernate", wrap(bodies.HibernateBody)
		case OpResume:
			method, path, body = "POST", "/clusters/"+id+"/resume", wrap(bodies.ResumeBody)
		default:
			diags.AddError("unknown cluster op", string(op.Kind))
			return diags
		}
		rsp, err := c.Do(ctx, client.Request{
			Method:      method,
			Path:        path,
			Body:        body,
			TerraformOp: crud.UpdateTerraformOp,
		})
		if err != nil {
			diags.AddError("cluster "+string(op.Kind)+" failed", err.Error())
			return diags
		}
		if taskID := stringFromBody(rsp.Body, "data", "id"); taskID != "" {
			if _, err := c.PollTask(ctx, taskID); err != nil {
				diags.AddError("cluster "+string(op.Kind)+" task failed", err.Error())
				return diags
			}
		}
	}
	return diags
}

// UpdateModelFromResponse populates the common fields on m from a
// canonical NC2 cluster response body.
func UpdateModelFromResponse(m *Model, body map[string]any) {
	data, ok := body["data"].(map[string]any)
	if !ok {
		return
	}
	if v, ok := data["id"].(string); ok {
		m.ID = types.StringValue(v)
	}
	if v, ok := data["organization_id"].(string); ok {
		m.OrganizationID = types.StringValue(v)
	}
	if v, ok := data["cloud_account_id"].(string); ok {
		m.CloudAccountID = types.StringValue(v)
	}
	if v, ok := data["name"].(string); ok {
		m.Name = types.StringValue(v)
	}
	if v, ok := data["region"].(string); ok {
		m.Region = types.StringValue(v)
	}
	if v, ok := data["use_case"].(string); ok {
		m.UseCase = types.StringValue(v)
	}
	if v, ok := data["host_access_ssh_key"].(string); ok {
		m.HostAccessSSHKey = types.StringValue(v)
	}
	if v, ok := data["license"].(string); ok {
		m.License = types.StringValue(v)
	}
	if v, ok := data["aos_version"].(string); ok {
		m.AOSVersion = types.StringValue(v)
	}
	if v, ok := data["software_tier"].(string); ok {
		m.SoftwareTier = types.StringValue(v)
	}
	if v, ok := data["state"].(string); ok {
		m.State = types.StringValue(v)
	}
	if v, ok := data["created_at"].(string); ok {
		m.CreatedAt = types.StringValue(v)
	}
	if v, ok := data["updated_at"].(string); ok {
		m.UpdatedAt = types.StringValue(v)
	}
	// Map state.state -> desired_state when desired_state is unknown.
	switch m.State.ValueString() {
	case "hibernated", "hibernating":
		if m.DesiredState.IsNull() || m.DesiredState.IsUnknown() {
			m.DesiredState = types.StringValue("hibernated")
		}
	case "running":
		if m.DesiredState.IsNull() || m.DesiredState.IsUnknown() {
			m.DesiredState = types.StringValue("running")
		}
	}
}

// stringFromBody walks a parsed JSON object and returns the string
// value at the supplied dotted-path keys, or "" if the path is
// missing or non-string.
func stringFromBody(body map[string]any, keys ...string) string {
	cur := any(body)
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur, ok = m[k]
		if !ok {
			return ""
		}
	}
	s, _ := cur.(string)
	return s
}

func wrap(body map[string]any) any {
	if body == nil {
		return nil
	}
	return map[string]any{"data": body}
}
