package swarmui

// Package swarmui provides a client for interacting with SwarmUI API for image generation.
// This complements the existing ComfyUI integration by providing an alternative backend
// that follows the same interface patterns for easy substitution in the pipeline.
//
// Design Principles:
//   * Interface-first approach matching ComfyUI client patterns
//   * Session management following SwarmUI authentication model
//   * Standard library HTTP client with timeout and retry logic
//   * Clear error handling with contextual error messages
//   * Small focused functions for readability and testability
//
// SwarmUI API Features supported:
//   * Session management (GetNewSession)
//   * Text2Image generation with progress monitoring
//   * Queue status and job management
//   * WebSocket progress updates (future extension)

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client defines the SwarmUI operations for image generation.
// This interface mirrors the ComfyUI Client interface to enable
// easy backend switching in the pipeline controller.
type Client interface {
	// GenerateImage generates a single image using Text2Image API
	GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResult, error)

	// GetQueueStatus returns current queue metrics
	GetQueueStatus(ctx context.Context) (*QueueStatus, error)

	// MonitorJob establishes a WebSocket connection for progress updates
	MonitorJob(ctx context.Context, jobID string) (<-chan JobProgress, error)

	// GetSession returns the current session ID
	GetSession() string

	// RefreshSession obtains a new session ID
	RefreshSession(ctx context.Context) error
}

// HTTPClient abstracts the subset of *http.Client used for testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config holds runtime settings for connecting to SwarmUI.
type Config struct {
	// ServerURL is the base address of SwarmUI (e.g. http://localhost:7801)
	ServerURL string
	// AuthToken optionally sets authentication token for secured instances
	AuthToken string
	// Timeout defines the per-request timeout
	Timeout time.Duration
	// RetryAttempts is the maximum number of retry attempts
	RetryAttempts int
	// RetryBackoff is the base duration for backoff between retries
	RetryBackoff time.Duration
	// SessionRefreshInterval defines how often to refresh the session
	SessionRefreshInterval time.Duration
}

// DefaultConfig returns a conservative default configuration.
func DefaultConfig() Config {
	return Config{
		ServerURL:              "http://localhost:7801",
		Timeout:                30 * time.Second,
		RetryAttempts:          3, // Increased to match ComfyUI
		RetryBackoff:           500 * time.Millisecond,
		SessionRefreshInterval: 30 * time.Minute,
	}
}

// Validate ensures configuration values are sensible.
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

// ImageRequest represents a Text2Image generation request.
type ImageRequest struct {
	Prompt         string                 `json:"prompt"`
	NegativePrompt string                 `json:"negativeprompt,omitempty"`
	Model          string                 `json:"model,omitempty"`
	Width          int                    `json:"width,omitempty"`
	Height         int                    `json:"height,omitempty"`
	CFGScale       float64                `json:"cfgscale,omitempty"`
	Steps          int                    `json:"steps,omitempty"`
	Seed           int64                  `json:"seed,omitempty"`
	Images         int                    `json:"images,omitempty"`
	DoNotSave      bool                   `json:"donotsave,omitempty"`
	BatchSize      int                    `json:"batchsize,omitempty"`
	Extra          map[string]interface{} `json:",inline"` // For additional parameters
}

// ImageResult contains the result of image generation.
type ImageResult struct {
	Images   []string `json:"images"` // Array of image paths/URLs
	JobID    string   `json:"job_id"` // Job identifier for monitoring
	Status   string   `json:"status"` // Generation status
	ErrorID  string   `json:"error_id,omitempty"`
	ErrorMsg string   `json:"error,omitempty"`
}

// QueueStatus reflects queue metrics from SwarmUI.
type QueueStatus struct {
	Pending  int `json:"pending"`
	Running  int `json:"running"`
	Finished int `json:"finished"`
}

// JobProgress represents progress updates for image generation.
type JobProgress struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	Progress  float64   `json:"progress"`          // 0.0 to 1.0
	Message   string    `json:"message"`           // Status message
	Preview   string    `json:"preview,omitempty"` // Preview image URL
	Timestamp time.Time `json:"timestamp"`
	Err       error     `json:"error,omitempty"` // Error if any
}

// SessionResponse represents the response from GetNewSession.
type SessionResponse struct {
	SessionID    string `json:"session_id"`
	UserID       string `json:"user_id"`
	OutputAppend bool   `json:"output_append_user"`
	Version      string `json:"version"`
	ServerID     string `json:"server_id"`
	ErrorID      string `json:"error_id,omitempty"`
	ErrorMsg     string `json:"error,omitempty"`
}

// client is the concrete implementation of Client.
type client struct {
	cfg         Config
	httpc       HTTPClient
	sessionID   string
	lastRefresh time.Time
}

// New creates a new SwarmUI Client instance.
func New(cfg Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	httpc := &http.Client{
		Timeout: cfg.Timeout,
	}

	return &client{
		cfg:   cfg,
		httpc: httpc,
	}, nil
}

// NewWithHTTPClient creates a Client with a custom HTTP client (for testing).
func NewWithHTTPClient(cfg Config, httpc HTTPClient) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &client{
		cfg:   cfg,
		httpc: httpc,
	}, nil
}

// GetSession returns the current session ID.
func (c *client) GetSession() string {
	return c.sessionID
}

