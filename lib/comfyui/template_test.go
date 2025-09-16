package comfyui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateManager(t *testing.T) {
	manager := NewTemplateManager()
	if manager == nil {
		t.Fatal("NewTemplateManager returned nil")
	}
}

func TestCreateBasicTemplate(t *testing.T) {
	template := CreateBasicTemplate("test_template", "pixel_art")

	if template.ID != "test_template" {
		t.Errorf("Expected ID 'test_template', got %s", template.ID)
	}

	if template.Name == "" {
		t.Error("Expected non-empty name")
	}

	if template.BaseWorkflow == nil {
		t.Fatal("Expected base workflow to be populated")
	}

	if len(template.Parameters) == 0 {
		t.Error("Expected template to have parameters")
	}

	if len(template.PromptSlots) == 0 {
		t.Error("Expected template to have prompt slots")
	}

	// Test specific parameters
	if _, exists := template.Parameters["positive_prompt"]; !exists {
		t.Error("Expected positive_prompt parameter")
	}

	if _, exists := template.Parameters["width"]; !exists {
		t.Error("Expected width parameter")
	}

	if _, exists := template.Parameters["height"]; !exists {
		t.Error("Expected height parameter")
	}
}

func TestValidateTemplate(t *testing.T) {
	manager := NewTemplateManager()
	tm := manager.(*templateManager)

	// Test valid template
	validTemplate := CreateBasicTemplate("valid", "anime")
	if err := tm.ValidateTemplate(validTemplate); err != nil {
		t.Errorf("Valid template failed validation: %v", err)
	}

	// Test nil template
	if err := tm.ValidateTemplate(nil); err == nil {
		t.Error("Expected error for nil template")
	}

	// Test template without ID
	invalidTemplate := CreateBasicTemplate("", "anime")
	invalidTemplate.ID = ""
	if err := tm.ValidateTemplate(invalidTemplate); err == nil {
		t.Error("Expected error for template without ID")
	}

	// Test template without name
	invalidTemplate = CreateBasicTemplate("test", "anime")
	invalidTemplate.Name = ""
	if err := tm.ValidateTemplate(invalidTemplate); err == nil {
		t.Error("Expected error for template without name")
	}

	// Test template without base workflow
	invalidTemplate = CreateBasicTemplate("test", "anime")
	invalidTemplate.BaseWorkflow = nil
	if err := tm.ValidateTemplate(invalidTemplate); err == nil {
		t.Error("Expected error for template without base workflow")
	}
}

func TestValidateParameters(t *testing.T) {
	manager := NewTemplateManager()
	tm := manager.(*templateManager)
	template := CreateBasicTemplate("test", "anime")

	// Test valid parameters
	validParams := map[string]interface{}{
		"positive_prompt": "a cute anime character",
		"width":           128,
		"height":          128,
		"steps":           20,
		"cfg_scale":       7.0,
		"sampler":         "euler_a",
	}

	if err := tm.ValidateParameters(template, validParams); err != nil {
		t.Errorf("Valid parameters failed validation: %v", err)
	}

	// Test missing required parameter
	invalidParams := map[string]interface{}{
		"width": 128,
		// Missing required positive_prompt
	}

	if err := tm.ValidateParameters(template, invalidParams); err == nil {
		t.Error("Expected error for missing required parameter")
	}

	// Test invalid parameter type
	invalidParams = map[string]interface{}{
		"positive_prompt": "test",
		"width":           "not_a_number", // Should be int
	}

	if err := tm.ValidateParameters(template, invalidParams); err == nil {
		t.Error("Expected error for invalid parameter type")
	}

	// Test parameter out of range
	invalidParams = map[string]interface{}{
		"positive_prompt": "test",
		"width":           1000, // Exceeds max value of 512
	}

	if err := tm.ValidateParameters(template, invalidParams); err == nil {
		t.Error("Expected error for parameter out of range")
	}

	// Test invalid enum value
	invalidParams = map[string]interface{}{
		"positive_prompt": "test",
		"sampler":         "invalid_sampler",
	}

	if err := tm.ValidateParameters(template, invalidParams); err == nil {
		t.Error("Expected error for invalid enum value")
	}

	// Test unknown parameter
	invalidParams = map[string]interface{}{
		"positive_prompt": "test",
		"unknown_param":   "value",
	}

	if err := tm.ValidateParameters(template, invalidParams); err == nil {
		t.Error("Expected error for unknown parameter")
	}
}

