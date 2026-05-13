package task

import "testing"

// TestUpdateModelFromResponse covers the body decoder.
func TestUpdateModelFromResponse(t *testing.T) {
	t.Parallel()
	m := model{}
	updateModelFromResponse(&m, map[string]any{
		"data": map[string]any{
			"id":     "t-1",
			"status": "done",
			"error": map[string]any{
				"code":    "",
				"message": "",
			},
		},
	})
	if m.ID.ValueString() != "t-1" || m.Status.ValueString() != "done" {
		t.Errorf("got %+v", m)
	}
}
