// Package redact tests the FR-002 sensitive-field pattern matcher.
package redact

import "testing"

// TestMatchPattern_PositiveCases covers every FR-002 pattern with multiple
// realistic field-name shapes (camelCase, snake_case, PascalCase, embedded).
// Case-insensitive substring matching is the FR-002 contract.
func TestMatchPattern_PositiveCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		// credential
		{"credential lowercase", "credential"},
		{"credentials plural", "credentials"},
		{"aws_credentials snake", "aws_credentials"},
		{"AWSCredential pascal", "AWSCredential"},
		{"Credential pascal", "Credential"},
		{"someCredentialField camel", "someCredentialField"},

		// password
		{"password lowercase", "password"},
		{"PASSWORD upper", "PASSWORD"},
		{"user_password snake", "user_password"},
		{"DbPassword pascal", "DbPassword"},
		{"passwordHash camel", "passwordHash"},

		// secret
		{"secret lowercase", "secret"},
		{"client_secret snake", "client_secret"},
		{"AppSecret pascal", "AppSecret"},
		{"secretKey camel", "secretKey"},

		// token
		{"token lowercase", "token"},
		{"access_token snake", "access_token"},
		{"BearerToken pascal", "BearerToken"},
		{"refreshToken camel", "refreshToken"},

		// private_key (matches both as full name and as substring)
		{"private_key snake", "private_key"},
		{"PrivateKey pascal", "PrivateKey"},
		{"ssh_private_key snake", "ssh_private_key"},
		{"someAccountPrivateKey camel", "someAccountPrivateKey"},

		// api_key
		{"api_key snake", "api_key"},
		{"ApiKey pascal", "ApiKey"},
		{"APIKey pascal-acronym", "APIKey"},
		{"my_api_key snake", "my_api_key"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if !MatchPattern(tc.input) {
				t.Errorf("MatchPattern(%q) = false, want true", tc.input)
			}
		})
	}
}

// TestMatchPattern_NegativeCases covers field names that look adjacent to the
// patterns but must not match — protects against false-positive redaction of
// non-sensitive operational fields. These are the cases that motivate FR-002a:
// new fields default non-sensitive until classified.
func TestMatchPattern_NegativeCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"description", "description"},
		{"name", "name"},
		{"id", "id"},
		{"created_at", "created_at"},
		{"updated_at", "updated_at"},
		{"region", "region"},
		{"cloud_provider", "cloud_provider"},
		{"status", "status"},
		{"organization_id", "organization_id"},
		{"cluster_id", "cluster_id"},
		{"host_id", "host_id"},
		// "key" alone must NOT match (only api_key / private_key are sensitive).
		{"key bare", "key"},
		{"ssh_key", "ssh_key"},
		// "tokenize" contains "token" — case proves substring matching is
		// genuinely what FR-002 requires; this case will be re-examined when we
		// pick "tokenize" up via FR-002a lint warnings in real code.
		// We document the current behavior here so a future change is intentional.
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if MatchPattern(tc.input) {
				t.Errorf("MatchPattern(%q) = true, want false", tc.input)
			}
		})
	}
}

// TestMatchPattern_DocumentedSubstringEdges pins the substring-match contract
// from FR-002: "credential" as a substring of "credentialed" still matches.
// Any future change to make matching word-boundary-aware MUST be a deliberate
// FR-002 amendment, not a silent code change.
func TestMatchPattern_DocumentedSubstringEdges(t *testing.T) {
	t.Parallel()

	pinned := map[string]bool{
		"credentialed":    true,  // contains "credential"
		"passwordless":    true,  // contains "password"
		"tokenize":        true,  // contains "token" — see negative-cases note
		"recommendations": false, // no pattern as substring
		"section":         false, // does NOT contain "secret"
	}

	for input, want := range pinned {
		input, want := input, want
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if got := MatchPattern(input); got != want {
				t.Errorf("MatchPattern(%q) = %v, want %v", input, got, want)
			}
		})
	}
}

// TestFR002Patterns_Stable ensures the exported pattern list does not silently
// shrink under refactoring; the count is part of the FR-002 contract.
func TestFR002Patterns_Stable(t *testing.T) {
	t.Parallel()

	want := []string{"credential", "password", "secret", "token", "private_key", "api_key"}
	if len(FR002Patterns) != len(want) {
		t.Fatalf("FR002Patterns has %d entries, want %d", len(FR002Patterns), len(want))
	}
	set := make(map[string]struct{}, len(FR002Patterns))
	for _, p := range FR002Patterns {
		set[p] = struct{}{}
	}
	for _, p := range want {
		if _, ok := set[p]; !ok {
			t.Errorf("FR002Patterns missing pattern %q", p)
		}
	}
}