func TestSaveAndLoadTemplate(t *testing.T) {
	manager := NewTemplateManager()

	// Create temporary directory
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test_template.json")

	// Create and save template
	original := CreateBasicTemplate("test_save_load", "pixel_art")

	if err := manager.SaveTemplate(original, templatePath); err != nil {
		t.Fatalf("SaveTemplate failed: %v", err)
	}

	// Load template
	loaded, err := manager.LoadTemplate(templatePath)
	if err != nil {
		t.Fatalf("LoadTemplate failed: %v", err)
	}

	// Verify loaded template
	if loaded.ID != original.ID {
		t.Errorf("Expected ID %s, got %s", original.ID, loaded.ID)
	}

	if loaded.Name != original.Name {
		t.Errorf("Expected name %s, got %s", original.Name, loaded.Name)
	}

	if len(loaded.Parameters) != len(original.Parameters) {
		t.Errorf("Expected %d parameters, got %d", len(original.Parameters), len(loaded.Parameters))
	}

	if len(loaded.PromptSlots) != len(original.PromptSlots) {
		t.Errorf("Expected %d prompt slots, got %d", len(original.PromptSlots), len(loaded.PromptSlots))
	}
}

func TestListTemplates(t *testing.T) {
	manager := NewTemplateManager()

	// Create temporary directory with templates
	tmpDir := t.TempDir()

	// Create test templates
	template1 := CreateBasicTemplate("template1", "anime")
	template2 := CreateBasicTemplate("template2", "pixel_art")

	template1Path := filepath.Join(tmpDir, "template1.json")
	template2Path := filepath.Join(tmpDir, "template2.json")

	if err := manager.SaveTemplate(template1, template1Path); err != nil {
		t.Fatalf("Save template1 failed: %v", err)
	}

	if err := manager.SaveTemplate(template2, template2Path); err != nil {
		t.Fatalf("Save template2 failed: %v", err)
	}

	// Create non-JSON file to test filtering
	nonJSONPath := filepath.Join(tmpDir, "not_a_template.txt")
	if err := os.WriteFile(nonJSONPath, []byte("not json"), 0o644); err != nil {
		t.Fatalf("Create non-JSON file failed: %v", err)
	}

	// List templates
	templates, err := manager.ListTemplates(tmpDir)
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	if len(templates) != 2 {
		t.Errorf("Expected 2 templates, got %d", len(templates))
	}

	// Verify template IDs
	foundIDs := make(map[string]bool)
	for _, tmpl := range templates {
		foundIDs[tmpl.ID] = true
	}

	if !foundIDs["template1"] {
		t.Error("Expected to find template1")
	}

	if !foundIDs["template2"] {
		t.Error("Expected to find template2")
	}
}

func TestInstantiateTemplate(t *testing.T) {
	manager := NewTemplateManager()
	template := CreateBasicTemplate("test_instantiate", "anime")

	params := map[string]interface{}{
		"positive_prompt": "a cute anime character, happy expression",
		"width":           256,
		"height":          256,
		"steps":           25,
		"cfg_scale":       8.0,
		"sampler":         "euler",
	}

	workflow, err := manager.InstantiateTemplate(template, params)
	if err != nil {
		t.Fatalf("InstantiateTemplate failed: %v", err)
	}

	if workflow == nil {
		t.Fatal("Expected instantiated workflow")
	}

	if workflow.Nodes == nil {
		t.Fatal("Expected workflow to have nodes")
	}

	// Verify that the workflow is separate from the template
	if workflow == template.BaseWorkflow {
		t.Error("Expected instantiated workflow to be a copy, not the original")
	}
}

