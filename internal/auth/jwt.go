// Package auth implements the FR-001 NC2 JWT minting + transparent
// refresh contract.
//
// The recipe (research item R-02) is documented in
// `openapi/openapi.json`'s `info.description` and the public Nutanix
// quick-start. It is unusual in two ways:
//
//   - The HMAC input is the `key_id` (NOT the JWT serialized form).
//   - The HMAC OUTPUT is base64-encoded, and that base64 string is
//     then used as the JWT signing secret. Forgetting the base64
//     step yields a syntactically valid JWT that the server rejects
//     with HTTP 401.
//
// The recipe is implemented as a pure function, MintJWT, taking a
// Credentials struct and an explicit `now` time so unit tests can
// pin the produced JWT byte-for-byte against golden fixtures.
//
// JWT payload (per R-02):
//
//	aud      = "https://apikeys.nutanix.com"
//	iat      = now()
//	exp      = iat + 5 minutes (Nutanix-recommended max)
//	iss      = <issuer UUID from credentials>
//	metadata = {"reason": "terraform-provider-nc2"}
//	context  = {}
//
// JWT header:
//
//	alg = HS512
//	kid = <key_id>
//	typ = JWT
//
// FR-001b: this package MUST NOT import any external secret-store
// client. Credentials arrive already-resolved from the provider
// configuration layer (FR-001a, internal/provider/config.go).
package auth

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// DefaultAudience is the fixed audience claim required by the
// Nutanix MyNutanix token exchange. Hard-coded by the API contract;
// not user-configurable.
const DefaultAudience = "https://apikeys.nutanix.com"

// DefaultLifetime is the JWT validity window. Nutanix recommends a
// maximum of 5 minutes; we use exactly that to maximize the time
// between forced refreshes while staying inside the supported window.
const DefaultLifetime = 5 * time.Minute

// MetadataReason is the fixed value of the `metadata.reason` claim
// the provider always emits, so audit trails on the NC2 side can
// distinguish provider-issued tokens from operator-issued ones.
const MetadataReason = "terraform-provider-nc2"

// Credentials bundles the three values FR-001a resolves from
// provider config / env / credentials file. APIKey and KeyID are
// secret material; Issuer is a non-secret organization UUID.
type Credentials struct {
	// APIKey is the per-issuer secret used as the HMAC key.
	APIKey string
	// KeyID identifies the API key on the NC2 side and doubles as
	// the input bytes for the HMAC operation per R-02.
	KeyID string
	// Issuer is the organization UUID that owns the API key; emitted
	// as the `iss` claim.
	Issuer string
}

// Validate returns an error when any required Credentials field is
// missing. Pure function.
func (c Credentials) Validate() error {
	if c.APIKey == "" {
		return errors.New("auth: APIKey is required")
	}
	if c.KeyID == "" {
		return errors.New("auth: KeyID is required")
	}
	if c.Issuer == "" {
		return errors.New("auth: Issuer is required")
	}
	return nil
}

// Token is the result of minting: the compact-serialized JWT and the
// absolute expiration timestamp. Callers cache the JWT until ExpiresAt
// is sufficiently close to `now` to refresh.
type Token struct {
	JWT       string
	ExpiresAt time.Time
}

// MintJWT mints a fresh NC2-compatible JWT from the supplied
// credentials and `now` timestamp. Pure function: same inputs always
// produce the same output (modulo the JWT library's internal field
// ordering, which it normalizes).
//
// Returns an error when:
//   - Credentials.Validate() fails (missing field).
//   - The HMAC computation fails (impossible with stdlib but
//     defensively checked).
//   - The JWT library cannot encode the token (impossible with the
//     fixed claim shape; defensively returned).
//
// The lifetime is fixed at DefaultLifetime; callers wanting a
// different window should add a wrapper. Per R-02 we treat 5 minutes
// as the contract's upper bound.
func MintJWT(creds Credentials, now time.Time) (Token, error) {
	if err := creds.Validate(); err != nil {
		return Token{}, err
	}

	secret, err := derivedSecret(creds.APIKey, creds.KeyID)
	if err != nil {
		return Token{}, fmt.Errorf("auth: deriving signing secret: %w", err)
	}

	exp := now.Add(DefaultLifetime)
	claims := jwt.MapClaims{
		"aud":      DefaultAudience,
		"iat":      now.Unix(),
		"exp":      exp.Unix(),
		"iss":      creds.Issuer,
		"metadata": map[string]any{"reason": MetadataReason},
		"context":  map[string]any{},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tok.Header["kid"] = creds.KeyID
	signed, err := tok.SignedString(secret)
	if err != nil {
		return Token{}, fmt.Errorf("auth: signing JWT: %w", err)
	}
	return Token{JWT: signed, ExpiresAt: exp}, nil
}

// derivedSecret implements the unusual R-02 recipe: HMAC-SHA512 with
// the API key as the HMAC key, KeyID as the HMAC message, then
// standard-base64-encode the digest. The resulting BYTES of that
// base64 string are the JWT signing secret.
func derivedSecret(apiKey, keyID string) ([]byte, error) {
	mac := hmac.New(sha512.New, []byte(apiKey))
	if _, err := mac.Write([]byte(keyID)); err != nil {
		return nil, err
	}
	digest := mac.Sum(nil)
	return []byte(base64.StdEncoding.EncodeToString(digest)), nil
}
