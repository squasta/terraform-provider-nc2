package organization

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestSensitive_Empty pins data-model.md §1: organization has no
// sensitive attributes.
func TestSensitive_Empty(t *testing.T) {
	t.Parallel()

	if got := Sensitive(); len(got) != 0 {
		t.Errorf("Sensitive() = %v; want empty", got)
	}
}

// TestOperationMappings_CoversOrganizationOperations pins FR-021
// for nc2_organization.
func TestOperationMappings_CoversOrganizationOperations(t *testing.T) {
	t.Parallel()

	want := map[string]bool{
		"CPanelWeb.Api.OrganizationController.create":     false,
		"CPanelWeb.Api.OrganizationController.show":       false,
		"CPanelWeb.Api.OrganizationController.update":     false,
		"CPanelWeb.Api.OrganizationController.update (2)": false,
		"CPanelWeb.Api.OrganizationController.terminate":  false,
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

// TestBuildBody_OmitsEmpty makes sure null/unknown attributes do not
// land in the request payload (otherwise NC2 would receive empty
// strings for `description` and reject the call).
func TestBuildBody_OmitsEmpty(t *testing.T) {
	t.Parallel()

	m := &model{
		Name:        types.StringValue("acme"),
		Description: types.StringNull(),
	}
	body := buildBody(m)
	if body["name"] != "acme" {
		t.Errorf("name = %v; want acme", body["name"])
	}
	if _, ok := body["description"]; ok {
		t.Errorf("description should be omitted when null; got %v", body["description"])
	}
}

// TestBuildBody_IncludesDescription pins the happy path.
func TestBuildBody_IncludesDescription(t *testing.T) {
	t.Parallel()

	m := &model{
		Name:        types.StringValue("acme"),
		Description: types.StringValue("test org"),
	}
	body := buildBody(m)
	if body["description"] != "test org" {
		t.Errorf("description = %v; want test org", body["description"])
	}
}

// TestDecodeInto_PopulatesAllFields pins the response shape.
func TestDecodeInto_PopulatesAllFields(t *testing.T) {
	t.Parallel()

	body := map[string]any{
		"data": map[string]any{
			"id":          "org-uuid",
			"name":        "acme",
			"description": "test org",
			"state":       "active",
			"created_at":  "2026-01-01T00:00:00Z",
			"updated_at":  "2026-01-02T00:00:00Z",
		},
	}
	var m model
	diags := decodeInto(body, &m)
	if diags.HasError() {
		t.Fatalf("decodeInto: %v", diags)
	}
	if m.ID.ValueString() != "org-uuid" {
		t.Errorf("ID = %q", m.ID.ValueString())
	}
	if m.State.ValueString() != "active" {
		t.Errorf("State = %q", m.State.ValueString())
	}
	if m.UpdatedAt.ValueString() != "2026-01-02T00:00:00Z" {
		t.Errorf("UpdatedAt = %q", m.UpdatedAt.ValueString())
	}
}

// TestDecodeInto_MissingDataEnvelopeErrors pins error handling.
func TestDecodeInto_MissingDataEnvelopeErrors(t *testing.T) {
	t.Parallel()

	var m model
	diags := decodeInto(map[string]any{}, &m)
	if !diags.HasError() {
		t.Errorf("expected error when data envelope missing")
	}
}
