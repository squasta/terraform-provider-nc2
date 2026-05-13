package clustershared

import (
	"reflect"
	"testing"
)

// TestRouteUpdate_OrderingMatchesDataModel pins the ordering from
// data-model.md §4 "Update routing".
func TestRouteUpdate_OrderingMatchesDataModel(t *testing.T) {
	t.Parallel()

	got, err := RouteUpdate(Diff{
		License:        true,
		SSHKey:         true,
		Capacity:       true,
		ResourceTags:   true,
		AccessPolicy:   true,
		GenericMutable: true,
	})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	want := []OperationKind{
		OpUpdateLicense,
		OpUpdateSSHKey,
		OpUpdateCapacity,
		OpUpdateResourceTags,
		OpUpdateAccessPolicy,
		OpGenericPatch,
	}
	gotKinds := make([]OperationKind, len(got))
	for i, op := range got {
		gotKinds[i] = op.Kind
	}
	if !reflect.DeepEqual(gotKinds, want) {
		t.Errorf("got %v; want %v", gotKinds, want)
	}
}

// TestRouteUpdate_DesiredStateAlone allows hibernate alone.
func TestRouteUpdate_DesiredStateAlone(t *testing.T) {
	t.Parallel()

	got, err := RouteUpdate(Diff{
		DesiredStateFrom: "running",
		DesiredStateTo:   "hibernated",
	})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if len(got) != 1 || got[0].Kind != OpHibernate {
		t.Errorf("got %+v; want [hibernate]", got)
	}
}

// TestRouteUpdate_ResumeFromHibernated routes to resume.
func TestRouteUpdate_ResumeFromHibernated(t *testing.T) {
	t.Parallel()

	got, err := RouteUpdate(Diff{
		DesiredStateFrom: "hibernated",
		DesiredStateTo:   "running",
	})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if len(got) != 1 || got[0].Kind != OpResume {
		t.Errorf("got %+v; want [resume]", got)
	}
}

// TestRouteUpdate_RejectsCombinedDiff rejects desired_state change
// combined with other field changes.
func TestRouteUpdate_RejectsCombinedDiff(t *testing.T) {
	t.Parallel()

	_, err := RouteUpdate(Diff{
		Capacity:         true,
		DesiredStateFrom: "running",
		DesiredStateTo:   "hibernated",
	})
	if err == nil {
		t.Errorf("want error; got nil")
	}
}

// TestRouteUpdate_RejectsBadDesiredState rejects unknown values.
func TestRouteUpdate_RejectsBadDesiredState(t *testing.T) {
	t.Parallel()

	_, err := RouteUpdate(Diff{
		DesiredStateFrom: "running",
		DesiredStateTo:   "frozen",
	})
	if err == nil {
		t.Errorf("want error; got nil")
	}
}

// TestRouteUpdate_NoOp returns an empty operation list.
func TestRouteUpdate_NoOp(t *testing.T) {
	t.Parallel()

	got, err := RouteUpdate(Diff{})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d ops; want 0", len(got))
	}
}

// TestCommonAttributeNames is a smoke check that the canonical name
// set covers every attribute documented in data-model.md §4.
func TestCommonAttributeNames(t *testing.T) {
	t.Parallel()

	got := CommonAttributeNames()
	required := []string{
		"id", "organization_id", "cloud_account_id", "name", "region",
		"capacity", "network", "redundancy", "license", "desired_state",
	}
	for _, r := range required {
		found := false
		for _, g := range got {
			if g == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing %q from CommonAttributeNames", r)
		}
	}
}
