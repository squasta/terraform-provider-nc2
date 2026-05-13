package shared

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/nutanix/terraform-provider-nc2/internal/client"
	"github.com/nutanix/terraform-provider-nc2/internal/resources/orgshared"
)

// Spec describes a single action invocation. Path and Method are
// the NC2 endpoint to call; Body is the request payload; PollTask
// indicates whether the response carries a task id that must be
// polled to terminal state.
type Spec struct {
	Method      string
	Path        string
	Body        map[string]any
	PollTask    bool
	TerraformOp string
}

// Invoke executes the action against the supplied client and writes
// the diagnostics + progress events. Reports the final Result via
// resp.Diagnostics's first warning (success) or error (failure).
func Invoke(ctx context.Context, c *client.Client, spec Spec, resp *action.InvokeResponse) {
	if c == nil {
		resp.Diagnostics.AddError(spec.TerraformOp, "no NC2 client; provider not configured")
		return
	}
	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("calling %s %s", spec.Method, spec.Path)})
	}
	body := wrap(spec.Body)
	rsp, err := c.Do(ctx, client.Request{
		Method:      spec.Method,
		Path:        spec.Path,
		Body:        body,
		TerraformOp: spec.TerraformOp,
	})
	if err != nil {
		resp.Diagnostics.AddError(spec.TerraformOp+" failed", err.Error())
		return
	}
	if !spec.PollTask {
		res := Result{HTTPStatus: rsp.Status}
		resp.Diagnostics.Append(diag.NewWarningDiagnostic(spec.TerraformOp+" succeeded", res.String()))
		return
	}
	taskID := stringFromBody(rsp.Body, "data", "id")
	if taskID == "" {
		res := Result{HTTPStatus: rsp.Status}
		resp.Diagnostics.Append(diag.NewWarningDiagnostic(spec.TerraformOp+" succeeded (no task)", res.String()))
		return
	}
	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("polling task %s", taskID)})
	}
	taskRes, err := c.PollTask(ctx, taskID)
	res := ResultFromTask(taskRes, rsp.Status)
	if err != nil {
		resp.Diagnostics.AddError(spec.TerraformOp+" task failed", res.String())
		return
	}
	resp.Diagnostics.Append(diag.NewWarningDiagnostic(spec.TerraformOp+" succeeded", res.String()))
}

// ClientFromConfigure pulls the NC2 client out of the framework's
// ProviderData. Returns nil if the bundle is missing or the wrong
// type — callers should add a diagnostic in that case.
func ClientFromConfigure(req action.ConfigureRequest, resp *action.ConfigureResponse, opName string) *client.Client {
	if req.ProviderData == nil {
		return nil
	}
	rt, ok := req.ProviderData.(orgshared.RuntimeAccessor)
	if !ok {
		resp.Diagnostics.AddError(opName+" Configure", fmt.Sprintf("unexpected ProviderData type %T", req.ProviderData))
		return nil
	}
	return rt.NC2Client()
}

// wrap is the canonical request envelope shared by every NC2 action.
func wrap(body map[string]any) any {
	if body == nil {
		return nil
	}
	return map[string]any{"data": body}
}

func stringFromBody(body map[string]any, keys ...string) string {
	cur := any(body)
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur, ok = m[k]
		if !ok {
			return ""
		}
	}
	s, _ := cur.(string)
	return s
}
