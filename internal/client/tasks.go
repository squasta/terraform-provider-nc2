package client

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// TaskStatus is the typed enum of NC2 task statuses surfaced to
// Terraform. The string values match the NC2 API verbatim.
type TaskStatus string

// Task lifecycle values. Pending / Running are non-terminal.
const (
	TaskPending  TaskStatus = "pending"
	TaskRunning  TaskStatus = "running"
	TaskDone     TaskStatus = "done"
	TaskFailed   TaskStatus = "failed"
	TaskCanceled TaskStatus = "canceled"
)

// IsTerminal reports whether a TaskStatus indicates the task has
// reached a final state and PollTask should return.
func (s TaskStatus) IsTerminal() bool {
	switch s {
	case TaskDone, TaskFailed, TaskCanceled:
		return true
	default:
		return false
	}
}

// TaskResult is what PollTask returns to the caller. ClusterID,
// ErrorCode, and ErrorMessage are extracted from the task envelope
// when present; callers consume them directly without re-parsing
// the response body.
type TaskResult struct {
	ID           string
	Status       TaskStatus
	ClusterID    string
	ErrorCode    string
	ErrorMessage string
	Body         map[string]any
}

// PollTask loops on `/tasks/{taskID}` at TaskPollInterval until the
// task reaches a terminal status or TaskMaxTimeout elapses (or the
// context is cancelled).
//
// On success (TaskDone), returns the TaskResult and a nil error.
// On TaskFailed / TaskCanceled, returns the TaskResult and an error
// summarizing the task's error envelope (so callers can surface it
// as a Terraform diagnostic without re-parsing).
// On timeout, returns the last seen TaskResult and a timeout error
// containing the task id.
func (c *Client) PollTask(ctx context.Context, taskID string) (TaskResult, error) {
	if taskID == "" {
		return TaskResult{}, errors.New("client: PollTask requires a non-empty taskID")
	}

	deadline := c.cfg.Clock().Add(c.cfg.TaskMaxTimeout)
	pollCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	var lastResult TaskResult
	for {
		rsp, err := c.do(pollCtx, Request{
			Method:      "GET",
			Path:        "/tasks/" + taskID,
			TerraformOp: "client.PollTask",
		}, false)
		if err != nil {
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.IsNotFound() {
				return lastResult, fmt.Errorf("client: PollTask: task %s not found", taskID)
			}
			return lastResult, fmt.Errorf("client: PollTask: %w", err)
		}
		lastResult = parseTaskResult(rsp.Body, taskID)
		if lastResult.Status.IsTerminal() {
			if lastResult.Status == TaskDone {
				return lastResult, nil
			}
			return lastResult, fmt.Errorf("client: task %s reached %s (code=%s, message=%s)",
				lastResult.ID, lastResult.Status, lastResult.ErrorCode, lastResult.ErrorMessage)
		}

		select {
		case <-pollCtx.Done():
			return lastResult, fmt.Errorf("client: PollTask: timeout waiting for task %s (last status=%s)",
				taskID, lastResult.Status)
		case <-time.After(c.cfg.TaskPollInterval):
		}
	}
}

// parseTaskResult pulls the documented task fields out of the
// canonical NC2 task response envelope. Pure function.
func parseTaskResult(body map[string]any, fallbackID string) TaskResult {
	res := TaskResult{ID: fallbackID, Body: body}
	data, ok := body["data"].(map[string]any)
	if !ok {
		return res
	}
	if v, ok := data["id"].(string); ok && v != "" {
		res.ID = v
	}
	if v, ok := data["status"].(string); ok {
		res.Status = TaskStatus(v)
	}
	if v, ok := data["cluster_id"].(string); ok {
		res.ClusterID = v
	}
	if errEnv, ok := data["error"].(map[string]any); ok {
		if v, ok := errEnv["code"].(string); ok {
			res.ErrorCode = v
		}
		if v, ok := errEnv["message"].(string); ok {
			res.ErrorMessage = v
		}
	}
	return res
}
