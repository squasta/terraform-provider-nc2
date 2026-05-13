package provider

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

// BuildTLSConfig constructs the *tls.Config the HTTP client should
// use. When caBundlePath is empty, returns nil — meaning "fall back
// to the OS trust store via http.DefaultTransport's defaults".
//
// When caBundlePath is set, the file is read, parsed as PEM, and
// APPENDED to a clone of the system trust store. The file MUST
// contain at least one PEM CERTIFICATE block; otherwise an error
// is returned describing where parsing failed.
//
// FR-003b/c: this function MUST NEVER set InsecureSkipVerify.
// There is no escape hatch; the parameter is simply not exposed.
func BuildTLSConfig(caBundlePath string) (*tls.Config, error) {
	if caBundlePath == "" {
		return nil, nil
	}
	data, err := os.ReadFile(caBundlePath)
	if err != nil {
		return nil, fmt.Errorf("provider: ca_bundle %s: %w", caBundlePath, err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("provider: ca_bundle %s is empty", caBundlePath)
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(data) {
		return nil, errors.New("provider: ca_bundle did not contain any valid PEM CERTIFICATE block")
	}
	return &tls.Config{
		RootCAs:    pool,
		MinVersion: tls.VersionTLS12,
	}, nil
}
