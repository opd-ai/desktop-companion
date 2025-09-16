package comfyui

// template.go provides workflow template management for dynamic prompt injection
// and parameter customization. This implements the workflow template system
// outlined in GIF_PLAN.md with support for parameterized workflows.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateWorkflow represents a parameterized ComfyUI workflow.
type TemplateWorkflow struct {
	ID           string                       `json:"id"`            // Template identifier
	Name         string                       `json:"name"`          // Human-readable name
	Description  string                       `json:"description"`   // Template description
	Version      string                       `json:"version"`       // Template version
	BaseWorkflow *Workflow                    `json:"base_workflow"` // Base workflow structure
	Parameters   map[string]TemplateParameter `json:"parameters"`    // Template parameters
	PromptSlots  []PromptSlot                 `json:"prompt_slots"`  // Prompt injection points
	OutputNodes  []string                     `json:"output_nodes"`  // Output node identifiers
	Metadata     TemplateMetadata             `json:"metadata"`      // Additional metadata
}

// TemplateParameter defines a template parameter with validation.
type TemplateParameter struct {
	Name         string      `json:"name"`                    // Parameter name
	Type         string      `json:"type"`                    // Parameter type (string, int, float, bool, enum)
	Description  string      `json:"description"`             // Parameter description
	Required     bool        `json:"required"`                // Whether parameter is required
	DefaultValue interface{} `json:"default_value,omitempty"` // Default value
	MinValue     *float64    `json:"min_value,omitempty"`     // Minimum value (for numeric types)
	MaxValue     *float64    `json:"max_value,omitempty"`     // Maximum value (for numeric types)
	EnumValues   []string    `json:"enum_values,omitempty"`   // Valid values (for enum type)
	Validation   string      `json:"validation,omitempty"`    // Validation regex (for string type)
}

// PromptSlot defines a prompt injection point in the workflow.
type PromptSlot struct {
	Name        string `json:"name"`        // Slot name (e.g., "positive_prompt", "negative_prompt")
	NodePath    string `json:"node_path"`   // Path to the node containing the prompt
	Field       string `json:"field"`       // Field name within the node
	Template    string `json:"template"`    // Go template string for prompt generation
	Description string `json:"description"` // Slot description
}

// TemplateMetadata contains additional template information.
type TemplateMetadata struct {
	Author     string   `json:"author,omitempty"`     // Template author
	Tags       []string `json:"tags,omitempty"`       // Template tags
	Category   string   `json:"category,omitempty"`   // Template category
	CreatedAt  string   `json:"created_at,omitempty"` // Creation timestamp
	UpdatedAt  string   `json:"updated_at,omitempty"` // Last update timestamp
	Compatible []string `json:"compatible,omitempty"` // Compatible ComfyUI versions
}

// TemplateManager manages workflow templates.
type TemplateManager interface {
	// LoadTemplate loads a workflow template from file
	LoadTemplate(path string) (*TemplateWorkflow, error)

	// SaveTemplate saves a workflow template to file
	SaveTemplate(template *TemplateWorkflow, path string) error

	// ListTemplates lists available templates in a directory
	ListTemplates(dir string) ([]*TemplateWorkflow, error)

	// InstantiateTemplate creates a workflow from template with parameters
	InstantiateTemplate(template *TemplateWorkflow, params map[string]interface{}) (*Workflow, error)

	// ValidateTemplate validates template structure and parameters
	ValidateTemplate(template *TemplateWorkflow) error

	// ValidateParameters validates parameter values against template
	ValidateParameters(template *TemplateWorkflow, params map[string]interface{}) error
}

// templateManager is the concrete implementation of TemplateManager.
type templateManager struct{}

// NewTemplateManager creates a new template manager instance.
func NewTemplateManager() TemplateManager {
	return &templateManager{}
}

// LoadTemplate loads a workflow template from file.
func (tm *templateManager) LoadTemplate(path string) (*TemplateWorkflow, error) {
	if path == "" {
		return nil, errors.New("template path required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read template file: %w", err)
	}

	var tmpl TemplateWorkflow
	if err := json.Unmarshal(data, &tmpl); err != nil {
		return nil, fmt.Errorf("parse template JSON: %w", err)
	}

	if err := tm.ValidateTemplate(&tmpl); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	return &tmpl, nil
}

