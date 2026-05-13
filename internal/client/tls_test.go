package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/nutanix/terraform-provider-nc2/internal/auth"
)

// TestTLS_RejectsUntrustedCertificate is the FR-003c runtime
// counterpart of internal/provider/tls_test.go::TestBuildTLSConfig_NeverSkipsVerify.
//
// Rationale: a request to an HTTPS server whose certificate is not in
// the supplied trust pool must fail with a TLS verification error
// (NOT a generic network error, NOT a 401 from the server). This pins
// the contract that nothing in the client overrides verification.
func TestTLS_RejectsUntrustedCertificate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(nil)
	defer srv.Close()

	c, err := New(Config{
		Credentials: auth.Credentials{APIKey: "ak", KeyID: "kid", Issuer: "iss"},
		BaseURL:     srv.URL,
		// Empty pool: nothing trusts the test server's self-signed cert.
		TLSConfig: &tls.Config{RootCAs: x509.NewCertPool(), MinVersion: tls.VersionTLS12},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	_, err = c.Do(context.Background(), Request{
		Method:      "GET",
		Path:        "/anything",
		TerraformOp: "tls_reject_test",
	})
	if err == nil {
		t.Fatalf("expected TLS verification error; got nil")
	}
	// The wrapped error should mention TLS / certificate / x509 — i.e.
	// the verification machinery rejected the handshake before any
	// application-layer status code was produced.
	msg := err.Error()
	if !(strings.Contains(msg, "x509") || strings.Contains(msg, "certificate") || strings.Contains(msg, "tls")) {
		t.Errorf("error %q does not look like a TLS verification failure", msg)
	}

	// Defensive: a TLS-verification path produces neither an HTTP
	// status nor a body, so the typed APIError must NOT be returned.
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		t.Errorf("expected non-APIError TLS failure; got APIError %+v", apiErr)
	}
}

// TestTLS_NoInsecureSkipVerifyInClientSource is the binary-level
// counterpart of FR-003c. It greps every .go file in the package for
// `InsecureSkipVerify` set to `true`. Anything matching is a
// regression and fails the test.
func TestTLS_NoInsecureSkipVerifyInClientSource(t *testing.T) {
	t.Parallel()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".go") {
			continue
		}
		// Skip this test file itself; the literal string is allowed
		// here for documentation / search purposes only.
		if ent.Name() == "tls_test.go" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, ent.Name()))
		if err != nil {
			t.Fatalf("ReadFile %s: %v", ent.Name(), err)
		}
		body := string(data)
		for _, bad := range []string{"InsecureSkipVerify: true", "InsecureSkipVerify:true"} {
			if strings.Contains(body, bad) {
				t.Errorf("FR-003c violation: %s contains %q", ent.Name(), bad)
			}
		}
	}
}