func TestParameterValidation(t *testing.T) {
	manager := NewTemplateManager()
	tm := manager.(*templateManager)

	tests := []struct {
		name      string
		parameter TemplateParameter
		value     interface{}
		wantErr   bool
	}{
		{
			name:      "valid string",
			parameter: TemplateParameter{Type: "string"},
			value:     "test string",
			wantErr:   false,
		},
		{
			name:      "invalid string type",
			parameter: TemplateParameter{Type: "string"},
			value:     123,
			wantErr:   true,
		},
		{
			name:      "valid int",
			parameter: TemplateParameter{Type: "int"},
			value:     42,
			wantErr:   false,
		},
		{
			name:      "valid int from float",
			parameter: TemplateParameter{Type: "int"},
			value:     42.0,
			wantErr:   false,
		},
		{
			name:      "invalid int type",
			parameter: TemplateParameter{Type: "int"},
			value:     "not a number",
			wantErr:   true,
		},
		{
			name:      "int below minimum",
			parameter: TemplateParameter{Type: "int", MinValue: floatPtr(10)},
			value:     5,
			wantErr:   true,
		},
		{
			name:      "int above maximum",
			parameter: TemplateParameter{Type: "int", MaxValue: floatPtr(100)},
			value:     200,
			wantErr:   true,
		},
		{
			name:      "valid float",
			parameter: TemplateParameter{Type: "float"},
			value:     3.14,
			wantErr:   false,
		},
		{
			name:      "valid float from int",
			parameter: TemplateParameter{Type: "float"},
			value:     42,
			wantErr:   false,
		},
		{
			name:      "invalid float type",
			parameter: TemplateParameter{Type: "float"},
			value:     "not a number",
			wantErr:   true,
		},
		{
			name:      "valid bool true",
			parameter: TemplateParameter{Type: "bool"},
			value:     true,
			wantErr:   false,
		},
		{
			name:      "valid bool false",
			parameter: TemplateParameter{Type: "bool"},
			value:     false,
			wantErr:   false,
		},
		{
			name:      "invalid bool type",
			parameter: TemplateParameter{Type: "bool"},
			value:     "not a bool",
			wantErr:   true,
		},
		{
			name:      "valid enum",
			parameter: TemplateParameter{Type: "enum", EnumValues: []string{"a", "b", "c"}},
			value:     "b",
			wantErr:   false,
		},
		{
			name:      "invalid enum value",
			parameter: TemplateParameter{Type: "enum", EnumValues: []string{"a", "b", "c"}},
			value:     "d",
			wantErr:   true,
		},
		{
			name:      "invalid enum type",
			parameter: TemplateParameter{Type: "enum", EnumValues: []string{"a", "b", "c"}},
			value:     123,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm.validateParameterValue(&tt.parameter, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateParameterValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyDefaultValues(t *testing.T) {
	manager := NewTemplateManager()
	tm := manager.(*templateManager)

	template := CreateBasicTemplate("test", "anime")

	// Test with no provided parameters (should get all defaults)
	params := map[string]interface{}{}
	result := tm.applyDefaultValues(template, params)

	// Should have default values for parameters with defaults
	if result["negative_prompt"] != "blurry, low quality, distorted" {
		t.Error("Expected default negative_prompt")
	}

	if result["width"] != 128 {
		t.Error("Expected default width")
	}

	// Should not have value for required parameter without default
	if _, exists := result["positive_prompt"]; exists {
		t.Error("Should not have value for required parameter without default")
	}

	// Test with provided parameters (should override defaults)
	params = map[string]interface{}{
		"positive_prompt": "custom prompt",
		"width":           256, // Override default
	}
	result = tm.applyDefaultValues(template, params)

	if result["positive_prompt"] != "custom prompt" {
		t.Error("Expected provided positive_prompt")
	}

	if result["width"] != 256 {
		t.Error("Expected overridden width")
	}

	if result["height"] != 128 {
		t.Error("Expected default height")
	}
}

func TestExecutePromptTemplate(t *testing.T) {
	manager := NewTemplateManager()
	tm := manager.(*templateManager)

	// Test simple template
	templateStr := "{{.character}}, {{.emotion}}"
	params := map[string]interface{}{
		"character": "anime girl",
		"emotion":   "happy",
	}

	result, err := tm.executePromptTemplate(templateStr, params)
	if err != nil {
		t.Fatalf("executePromptTemplate failed: %v", err)
	}

	expected := "anime girl, happy"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test conditional template
	templateStr = "{{.character}}{{if .style}}, {{.style}} style{{end}}"
	params = map[string]interface{}{
		"character": "anime girl",
		"style":     "pixel art",
	}

	result, err = tm.executePromptTemplate(templateStr, params)
	if err != nil {
		t.Fatalf("executePromptTemplate with conditional failed: %v", err)
	}

	expected = "anime girl, pixel art style"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test empty template
	result, err = tm.executePromptTemplate("", params)
	if err != nil {
		t.Fatalf("executePromptTemplate with empty template failed: %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty result for empty template, got %q", result)
	}
}
