package comfyui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestFileWorkflowDiscoverer_DiscoverTemplates tests valid and invalid template discovery.
func TestFileWorkflowDiscoverer_DiscoverTemplates(t *testing.T) {
	dir := t.TempDir()
	valid := `{"base_workflow":{"id":"wf1","nodes":{"n1":"foo"}},"parameters":{},"output_nodes":["out"]}`
	invalid := `{not json}`
	os.WriteFile(filepath.Join(dir, "valid.json"), []byte(valid), 0644)
	os.WriteFile(filepath.Join(dir, "invalid.json"), []byte(invalid), 0644)
	d := &FileWorkflowDiscoverer{}
	ctx := context.Background()
	tmpls, err := d.DiscoverTemplates(ctx, dir)
	if err == nil && len(tmpls) != 1 {
		t.Errorf("expected 1 valid template, got %d", len(tmpls))
	}
}

// TestFileWorkflowDiscoverer_ValidateTemplate tests required fields and placeholder validation.
func TestFileWorkflowDiscoverer_ValidateTemplate(t *testing.T) {
	d := &FileWorkflowDiscoverer{}
	ctx := context.Background()
	tmpl := &WorkflowTemplate{
		BaseWorkflow: &Workflow{ID: "wf1", Nodes: map[string]interface{}{"n1": "foo {{MISSING}}"}},
		Parameters:   map[string]Parameter{},
		OutputNodes:  []string{"out"},
	}
	err := d.ValidateTemplate(ctx, tmpl)
	if err == nil || err.Error() != "missing placeholder substitutions: [MISSING]" {
		t.Errorf("expected missing placeholder error, got %v", err)
	}
	// Fix placeholder
	tmpl.Parameters["MISSING"] = Parameter{}
	err = d.ValidateTemplate(ctx, tmpl)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// Test missing output nodes
	tmpl.OutputNodes = nil
	err = d.ValidateTemplate(ctx, tmpl)
	if err == nil || err.Error() != "workflow missing output nodes" {
		t.Errorf("expected missing output nodes error, got %v", err)
	}
}

// TestFileWorkflowDiscoverer_ContextCancel tests context cancellation during discovery.
func TestFileWorkflowDiscoverer_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	valid := `{"base_workflow":{"id":"wf1","nodes":{"n1":"foo"}},"parameters":{},"output_nodes":["out"]}`
	os.WriteFile(filepath.Join(dir, "valid.json"), []byte(valid), 0644)
	d := &FileWorkflowDiscoverer{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := d.DiscoverTemplates(ctx, dir)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Errorf("expected context canceled error, got %v", err)
	}
}
