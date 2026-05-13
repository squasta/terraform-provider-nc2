package cluster

import "testing"

// TestUpdateModelFromResponse covers the body decoder.
func TestUpdateModelFromResponse(t *testing.T) {
	t.Parallel()

	m := model{}
	updateModelFromResponse(&m, map[string]any{
		"data": map[string]any{
			"id":             "c-1",
			"name":           "demo",
			"cloud_provider": "aws",
			"region":         "us-east-1",
			"state":          "running",
		},
	})
	if m.ID.ValueString() != "c-1" || m.CloudProvider.ValueString() != "aws" {
		t.Errorf("got %+v", m)
	}
}
