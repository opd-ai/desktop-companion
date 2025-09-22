package backends

// Package backends provides a unified interface for different AI image generation
// backends, including ComfyUI and SwarmUI. This abstraction allows the pipeline
// to work with either backend transparently, following the project's interface-first
// design philosophy.
//
// Design Principles:
//   * Unified interface that abstracts backend-specific details
//   * Factory pattern for backend creation and configuration
//   * Error handling that preserves backend-specific error context
//   * Support for backend-specific capabilities and features
//   * Easy extension for future backend integrations

import (
	"context"
	"fmt"
	"time"

	"github.com/opd-ai/desktop-companion/lib/comfyui"
	"github.com/opd-ai/desktop-companion/lib/swarmui"
)

// Backend represents a unified interface for AI image generation backends.
type Backend interface {
	// GenerateImage generates images based on the provided request
	GenerateImage(ctx context.Context, req *GenerateRequest) (*GenerateResult, error)

	// GetQueueStatus returns current queue status information
	GetQueueStatus(ctx context.Context) (*QueueStatus, error)

	// MonitorJob provides real-time progress updates for a generation job
	MonitorJob(ctx context.Context, jobID string) (<-chan JobProgress, error)

	// GetBackendInfo returns information about the backend type and capabilities
	GetBackendInfo() *BackendInfo

	// Close cleans up any resources used by the backend
	Close() error
}

// BackendType represents the type of backend.
type BackendType string

const (
	BackendTypeComfyUI BackendType = "comfyui"
	BackendTypeSwarmUI BackendType = "swarmui"
)

// BackendInfo provides information about a backend's capabilities.
type BackendInfo struct {
	Type         BackendType            `json:"type"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version,omitempty"`
	Capabilities []string               `json:"capabilities"`
	Features     map[string]interface{} `json:"features,omitempty"`
}

// GenerateRequest represents a unified image generation request.
type GenerateRequest struct {
	// Core parameters supported by all backends
	Prompt         string  `json:"prompt"`
	NegativePrompt string  `json:"negative_prompt,omitempty"`
	Width          int     `json:"width,omitempty"`
	Height         int     `json:"height,omitempty"`
	Steps          int     `json:"steps,omitempty"`
	CFGScale       float64 `json:"cfg_scale,omitempty"`
	Seed           int64   `json:"seed,omitempty"`
	Images         int     `json:"images,omitempty"`

	// Advanced parameters
	Model     string `json:"model,omitempty"`
	Style     string `json:"style,omitempty"`
	Quality   string `json:"quality,omitempty"`
	BatchSize int    `json:"batch_size,omitempty"`
	DoNotSave bool   `json:"do_not_save,omitempty"`

	// Backend-specific parameters
	BackendParams map[string]interface{} `json:"backend_params,omitempty"`

	// Workflow parameters (for ComfyUI)
	Workflow *WorkflowParams `json:"workflow,omitempty"`
}

// WorkflowParams contains ComfyUI-specific workflow parameters.
type WorkflowParams struct {
	ID    string                 `json:"id,omitempty"`
	Nodes map[string]interface{} `json:"nodes,omitempty"`
	Meta  map[string]interface{} `json:"meta,omitempty"`
}

// GenerateResult represents a unified image generation result.
type GenerateResult struct {
	JobID     string          `json:"job_id"`
	Images    []string        `json:"images"`
	Status    string          `json:"status"`
	StartTime time.Time       `json:"start_time"`
	EndTime   time.Time       `json:"end_time,omitempty"`
	Metadata  *ResultMetadata `json:"metadata,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// ResultMetadata contains additional information about the generation result.
type ResultMetadata struct {
	Backend        BackendType            `json:"backend"`
	Model          string                 `json:"model,omitempty"`
	ActualParams   map[string]interface{} `json:"actual_params,omitempty"`
	GenerationTime time.Duration          `json:"generation_time"`
	BackendJobID   string                 `json:"backend_job_id,omitempty"`
}

// QueueStatus represents unified queue status information.
type QueueStatus struct {
	Pending  int         `json:"pending"`
	Running  int         `json:"running"`
	Finished int         `json:"finished"`
	Backend  BackendType `json:"backend"`
}

// JobProgress represents unified job progress information.
type JobProgress struct {
	JobID     string      `json:"job_id"`
	Status    string      `json:"status"`
	Progress  float64     `json:"progress"` // 0.0 to 1.0
	Message   string      `json:"message"`
	Preview   string      `json:"preview,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Backend   BackendType `json:"backend"`
	Error     error       `json:"error,omitempty"`
}

// Config represents unified backend configuration.
type Config struct {
	Type    BackendType    `json:"type"`
	ComfyUI *ComfyUIConfig `json:"comfyui,omitempty"`
	SwarmUI *SwarmUIConfig `json:"swarmui,omitempty"`
}

// ComfyUIConfig contains ComfyUI-specific configuration.
type ComfyUIConfig struct {
	ServerURL     string        `json:"server_url"`
	APIKey        string        `json:"api_key,omitempty"`
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
	RetryBackoff  time.Duration `json:"retry_backoff"`
	WSPath        string        `json:"ws_path,omitempty"`
}

