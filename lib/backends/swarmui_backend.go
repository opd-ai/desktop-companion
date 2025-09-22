package backends

import (
	"context"
	"fmt"
	"time"

	"github.com/opd-ai/desktop-companion/lib/swarmui"
)

// swarmUIBackend wraps a SwarmUI client to implement the unified Backend interface.
type swarmUIBackend struct {
	client swarmui.Client
	config *SwarmUIConfig
}

// GenerateImage implements Backend interface for SwarmUI.
func (b *swarmUIBackend) GenerateImage(ctx context.Context, req *GenerateRequest) (*GenerateResult, error) {
	startTime := time.Now()

	// Convert unified request to SwarmUI format
	swarmReq := b.convertToSwarmUIRequest(req)

	// Generate image using SwarmUI
	result, err := b.client.GenerateImage(ctx, swarmReq)
	if err != nil {
		return nil, fmt.Errorf("generating image: %w", err)
	}

	endTime := time.Now()

	return &GenerateResult{
		JobID:     result.JobID,
		Images:    result.Images,
		Status:    result.Status,
		StartTime: startTime,
		EndTime:   endTime,
		Metadata: &ResultMetadata{
			Backend:        BackendTypeSwarmUI,
			Model:          req.Model,
			GenerationTime: endTime.Sub(startTime),
			BackendJobID:   result.JobID,
		},
		Error: result.ErrorMsg,
	}, nil
}

// GetQueueStatus implements Backend interface for SwarmUI.
func (b *swarmUIBackend) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
	status, err := b.client.GetQueueStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &QueueStatus{
		Pending:  status.Pending,
		Running:  status.Running,
		Finished: status.Finished,
		Backend:  BackendTypeSwarmUI,
	}, nil
}

// MonitorJob implements Backend interface for SwarmUI.
func (b *swarmUIBackend) MonitorJob(ctx context.Context, jobID string) (<-chan JobProgress, error) {
	swarmProgressChan, err := b.client.MonitorJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Convert SwarmUI progress to unified format
	progressChan := make(chan JobProgress, 1)
	go func() {
		defer close(progressChan)
		for swarmProgress := range swarmProgressChan {
			unifiedProgress := JobProgress{
				JobID:     jobID,
				Status:    swarmProgress.Status,
				Progress:  swarmProgress.Progress,
				Message:   swarmProgress.Message,
				Preview:   swarmProgress.Preview,
				Timestamp: swarmProgress.Timestamp,
				Backend:   BackendTypeSwarmUI,
				Error:     swarmProgress.Err,
			}

			select {
			case progressChan <- unifiedProgress:
			case <-ctx.Done():
				return
			}
		}
	}()

	return progressChan, nil
}

// GetBackendInfo implements Backend interface for SwarmUI.
func (b *swarmUIBackend) GetBackendInfo() *BackendInfo {
	return &BackendInfo{
		Type: BackendTypeSwarmUI,
		Name: "SwarmUI",
		Capabilities: []string{
			"text2image",
			"session_management",
			"progress_monitoring",
			"multiple_models",
		},
		Features: map[string]interface{}{
			"supports_sessions":    true,
			"supports_websocket":   true,
			"supports_auth":        true,
			"auto_session_refresh": true,
		},
	}
}

// Close implements Backend interface for SwarmUI.
func (b *swarmUIBackend) Close() error {
	// SwarmUI client doesn't require explicit cleanup
	return nil
}

// convertToSwarmUIRequest converts a unified GenerateRequest to SwarmUI format.
func (b *swarmUIBackend) convertToSwarmUIRequest(req *GenerateRequest) *swarmui.ImageRequest {
	swarmReq := &swarmui.ImageRequest{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Model:          req.Model,
		Width:          req.Width,
		Height:         req.Height,
		CFGScale:       req.CFGScale,
		Steps:          req.Steps,
		Seed:           req.Seed,
		Images:         req.Images,
		DoNotSave:      req.DoNotSave,
		BatchSize:      req.BatchSize,
	}

	// Set defaults if not specified
	if swarmReq.Images == 0 {
		swarmReq.Images = 1
	}

	// Add backend-specific parameters
	if req.BackendParams != nil {
		swarmReq.Extra = make(map[string]interface{})
		for k, v := range req.BackendParams {
			swarmReq.Extra[k] = v
		}
	}

	// Apply quality settings
	if req.Quality != "" {
		if swarmReq.Extra == nil {
			swarmReq.Extra = make(map[string]interface{})
		}
		swarmReq.Extra["quality"] = req.Quality
	}

	// Apply style settings
	if req.Style != "" {
		if swarmReq.Extra == nil {
			swarmReq.Extra = make(map[string]interface{})
		}
		swarmReq.Extra["style"] = req.Style
	}

	return swarmReq
}