// SaveTemplate saves a workflow template to file.
func (tm *templateManager) SaveTemplate(tmpl *TemplateWorkflow, path string) error {
	if tmpl == nil {
		return errors.New("template is nil")
	}
	if path == "" {
		return errors.New("template path required")
	}

	if err := tm.ValidateTemplate(tmpl); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	// Create directory if it doesn't exist
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create template directory: %w", err)
		}
	}

	data, err := json.MarshalIndent(tmpl, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal template JSON: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write template file: %w", err)
	}

	return nil
}

// ListTemplates lists available templates in a directory.
func (tm *templateManager) ListTemplates(dir string) ([]*TemplateWorkflow, error) {
	if dir == "" {
		return nil, errors.New("template directory required")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read template directory: %w", err)
	}

	var templates []*TemplateWorkflow
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		templatePath := filepath.Join(dir, entry.Name())
		tmpl, err := tm.LoadTemplate(templatePath)
		if err != nil {
			// Log error but continue with other templates
			continue
		}

		templates = append(templates, tmpl)
	}

	return templates, nil
}

// InstantiateTemplate creates a workflow from template with parameters.
func (tm *templateManager) InstantiateTemplate(tmpl *TemplateWorkflow, params map[string]interface{}) (*Workflow, error) {
	if tmpl == nil {
		return nil, errors.New("template is nil")
	}
	if params == nil {
		params = make(map[string]interface{})
	}

	// Validate parameters
	if err := tm.ValidateParameters(tmpl, params); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w", err)
	}

	// Apply default values for missing parameters
	finalParams := tm.applyDefaultValues(tmpl, params)

	// Deep copy base workflow
	workflow, err := tm.copyWorkflow(tmpl.BaseWorkflow)
	if err != nil {
		return nil, fmt.Errorf("copy base workflow: %w", err)
	}

	// Generate and inject prompts
	if err := tm.injectPrompts(workflow, tmpl, finalParams); err != nil {
		return nil, fmt.Errorf("inject prompts: %w", err)
	}

	// Apply parameter substitutions to workflow nodes
	if err := tm.applyParameterSubstitutions(workflow, tmpl, finalParams); err != nil {
		return nil, fmt.Errorf("apply parameter substitutions: %w", err)
	}

	return workflow, nil
}

// ValidateTemplate validates template structure and parameters.
func (tm *templateManager) ValidateTemplate(tmpl *TemplateWorkflow) error {
	if tmpl == nil {
		return errors.New("template is nil")
	}

	if tmpl.ID == "" {
		return errors.New("template ID required")
	}

	if tmpl.Name == "" {
		return errors.New("template name required")
	}

	if tmpl.BaseWorkflow == nil {
		return errors.New("base workflow required")
	}

	if tmpl.BaseWorkflow.Nodes == nil || len(tmpl.BaseWorkflow.Nodes) == 0 {
		return errors.New("base workflow must have nodes")
	}

	// Validate parameters
	for paramName, param := range tmpl.Parameters {
		if param.Name == "" {
			param.Name = paramName
		}

		if err := tm.validateParameter(&param); err != nil {
			return fmt.Errorf("invalid parameter %s: %w", paramName, err)
		}
	}

	// Validate prompt slots
	for i, slot := range tmpl.PromptSlots {
		if err := tm.validatePromptSlot(&slot); err != nil {
			return fmt.Errorf("invalid prompt slot %d: %w", i, err)
		}
	}

	return nil
}

