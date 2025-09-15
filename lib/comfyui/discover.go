// Package comfyui provides ComfyUI integration primitives.
package comfyui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Import Workflow from comfyui package. Define WorkflowTemplate and Parameter minimally here for compilation.
// WorkflowTemplate wraps a Workflow and adds parameters/output nodes for validation.
type WorkflowTemplate struct {
	BaseWorkflow *Workflow
	Parameters   map[string]Parameter
	OutputNodes  []string
}

// Parameter is a placeholder for future type details.
type Parameter struct{}

// WorkflowDiscoverer defines methods for discovering and validating workflow templates.
type WorkflowDiscoverer interface {
	DiscoverTemplates(ctx context.Context, dir string) ([]*WorkflowTemplate, error)
	ValidateTemplate(ctx context.Context, tmpl *WorkflowTemplate) error
}

// FileWorkflowDiscoverer implements WorkflowDiscoverer using the filesystem.
type FileWorkflowDiscoverer struct{}

// DiscoverTemplates scans dir for .json workflow templates and parses them.
func (f *FileWorkflowDiscoverer) DiscoverTemplates(ctx context.Context, dir string) ([]*WorkflowTemplate, error) {
	var templates []*WorkflowTemplate
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return fmt.Errorf("read %s: %w", path, readErr)
		}
		var tmpl WorkflowTemplate
		if err := json.Unmarshal(data, &tmpl); err != nil {
			return fmt.Errorf("invalid JSON in %s: %w", path, err)
		}
		templates = append(templates, &tmpl)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// ValidateTemplate checks for required fields and placeholder tokens in a WorkflowTemplate.
func (f *FileWorkflowDiscoverer) ValidateTemplate(ctx context.Context, tmpl *WorkflowTemplate) error {
	if tmpl == nil || tmpl.BaseWorkflow == nil {
		return errors.New("missing base workflow")
	}
	wf := tmpl.BaseWorkflow
	if wf.ID == "" {
		return errors.New("workflow missing ID")
	}
	if len(wf.Nodes) == 0 {
		return errors.New("workflow missing nodes")
	}
	// Connections field not present; only check nodes and meta.
	// Check for placeholder tokens in nodes and metadata
	missing := findMissingPlaceholders(wf.Nodes, tmpl.Parameters)
	if len(missing) > 0 {
		return fmt.Errorf("missing placeholder substitutions: %v", missing)
	}
	// Check output nodes
	if len(tmpl.OutputNodes) == 0 {
		return errors.New("workflow missing output nodes")
	}
	return nil
}

// findMissingPlaceholders returns a list of placeholder tokens in nodes not present in parameters.
func findMissingPlaceholders(nodes map[string]interface{}, params map[string]Parameter) []string {
	var missing []string
	for _, v := range nodes {
		s := fmt.Sprintf("%v", v)
		tokens := findTokens(s)
		for _, token := range tokens {
			if _, ok := params[token]; !ok {
				missing = append(missing, token)
			}
		}
	}
	return missing
}

// findTokens extracts {{KEY}} tokens from a string.
func findTokens(s string) []string {
	var tokens []string
	start := 0
	for {
		i := strings.Index(s[start:], "{{")
		if i == -1 {
			break
		}
		j := strings.Index(s[start+i:], "}}")
		if j == -1 {
			break
		}
		token := s[start+i+2 : start+i+j]
		tokens = append(tokens, token)
		start += i + j + 2
	}
	return tokens
}
