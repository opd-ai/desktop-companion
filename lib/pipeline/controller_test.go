package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/opd-ai/desktop-companion/lib/comfyui"
)

// mockComfyUIClient implements comfyui.Client for testing.
type mockComfyUIClient struct {
	submitWorkflowFunc func(ctx context.Context, wf *comfyui.Workflow) (*comfyui.Job, error)
	monitorJobFunc     func(ctx context.Context, jobID string) (<-chan comfyui.JobProgress, error)
	getResultFunc      func(ctx context.Context, jobID string) (*comfyui.JobResult, error)
	getQueueStatusFunc func(ctx context.Context) (*comfyui.QueueStatus, error)
}

func (m *mockComfyUIClient) SubmitWorkflow(ctx context.Context, wf *comfyui.Workflow) (*comfyui.Job, error) {
	if m.submitWorkflowFunc != nil {
		return m.submitWorkflowFunc(ctx, wf)
	}
	return &comfyui.Job{ID: "test-job-id", Status: "queued"}, nil
}

func (m *mockComfyUIClient) MonitorJob(ctx context.Context, jobID string) (<-chan comfyui.JobProgress, error) {
	if m.monitorJobFunc != nil {
		return m.monitorJobFunc(ctx, jobID)
	}
	ch := make(chan comfyui.JobProgress, 1)
	ch <- comfyui.JobProgress{JobID: jobID, Status: "completed", Progress: 1.0}
	close(ch)
	return ch, nil
}

func (m *mockComfyUIClient) GetResult(ctx context.Context, jobID string) (*comfyui.JobResult, error) {
	if m.getResultFunc != nil {
		return m.getResultFunc(ctx, jobID)
	}
	return &comfyui.JobResult{
		JobID:  jobID,
		Status: "completed",
		Artifacts: []comfyui.Artifact{
			{Filename: "frame_0.png", MIME: "image/png", Data: []byte("fake-png-data")},
			{Filename: "frame_1.png", MIME: "image/png", Data: []byte("fake-png-data")},
		},
	}, nil
}

func (m *mockComfyUIClient) GetQueueStatus(ctx context.Context) (*comfyui.QueueStatus, error) {
	if m.getQueueStatusFunc != nil {
		return m.getQueueStatusFunc(ctx)
	}
	return &comfyui.QueueStatus{Pending: 0, Running: 0, Finished: 1}, nil
}

func TestNewController(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}

	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	if controller == nil {
		t.Fatal("NewController returned nil")
	}
}

func TestNewControllerValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *PipelineConfig
		client  comfyui.Client
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			client:  &mockComfyUIClient{},
			wantErr: true,
		},
		{
			name:    "nil client",
			config:  DefaultPipelineConfig(),
			client:  nil,
			wantErr: true,
		},
		{
			name:    "valid inputs",
			config:  DefaultPipelineConfig(),
			client:  &mockComfyUIClient{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewController(tt.config, tt.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewController() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProcessCharacterValidation(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	// Test nil config
	_, err = controller.ProcessCharacter(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil character config")
	}
}

func TestBuildPrompts(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	pipelineController := controller.(*pipelineController)

	charConfig := DefaultCharacterConfig("romance_tsundere")
	charConfig.Character.Description = "A cute anime girl"
	charConfig.Character.Style = "anime"

	// Test positive prompt building
	positivePrompt := pipelineController.buildPositivePrompt(charConfig, "shy")
	if positivePrompt == "" {
		t.Error("Expected non-empty positive prompt")
	}

	// Should include character description
	if !contains(positivePrompt, "cute anime girl") {
		t.Error("Expected character description in prompt")
	}

	// Should include state modifier
	if !contains(positivePrompt, "blushing") || !contains(positivePrompt, "bashful") {
		t.Error("Expected shy state modifiers in prompt")
	}

	// Test negative prompt building
	negativePrompt := pipelineController.buildNegativePrompt(charConfig, "shy")
	if negativePrompt == "" {
		t.Error("Expected non-empty negative prompt")
	}

	// Should include basic negative terms
	if !contains(negativePrompt, "blurry") {
		t.Error("Expected basic negative terms in prompt")
	}
}

func TestCreateWorkflowForState(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	pipelineController := controller.(*pipelineController)
	charConfig := DefaultCharacterConfig("test")

	workflow, err := pipelineController.createWorkflowForState(charConfig, "idle")
	if err != nil {
		t.Fatalf("createWorkflowForState failed: %v", err)
	}

	if workflow.ID == "" {
		t.Error("Expected workflow to have an ID")
	}

	if workflow.Nodes == nil {
		t.Fatal("Expected workflow to have nodes")
	}

	// Check for required nodes
	if _, exists := workflow.Nodes["prompt"]; !exists {
		t.Error("Expected workflow to have prompt node")
	}

	if _, exists := workflow.Nodes["generation"]; !exists {
		t.Error("Expected workflow to have generation node")
	}

	// Check metadata
	if workflow.Meta == nil {
		t.Fatal("Expected workflow to have metadata")
	}

	if workflow.Meta["archetype"] != "test" {
		t.Error("Expected archetype in metadata")
	}

	if workflow.Meta["state"] != "idle" {
		t.Error("Expected state in metadata")
	}
}

func TestProcessBatchValidation(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	// Test empty configs
	_, err = controller.ProcessBatch(context.Background(), []*CharacterConfig{})
	if err == nil {
		t.Error("Expected error for empty character configs")
	}
}

func TestValidateAssetsValidation(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	// Test empty asset path
	_, err = controller.ValidateAssets(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty asset path")
	}
}

func TestDeployAssetsValidation(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	// Test nil result
	err = controller.DeployAssets(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil process result")
	}

	// Test failed result
	failedResult := &ProcessResult{
		Character: "test",
		Success:   false,
	}
	err = controller.DeployAssets(context.Background(), failedResult)
	if err == nil {
		t.Error("Expected error for failed process result")
	}
}

func TestIsRetryableError(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	pipelineController := controller.(*pipelineController)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"timeout error", fmt.Errorf("request timeout"), true},
		{"connection error", fmt.Errorf("connection refused"), true},
		{"server error", fmt.Errorf("server error 500"), true},
		{"queue error", fmt.Errorf("queue full"), true},
		{"validation error", fmt.Errorf("invalid format"), false},
		{"file not found", fmt.Errorf("file not found"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pipelineController.isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateBatchSummary(t *testing.T) {
	config := DefaultPipelineConfig()
	client := &mockComfyUIClient{}
	controller, err := NewController(config, client)
	if err != nil {
		t.Fatalf("NewController failed: %v", err)
	}

	pipelineController := controller.(*pipelineController)

	// Create test character results
	characters := map[string]*ProcessResult{
		"character1": {
			Character: "character1",
			Success:   true,
			GeneratedAssets: map[string]*GeneratedAsset{
				"idle":    {State: "idle", Metrics: &AssetMetrics{}},
				"talking": {State: "talking", Metrics: &AssetMetrics{}},
			},
			ProcessingTime: 2 * time.Second,
		},
		"character2": {
			Character: "character2",
			Success:   false,
			GeneratedAssets: map[string]*GeneratedAsset{
				"idle": {State: "idle", Metrics: nil}, // Failed asset
			},
			ProcessingTime: 1 * time.Second,
		},
	}

	summary := pipelineController.generateBatchSummary(characters)

	if summary.TotalCharacters != 2 {
		t.Errorf("Expected 2 total characters, got %d", summary.TotalCharacters)
	}

	if summary.SuccessfulCharacters != 1 {
		t.Errorf("Expected 1 successful character, got %d", summary.SuccessfulCharacters)
	}

	if summary.FailedCharacters != 1 {
		t.Errorf("Expected 1 failed character, got %d", summary.FailedCharacters)
	}

	if summary.TotalAssets != 3 {
		t.Errorf("Expected 3 total assets, got %d", summary.TotalAssets)
	}

	if summary.SuccessfulAssets != 2 {
		t.Errorf("Expected 2 successful assets, got %d", summary.SuccessfulAssets)
	}

	if summary.FailedAssets != 1 {
		t.Errorf("Expected 1 failed asset, got %d", summary.FailedAssets)
	}

	expectedSuccessRate := 0.5 // 1 out of 2 characters successful
	if summary.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected success rate %.2f, got %.2f", expectedSuccessRate, summary.SuccessRate)
	}

	expectedAvgTime := time.Duration(1500) * time.Millisecond // (2s + 1s) / 2
	if summary.AverageProcessingTime != expectedAvgTime {
		t.Errorf("Expected average processing time %v, got %v", expectedAvgTime, summary.AverageProcessingTime)
	}
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", true},
		{"hello world", "lo wo", true},
		{"hello world", "xyz", false},
		{"", "test", false},
		{"test", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s contains %s", tt.s, tt.substr), func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("contains(%q, %q) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
