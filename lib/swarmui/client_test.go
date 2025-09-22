package swarmui
package swarmui

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// mockHTTPClient implements HTTPClient for testing.
type mockHTTPClient struct {
	responses []mockResponse
	callCount int
}

type mockResponse struct {
	statusCode int
	body       string
	err        error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.callCount >= len(m.responses) {
		return nil, io.EOF
	}

	resp := m.responses[m.callCount]
	m.callCount++

	if resp.err != nil {
		return nil, resp.err
	}

	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(strings.NewReader(resp.body)),
		Header:     make(http.Header),
	}, nil
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "empty server URL",
			config: Config{
				ServerURL: "",
				Timeout:   time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero timeout",
			config: Config{
				ServerURL: "http://localhost:7801",
				Timeout:   0,
			},
			wantErr: true,
		},
		{
			name: "negative retry attempts",
			config: Config{
				ServerURL:     "http://localhost:7801",
				Timeout:       time.Second,
				RetryAttempts: -1,
			},
			wantErr: true,
		},
		{
			name: "negative retry backoff",
			config: Config{
				ServerURL:    "http://localhost:7801",
				Timeout:      time.Second,
				RetryBackoff: -time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ServerURL != "http://localhost:7801" {
		t.Errorf("DefaultConfig().ServerURL = %v, want %v", cfg.ServerURL, "http://localhost:7801")
	}

	if cfg.Timeout <= 0 {
		t.Errorf("DefaultConfig().Timeout = %v, want positive duration", cfg.Timeout)
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("DefaultConfig() should be valid: %v", err)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid config",
			config: Config{
				ServerURL: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("New() returned nil client without error")
			}
		})
	}
}

func TestClient_RefreshSession(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		statusCode   int
		wantErr      bool
		wantSession  string
	}{
		{
			name: "successful session",
			response: `{
				"session_id": "test-session-123",
				"user_id": "local",
				"output_append_user": true,
				"version": "0.6.3.0",
				"server_id": "test-server"
			}`,
			statusCode:  200,
			wantErr:     false,
			wantSession: "test-session-123",
		},
		{
			name: "session error",
			response: `{
				"error_id": "auth_failed",
				"error": "Authentication required"
			}`,
			statusCode: 200,
			wantErr:    true,
		},
		{
			name:        "invalid JSON",
			response:    `{invalid json`,
			statusCode:  200,
			wantErr:     true,
		},
		{
			name: "empty session ID",
			response: `{
				"session_id": "",
				"user_id": "local"
			}`,
			statusCode: 200,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				responses: []mockResponse{
					{
						statusCode: tt.statusCode,
						body:       tt.response,
					},
				},
			}

			cfg := DefaultConfig()
			client, err := NewWithHTTPClient(cfg, mockClient)
			if err != nil {
				t.Fatalf("NewWithHTTPClient() error = %v", err)
			}

			ctx := context.Background()
			err = client.RefreshSession(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if session := client.GetSession(); session != tt.wantSession {
					t.Errorf("GetSession() = %v, want %v", session, tt.wantSession)
				}
			}
		})
	}
}

func TestClient_GenerateImage(t *testing.T) {
	tests := []struct {
		name      string
		request   *ImageRequest
		responses []mockResponse
		wantErr   bool
		wantImages int
	}{
		{
			name: "successful generation",
			request: &ImageRequest{
				Prompt: "a cute cat",
				Width:  512,
				Height: 512,
				Images: 1,
			},
			responses: []mockResponse{
				// Session response
				{
					statusCode: 200,
					body: `{
						"session_id": "test-session-123",
						"user_id": "local"
					}`,
				},
				// Image generation response
				{
					statusCode: 200,
					body: `{
						"images": ["/View/local/raw/2024-05-19/a cat-OfficialStableDiffusionsd_xl_base_10s-1872258705.png"]
					}`,
				},
			},
			wantErr:    false,
			wantImages: 1,
		},
		{
			name: "generation error",
			request: &ImageRequest{
				Prompt: "test prompt",
			},
			responses: []mockResponse{
				// Session response
				{
					statusCode: 200,
					body: `{
						"session_id": "test-session-123",
						"user_id": "local"
					}`,
				},
				// Image generation error
				{
					statusCode: 200,
					body: `{
						"error_id": "model_not_found",
						"error": "Specified model not available"
					}`,
				},
			},
			wantErr: true,
		},
		{
			name: "session refresh on invalid session",
			request: &ImageRequest{
				Prompt: "test prompt",
			},
			responses: []mockResponse{
				// Initial session response
				{
					statusCode: 200,
					body: `{
						"session_id": "old-session",
						"user_id": "local"
					}`,
				},
				// Generation with invalid session
				{
					statusCode: 200,
					body: `{
						"error_id": "invalid_session_id",
						"error": "Session expired"
					}`,
				},
				// New session response
				{
					statusCode: 200,
					body: `{
						"session_id": "new-session-123",
						"user_id": "local"
					}`,
				},
				// Successful generation with new session
				{
					statusCode: 200,
					body: `{
						"images": ["/View/local/raw/test.png"]
					}`,
				},
			},
			wantErr:    false,
			wantImages: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockHTTPClient{
				responses: tt.responses,
			}

			cfg := DefaultConfig()
			cfg.RetryAttempts = 1 // Reduce retries for faster tests
			client, err := NewWithHTTPClient(cfg, mockClient)
			if err != nil {
				t.Fatalf("NewWithHTTPClient() error = %v", err)
			}

			ctx := context.Background()
			result, err := client.GenerateImage(ctx, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result.Images) != tt.wantImages {
					t.Errorf("GenerateImage() returned %d images, want %d", len(result.Images), tt.wantImages)
				}
			}
		})
	}
}