// ValidateParameters validates parameter values against template.
func (tm *templateManager) ValidateParameters(tmpl *TemplateWorkflow, params map[string]interface{}) error {
	if tmpl == nil {
		return errors.New("template is nil")
	}

	// Check required parameters
	for paramName, param := range tmpl.Parameters {
		value, exists := params[paramName]

		if param.Required && !exists {
			return fmt.Errorf("required parameter missing: %s", paramName)
		}

		if exists {
			if err := tm.validateParameterValue(&param, value); err != nil {
				return fmt.Errorf("invalid value for parameter %s: %w", paramName, err)
			}
		}
	}

	// Check for unknown parameters
	for paramName := range params {
		if _, exists := tmpl.Parameters[paramName]; !exists {
			return fmt.Errorf("unknown parameter: %s", paramName)
		}
	}

	return nil
}

// validateParameter validates a parameter definition.
func (tm *templateManager) validateParameter(param *TemplateParameter) error {
	if param.Name == "" {
		return errors.New("parameter name required")
	}

	validTypes := []string{"string", "int", "float", "bool", "enum"}
	validType := false
	for _, t := range validTypes {
		if param.Type == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid parameter type: %s", param.Type)
	}

	if param.Type == "enum" && len(param.EnumValues) == 0 {
		return errors.New("enum parameter must have enum_values")
	}

	if param.MinValue != nil && param.MaxValue != nil && *param.MinValue > *param.MaxValue {
		return errors.New("min_value cannot be greater than max_value")
	}

	// Validate default value if provided
	if param.DefaultValue != nil {
		if err := tm.validateParameterValue(param, param.DefaultValue); err != nil {
			return fmt.Errorf("invalid default value: %w", err)
		}
	}

	return nil
}

// validatePromptSlot validates a prompt slot definition.
func (tm *templateManager) validatePromptSlot(slot *PromptSlot) error {
	if slot.Name == "" {
		return errors.New("prompt slot name required")
	}

	if slot.NodePath == "" {
		return errors.New("prompt slot node path required")
	}

	if slot.Field == "" {
		return errors.New("prompt slot field required")
	}

	// Validate template syntax
	if slot.Template != "" {
		_, err := template.New(slot.Name).Parse(slot.Template)
		if err != nil {
			return fmt.Errorf("invalid template syntax: %w", err)
		}
	}

	return nil
}

// validateParameterValue validates a parameter value against its definition.
func (tm *templateManager) validateParameterValue(param *TemplateParameter, value interface{}) error {
	switch param.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		// TODO: Add regex validation if param.Validation is set

	case "int":
		var intVal int64
		switch v := value.(type) {
		case int:
			intVal = int64(v)
		case int64:
			intVal = v
		case float64:
			intVal = int64(v)
		default:
			return fmt.Errorf("expected int, got %T", value)
		}

		if param.MinValue != nil && float64(intVal) < *param.MinValue {
			return fmt.Errorf("value %d below minimum %g", intVal, *param.MinValue)
		}
		if param.MaxValue != nil && float64(intVal) > *param.MaxValue {
			return fmt.Errorf("value %d above maximum %g", intVal, *param.MaxValue)
		}

	case "float":
		var floatVal float64
		switch v := value.(type) {
		case float64:
			floatVal = v
		case int:
			floatVal = float64(v)
		case int64:
			floatVal = float64(v)
		default:
			return fmt.Errorf("expected float, got %T", value)
		}

		if param.MinValue != nil && floatVal < *param.MinValue {
			return fmt.Errorf("value %g below minimum %g", floatVal, *param.MinValue)
		}
		if param.MaxValue != nil && floatVal > *param.MaxValue {
			return fmt.Errorf("value %g above maximum %g", floatVal, *param.MaxValue)
		}

	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}

	case "enum":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for enum, got %T", value)
		}

		validValue := false
		for _, enumVal := range param.EnumValues {
			if strVal == enumVal {
				validValue = true
				break
			}
		}
		if !validValue {
			return fmt.Errorf("invalid enum value %s, must be one of: %v", strVal, param.EnumValues)
		}
	}

	return nil
}

// applyDefaultValues applies default values for missing parameters.
func (tm *templateManager) applyDefaultValues(tmpl *TemplateWorkflow, params map[string]interface{}) map[string]interface{} {
	finalParams := make(map[string]interface{})

	// Copy provided parameters
	for k, v := range params {
		finalParams[k] = v
	}

	// Apply defaults for missing parameters
	for paramName, param := range tmpl.Parameters {
		if _, exists := finalParams[paramName]; !exists && param.DefaultValue != nil {
			finalParams[paramName] = param.DefaultValue
		}
	}

	return finalParams
}

