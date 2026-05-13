package shared

import (
	"testing"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
)

// TestResultFromTask_Success covers the happy path.
func TestResultFromTask_Success(t *testing.T) {
	t.Parallel()
	r := ResultFromTask(client.TaskResult{ID: "t-1", Status: client.TaskDone}, 200)
	if r.TaskID != "t-1" || r.Status != "done" || r.HTTPStatus != 200 {
		t.Errorf("got %+v", r)
	}
	if r.String() != "task=t-1 status=done http=200" {
		t.Errorf("string=%q", r.String())
	}
}

// TestResultFromTask_Error preserves error code / message.
func TestResultFromTask_Error(t *testing.T) {
	t.Parallel()
	r := ResultFromTask(client.TaskResult{ID: "t-2", Status: client.TaskFailed, ErrorCode: "EBOOM", ErrorMessage: "kaboom"}, 200)
	if r.ErrorCode != "EBOOM" || r.ErrorMessage != "kaboom" {
		t.Errorf("got %+v", r)
	}
	if got := r.String(); got != "task=t-2 status=failed http=200 code=EBOOM message=kaboom" {
		t.Errorf("string=%q", got)
	}
}
