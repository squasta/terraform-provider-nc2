package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nutanix/terraform-provider-nc2/internal/auth"
)

// TestPollTask_Success transitions pending → done.
func TestPollTask_Success(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.WriteHeader(200)
		if n == 1 {
			_, _ = w.Write([]byte(`{"data":{"id":"t-1","status":"pending"}}`))
		} else {
			_, _ = w.Write([]byte(`{"data":{"id":"t-1","status":"done","cluster_id":"c-1"}}`))
		}
	}))
	defer srv.Close()

	c, _ := New(Config{
		Credentials:      auth.Credentials{APIKey: "ak", KeyID: "kid", Issuer: "iss"},
		BaseURL:          srv.URL,
		TaskPollInterval: 1 * time.Millisecond,
		TaskMaxTimeout:   5 * time.Second,
		RequestTimeout:   2 * time.Second,
	})
	res, err := c.PollTask(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("PollTask: %v", err)
	}
	if res.Status != TaskDone || res.ClusterID != "c-1" {
		t.Errorf("got %+v", res)
	}
}

// TestPollTask_Failed surfaces the task error.
func TestPollTask_Failed(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"data":{"id":"t-2","status":"failed","error":{"code":"E","message":"nope"}}}`))
	}))
	defer srv.Close()

	c, _ := New(Config{
		Credentials:      auth.Credentials{APIKey: "ak", KeyID: "kid", Issuer: "iss"},
		BaseURL:          srv.URL,
		TaskPollInterval: 1 * time.Millisecond,
		TaskMaxTimeout:   5 * time.Second,
		RequestTimeout:   2 * time.Second,
	})
	_, err := c.PollTask(context.Background(), "t-2")
	if err == nil {
		t.Fatalf("expected failure")
	}
}

// TestPollTask_EmptyID guard.
func TestPollTask_EmptyID(t *testing.T) {
	t.Parallel()
	c, _ := New(Config{Credentials: auth.Credentials{APIKey: "a", KeyID: "k", Issuer: "i"}})
	if _, err := c.PollTask(context.Background(), ""); err == nil {
		t.Errorf("expected error")
	}
}

// TestParseTaskResult covers the pure parser.
func TestParseTaskResult(t *testing.T) {
	t.Parallel()
	body := map[string]any{
		"data": map[string]any{
			"id":         "t",
			"status":     "done",
			"cluster_id": "c",
			"error":      map[string]any{"code": "E", "message": "m"},
		},
	}
	r := parseTaskResult(body, "fallback")
	if r.ID != "t" || r.Status != TaskDone || r.ClusterID != "c" || r.ErrorCode != "E" {
		t.Errorf("%+v", r)
	}
}
