package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestMintJWT_Validate covers the FR-001 "missing inputs" guard.
func TestMintJWT_ValidateMissing(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		creds Credentials
	}{
		{"missing api_key", Credentials{KeyID: "kid-1", Issuer: "iss-1"}},
		{"missing key_id", Credentials{APIKey: "ak", Issuer: "iss-1"}},
		{"missing issuer", Credentials{APIKey: "ak", KeyID: "kid-1"}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := MintJWT(tc.creds, time.Now()); err == nil {
				t.Errorf("expected error for %s", tc.name)
			}
		})
	}
}

// TestMintJWT_Deterministic verifies the recipe is deterministic
// w.r.t. (creds, now) — the same inputs produce a JWT that signs
// with the same secret derivation. We re-parse + re-verify the JWT
// using the same recipe to round-trip the contract.
func TestMintJWT_Deterministic(t *testing.T) {
	t.Parallel()

	creds := Credentials{APIKey: "deadbeef", KeyID: "key-uuid", Issuer: "iss-uuid"}
	now := time.Date(2026, 5, 13, 12, 0, 0, 0, time.UTC)

	tok, err := MintJWT(creds, now)
	if err != nil {
		t.Fatalf("MintJWT: %v", err)
	}
	if tok.JWT == "" {
		t.Fatalf("JWT is empty")
	}
	if tok.ExpiresAt.Sub(now) != DefaultLifetime {
		t.Errorf("ExpiresAt should be %v after now; got %v", DefaultLifetime, tok.ExpiresAt.Sub(now))
	}
	if strings.Count(tok.JWT, ".") != 2 {
		t.Errorf("JWT compact form should have 3 segments (2 dots); got %q", tok.JWT)
	}

	parsed, err := jwt.Parse(tok.JWT, func(token *jwt.Token) (any, error) {
		secret, _ := derivedSecret(creds.APIKey, creds.KeyID)
		return secret, nil
	}, jwt.WithValidMethods([]string{"HS512"}),
		jwt.WithoutClaimsValidation(),
	)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if got := parsed.Header["kid"]; got != creds.KeyID {
		t.Errorf("kid header = %v; want %s", got, creds.KeyID)
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims have unexpected type %T", parsed.Claims)
	}
	if claims["aud"] != DefaultAudience {
		t.Errorf("aud = %v; want %s", claims["aud"], DefaultAudience)
	}
	if claims["iss"] != creds.Issuer {
		t.Errorf("iss = %v; want %s", claims["iss"], creds.Issuer)
	}
	if md, ok := claims["metadata"].(map[string]any); !ok || md["reason"] != MetadataReason {
		t.Errorf("metadata.reason = %v; want %s", claims["metadata"], MetadataReason)
	}
}
