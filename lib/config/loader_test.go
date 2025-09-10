package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Test data structures for testing JSON operations
type TestConfig struct {
	Name    string            `json:"name"`
	Value   int               `json:"value"`
	Options map[string]string `json:"options"`
}

type InvalidConfig struct {
	Channel chan int `json:"channel"` // Channels are not JSON serializable
}

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		want     string
	}{
		{
			name:     "AbsolutePath",
			basePath: "/tmp/config",
			want:     "/tmp/config",
		},
		{
			name:     "RelativePath",
			basePath: "config",
			want:     "config",
		},
		{
			name:     "EmptyPath",
			basePath: "",
			want:     "",
		},
		{
			name:     "PathWithSpaces",
			basePath: "/path with spaces/config",
			want:     "/path with spaces/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := New(tt.basePath)
			if loader == nil {
				t.Error("New() returned nil")
			}
			if loader.basePath != tt.want {
				t.Errorf("New() basePath = %v, want %v", loader.basePath, tt.want)
			}
		})
	}
}

func TestLoader_LoadJSON(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test config file
	testConfig := TestConfig{
		Name:  "test",
		Value: 42,
		Options: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	validConfigPath := filepath.Join(tempDir, "valid.json")
	validData, _ := json.MarshalIndent(testConfig, "", "  ")
	if err := os.WriteFile(validConfigPath, validData, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create invalid JSON file
	invalidConfigPath := filepath.Join(tempDir, "invalid.json")
	invalidData := []byte(`{"name": "test", "value":}`) // Missing value
	if err := os.WriteFile(invalidConfigPath, invalidData, 0644); err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	loader := New(tempDir)

	tests := []struct {
		name     string
		filename string
		target   interface{}
		wantErr  bool
		validate func(t *testing.T, target interface{})
	}{
		{
			name:     "ValidJSONFile",
			filename: "valid.json",
			target:   &TestConfig{},
			wantErr:  false,
			validate: func(t *testing.T, target interface{}) {
				config := target.(*TestConfig)
				if config.Name != "test" {
					t.Errorf("Expected name 'test', got '%s'", config.Name)
				}
				if config.Value != 42 {
					t.Errorf("Expected value 42, got %d", config.Value)
				}
				if len(config.Options) != 2 {
					t.Errorf("Expected 2 options, got %d", len(config.Options))
				}
			},
		},
		{
			name:     "NonExistentFile",
			filename: "nonexistent.json",
			target:   &TestConfig{},
			wantErr:  true,
			validate: nil,
		},
		{
			name:     "InvalidJSONSyntax",
			filename: "invalid.json",
			target:   &TestConfig{},
			wantErr:  true,
			validate: nil,
		},
		{
			name:     "NilTarget",
			filename: "valid.json",
			target:   nil,
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.LoadJSON(tt.filename, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, tt.target)
			}
		})
	}
}

func TestLoader_SaveJSON(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config_save_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	loader := New(tempDir)

	testConfig := TestConfig{
		Name:  "saved_test",
		Value: 123,
		Options: map[string]string{
			"option1": "value1",
		},
	}

	tests := []struct {
		name     string
		filename string
		data     interface{}
		wantErr  bool
		validate func(t *testing.T, filename string)
	}{
		{
			name:     "ValidData",
			filename: "saved.json",
			data:     testConfig,
			wantErr:  false,
			validate: func(t *testing.T, filename string) {
				// Verify file was created and contains correct data
				fullPath := filepath.Join(tempDir, filename)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Error("File was not created")
					return
				}

				// Load and verify content
				var loaded TestConfig
				err := loader.LoadJSON(filename, &loaded)
				if err != nil {
					t.Errorf("Failed to load saved file: %v", err)
					return
				}

				if loaded.Name != testConfig.Name {
					t.Errorf("Expected name '%s', got '%s'", testConfig.Name, loaded.Name)
				}
				if loaded.Value != testConfig.Value {
					t.Errorf("Expected value %d, got %d", testConfig.Value, loaded.Value)
				}
			},
		},
		{
			name:     "SubdirectoryCreation",
			filename: "subdir/nested.json",
			data:     testConfig,
			wantErr:  false,
			validate: func(t *testing.T, filename string) {
				fullPath := filepath.Join(tempDir, filename)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Error("Nested file was not created")
				}
				// Verify directory was created
				dirPath := filepath.Join(tempDir, "subdir")
				if _, err := os.Stat(dirPath); os.IsNotExist(err) {
					t.Error("Subdirectory was not created")
				}
			},
		},
		{
			name:     "InvalidData",
			filename: "invalid.json",
			data:     InvalidConfig{Channel: make(chan int)},
			wantErr:  true,
			validate: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.SaveJSON(tt.filename, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil && err == nil {
				tt.validate(t, tt.filename)
			}
		})
	}
}

