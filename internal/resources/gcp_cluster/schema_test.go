package gcp_cluster //nolint:revive,staticcheck

import (
	"context"
	"testing"
)

// TestSchema_NoAccessPolicy pins FR-010a: nc2_gcp_cluster MUST NOT
// expose `access_policy` (AWS-only).
func TestSchema_NoAccessPolicy(t *testing.T) {
	t.Parallel()

	s := resourceSchema(context.Background())
	if _, ok := s.Attributes["access_policy"]; ok {
		t.Errorf("nc2_gcp_cluster schema includes access_policy; FR-010a says it must be AWS-only")
	}
}

// TestSchema_NoDesiredState pins FR-012: nc2_gcp_cluster MUST NOT
// expose `desired_state` — hibernate / resume is AWS-only.
func TestSchema_NoDesiredState(t *testing.T) {
	t.Parallel()

	s := resourceSchema(context.Background())
	if _, ok := s.Attributes["desired_state"]; ok {
		t.Errorf("nc2_gcp_cluster schema includes desired_state; FR-012 says hibernate/resume must be AWS-only")
	}
}

// TestSchema_HasCommonAttributes pins the shared baseline (excluding
// the AWS-only `desired_state`).
func TestSchema_HasCommonAttributes(t *testing.T) {
	t.Parallel()

	s := resourceSchema(context.Background())
	for _, attr := range []string{"id", "name", "region", "license", "state"} {
		if _, ok := s.Attributes[attr]; !ok {
			t.Errorf("schema missing required common attribute %q", attr)
		}
	}
}
