package auth

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestTokenManager_CachesAndRefreshes covers the cache-hit path,
// the cache-miss-on-expiry path, and the Invalidate forced-refresh.
func TestTokenManager_CachesAndRefreshes(t *testing.T) {
	t.Parallel()

	creds := Credentials{APIKey: "ak", KeyID: "kid", Issuer: "iss"}
	now := time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC)

	mu := sync.Mutex{}
	clock := func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return now
	}
	advance := func(d time.Duration) {
		mu.Lock()
		now = now.Add(d)
		mu.Unlock()
	}

	tm := NewTokenManager(creds, clock)
	first, err := tm.Token(context.Background())
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if first == "" {
		t.Fatalf("first token empty")
	}
	cached, err := tm.Token(context.Background())
	if err != nil {
		t.Fatalf("Token2: %v", err)
	}
	if cached != first {
		t.Errorf("token not cached: %q vs %q", first, cached)
	}

	advance(DefaultLifetime + time.Second)
	refreshed, err := tm.Token(context.Background())
	if err != nil {
		t.Fatalf("Token3: %v", err)
	}
	if refreshed == first {
		t.Errorf("expected refreshed token to differ; both %q", first)
	}

	tm.Invalidate()
	advance(time.Second)
	postInvalid, err := tm.Token(context.Background())
	if err != nil {
		t.Fatalf("Token4: %v", err)
	}
	if postInvalid == refreshed {
		t.Errorf("Invalidate did not force refresh; both %q", refreshed)
	}
}

// TestTokenManager_ContextCancel verifies ctx cancellation surfaces.
func TestTokenManager_ContextCancel(t *testing.T) {
	t.Parallel()

	tm := NewTokenManager(Credentials{APIKey: "ak", KeyID: "kid", Issuer: "iss"}, time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := tm.Token(ctx); err == nil {
		t.Errorf("expected ctx cancellation error")
	}
}
