// Package client implements the NC2 HTTP client used by every
// resource, data source, and action in the provider.
//
// Layered responsibilities:
//
//   - JWT minting + refresh: via internal/auth.TokenManager
//     (FR-001). Authorization header: `Bearer <jwt>`.
//   - Strict TLS verification: never sets InsecureSkipVerify
//     (FR-003b/c); the provider's `ca_bundle` PEM is appended to
//     the system trust store via internal/provider.BuildTLSConfig.
//   - Single audit record per call (FR-020a..d): emitted via
//     internal/audit.Record on success and failure alike.
//   - Async task polling (FR-003, FR-008): PollTask(ctx, taskID)
//     loops on `/tasks/{id}` until the server reports a terminal
//     state, honoring task_poll_interval_seconds and
//     task_max_timeout_seconds from the provider config.
//   - Error mapping (FR-020): MapError translates HTTP envelopes
//     into typed APIError values that callers can `errors.As` on.
//
// The package's only mutable state is the http.Client connection
// pool (stdlib) and the cached JWT inside auth.TokenManager. Both
// are explicitly carried in the Client struct; no globals.
package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nutanix/terraform-provider-nc2/internal/audit"
	"github.com/nutanix/terraform-provider-nc2/internal/auth"
)

// DefaultBaseURL is the production NC2 API endpoint. Override via
// Config.BaseURL for staging / test environments.
const DefaultBaseURL = "https://cloud.nutanix.com/api/v2"

// DefaultUserAgent is the User-Agent header sent on every request.
// The version suffix is filled in by the provider package at build
// time via Config.UserAgent.
const DefaultUserAgent = "terraform-provider-nc2/dev"

// DefaultTaskPollInterval is the time between polls of the same
// task. Honors FR-003 / FR-008.
const DefaultTaskPollInterval = 5 * time.Second

// DefaultTaskMaxTimeout is the upper bound on how long PollTask
// waits for a single task to reach a terminal state.
const DefaultTaskMaxTimeout = 60 * time.Minute

// DefaultRequestTimeout is the per-request HTTP timeout.
const DefaultRequestTimeout = 60 * time.Second

// Config bundles everything the Client needs at construction time.
// All fields are documented; sensible defaults apply when zero.
type Config struct {
	// Credentials drive JWT minting. Required.
	Credentials auth.Credentials

	// BaseURL overrides DefaultBaseURL. Optional.
	BaseURL string

	// UserAgent overrides DefaultUserAgent. Optional.
	UserAgent string

	// TLSConfig is the *tls.Config the underlying http.Transport
	// uses. The provider package builds it via BuildTLSConfig.
	// Optional; when nil, the system default trust store applies
	// (which is fine for production but means no `ca_bundle`).
	TLSConfig *tls.Config

	// TaskPollInterval overrides DefaultTaskPollInterval.
	TaskPollInterval time.Duration

	// TaskMaxTimeout overrides DefaultTaskMaxTimeout.
	TaskMaxTimeout time.Duration

	// RequestTimeout overrides DefaultRequestTimeout. Set to 0 to
	// disable the per-request timeout (use ctx.Deadline instead).
	RequestTimeout time.Duration

	// Clock supplies time.Now (and is used for retry / backoff
	// decisions). NewClient defaults to time.Now.
	Clock func() time.Time
}

// Client is the typed NC2 HTTP client. Safe for concurrent use.
type Client struct {
	cfg     Config
	tokens  *auth.TokenManager
	http    *http.Client
	baseURL *url.URL
}

// New constructs a Client from a fully-resolved Config. Returns an
// error when Credentials are invalid or BaseURL fails to parse.
func New(cfg Config) (*Client, error) {
	if err := cfg.Credentials.Validate(); err != nil {
		return nil, err
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}
	if cfg.TaskPollInterval == 0 {
		cfg.TaskPollInterval = DefaultTaskPollInterval
	}
	if cfg.TaskMaxTimeout == 0 {
		cfg.TaskMaxTimeout = DefaultTaskMaxTimeout
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = DefaultRequestTimeout
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}

	base, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("client: parse BaseURL: %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig:       cfg.TLSConfig,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		cfg:     cfg,
		tokens:  auth.NewTokenManager(cfg.Credentials, cfg.Clock),
		http:    &http.Client{Timeout: cfg.RequestTimeout, Transport: transport},
		baseURL: base,
	}, nil
}

