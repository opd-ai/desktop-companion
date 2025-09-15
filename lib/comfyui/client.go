package comfyui

// Package comfyui provides a minimal client integration with a local ComfyUI
// instance. This is the first incremental step toward the full animation asset
// generation pipeline outlined in GIF_PLAN.md. The goal of this initial commit
// is to establish stable, well‑tested primitives (configuration, interfaces,
// HTTP client with retry) that future stages (WebSocket monitoring, workflow
// templating, batch queue) can build upon without churn.
//
// Design Principles applied here:
//   * Interface first: allows test doubles for higher level pipeline logic.
//   * Minimal surface: only the endpoints required by early pipeline code.
//   * Standard library only: net/http + json are sufficient now.
//   * Clear error wrapping: all failures return contextualised errors.
//   * Small focused functions (<30 LOC) for readability & testability.
//
// Future extensions (not yet implemented here):
//   * WebSocket progress monitoring (MonitorJob)
//   * Workflow template expansion / parameter injection
//   * Queue status streaming + backpressure controls
//   * Result retrieval & asset decoding

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"
    "strings"
)

// Client defines the minimal ComfyUI operations required by the first
// pipeline increment. Additional methods (MonitorJob, GetResult) will be added
// as subsequent tasks deliver those capabilities.
// Client is the public interface implemented by the ComfyUI HTTP client.
//
// Methods are intentionally minimal for the first milestone. Additional
// capabilities (MonitorJob, GetResult, CancelJob) will be appended in later
// tasks; keeping the interface small reduces churn risk.
type Client interface {
    // SubmitWorkflow submits a workflow graph to the ComfyUI server and
    // returns a Job descriptor on success.
    SubmitWorkflow(ctx context.Context, wf *Workflow) (*Job, error)
    // GetQueueStatus returns current queue metrics (pending/running/finished).
    GetQueueStatus(ctx context.Context) (*QueueStatus, error)
}

// HTTPClient abstracts the subset of *http.Client used. This enables tests to
// inject a fake without resorting to global overrides.
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// Config holds runtime settings for connecting to a ComfyUI server.
type Config struct {
    // ServerURL is the base address of ComfyUI (e.g. http://localhost:8188).
    ServerURL     string
    // APIKey optionally sets an Authorization: Bearer header when non-empty.
    APIKey        string
    // Timeout defines the per-request timeout applied via http.Client.
    Timeout       time.Duration
    // RetryAttempts is the maximum number of retry attempts for transient
    // errors (the total requests attempted will be RetryAttempts+1).
    RetryAttempts int
    // RetryBackoff is the base duration used for linear backoff between
    // retries (multiplied by attempt index starting from 1).
    RetryBackoff  time.Duration
}

// DefaultConfig returns a conservative default configuration.
// DefaultConfig returns a baseline configuration targeting a local developer
// machine. Production setups should tune timeouts and retry counts explicitly.
func DefaultConfig() Config {
    return Config{
        ServerURL:     "http://localhost:8188",
        Timeout:       30 * time.Second,
        RetryAttempts: 2,
        RetryBackoff:  500 * time.Millisecond,
    }
}

// Validate ensures configuration values are sensible.
// Validate returns an error when configuration values are invalid so callers
// can fail fast before performing any network operations.
func (c Config) Validate() error {
    if c.ServerURL == "" {
        return errors.New("server URL required")
    }
    if c.Timeout <= 0 {
        return errors.New("timeout must be positive")
    }
    if c.RetryAttempts < 0 {
        return errors.New("retry attempts cannot be negative")
    }
    if c.RetryBackoff < 0 {
        return errors.New("retry backoff cannot be negative")
    }
    return nil
}

// Workflow represents a minimal workflow payload. We store arbitrary node data
// so callers can construct detailed graphs without this package needing to
// understand internal schema yet.
type Workflow struct {
    ID    string                 `json:"id"`              // Optional identifier (client-side reference)
    Nodes map[string]interface{} `json:"nodes"`           // Arbitrary workflow node graph
    Meta  map[string]interface{} `json:"meta,omitempty"`  // Optional supplemental metadata
}