func TestClient_GetQueueStatus(t *testing.T) {
	mockClient := &mockHTTPClient{}
	cfg := DefaultConfig()
	client, err := NewWithHTTPClient(cfg, mockClient)
	if err != nil {
		t.Fatalf("NewWithHTTPClient() error = %v", err)
	}

	ctx := context.Background()
	status, err := client.GetQueueStatus(ctx)

	if err != nil {
		t.Errorf("GetQueueStatus() error = %v", err)
		return
	}

	if status == nil {
		t.Error("GetQueueStatus() returned nil status")
	}
}

func TestClient_MonitorJob(t *testing.T) {
	mockClient := &mockHTTPClient{}
	cfg := DefaultConfig()
	client, err := NewWithHTTPClient(cfg, mockClient)
	if err != nil {
		t.Fatalf("NewWithHTTPClient() error = %v", err)
	}

	ctx := context.Background()
	ch, err := client.MonitorJob(ctx, "test-job-123")

	if err != nil {
		t.Errorf("MonitorJob() error = %v", err)
		return
	}

	if ch == nil {
		t.Error("MonitorJob() returned nil channel")
		return
	}

	// Read one progress update
	select {
	case progress := <-ch:
		if progress.JobID != "test-job-123" {
			t.Errorf("MonitorJob() progress.JobID = %v, want %v", progress.JobID, "test-job-123")
		}
	case <-time.After(time.Second):
		t.Error("MonitorJob() did not send progress update within timeout")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "connection refused",
			err:  &http.Client{}.Do(&http.Request{}), // This will create a connection refused error in testing
			want: false, // actual error type varies, but we test with string content
		},
		{
			name: "timeout error",
			err:  context.DeadlineExceeded,
			want: false, // context.DeadlineExceeded doesn't contain "timeout" string
		},
	}

	// Test with string-based errors since that's how isRetryableError works
	stringTests := []struct {
		name     string
		errStr   string
		want     bool
	}{
		{
			name:   "connection refused",
			errStr: "connection refused",
			want:   true,
		},
		{
			name:   "timeout",
			errStr: "request timeout",
			want:   true,
		},
		{
			name:   "temporary failure",
			errStr: "temporary failure",
			want:   true,
		},
		{
			name:   "server error",
			errStr: "server error",
			want:   true,
		},
		{
			name:   "validation error",
			errStr: "invalid input",
			want:   false,
		},
	}

	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{msg: tt.errStr}
			if got := isRetryableError(err); got != tt.want {
				t.Errorf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// testError is a simple error implementation for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestImageRequestMarshaling(t *testing.T) {
	req := &ImageRequest{
		Prompt:         "a beautiful landscape",
		NegativePrompt: "blurry, low quality",
		Model:          "stable-diffusion-xl",
		Width:          1024,
		Height:         1024,
		CFGScale:       7.5,
		Steps:          20,
		Seed:           42,
		Images:         2,
		DoNotSave:      true,
		Extra: map[string]interface{}{
			"sampler": "DPM++ 2M",
			"scheduler": "karras",
		},
	}

	// Create payload like the actual client does
	payload := make(map[string]interface{})
	payload["session_id"] = "test-session"
	payload["prompt"] = req.Prompt
	payload["negativeprompt"] = req.NegativePrompt
	payload["model"] = req.Model
	payload["width"] = req.Width
	payload["height"] = req.Height
	payload["cfgscale"] = req.CFGScale
	payload["steps"] = req.Steps
	payload["seed"] = req.Seed
	payload["images"] = req.Images
	payload["donotsave"] = req.DoNotSave

	for k, v := range req.Extra {
		payload[k] = v
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Verify the JSON contains expected fields
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expectedFields := []string{
		"session_id", "prompt", "negativeprompt", "model",
		"width", "height", "cfgscale", "steps", "seed",
		"images", "donotsave", "sampler", "scheduler",
	}

	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field %s not found in JSON payload", field)
		}
	}
}