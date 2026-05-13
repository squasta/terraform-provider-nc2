package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nutanix/terraform-provider-nc2/internal/auth"
)

func newTestServer(t *testing.T, status int, body string) (*httptest.Server, *int) {
	t.Helper()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.Header.Get("Authorization"); !strings.HasPrefix(got, "Bearer ") {
			t.Errorf("missing Bearer auth header: %q", got)
		}
		if got := r.Header.Get("User-Agent"); !strings.Contains(got, "terraform-provider-nc2") {
			t.Errorf("missing or wrong User-Agent: %q", got)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func newClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := New(Config{
		Credentials:    auth.Credentials{APIKey: "ak", KeyID: "kid", Issuer: "iss"},
		BaseURL:        baseURL,
		UserAgent:      "terraform-provider-nc2/test",
		RequestTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

// TestClient_DoSuccess covers the 200 happy path: auth header set,
// JSON envelope decoded.
func TestClient_DoSuccess(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t, 200, `{"data":{"id":"x-1","name":"demo"}}`)
	c := newClient(t, srv.URL)

	rsp, err := c.Do(context.Background(), Request{Method: "GET", Path: "/x", TerraformOp: "test.read"})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if rsp.Status != 200 {
		t.Errorf("Status=%d", rsp.Status)
	}
	data, _ := rsp.Body["data"].(map[string]any)
	if data["id"] != "x-1" || data["name"] != "demo" {
		t.Errorf("body parse: %+v", rsp.Body)
	}
}

// TestClient_DoErrorMapping covers FR-020 error envelope translation.
func TestClient_DoErrorMapping(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t, 422, `{"error":{"code":"E_BAD","message":"bad input"}}`)
	c := newClient(t, srv.URL)
	_, err := c.Do(context.Background(), Request{Method: "POST", Path: "/y", TerraformOp: "test.create"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T (%v)", err, err)
	}
	if apiErr.Status != 422 || apiErr.Code != "E_BAD" {
		t.Errorf("got %+v", apiErr)
	}
}

// TestClient_NotFound returns IsNotFound true.
func TestClient_NotFound(t *testing.T) {
	t.Parallel()

	srv, _ := newTestServer(t, 404, `{"error":{"code":"E_NOTFOUND"}}`)
	c := newClient(t, srv.URL)
	_, err := c.Do(context.Background(), Request{Method: "GET", Path: "/missing", TerraformOp: "test.read"})
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !apiErr.IsNotFound() {
		t.Errorf("expected IsNotFound; got %v", err)
	}
}

// TestClient_RequiresValidCredentials rejects empty creds at New.
func TestClient_RequiresValidCredentials(t *testing.T) {
	t.Parallel()

	if _, err := New(Config{}); err == nil {
		t.Errorf("expected error from empty credentials")
	}
}

// TestClient_BodyJSONEncoded round-trips a body through the wire.
func TestClient_BodyJSONEncoded(t *testing.T) {
	t.Parallel()

	got := make(chan map[string]any, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]any
		_ = json.NewDecoder(r.Body).Decode(&m)
		got <- m
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	c := newClient(t, srv.URL)
	_, err := c.Do(context.Background(), Request{
		Method:      "POST",
		Path:        "/z",
		Body:        map[string]any{"hello": "world"},
		TerraformOp: "test.create",
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	body := <-got
	if body["hello"] != "world" {
		t.Errorf("body roundtrip failed: %+v", body)
	}
}
