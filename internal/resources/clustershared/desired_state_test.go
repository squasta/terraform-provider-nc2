package clustershared

import "testing"

// TestMapDesiredStateTransition_Cases pins all branches of the
// pure helper.
func TestMapDesiredStateTransition_Cases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		prev    string
		next    string
		want    OperationKind
		wantErr bool
	}{
		{"noop_same", "running", "running", OpNoop, false},
		{"noop_empty_next", "running", "", OpNoop, false},
		{"hibernate", "running", "hibernated", OpHibernate, false},
		{"resume", "hibernated", "running", OpResume, false},
		{"resume_empty_prev_is_noop", "", "running", OpNoop, false},
		{"hibernate_unknown_prev", "frobnicating", "hibernated", "", true},
		{"resume_unknown_prev", "frobnicating", "running", "", true},
		{"unknown_next", "running", "destroyed", "", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := MapDesiredStateTransition(tc.prev, tc.next)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