// copyWorkflow creates a deep copy of a workflow.
func (tm *templateManager) copyWorkflow(workflow *Workflow) (*Workflow, error) {
	if workflow == nil {
		return nil, errors.New("workflow is nil")
	}

	// Use JSON marshaling/unmarshaling for deep copy
	data, err := json.Marshal(workflow)
	if err != nil {
		return nil, fmt.Errorf("marshal workflow: %w", err)
	}

	var copy Workflow
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, fmt.Errorf("unmarshal workflow: %w", err)
	}

	return &copy, nil
}

// injectPrompts generates and injects prompts into workflow nodes.
func (tm *templateManager) injectPrompts(workflow *Workflow, tmpl *TemplateWorkflow, params map[string]interface{}) error {
	for _, slot := range tmpl.PromptSlots {
		// Generate prompt from template
		promptValue, err := tm.executePromptTemplate(slot.Template, params)
		if err != nil {
			return fmt.Errorf("execute prompt template %s: %w", slot.Name, err)
		}

		// Inject into workflow
		if err := tm.setNodeField(workflow, slot.NodePath, slot.Field, promptValue); err != nil {
			return fmt.Errorf("inject prompt %s: %w", slot.Name, err)
		}
	}

	return nil
}

// executePromptTemplate executes a prompt template with parameters.
func (tm *templateManager) executePromptTemplate(templateStr string, params map[string]interface{}) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	tmpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, params); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return result.String(), nil
}

// applyParameterSubstitutions applies parameter substitutions to workflow nodes.
func (tm *templateManager) applyParameterSubstitutions(workflow *Workflow, tmpl *TemplateWorkflow, params map[string]interface{}) error {
	// This is a simplified implementation that replaces parameter placeholders
	// in workflow nodes. A full implementation would support more sophisticated
	// parameter binding mechanisms.

	data, err := json.Marshal(workflow.Nodes)
	if err != nil {
		return fmt.Errorf("marshal nodes: %w", err)
	}

	nodeStr := string(data)

	// Replace parameter placeholders
	for paramName, value := range params {
		placeholder := fmt.Sprintf("{{.%s}}", paramName)
		replacement := fmt.Sprintf("%v", value)
		nodeStr = strings.ReplaceAll(nodeStr, placeholder, replacement)
	}

	if err := json.Unmarshal([]byte(nodeStr), &workflow.Nodes); err != nil {
		return fmt.Errorf("unmarshal nodes: %w", err)
	}

	return nil
}

// setNodeField sets a field value in a workflow node.
func (tm *templateManager) setNodeField(workflow *Workflow, nodePath, field string, value interface{}) error {
	// Parse node path (e.g., "prompt.positive" -> node "prompt", field "positive")
	pathParts := strings.Split(nodePath, ".")
	if len(pathParts) < 1 {
		return fmt.Errorf("invalid node path: %s", nodePath)
	}

	nodeID := pathParts[0]
	node, exists := workflow.Nodes[nodeID]
	if !exists {
		return fmt.Errorf("node not found: %s", nodeID)
	}

	// Navigate to nested field if path has multiple parts
	current := node
	for i := 1; i < len(pathParts)-1; i++ {
		nodeMap, ok := current.(map[string]interface{})
		if !ok {
			return fmt.Errorf("node %s is not a map", strings.Join(pathParts[:i+1], "."))
		}

		next, exists := nodeMap[pathParts[i]]
		if !exists {
			// Create intermediate map if it doesn't exist
			next = make(map[string]interface{})
			nodeMap[pathParts[i]] = next
		}
		current = next
	}

	// Set the final field value
	nodeMap, ok := current.(map[string]interface{})
	if !ok {
		return fmt.Errorf("target node is not a map")
	}

	nodeMap[field] = value

	return nil
}

