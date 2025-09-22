package backends

import (
	"context"
	"fmt"
	"time"

	"github.com/opd-ai/desktop-companion/lib/comfyui"
)

// comfyUIBackend wraps a ComfyUI client to implement the unified Backend interface.
type comfyUIBackend struct {
	client comfyui.Client
	config *ComfyUIConfig
}

// GenerateImage implements Backend interface for ComfyUI.
func (b *comfyUIBackend) GenerateImage(ctx context.Context, req *GenerateRequest) (*GenerateResult, error) {
	startTime := time.Now()

	// Convert unified request to ComfyUI workflow format
	workflow, err := b.convertToWorkflow(req)
	if err != nil {
		return nil, fmt.Errorf("converting request to workflow: %w", err)
	}

	// Submit workflow to ComfyUI
	job, err := b.client.SubmitWorkflow(ctx, workflow)
	if err != nil {
		return nil, fmt.Errorf("submitting workflow: %w", err)
	}

	// Get result
	result, err := b.client.GetResult(ctx, job.ID)
	if err != nil {
		return nil, fmt.Errorf("getting result: %w", err)
	}

	endTime := time.Now()

	// Convert ComfyUI result to unified format
	images := make([]string, len(result.Artifacts))
	for i, artifact := range result.Artifacts {
		images[i] = artifact.Filename
	}

	return &GenerateResult{
		JobID:     job.ID,
		Images:    images,
		Status:    "completed",
		StartTime: startTime,
		EndTime:   endTime,
		Metadata: &ResultMetadata{
			Backend:        BackendTypeComfyUI,
			Model:          req.Model,
			GenerationTime: endTime.Sub(startTime),
			BackendJobID:   job.ID,
		},
	}, nil
}

// GetQueueStatus implements Backend interface for ComfyUI.
func (b *comfyUIBackend) GetQueueStatus(ctx context.Context) (*QueueStatus, error) {
	status, err := b.client.GetQueueStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &QueueStatus{
		Pending:  status.Pending,
		Running:  status.Running,
		Finished: status.Finished,
		Backend:  BackendTypeComfyUI,
	}, nil
}

// MonitorJob implements Backend interface for ComfyUI.
func (b *comfyUIBackend) MonitorJob(ctx context.Context, jobID string) (<-chan JobProgress, error) {
	comfyProgressChan, err := b.client.MonitorJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Convert ComfyUI progress to unified format
	progressChan := make(chan JobProgress, 1)
	go func() {
		defer close(progressChan)
		for comfyProgress := range comfyProgressChan {
			unifiedProgress := JobProgress{
				JobID:     jobID,
				Status:    comfyProgress.Status,
				Progress:  comfyProgress.Progress,
				Message:   comfyProgress.Message,
				Timestamp: time.Now(), // ComfyUI doesn't provide timestamp, use current time
				Backend:   BackendTypeComfyUI,
				Error:     comfyProgress.Err,
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

// GetBackendInfo implements Backend interface for ComfyUI.
func (b *comfyUIBackend) GetBackendInfo() *BackendInfo {
	return &BackendInfo{
		Type: BackendTypeComfyUI,
		Name: "ComfyUI",
		Capabilities: []string{
			"text2image",
			"workflow_support",
			"progress_monitoring",
			"queue_management",
		},
		Features: map[string]interface{}{
			"supports_workflows": true,
			"supports_websocket": true,
			"supports_queue":     true,
		},
	}
}

// Close implements Backend interface for ComfyUI.
func (b *comfyUIBackend) Close() error {
	// ComfyUI client doesn't require explicit cleanup
	return nil
}

// convertToWorkflow converts a unified GenerateRequest to ComfyUI workflow format.
func (b *comfyUIBackend) convertToWorkflow(req *GenerateRequest) (*comfyui.Workflow, error) {
	// If workflow is explicitly provided, use it
	if req.Workflow != nil {
		return &comfyui.Workflow{
			ID:    req.Workflow.ID,
			Nodes: req.Workflow.Nodes,
			Meta:  req.Workflow.Meta,
		}, nil
	}

	// Otherwise, create a basic workflow from the request parameters
	// This is a simplified implementation - in practice, you'd want
	// more sophisticated workflow generation
	nodes := make(map[string]interface{})

	// Add text prompt node
	nodes["text_prompt"] = map[string]interface{}{
		"class_type": "CLIPTextEncode",
		"inputs": map[string]interface{}{
			"text": req.Prompt,
		},
	}

	// Add negative prompt node if provided
	if req.NegativePrompt != "" {
		nodes["negative_prompt"] = map[string]interface{}{
			"class_type": "CLIPTextEncode",
			"inputs": map[string]interface{}{
				"text": req.NegativePrompt,
			},
		}
	}

	// Add sampling parameters
	samplerNode := map[string]interface{}{
		"class_type": "KSampler",
		"inputs": map[string]interface{}{
			"seed":  req.Seed,
			"steps": req.Steps,
		},
	}

	if req.CFGScale > 0 {
		samplerNode["inputs"].(map[string]interface{})["cfg"] = req.CFGScale
	}

	nodes["sampler"] = samplerNode

	// Add image dimensions
	if req.Width > 0 && req.Height > 0 {
		nodes["empty_latent"] = map[string]interface{}{
			"class_type": "EmptyLatentImage",
			"inputs": map[string]interface{}{
				"width":      req.Width,
				"height":     req.Height,
				"batch_size": req.Images,
			},
		}
	}

	// Create workflow metadata
	meta := map[string]interface{}{
		"generated_from":  "unified_backend",
		"original_prompt": req.Prompt,
	}

	if req.Model != "" {
		meta["model"] = req.Model
	}

	// Add any backend-specific parameters
	for k, v := range req.BackendParams {
		meta[k] = v
	}

	return &comfyui.Workflow{
		ID:    fmt.Sprintf("generated_%d", time.Now().UnixNano()),
		Nodes: nodes,
		Meta:  meta,
	}, nil
}