// Job returned after submission.
type Job struct {
    ID     string `json:"id"`     // Unique server-assigned job id
    Status string `json:"status"` // Initial status (e.g. queued)
}

// QueueStatus reflects lightweight queue metrics from the server.
type QueueStatus struct {
    Pending  int `json:"pending"`
    Running  int `json:"running"`
    Finished int `json:"finished"`
}

// client is the concrete implementation of Client.
type client struct {
    cfg    Config
    httpc  HTTPClient
}

// New creates a new Client instance.
// New constructs a Client using the provided configuration.
func New(cfg Config) (Client, error) {
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid comfyui config: %w", err)
    }
    hc := &http.Client{Timeout: cfg.Timeout}
    return &client{cfg: cfg, httpc: hc}, nil
}

// SubmitWorkflow posts a workflow JSON to the ComfyUI /api/workflows endpoint
// (endpoint name chosen based on planned API; adjust when actual endpoint is
// confirmed). It retries on transient network / 5xx errors.
func (c *client) SubmitWorkflow(ctx context.Context, wf *Workflow) (*Job, error) {
    if wf == nil {
        return nil, errors.New("workflow is nil")
    }
    body, err := json.Marshal(wf)
    if err != nil {
        return nil, fmt.Errorf("marshal workflow: %w", err)
    }
    var lastErr error
    attempts := c.cfg.RetryAttempts + 1
    for i := 0; i < attempts; i++ {
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        default:
        }
        job, err := c.postWorkflow(ctx, body)
        if err == nil {
            return job, nil
        }
        if !isRetryable(err) || i == attempts-1 {
            return nil, err
        }
        lastErr = err
        sleep := c.cfg.RetryBackoff * time.Duration(i+1)
        timer := time.NewTimer(sleep)
        select {
        case <-ctx.Done():
            timer.Stop()
            return nil, ctx.Err()
        case <-timer.C:
        }
    }
    return nil, lastErr
}

func (c *client) postWorkflow(ctx context.Context, body []byte) (*Job, error) {
    url := c.cfg.ServerURL + "/api/workflows"
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    if c.cfg.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
    }
    resp, err := c.httpc.Do(req)
    if err != nil {
        return nil, fmt.Errorf("post workflow: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 500 {
        return nil, fmt.Errorf("server error %d", resp.StatusCode)
    }
    if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
    }
    var job Job
    if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
        return nil, fmt.Errorf("decode job response: %w", err)
    }
    if job.ID == nilString(job.ID) { // ensure non-empty
        return nil, errors.New("empty job id in response")
    }
    return &job, nil
}

// GetQueueStatus retrieves current queue metrics.
func (c *client) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
    url := c.cfg.ServerURL + "/api/queue"
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    if c.cfg.APIKey != "" {
        req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
    }
    resp, err := c.httpc.Do(req)
    if err != nil {
        return nil, fmt.Errorf("get queue status: %w", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
        return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
    }
    var qs QueueStatus
    if err := json.NewDecoder(resp.Body).Decode(&qs); err != nil {
        return nil, fmt.Errorf("decode queue status: %w", err)
    }
    return &qs, nil
}

// nilString returns empty string; helper used only to make intent explicit when checking job id.
func nilString(s string) string { return "" }

// isRetryable determines if an error should trigger a retry. For now we treat
// explicit server error wrappers (contains 'server error') and network layer
// errors as retryable. This can evolve with richer error types later.
func isRetryable(err error) bool {
    if err == nil { return false }
    // Simple substring check keeps logic lightweight; can refine later.
    es := err.Error()
    if strings.Contains(es, "server error") || strings.Contains(es, "timeout") || strings.Contains(es, "connection refused") {
        return true
    }
    return false
}
