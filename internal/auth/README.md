# `internal/auth`

JWT minting and transparent refresh for the NC2 API (FR-001, R-02).

## Public API

| Symbol | Kind | Purpose |
|---|---|---|
| `Credentials` | struct | `{APIKey, KeyID, Issuer}` triple resolved by the provider config layer. |
| `Credentials.Validate()` | method | Returns an error when any required field is missing. |
| `Token` | struct | `{JWT, ExpiresAt}` returned by `MintJWT`. |
| `MintJWT(creds, now)` | func | Pure: returns a freshly-signed JWT and its expiry timestamp. |
| `TokenManager` | struct | Thread-safe JWT cache + lazy refresher. |
| `NewTokenManager(creds, clock)` | func | Constructor; pass `time.Now` in production, a fake clock in tests. |
| `(*TokenManager).Token(ctx)` | method | Returns a JWT valid for at least 30s; mints a fresh one on expiry. |
| `(*TokenManager).Invalidate()` | method | Forces the next `Token` call to refresh; called by `internal/client` on a 401. |
| `(*TokenManager).ExpiresAt()` | method | Diagnostic accessor for the cached JWT's expiry. |

## JWT recipe (per R-02)

The recipe is documented in `openapi/openapi.json#info.description` and the public Nutanix quick-start:

1. Build the JWT payload:
   - `aud = "https://apikeys.nutanix.com"`
   - `iat = now()`
   - `exp = iat + 5 minutes`
   - `iss = <issuer UUID>`
   - `metadata = {"reason": "terraform-provider-nc2"}`
   - `context  = {}`
2. Compute the signing secret as `base64_standard(HMAC_SHA512(api_key, key_id))`. Note the unusual two-step: HMAC input is the `key_id` bytes; the OUTPUT bytes are then base64-encoded and used as the JWT secret.
3. Encode with `HS512` and header `{"kid": "<key_id>", "typ": "JWT"}`.

Forgetting the base64 step at step 2 produces a syntactically valid JWT that the server rejects with HTTP 401.

## FR-001b: no external secret stores

This package MUST NOT import any external-secret-store client (Vault, AWS Secrets Manager, Azure Key Vault, etc.). Credentials arrive already-resolved from `internal/provider/config.go`. A CI import-graph check guards this invariant.

## Concurrency

`TokenManager` is safe for concurrent use. The cache is guarded by a single `sync.RWMutex` so the steady-state `Token` call is wait-free; only refreshes serialize on the write lock. Verified by `go test -race`.
