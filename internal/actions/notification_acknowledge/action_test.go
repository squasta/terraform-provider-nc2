package notification_acknowledge //nolint:revive,staticcheck

import "testing"

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()
	want := map[string]bool{
		"CPanelWeb.Api.NotificationController.update (2)": false,
		"CPanelWeb.Api.NotificationController.update":     false,
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
