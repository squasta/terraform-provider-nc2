package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// refreshSlack is how long before ExpiresAt the manager considers
// a cached token "expired" and mints a fresh one. With a 5-minute
// lifetime, refreshing 30 seconds early absorbs typical clock skew
// and request-in-flight time without forcing a refresh on every call.
const refreshSlack = 30 * time.Second

// TokenManager mints and caches JWTs on demand. The only mutable
// state in the auth library; guarded by a sync.RWMutex per the
// Constitution Principle III "effects-at-the-edge" pattern.
//
// The manager carries a `clock` function so unit tests can pin the
// "current time" deterministically. Production callers pass
// `time.Now`.
type TokenManager struct {
	creds Credentials
	clock func() time.Time

	mu     sync.RWMutex
	cached Token
}

// NewTokenManager returns a TokenManager for the supplied
// credentials. The clock function MUST NOT be nil; pass `time.Now`
// in production.
func NewTokenManager(creds Credentials, clock func() time.Time) *TokenManager {
	if clock == nil {
		clock = time.Now
	}
	return &TokenManager{creds: creds, clock: clock}
}

// Token returns a JWT valid for at least refreshSlack from now.
// Concurrent callers share a single cached token; on expiry exactly
// one goroutine mints the replacement and the others reuse the
// freshly-cached value.
//
// Returns an error from MintJWT (invalid credentials, HMAC failure,
// JWT library failure). Honors ctx cancellation when waiting for
// the write lock.
func (m *TokenManager) Token(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if tok, ok := m.cachedFresh(); ok {
		return tok, nil
	}
	return m.refresh(ctx)
}

// Invalidate forces the next Token call to mint a fresh JWT.
// Used by the HTTP client on a 401 response to recover from a
// server-side token expiry that the client could not predict.
func (m *TokenManager) Invalidate() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cached = Token{}
}

// cachedFresh returns the cached JWT when it is still valid for
// at least refreshSlack from now. Reads use the read lock so
// concurrent Token calls are wait-free in the steady state.
func (m *TokenManager) cachedFresh() (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.cached.JWT == "" {
		return "", false
	}
	if m.clock().Add(refreshSlack).After(m.cached.ExpiresAt) {
		return "", false
	}
	return m.cached.JWT, true
}

// refresh mints a fresh JWT under the write lock. A second pass of
// the cache check inside the critical section deduplicates concurrent
// refresh attempts.
func (m *TokenManager) refresh(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cached.JWT != "" && m.clock().Add(refreshSlack).Before(m.cached.ExpiresAt) {
		return m.cached.JWT, nil
	}
	tok, err := MintJWT(m.creds, m.clock())
	if err != nil {
		return "", fmt.Errorf("auth: refreshing JWT: %w", err)
	}
	m.cached = tok
	return tok.JWT, nil
}

// ExpiresAt returns the cached JWT's expiry, or the zero time if no
// JWT has been minted yet. Diagnostic only; not part of any contract.
func (m *TokenManager) ExpiresAt() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cached.ExpiresAt
}