func TestLoader_FileExists(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config_exists_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	loader := New(tempDir)

	// Create an existing file
	existingFile := "existing.json"
	existingPath := filepath.Join(tempDir, existingFile)
	if err := os.WriteFile(existingPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "ExistingFile",
			filename: existingFile,
			want:     true,
		},
		{
			name:     "NonExistentFile",
			filename: "nonexistent.json",
			want:     false,
		},
		{
			name:     "EmptyFilename",
			filename: "",
			want:     true, // Empty filename resolves to basePath directory which exists
		},
		{
			name:     "SubdirectoryFile",
			filename: "subdir/file.json",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loader.FileExists(tt.filename); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoader_GetFullPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		filename string
		want     string
	}{
		{
			name:     "SimpleFile",
			basePath: "/config",
			filename: "test.json",
			want:     "/config/test.json",
		},
		{
			name:     "NestedFile",
			basePath: "/config",
			filename: "subdir/test.json",
			want:     "/config/subdir/test.json",
		},
		{
			name:     "EmptyBasePath",
			basePath: "",
			filename: "test.json",
			want:     "test.json",
		},
		{
			name:     "EmptyFilename",
			basePath: "/config",
			filename: "",
			want:     "/config",
		},
		{
			name:     "WindowsStylePath",
			basePath: "C:\\config",
			filename: "test.json",
			want:     filepath.Join("C:\\config", "test.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := New(tt.basePath)
			if got := loader.GetFullPath(tt.filename); got != tt.want {
				t.Errorf("GetFullPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoader_ListFiles(t *testing.T) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "config_list_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	loader := New(tempDir)

	// Create test files with different extensions
	testFiles := []string{
		"config1.json",
		"config2.json",
		"settings.txt",
		"data.xml",
		"empty.json",
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tempDir, filename)
		content := "{}"
		if filepath.Ext(filename) == ".txt" {
			content = "text content"
		} else if filepath.Ext(filename) == ".xml" {
			content = "<xml></xml>"
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	tests := []struct {
		name      string
		extension string
		wantCount int
		wantFiles []string
		wantErr   bool
	}{
		{
			name:      "JSONFiles",
			extension: "json",
			wantCount: 3,
			wantFiles: []string{"config1.json", "config2.json", "empty.json"},
			wantErr:   false,
		},
		{
			name:      "TXTFiles",
			extension: "txt",
			wantCount: 1,
			wantFiles: []string{"settings.txt"},
			wantErr:   false,
		},
		{
			name:      "XMLFiles",
			extension: "xml",
			wantCount: 1,
			wantFiles: []string{"data.xml"},
			wantErr:   false,
		},
		{
			name:      "NonExistentExtension",
			extension: "pdf",
			wantCount: 0,
			wantFiles: []string{},
			wantErr:   false,
		},
		{
			name:      "EmptyExtension",
			extension: "",
			wantCount: 0,
			wantFiles: []string{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loader.ListFiles(tt.extension)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantCount {
				t.Errorf("ListFiles() returned %d files, want %d", len(got), tt.wantCount)
			}

			// Verify expected files are present (order independent)
			for _, expectedFile := range tt.wantFiles {
				found := false
				for _, actualFile := range got {
					if actualFile == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file '%s' not found in results: %v", expectedFile, got)
				}
			}
		})
	}
}

// Table-driven test for complete workflow: Save -> Load -> Verify
func TestLoader_CompleteWorkflow(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_workflow_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	loader := New(tempDir)

	workflowTests := []struct {
		name     string
		filename string
		saveData TestConfig
		loadData TestConfig
	}{
		{
			name:     "BasicWorkflow",
			filename: "workflow1.json",
			saveData: TestConfig{
				Name:  "workflow_test",
				Value: 999,
				Options: map[string]string{
					"env": "test",
				},
			},
		},
		{
			name:     "ComplexDataWorkflow",
			filename: "workflow2.json",
			saveData: TestConfig{
				Name:  "complex_workflow",
				Value: -42,
				Options: map[string]string{
					"key1":   "value with spaces",
					"key2":   "unicode: 测试",
					"empty":  "",
					"number": "123",
				},
			},
		},
	}

	for _, tt := range workflowTests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Save data
			err := loader.SaveJSON(tt.filename, tt.saveData)
			if err != nil {
				t.Fatalf("Failed to save JSON: %v", err)
			}

			// Step 2: Verify file exists
			if !loader.FileExists(tt.filename) {
				t.Error("File should exist after saving")
			}

			// Step 3: Load data
			err = loader.LoadJSON(tt.filename, &tt.loadData)
			if err != nil {
				t.Fatalf("Failed to load JSON: %v", err)
			}

			// Step 4: Verify data integrity
			if tt.loadData.Name != tt.saveData.Name {
				t.Errorf("Name mismatch: saved '%s', loaded '%s'", tt.saveData.Name, tt.loadData.Name)
			}
			if tt.loadData.Value != tt.saveData.Value {
				t.Errorf("Value mismatch: saved %d, loaded %d", tt.saveData.Value, tt.loadData.Value)
			}
			if len(tt.loadData.Options) != len(tt.saveData.Options) {
				t.Errorf("Options count mismatch: saved %d, loaded %d", len(tt.saveData.Options), len(tt.loadData.Options))
			}
			for key, savedValue := range tt.saveData.Options {
				if loadedValue, exists := tt.loadData.Options[key]; !exists || loadedValue != savedValue {
					t.Errorf("Option '%s' mismatch: saved '%s', loaded '%s'", key, savedValue, loadedValue)
				}
			}

			// Step 5: Verify full path
			expectedPath := filepath.Join(tempDir, tt.filename)
			actualPath := loader.GetFullPath(tt.filename)
			if actualPath != expectedPath {
				t.Errorf("GetFullPath() = '%s', want '%s'", actualPath, expectedPath)
			}
		})
	}
}

// Test error conditions and edge cases
func TestLoader_ErrorConditions(t *testing.T) {
	// Create temporary directory with restricted permissions
	tempDir, err := os.MkdirTemp("", "config_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	loader := New(tempDir)

	t.Run("SaveToReadOnlyDirectory", func(t *testing.T) {
		// Create a subdirectory and make it read-only
		readOnlyDir := filepath.Join(tempDir, "readonly")
		if err := os.Mkdir(readOnlyDir, 0755); err != nil {
			t.Fatalf("Failed to create readonly dir: %v", err)
		}
		if err := os.Chmod(readOnlyDir, 0444); err != nil {
			t.Fatalf("Failed to make dir readonly: %v", err)
		}

		// Restore permissions for cleanup
		defer os.Chmod(readOnlyDir, 0755)

		restrictedLoader := New(readOnlyDir)
		testData := TestConfig{Name: "test", Value: 1}

		err := restrictedLoader.SaveJSON("test.json", testData)
		if err == nil {
			t.Error("Expected error when saving to read-only directory")
		}
	})

	t.Run("LoadFromEmptyFile", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.json")
		if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		var config TestConfig
		err := loader.LoadJSON("empty.json", &config)
		if err == nil {
			t.Error("Expected error when loading empty JSON file")
		}
	})
}