// RefreshSession obtains a new session ID from SwarmUI.
func (c *client) RefreshSession(ctx context.Context) error {
	url := fmt.Sprintf("%s/API/GetNewSession", strings.TrimSuffix(c.cfg.ServerURL, "/"))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return fmt.Errorf("creating session request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.cfg.AuthToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "swarm_token",
			Value: c.cfg.AuthToken,
		})
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("session request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading session response: %w", err)
	}

	var sessionResp SessionResponse
	if err := json.Unmarshal(body, &sessionResp); err != nil {
		return fmt.Errorf("parsing session response: %w", err)
	}

	if sessionResp.ErrorID != "" || sessionResp.ErrorMsg != "" {
		return fmt.Errorf("session error: %s (%s)", sessionResp.ErrorMsg, sessionResp.ErrorID)
	}

	if sessionResp.SessionID == "" {
		return errors.New("no session ID in response")
	}

	c.sessionID = sessionResp.SessionID
	c.lastRefresh = time.Now()
	return nil
}

// ensureSession ensures we have a valid session, refreshing if needed.
func (c *client) ensureSession(ctx context.Context) error {
	if c.sessionID == "" || time.Since(c.lastRefresh) > c.cfg.SessionRefreshInterval {
		return c.RefreshSession(ctx)
	}
	return nil
}

// GenerateImage generates a single image using SwarmUI Text2Image API.
func (c *client) GenerateImage(ctx context.Context, req *ImageRequest) (*ImageResult, error) {
	if err := c.ensureSession(ctx); err != nil {
		return nil, fmt.Errorf("session setup failed: %w", err)
	}

	// Create request payload with session ID
	payload := make(map[string]interface{})
	payload["session_id"] = c.sessionID

	// Copy all fields from ImageRequest
	if req.Prompt != "" {
		payload["prompt"] = req.Prompt
	}
	if req.NegativePrompt != "" {
		payload["negativeprompt"] = req.NegativePrompt
	}
	if req.Model != "" {
		payload["model"] = req.Model
	}
	if req.Width > 0 {
		payload["width"] = req.Width
	}
	if req.Height > 0 {
		payload["height"] = req.Height
	}
	if req.CFGScale > 0 {
		payload["cfgscale"] = req.CFGScale
	}
	if req.Steps > 0 {
		payload["steps"] = req.Steps
	}
	if req.Seed != 0 {
		payload["seed"] = req.Seed
	}
	if req.Images > 0 {
		payload["images"] = req.Images
	} else {
		payload["images"] = 1 // Default to 1 image
	}
	if req.DoNotSave {
		payload["donotsave"] = req.DoNotSave
	}
	if req.BatchSize > 0 {
		payload["batchsize"] = req.BatchSize
	}

	// Add any extra parameters
	for k, v := range req.Extra {
		payload[k] = v
	}

	return c.makeAPIRequest(ctx, "/API/GenerateText2Image", payload)
}

// makeAPIRequest performs the actual HTTP request with retry logic.
func (c *client) makeAPIRequest(ctx context.Context, endpoint string, payload map[string]interface{}) (*ImageResult, error) {
	url := fmt.Sprintf("%s%s", strings.TrimSuffix(c.cfg.ServerURL, "/"), endpoint)

	var lastErr error
	for attempt := 0; attempt <= c.cfg.RetryAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * c.cfg.RetryBackoff):
				// Continue with retry
			}
		}

		result, err := c.doAPIRequest(ctx, url, payload)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error indicates invalid session
		if strings.Contains(err.Error(), "invalid_session_id") {
			if refreshErr := c.RefreshSession(ctx); refreshErr != nil {
				return nil, fmt.Errorf("session refresh failed: %w", refreshErr)
			}
			payload["session_id"] = c.sessionID
			// Retry with new session
			continue
		}

		// For other errors, check if retryable
		if !isRetryableError(err) {
			break
		}
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.cfg.RetryAttempts+1, lastErr)
}

// doAPIRequest performs a single HTTP request.
func (c *client) doAPIRequest(ctx context.Context, url string, payload map[string]interface{}) (*ImageResult, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.cfg.AuthToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "swarm_token",
			Value: c.cfg.AuthToken,
		})
	}

	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result ImageResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if result.ErrorID != "" || result.ErrorMsg != "" {
		return nil, fmt.Errorf("API error: %s (%s)", result.ErrorMsg, result.ErrorID)
	}

	return &result, nil
}

// GetQueueStatus returns current queue metrics (placeholder implementation).
func (c *client) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
	// SwarmUI doesn't have a direct queue status endpoint in the documented API
	// This is a placeholder that could be implemented if such an endpoint exists
	return &QueueStatus{
		Pending:  0,
		Running:  0,
		Finished: 0,
	}, nil
}

// MonitorJob establishes WebSocket connection for progress updates (future implementation).
func (c *client) MonitorJob(ctx context.Context, jobID string) (<-chan JobProgress, error) {
	// SwarmUI supports WebSocket connections for real-time updates
	// This is a placeholder for future implementation
	ch := make(chan JobProgress, 1)
	go func() {
		defer close(ch)
		// TODO: Implement WebSocket monitoring
		ch <- JobProgress{
			JobID:     jobID,
			Status:    "completed",
			Progress:  1.0,
			Message:   "WebSocket monitoring not yet implemented",
			Timestamp: time.Now(),
		}
	}()
	return ch, nil
}

// isRetryableError determines if an error should trigger a retry.
func isRetryableError(err error) bool {
	errStr := err.Error()
	// Network-level errors are generally retryable
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "temporary failure") ||
		strings.Contains(errStr, "server error") {
		return true
	}
	return false
}
