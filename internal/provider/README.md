# `internal/provider`

The `nc2` provider's framework wiring: schema, credential resolution, TLS configuration, and the `Runtime` bundle handed to every resource / data source / action.

## Files

- `provider.go` — `New() provider.Provider`. Wires the framework's `Provider` interface, exposes the schema, runs `Configure`, and returns the per-story constructor slices.
- `config.go` — FR-001a credential resolution (block > env > credentials_file). Pure function `Resolve(ctx, block, env, file)` returns a `ProviderConfig` and `diag.Diagnostics`. INI-style credentials_file parser at `parseINI`.
- `tls.go` — `BuildTLSConfig(caBundlePath)` clones the system trust store and appends the operator's PEM bundle (FR-003b). NEVER sets `InsecureSkipVerify` (FR-003b/c).
- `registry_us2.go`, `registry_us3.go`, `registry_us5.go` — per-story `init()` blocks that append the per-package constructors to `resourceRegistry` / `dataSourceRegistry` / `actionRegistry`.

## Schema

Mirrors `specs/001-nutanix-nc2-provider/contracts/provider.json`. All credential attributes (`api_key`, `key_id`, `issuer`) are marked `Sensitive: true`.

## Runtime bundle

`Runtime{client *client.Client}` implements both `orgshared.RuntimeAccessor` and `dsshared.RuntimeAccessor` through the single `NC2Client()` getter. The bundle is set on every framework `*Data` field (`ResourceData`, `DataSourceData`, `EphemeralResourceData`, `ActionData`) so every package can extract the wired-up client without any package-level globals.

## FR-001a precedence (with conflict warnings)

`Resolve` walks each credential attribute through three sources in order:

1. **Provider block** (`api_key = "..."` in the `provider "nc2" {}` block).
2. **Environment** (`NC2_API_KEY`, `NC2_KEY_ID`, `NC2_ISSUER`, `NC2_PROFILE`, `NC2_CREDENTIALS_FILE`).
3. **Credentials file** (`~/.nc2/credentials` by default, INI sections per profile).

When the same attribute is supplied by more than one source AND the values disagree, a non-fatal `Warning` diagnostic names every contributing source and uses the higher-precedence value.

## FR-003b/c invariants

- `tls.go` has zero references to `InsecureSkipVerify`. There is no escape hatch in any build (verified by `tls_test.go::TestBuildTLSConfig_NeverSkipsVerify` which greps the source).
- An invalid `ca_bundle` PEM produces a clear plan-time error containing the file path.
