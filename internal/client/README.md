# `internal/client`

Typed HTTP client for the NC2 v2 API. Sits on top of `internal/auth` (JWT minting + refresh) and `internal/audit` (per-call structured log emission).

## Public API

| Symbol | Kind | Purpose |
|---|---|---|
| `Client` | struct | The HTTP client. Safe for concurrent use. |
| `Config` | struct | All inputs to `New`: credentials, BaseURL, UserAgent, TLSConfig, task polling timings, request timeout. |
| `New(Config)` | func | Constructor. Returns an error on invalid credentials or BaseURL. |
| `Request` | struct | `{Method, Path, Body, Query, TerraformOp}`. |
| `Response` | struct | `{Status, Body, Header}`. `Body` is the parsed JSON envelope. |
| `(*Client).Do(ctx, Request)` | method | Issues the request, attaches `Authorization: Bearer <jwt>`, emits one audit record, decodes JSON. On 401 invalidates the cached JWT and retries once. |
| `(*Client).PollTask(ctx, taskID)` | method | Polls `/tasks/{id}` until terminal or `TaskMaxTimeout`. |
| `TaskResult` | struct | `{ID, Status, ClusterID, ErrorCode, ErrorMessage, Body}`. |
| `TaskStatus` (`TaskPending`, `TaskRunning`, `TaskDone`, `TaskFailed`, `TaskCanceled`) | typed enum | Matches NC2 status string verbatim. |
| `APIError` | struct | Typed error for every non-2xx response. Implements `Error()`, `IsNotFound()`, `IsConflict()`, `IsUnauthorized()`. |
| `MapError(*http.Response, body, endpoint)` | func | Pure: translates an HTTP envelope into an `*APIError`. |

## Contract

- Every NC2 call emits exactly one `audit.Record` (FR-020a..d), regardless of success or failure.
- `Authorization: Bearer <jwt>` carries a JWT minted by `internal/auth.TokenManager`.
- `User-Agent` is `terraform-provider-nc2/<version>` (from `Config.UserAgent`).
- `X-Correlation-Id` is a 16-byte hex; the same value lands in the audit record.
- TLS verification is mandatory (FR-003b/c). The client never sets `InsecureSkipVerify`; the only customization surface is `Config.TLSConfig.RootCAs`, populated by `internal/provider.BuildTLSConfig` from the operator's `ca_bundle`.

## Mutable state

The only mutable state in the package is the `http.Client` connection pool (stdlib) and the cached JWT inside `auth.TokenManager`. Both are explicitly carried in the `Client` struct; no globals. Verified by `go test -race ./internal/client/...`.

## Async task model (FR-003, FR-008)

`PollTask` honors `Config.TaskPollInterval` and `Config.TaskMaxTimeout`. On terminal `TaskFailed` / `TaskCanceled`, it returns the typed `TaskResult` and an error summarizing the task's `error.code` / `error.message`, so the caller can surface it as a Terraform diagnostic without re-parsing the body.
