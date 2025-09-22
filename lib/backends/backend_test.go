package backends

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	tests := []struct {
		name        string
		backendType BackendType
		wantNil     bool
	}{
		{
			name:        "ComfyUI config",
			backendType: BackendTypeComfyUI,
			wantNil:     false,
		},
		{
			name:        "SwarmUI config",
			backendType: BackendTypeSwarmUI,
			wantNil:     false,
		},
		{
			name:        "invalid backend type",
			backendType: BackendType("invalid"),
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig(tt.backendType)
			if tt.wantNil && cfg != nil {
				t.Errorf("DefaultConfig() = %v, want nil", cfg)
			}
			if !tt.wantNil && cfg == nil {
				t.Error("DefaultConfig() = nil, want non-nil")
			}
			if !tt.wantNil && cfg.Type != tt.backendType {
				t.Errorf("DefaultConfig().Type = %v, want %v", cfg.Type, tt.backendType)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid ComfyUI config",
			config: &Config{
				Type: BackendTypeComfyUI,
				ComfyUI: &ComfyUIConfig{
					ServerURL:     "http://localhost:8188",
					Timeout:       30 * time.Second,
					RetryAttempts: 2,
					RetryBackoff:  500 * time.Millisecond,
				},
			},
			wantErr: false,
		},
		{
			name: "valid SwarmUI config",
			config: &Config{
				Type: BackendTypeSwarmUI,
				SwarmUI: &SwarmUIConfig{
					ServerURL:              "http://localhost:7801",
					Timeout:                30 * time.Second,
					RetryAttempts:          2,
					RetryBackoff:           500 * time.Millisecond,
					SessionRefreshInterval: 30 * time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "ComfyUI config without ComfyUI settings",
			config: &Config{
				Type:    BackendTypeComfyUI,
				ComfyUI: nil,
			},
			wantErr: true,
		},
		{
			name: "SwarmUI config without SwarmUI settings",
			config: &Config{
				Type:    BackendTypeSwarmUI,
				SwarmUI: nil,
			},
			wantErr: true,
		},
		{
			name: "invalid backend type",
			config: &Config{
				Type: BackendType("invalid"),
			},
			wantErr: true,
		},
		{
			name: "ComfyUI config with invalid settings",
			config: &Config{
				Type: BackendTypeComfyUI,
				ComfyUI: &ComfyUIConfig{
					ServerURL: "", // Invalid: empty server URL
					Timeout:   30 * time.Second,
				},
			},
			wantErr: true,
		},
		{
			name: "SwarmUI config with invalid settings",
			config: &Config{
				Type: BackendTypeSwarmUI,
				SwarmUI: &SwarmUIConfig{
					ServerURL: "", // Invalid: empty server URL
					Timeout:   30 * time.Second,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewBackend(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "invalid backend type",
			config: &Config{
				Type: BackendType("invalid"),
			},
			wantErr: true,
		},
		{
			name: "ComfyUI config without ComfyUI settings",
			config: &Config{
				Type:    BackendTypeComfyUI,
				ComfyUI: nil,
			},
			wantErr: true,
		},
		{
			name: "SwarmUI config without SwarmUI settings",
			config: &Config{
				Type:    BackendTypeSwarmUI,
				SwarmUI: nil,
			},
			wantErr: true,
		},
		// Note: We can't easily test successful backend creation without
		// mock clients or actual running services
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewBackend(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBackend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && backend == nil {
				t.Error("NewBackend() returned nil backend without error")
			}
			if backend != nil {
				_ = backend.Close() // Clean up
			}
		})
	}
}

func TestBackendInfo(t *testing.T) {
	tests := []struct {
		name         string
		backendType  BackendType
		expectedName string
		expectedCaps []string
	}{
		{
			name:         "ComfyUI backend info",
			backendType:  BackendTypeComfyUI,
			expectedName: "ComfyUI",
			expectedCaps: []string{
				"text2image",
				"workflow_support",
				"progress_monitoring",
				"queue_management",
			},
		},
		{
			name:         "SwarmUI backend info",
			backendType:  BackendTypeSwarmUI,
			expectedName: "SwarmUI",
			expectedCaps: []string{
				"text2image",
				"session_management",
				"progress_monitoring",
				"multiple_models",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock backend to test GetBackendInfo
			var backend Backend
			switch tt.backendType {
			case BackendTypeComfyUI:
				backend = &comfyUIBackend{}
			case BackendTypeSwarmUI:
				backend = &swarmUIBackend{}
			}

			info := backend.GetBackendInfo()
			if info.Type != tt.backendType {
				t.Errorf("GetBackendInfo().Type = %v, want %v", info.Type, tt.backendType)
			}
			if info.Name != tt.expectedName {
				t.Errorf("GetBackendInfo().Name = %v, want %v", info.Name, tt.expectedName)
			}

			// Check that all expected capabilities are present
			capabilityMap := make(map[string]bool)
			for _, cap := range info.Capabilities {
				capabilityMap[cap] = true
			}

			for _, expectedCap := range tt.expectedCaps {
				if !capabilityMap[expectedCap] {
					t.Errorf("GetBackendInfo().Capabilities missing %v", expectedCap)
				}
			}
		})
	}
}

func TestConvertToSwarmUIRequest(t *testing.T) {
	backend := &swarmUIBackend{}

	req := &GenerateRequest{
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
		BatchSize:      1,
		Quality:        "high",
		Style:          "anime",
		BackendParams: map[string]interface{}{
			"sampler":   "DPM++ 2M",
			"scheduler": "karras",
		},
	}

	swarmReq := backend.convertToSwarmUIRequest(req)

	if swarmReq.Prompt != req.Prompt {
		t.Errorf("convertToSwarmUIRequest().Prompt = %v, want %v", swarmReq.Prompt, req.Prompt)
	}
	if swarmReq.NegativePrompt != req.NegativePrompt {
		t.Errorf("convertToSwarmUIRequest().NegativePrompt = %v, want %v", swarmReq.NegativePrompt, req.NegativePrompt)
	}
	if swarmReq.Model != req.Model {
		t.Errorf("convertToSwarmUIRequest().Model = %v, want %v", swarmReq.Model, req.Model)
	}
	if swarmReq.Width != req.Width {
		t.Errorf("convertToSwarmUIRequest().Width = %v, want %v", swarmReq.Width, req.Width)
	}
	if swarmReq.Height != req.Height {
		t.Errorf("convertToSwarmUIRequest().Height = %v, want %v", swarmReq.Height, req.Height)
	}
	if swarmReq.CFGScale != req.CFGScale {
		t.Errorf("convertToSwarmUIRequest().CFGScale = %v, want %v", swarmReq.CFGScale, req.CFGScale)
	}
	if swarmReq.Steps != req.Steps {
		t.Errorf("convertToSwarmUIRequest().Steps = %v, want %v", swarmReq.Steps, req.Steps)
	}
	if swarmReq.Seed != req.Seed {
		t.Errorf("convertToSwarmUIRequest().Seed = %v, want %v", swarmReq.Seed, req.Seed)
	}
	if swarmReq.Images != req.Images {
		t.Errorf("convertToSwarmUIRequest().Images = %v, want %v", swarmReq.Images, req.Images)
	}
	if swarmReq.DoNotSave != req.DoNotSave {
		t.Errorf("convertToSwarmUIRequest().DoNotSave = %v, want %v", swarmReq.DoNotSave, req.DoNotSave)
	}
	if swarmReq.BatchSize != req.BatchSize {
		t.Errorf("convertToSwarmUIRequest().BatchSize = %v, want %v", swarmReq.BatchSize, req.BatchSize)
	}

	// Check extra parameters
	if swarmReq.Extra["sampler"] != "DPM++ 2M" {
		t.Errorf("convertToSwarmUIRequest().Extra[sampler] = %v, want %v", swarmReq.Extra["sampler"], "DPM++ 2M")
	}
	if swarmReq.Extra["scheduler"] != "karras" {
		t.Errorf("convertToSwarmUIRequest().Extra[scheduler] = %v, want %v", swarmReq.Extra["scheduler"], "karras")
	}
	if swarmReq.Extra["quality"] != "high" {
		t.Errorf("convertToSwarmUIRequest().Extra[quality] = %v, want %v", swarmReq.Extra["quality"], "high")
	}
	if swarmReq.Extra["style"] != "anime" {
		t.Errorf("convertToSwarmUIRequest().Extra[style] = %v, want %v", swarmReq.Extra["style"], "anime")
	}
}

func TestConvertToSwarmUIRequestDefaults(t *testing.T) {
	backend := &swarmUIBackend{}

	req := &GenerateRequest{
		Prompt: "test prompt",
		// All other fields default/empty
	}

	swarmReq := backend.convertToSwarmUIRequest(req)

	// Should default to 1 image
	if swarmReq.Images != 1 {
		t.Errorf("convertToSwarmUIRequest().Images = %v, want 1", swarmReq.Images)
	}
}
