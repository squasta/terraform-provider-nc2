# Authentication

The NC2 provider authenticates by minting a short-lived JWT against
your MyNutanix API key (R-02). Three credentials are required for
every call:

| Attribute | Provider attr | Env var | Credentials-file key |
|---|---|---|---|
| API key (sensitive) | `api_key` | `NC2_API_KEY` | `api_key` |
| API key id (sensitive) | `key_id` | `NC2_KEY_ID` | `key_id` |
| Issuer / org UUID (sensitive) | `issuer` | `NC2_ISSUER` | `issuer` |

## Resolution precedence (FR-001a)

For each credential attribute the provider picks the first non-empty
value in this order:

1. The provider block (e.g. `provider "nc2" { api_key = ... }`).
2. The environment variable.
3. The credentials file profile (default profile is `default`).

If two sources disagree (e.g. provider block sets `api_key=A` and
the env sets `NC2_API_KEY=B`), the provider emits a diagnostic
**warning** at Configure time, naming both sources, and proceeds
with the higher-precedence value. There is no warning when sources
agree.

If no source supplies a credential, the provider emits an
**error** (not a warning) listing every attribute that is missing.

## Credentials file

Path: `~/.nc2/credentials` by default. Override with the
`credentials_file` provider attribute or `NC2_CREDENTIALS_FILE`
environment variable.

Format (INI-like, profiles in `[brackets]`):

```ini
[default]
api_key = ...
key_id  = ...
issuer  = ...

[staging]
api_key = ...
key_id  = ...
issuer  = ...
```

Select a non-default profile with `provider "nc2" { profile =
"staging" }` or `NC2_PROFILE=staging`.

The file is read in plaintext; restrict its permissions
(`chmod 600 ~/.nc2/credentials`).

## No external secret stores (FR-001b)

The provider deliberately does **not** import any native client for
external secret stores (HashiCorp Vault, AWS Secrets Manager,
Azure Key Vault, GCP Secret Manager, 1Password, etc.). Two
binary-level CI gates enforce this:

- `internal/provider/imports_test.go` walks the import graph of
  every internal package and fails on any path matching `vault`,
  `secretsmanager`, `keyvault`, `secret-manager`, `1password`.
- `tests/integration/no_external_secret_stores_test.go` walks the
  binary's `go list -deps` graph (scoped to the main package) for
  the same patterns.

If you want to source NC2 credentials from a secret store, do the
fetch in a step **outside** Terraform (e.g. your CI runner, a
sidecar process, a `local-exec`-driven `vault kv get`) and feed the
plaintext value into one of the three sources above. This keeps the
provider's supply-chain surface minimal (FR-032a).

## JWT recipe (R-02, internal)

For reference; you do not interact with this directly:

- HS512 (HMAC-SHA512) over the canonical JWT header + payload.
- Secret: `HMAC-SHA512(api_key, key_id)` (raw bytes).
- Header: `{"alg": "HS512", "typ": "JWT", "kid": "<key_id>"}`.
- Claims: `iss=<issuer>`, `aud=https://apikeys.nutanix.com`,
  `iat=now`, `exp=now+5m`, plus a `metadata.reason` field set to
  `terraform-provider-nc2`.
- The internal `auth.TokenManager` caches the JWT and refreshes
  ~30 s before expiry, with one forced refresh on a 401 response.
