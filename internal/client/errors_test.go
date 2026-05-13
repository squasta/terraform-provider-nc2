package client

import (
	"errors"
	"net/http"
	"testing"
)

// TestMapError_ParsesEnvelope covers the canonical NC2 error body.
func TestMapError_ParsesEnvelope(t *testing.T) {
	t.Parallel()

	body := []byte(`{"error":{"code":"E_BOOM","message":"kaboom"}}`)
	err := MapError(&http.Response{StatusCode: 422}, body, "/x")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.Status != 422 || apiErr.Code != "E_BOOM" || apiErr.Message != "kaboom" || apiErr.Endpoint != "/x" {
		t.Errorf("got %+v", apiErr)
	}
	if got := apiErr.Error(); got == "" {
		t.Errorf("Error() empty: %s", got)
	}
}

// TestMapError_FallbackBodyAsMessage covers non-JSON bodies.
func TestMapError_FallbackBodyAsMessage(t *testing.T) {
	t.Parallel()

	err := MapError(&http.Response{StatusCode: 500}, []byte("internal failure"), "/y")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError")
	}
	if apiErr.Message != "internal failure" {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

// TestAPIError_Predicates pin the documented helpers.
func TestAPIError_Predicates(t *testing.T) {
	t.Parallel()

	if !(&APIError{Status: 404}).IsNotFound() {
		t.Errorf("404 should be NotFound")
	}
	if !(&APIError{Status: 401}).IsUnauthorized() {
		t.Errorf("401 should be Unauthorized")
	}
	if !(&APIError{Status: 409}).IsConflict() {
		t.Errorf("409 should be Conflict")
	}
	if (&APIError{Status: 200}).IsNotFound() {
		t.Errorf("200 should not be NotFound")
	}
}
