package aws_cluster //nolint:revive,staticcheck

import (
	"context"
	"testing"
)

// TestSchema_AccessPolicyPresent pins FR-010a: nc2_aws_cluster
// exposes the AWS-only access_policy attribute.
func TestSchema_AccessPolicyPresent(t *testing.T) {
	t.Parallel()

	s := resourceSchema(context.Background())
	if _, ok := s.Attributes["access_policy"]; !ok {
		t.Errorf("nc2_aws_cluster schema missing access_policy")
	}
}

// TestSchema_HasCommonAttributes pins the shared baseline.
func TestSchema_HasCommonAttributes(t *testing.T) {
	t.Parallel()

	s := resourceSchema(context.Background())
	for _, attr := range []string{"id", "name", "region", "license", "desired_state", "state"} {
		if _, ok := s.Attributes[attr]; !ok {
			t.Errorf("schema missing required common attribute %q", attr)
		}
	}
}
