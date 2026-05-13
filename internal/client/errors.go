package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError is the typed error every non-2xx NC2 response surfaces
// as. Callers can `errors.As(err, &apiErr)` to inspect the HTTP
// status, NC2 error code (when present), endpoint, and the parsed
// response body.
type APIError struct {
	Status   int
	Code     string
	Message  string
	Endpoint string
	Body     map[string]any
}

// Error implements the error interface. The format is stable and
// machine-friendly; tests pin it.
func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code != "" {
		return fmt.Sprintf("nc2 api %d on %s (code=%s): %s", e.Status, e.Endpoint, e.Code, e.Message)
	}
	return fmt.Sprintf("nc2 api %d on %s: %s", e.Status, e.Endpoint, e.Message)
}

// IsNotFound reports whether the error is a 404, used by the per-
// resource Read methods to drop a vanished resource from state per
// FR-009.
func (e *APIError) IsNotFound() bool {
	return e != nil && e.Status == http.StatusNotFound
}

// IsConflict reports whether the error is a 409 — typically used by
// resources that need to retry on a transient state-machine race.
func (e *APIError) IsConflict() bool {
	return e != nil && e.Status == http.StatusConflict
}

// IsUnauthorized reports whether the error is a 401, the trigger
// for the client's one-shot JWT refresh + retry.
func (e *APIError) IsUnauthorized() bool {
	return e != nil && e.Status == http.StatusUnauthorized
}

// MapError translates an HTTP response into the typed APIError. The
// raw body is parsed for `error.code` / `error.message` if present;
// otherwise a stringified body is used as the message.
//
// Pure function (no I/O); the *http.Response is only inspected for
// its StatusCode and Header values. Callers MUST have already drained
// and closed the body before calling.
func MapError(resp *http.Response, body []byte, endpoint string) error {
	apiErr := &APIError{
		Status:   resp.StatusCode,
		Endpoint: endpoint,
	}
	var parsed map[string]any
	if len(body) > 0 {
		_ = json.Unmarshal(body, &parsed)
	}
	apiErr.Body = parsed

	if e, ok := parsed["error"].(map[string]any); ok {
		if c, ok := e["code"].(string); ok {
			apiErr.Code = c
		}
		if m, ok := e["message"].(string); ok {
			apiErr.Message = m
		}
	}
	if apiErr.Message == "" {
		if len(body) > 0 {
			apiErr.Message = string(body)
		} else {
			apiErr.Message = http.StatusText(resp.StatusCode)
		}
	}
	return apiErr
}
