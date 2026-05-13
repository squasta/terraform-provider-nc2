package cloud_account //nolint:revive,staticcheck

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestSensitive_ListsExplicitCredentialPaths pins data-model.md §2.
func TestSensitive_ListsExplicitCredentialPaths(t *testing.T) {
	t.Parallel()

	got := Sensitive()
	want := map[string]bool{
		"credentials.aws.access_key_id":        false,
		"credentials.aws.secret_access_key":    false,
		"credentials.azure.client_secret":      false,
		"credentials.gcp.service_account_json": false,
	}
	for _, p := range got {
		if _, ok := want[p]; ok {
			want[p] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("Sensitive() missing %q", k)
		}
	}
}

// TestOperationMappings_CoversCloudAccountOperations pins FR-021.
func TestOperationMappings_CoversCloudAccountOperations(t *testing.T) {
	t.Parallel()

	want := map[string]bool{
		"CPanelWeb.Api.OrganizationController.create_cloud_account":  false,
		"CPanelWeb.Api.CloudAccountController.show":                  false,
		"CPanelWeb.Api.CloudAccountController.update":                false,
		"CPanelWeb.Api.CloudAccountController.update (2)":            false,
		"CPanelWeb.Api.CloudAccountController.update_credentials":    false,
	}
	for _, m := range OperationMappings {
		if _, ok := want[m.OperationID]; ok {
			want[m.OperationID] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("OperationMappings missing %q", k)
		}
	}
}

// TestMapsEqual_IdenticalMaps pins the diff helper.
func TestMapsEqual_IdenticalMaps(t *testing.T) {
	t.Parallel()

	a, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"access_key_id":     types.StringValue("AKIA..."),
		"secret_access_key": types.StringValue("sek"),
	})
	b, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"access_key_id":     types.StringValue("AKIA..."),
		"secret_access_key": types.StringValue("sek"),
	})
	if !mapsEqual(a, b) {
		t.Errorf("expected mapsEqual(a, b) = true")
	}
}

// TestMapsEqual_DifferentValuesDriveCredentialsRotation.
func TestMapsEqual_DifferentValuesDriveCredentialsRotation(t *testing.T) {
	t.Parallel()

	a, _ := types.MapValue(types.StringType, map[string]attr.Value{"k": types.StringValue("v1")})
	b, _ := types.MapValue(types.StringType, map[string]attr.Value{"k": types.StringValue("v2")})
	if mapsEqual(a, b) {
		t.Errorf("expected mapsEqual to return false on changed values")
	}
}

// TestCredentialsRequestBody_FlattensMap pins the body shape sent to
// /update-credentials.
func TestCredentialsRequestBody_FlattensMap(t *testing.T) {
	t.Parallel()

	creds, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"access_key_id":     types.StringValue("AKIA0000"),
		"secret_access_key": types.StringValue("sek"),
	})
	m := &model{Credentials: creds}
	body, diags := credentialsRequestBody(context.Background(), m)
	if diags.HasError() {
		t.Fatalf("diags: %v", diags)
	}
	if body["access_key_id"] != "AKIA0000" || body["secret_access_key"] != "sek" {
		t.Errorf("body = %+v", body)
	}
}

// TestDecodeInto_PreservesCredentialsAcrossRefresh ensures NC2's
// silence on credentials does not blow away state.
func TestDecodeInto_PreservesCredentialsAcrossRefresh(t *testing.T) {
	t.Parallel()

	creds, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"access_key_id": types.StringValue("AKIA1234"),
	})
	m := &model{Credentials: creds}
	body := map[string]any{
		"data": map[string]any{
			"id":              "ca-uuid",
			"organization_id": "org-uuid",
			"cloud_provider":  "aws",
			"name":            "acme-aws",
			"state":           "active",
		},
	}
	diags := decodeInto(context.Background(), body, m)
	if diags.HasError() {
		t.Fatalf("decodeInto: %v", diags)
	}
	if m.ID.ValueString() != "ca-uuid" {
		t.Errorf("ID = %q", m.ID.ValueString())
	}
	if !mapsEqual(m.Credentials, creds) {
		t.Errorf("Credentials should be preserved when API does not echo them")
	}
}

// TestDecodeInto_FillsEmptyCredentialsMapWhenAbsent pins the
// "first read" path.
func TestDecodeInto_FillsEmptyCredentialsMapWhenAbsent(t *testing.T) {
	t.Parallel()

	m := &model{}
	body := map[string]any{
		"data": map[string]any{"id": "ca-uuid", "name": "acme"},
	}
	diags := decodeInto(context.Background(), body, m)
	if diags.HasError() {
		t.Fatalf("decodeInto: %v", diags)
	}
	if m.Credentials.IsNull() || m.Credentials.IsUnknown() {
		t.Errorf("expected empty (non-null) Credentials map; got null/unknown")
	}
	if len(m.Credentials.Elements()) != 0 {
		t.Errorf("expected empty Credentials map; got %d elements", len(m.Credentials.Elements()))
	}
}