// CreateBasicTemplate creates a basic workflow template for character generation.
func CreateBasicTemplate(templateID, style string) *TemplateWorkflow {
	return &TemplateWorkflow{
		ID:          templateID,
		Name:        fmt.Sprintf("Basic Character Generation (%s)", style),
		Description: fmt.Sprintf("Basic template for generating %s style character animations", style),
		Version:     "1.0.0",
		BaseWorkflow: &Workflow{
			ID: templateID + "_base",
			Nodes: map[string]interface{}{
				"prompt": map[string]interface{}{
					"positive": "{{.positive_prompt}}",
					"negative": "{{.negative_prompt}}",
				},
				"generation": map[string]interface{}{
					"width":     "{{.width}}",
					"height":    "{{.height}}",
					"steps":     "{{.steps}}",
					"cfg_scale": "{{.cfg_scale}}",
					"sampler":   "{{.sampler}}",
					"scheduler": "{{.scheduler}}",
					"seed":      "{{.seed}}",
				},
			},
		},
		Parameters: map[string]TemplateParameter{
			"positive_prompt": {
				Name:        "positive_prompt",
				Type:        "string",
				Description: "Positive prompt for generation",
				Required:    true,
			},
			"negative_prompt": {
				Name:         "negative_prompt",
				Type:         "string",
				Description:  "Negative prompt for generation",
				Required:     false,
				DefaultValue: "blurry, low quality, distorted",
			},
			"width": {
				Name:         "width",
				Type:         "int",
				Description:  "Image width",
				Required:     false,
				DefaultValue: 128,
				MinValue:     floatPtr(64),
				MaxValue:     floatPtr(512),
			},
			"height": {
				Name:         "height",
				Type:         "int",
				Description:  "Image height",
				Required:     false,
				DefaultValue: 128,
				MinValue:     floatPtr(64),
				MaxValue:     floatPtr(512),
			},
			"steps": {
				Name:         "steps",
				Type:         "int",
				Description:  "Generation steps",
				Required:     false,
				DefaultValue: 20,
				MinValue:     floatPtr(1),
				MaxValue:     floatPtr(100),
			},
			"cfg_scale": {
				Name:         "cfg_scale",
				Type:         "float",
				Description:  "CFG scale",
				Required:     false,
				DefaultValue: 7.0,
				MinValue:     floatPtr(1.0),
				MaxValue:     floatPtr(20.0),
			},
			"sampler": {
				Name:         "sampler",
				Type:         "enum",
				Description:  "Sampling method",
				Required:     false,
				DefaultValue: "euler_a",
				EnumValues:   []string{"euler_a", "euler", "dpm_2", "dpm_2_ancestral", "heun", "dpm_pp_2s_ancestral"},
			},
			"scheduler": {
				Name:         "scheduler",
				Type:         "enum",
				Description:  "Scheduler type",
				Required:     false,
				DefaultValue: "normal",
				EnumValues:   []string{"normal", "karras", "exponential", "sgm_uniform"},
			},
			"seed": {
				Name:         "seed",
				Type:         "int",
				Description:  "Random seed (-1 for random)",
				Required:     false,
				DefaultValue: -1,
				MinValue:     floatPtr(-1),
			},
		},
		PromptSlots: []PromptSlot{
			{
				Name:        "positive_prompt",
				NodePath:    "prompt",
				Field:       "positive",
				Template:    "{{.character_description}}, {{.state_modifier}}, {{.style_prompt}}",
				Description: "Main positive prompt",
			},
			{
				Name:        "negative_prompt",
				NodePath:    "prompt",
				Field:       "negative",
				Template:    "{{.negative_prompt}}{{if .style_negative}}, {{.style_negative}}{{end}}",
				Description: "Main negative prompt",
			},
		},
		OutputNodes: []string{"generation"},
		Metadata: TemplateMetadata{
			Category: "character_generation",
			Tags:     []string{style, "basic", "animation"},
		},
	}
}

// floatPtr returns a pointer to a float64 value.
func floatPtr(f float64) *float64 {
	return &f
}