// SwarmUIConfig contains SwarmUI-specific configuration.
type SwarmUIConfig struct {
	ServerURL              string        `json:"server_url"`
	AuthToken              string        `json:"auth_token,omitempty"`
	Timeout                time.Duration `json:"timeout"`
	RetryAttempts          int           `json:"retry_attempts"`
	RetryBackoff           time.Duration `json:"retry_backoff"`
	SessionRefreshInterval time.Duration `json:"session_refresh_interval"`
}

// NewBackend creates a new backend instance based on the provided configuration.
func NewBackend(cfg *Config) (Backend, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	switch cfg.Type {
	case BackendTypeComfyUI:
		return NewComfyUIBackend(cfg.ComfyUI)
	case BackendTypeSwarmUI:
		return NewSwarmUIBackend(cfg.SwarmUI)
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", cfg.Type)
	}
}

// NewComfyUIBackend creates a new ComfyUI backend wrapper.
func NewComfyUIBackend(cfg *ComfyUIConfig) (Backend, error) {
	if cfg == nil {
		return nil, fmt.Errorf("ComfyUI config is required")
	}

	comfyConfig := comfyui.Config{
		ServerURL:     cfg.ServerURL,
		APIKey:        cfg.APIKey,
		Timeout:       cfg.Timeout,
		RetryAttempts: cfg.RetryAttempts,
		RetryBackoff:  cfg.RetryBackoff,
		WSPath:        cfg.WSPath,
	}

	client, err := comfyui.New(comfyConfig)
	if err != nil {
		return nil, fmt.Errorf("creating ComfyUI client: %w", err)
	}

	return &comfyUIBackend{
		client: client,
		config: cfg,
	}, nil
}

// NewSwarmUIBackend creates a new SwarmUI backend wrapper.
func NewSwarmUIBackend(cfg *SwarmUIConfig) (Backend, error) {
	if cfg == nil {
		return nil, fmt.Errorf("SwarmUI config is required")
	}

	swarmConfig := swarmui.Config{
		ServerURL:              cfg.ServerURL,
		AuthToken:              cfg.AuthToken,
		Timeout:                cfg.Timeout,
		RetryAttempts:          cfg.RetryAttempts,
		RetryBackoff:           cfg.RetryBackoff,
		SessionRefreshInterval: cfg.SessionRefreshInterval,
	}

	client, err := swarmui.New(swarmConfig)
	if err != nil {
		return nil, fmt.Errorf("creating SwarmUI client: %w", err)
	}

	return &swarmUIBackend{
		client: client,
		config: cfg,
	}, nil
}

// DefaultConfig returns a default configuration for the specified backend type.
func DefaultConfig(backendType BackendType) *Config {
	switch backendType {
	case BackendTypeComfyUI:
		comfyDefault := comfyui.DefaultConfig()
		return &Config{
			Type: BackendTypeComfyUI,
			ComfyUI: &ComfyUIConfig{
				ServerURL:     comfyDefault.ServerURL,
				Timeout:       comfyDefault.Timeout,
				RetryAttempts: comfyDefault.RetryAttempts,
				RetryBackoff:  comfyDefault.RetryBackoff,
				WSPath:        comfyDefault.WSPath,
			},
		}
	case BackendTypeSwarmUI:
		swarmDefault := swarmui.DefaultConfig()
		return &Config{
			Type: BackendTypeSwarmUI,
			SwarmUI: &SwarmUIConfig{
				ServerURL:              swarmDefault.ServerURL,
				Timeout:                swarmDefault.Timeout,
				RetryAttempts:          swarmDefault.RetryAttempts,
				RetryBackoff:           swarmDefault.RetryBackoff,
				SessionRefreshInterval: swarmDefault.SessionRefreshInterval,
			},
		}
	default:
		return nil
	}
}

// ValidateConfig validates backend configuration.
func ValidateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}

	switch cfg.Type {
	case BackendTypeComfyUI:
		if cfg.ComfyUI == nil {
			return fmt.Errorf("ComfyUI config is required for ComfyUI backend")
		}
		comfyConfig := comfyui.Config{
			ServerURL:     cfg.ComfyUI.ServerURL,
			APIKey:        cfg.ComfyUI.APIKey,
			Timeout:       cfg.ComfyUI.Timeout,
			RetryAttempts: cfg.ComfyUI.RetryAttempts,
			RetryBackoff:  cfg.ComfyUI.RetryBackoff,
			WSPath:        cfg.ComfyUI.WSPath,
		}
		return comfyConfig.Validate()
	case BackendTypeSwarmUI:
		if cfg.SwarmUI == nil {
			return fmt.Errorf("SwarmUI config is required for SwarmUI backend")
		}
		swarmConfig := swarmui.Config{
			ServerURL:              cfg.SwarmUI.ServerURL,
			AuthToken:              cfg.SwarmUI.AuthToken,
			Timeout:                cfg.SwarmUI.Timeout,
			RetryAttempts:          cfg.SwarmUI.RetryAttempts,
			RetryBackoff:           cfg.SwarmUI.RetryBackoff,
			SessionRefreshInterval: cfg.SwarmUI.SessionRefreshInterval,
		}
		return swarmConfig.Validate()
	default:
		return fmt.Errorf("unsupported backend type: %s", cfg.Type)
	}
}
