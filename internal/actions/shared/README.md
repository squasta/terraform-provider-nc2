# `internal/actions/shared`

Pure helpers shared by every NC2 action package.

## Public API

- `type Result struct { TaskID, Status, HTTPStatus, ErrorCode, ErrorMessage }`
- `func ResultFromTask(t client.TaskResult, httpStatus int) Result`
- `type Spec struct { Method, Path, Body, PollTask, TerraformOp }`
- `func Invoke(ctx, c, spec, *action.InvokeResponse)` — handles the
  full call → poll → progress → diagnostics flow.
- `func ClientFromConfigure(req, resp, opName) *client.Client`
