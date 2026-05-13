// Package shared houses the small, pure helpers reused by every NC2
// action package. Actions in the framework only emit Diagnostics +
// progress events; this package provides a typed Result struct so
// the shape of what we report stays consistent and so per-action
// audit / log output (FR-020d) is straightforward to format.
package shared

import (
	"fmt"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// Result captures the outcome of an action invocation. Mirrors the
// `result.*` shape described in data-model.md §11. Actions in the
// framework cannot set output values on the configuration, so this
// struct is consumed by the per-action progress / audit messages.
type Result struct {
	TaskID       string
	Status       string
	HTTPStatus   int
	ErrorCode    string
	ErrorMessage string
}

// ResultFromTask builds a Result from the (TaskResult, http_status)
// pair returned by the typed client. Pure function — no I/O.
func ResultFromTask(t client.TaskResult, httpStatus int) Result {
	return Result{
		TaskID:       t.ID,
		Status:       string(t.Status),
		HTTPStatus:   httpStatus,
		ErrorCode:    t.ErrorCode,
		ErrorMessage: t.ErrorMessage,
	}
}

// String renders a Result as a single human-readable line. Used as
// the Diagnostics summary on a successful action invocation.
func (r Result) String() string {
	if r.ErrorCode == "" && r.ErrorMessage == "" {
		return fmt.Sprintf("task=%s status=%s http=%d", r.TaskID, r.Status, r.HTTPStatus)
	}
	return fmt.Sprintf("task=%s status=%s http=%d code=%s message=%s",
		r.TaskID, r.Status, r.HTTPStatus, r.ErrorCode, r.ErrorMessage)
}
