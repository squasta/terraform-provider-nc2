# `internal/resources/orgshared`

Cross-package types shared by every resource and action package. Currently just the `RuntimeAccessor` interface used to extract the configured NC2 client from the framework's `ProviderData`.

Lives here (rather than in `internal/provider`) to avoid an import cycle: `internal/provider` itself wires up the per-resource constructors and so cannot be depended on by them.

## Public API

| Symbol | Kind | Purpose |
|---|---|---|
| `RuntimeAccessor` | interface | `NC2Client() *client.Client`. Implemented by `provider.Runtime`. |