// Request is the typed input to Client.Do. Body MAY be nil for
// GET / DELETE; otherwise it is JSON-encoded.
type Request struct {
	Method      string
	Path        string
	Body        any
	Query       url.Values
	TerraformOp string
}

// Response is the typed output of Client.Do. Status is the HTTP
// status code; Body is the parsed JSON envelope (always
// `map[string]any` because every NC2 endpoint returns an object).
type Response struct {
	Status int
	Body   map[string]any
	Header http.Header
}

// Do issues the supplied Request, attaches the Authorization header,
// emits one audit record, and decodes the JSON response body.
//
// On HTTP 401 the client invalidates the cached JWT and retries the
// request exactly once. Any other non-2xx status surfaces as an
// *APIError via MapError, which callers can `errors.As` against.
func (c *Client) Do(ctx context.Context, req Request) (Response, error) {
	rsp, err := c.do(ctx, req, false)
	if err == nil {
		return rsp, nil
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		c.tokens.Invalidate()
		return c.do(ctx, req, true)
	}
	return rsp, err
}

func (c *Client) do(ctx context.Context, req Request, isRetry bool) (Response, error) {
	corrID := newCorrelationID()
	startedAt := c.cfg.Clock()

	url := *c.baseURL
	url.Path = strings.TrimRight(url.Path, "/") + "/" + strings.TrimLeft(req.Path, "/")
	if len(req.Query) > 0 {
		url.RawQuery = req.Query.Encode()
	}

	var bodyReader io.Reader
	if req.Body != nil {
		buf, err := json.Marshal(req.Body)
		if err != nil {
			return Response{}, fmt.Errorf("client: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url.String(), bodyReader)
	if err != nil {
		return Response{}, fmt.Errorf("client: build request: %w", err)
	}

	tok, err := c.tokens.Token(ctx)
	if err != nil {
		return Response{}, fmt.Errorf("client: minting JWT: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+tok)
	httpReq.Header.Set("User-Agent", c.cfg.UserAgent)
	httpReq.Header.Set("Accept", "application/json")
	if bodyReader != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	httpReq.Header.Set("X-Correlation-Id", corrID)

	httpRsp, err := c.http.Do(httpReq)
	latency := c.cfg.Clock().Sub(startedAt).Milliseconds()
	if err != nil {
		audit.Record(ctx, audit.AuditRecord{
			CorrelationID: corrID,
			TerraformOp:   req.TerraformOp,
			HTTPMethod:    req.Method,
			Path:          req.Path,
			Status:        0,
			LatencyMs:     latency,
		})
		return Response{}, fmt.Errorf("client: HTTP %s %s: %w", req.Method, req.Path, err)
	}
	defer httpRsp.Body.Close()

	rawBody, err := io.ReadAll(httpRsp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("client: read body: %w", err)
	}

	var parsed map[string]any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &parsed); err != nil {
			parsed = map[string]any{"raw": string(rawBody)}
		}
	}

	rec := audit.AuditRecord{
		CorrelationID: corrID,
		TerraformOp:   req.TerraformOp,
		HTTPMethod:    req.Method,
		Path:          req.Path,
		Status:        httpRsp.StatusCode,
		LatencyMs:     latency,
	}
	if taskID := stringFromBody(parsed, "data", "id"); taskID != "" {
		rec.NC2TaskID = taskID
	}
	if errCode := stringFromBody(parsed, "error", "code"); errCode != "" {
		rec.NC2ErrorCode = errCode
	}
	audit.Record(ctx, rec)

	if httpRsp.StatusCode >= 400 {
		_ = isRetry
		return Response{Status: httpRsp.StatusCode, Body: parsed, Header: httpRsp.Header},
			MapError(httpRsp, rawBody, req.Path)
	}

	return Response{Status: httpRsp.StatusCode, Body: parsed, Header: httpRsp.Header}, nil
}

// stringFromBody walks parsed JSON for a dotted-path string. Returns
// "" on any miss.
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

// newCorrelationID returns a 16-byte random hex id used as the
// audit-record CorrelationID and the X-Correlation-Id header.
func newCorrelationID() string {
	var buf [16]byte
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}
