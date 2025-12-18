package template

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"prompter-cli/internal/interfaces"
)

func TestProcessor_LoadTemplate(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir := t.TempDir()
	
	// Create pre and post directories
	preDir := filepath.Join(tempDir, "pre")
	postDir := filepath.Join(tempDir, "post")
	
	if err := os.MkdirAll(preDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(postDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create test template files
	testTemplate := "Hello {{.Prompt}}!"
	
	preTemplatePath := filepath.Join(preDir, "test-template.md")
	if err := os.WriteFile(preTemplatePath, []byte(testTemplate), 0644); err != nil {
		t.Fatal(err)
	}
	
	postTemplatePath := filepath.Join(postDir, "Another-Template.md")
	if err := os.WriteFile(postTemplatePath, []byte(testTemplate), 0644); err != nil {
		t.Fatal(err)
	}
	
	processor := NewProcessor(tempDir)
	
	tests := []struct {
		name        string
		templateName string
		wantError   bool
	}{
		{
			name:        "load template by exact name",
			templateName: "test-template",
			wantError:   false,
		},
		{
			name:        "load template case insensitive",
			templateName: "TEST-TEMPLATE",
			wantError:   false,
		},
		{
			name:        "load template from post directory",
			templateName: "another-template",
			wantError:   false,
		},
		{
			name:        "load template case insensitive from post",
			templateName: "ANOTHER-TEMPLATE",
			wantError:   false,
		},
		{
			name:        "template not found",
			templateName: "nonexistent",
			wantError:   true,
		},
		{
			name:        "load by absolute path",
			templateName: preTemplatePath,
			wantError:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := processor.LoadTemplate(tt.templateName)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if tmpl == nil {
				t.Errorf("expected template but got nil")
			}
		})
	}
}

func TestProcessor_LoadTemplate_WithDefaultTemplates(t *testing.T) {
	// Create a temporary directory for test templates
	tempDir := t.TempDir()
	
	// Create pre directory
	preDir := filepath.Join(tempDir, "pre")
	if err := os.MkdirAll(preDir, 0755); err != nil {
		t.Fatal(err)
	}
	
	// Create test template files including default templates
	testTemplate := "Hello {{.Prompt}}!"
	
	// Regular template
	regularPath := filepath.Join(preDir, "regular.md")
	if err := os.WriteFile(regularPath, []byte(testTemplate), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Default template
	defaultPath := filepath.Join(preDir, "example.default.md")
	if err := os.WriteFile(defaultPath, []byte(testTemplate), 0644); err != nil {
		t.Fatal(err)
	}
	
	processor := NewProcessor(tempDir)
	
	tests := []struct {
		name        string
		templateName string
		wantError   bool
		description string
	}{
		{
			name:        "load regular template",
			templateName: "regular",
			wantError:   false,
			description: "should load regular template by name",
		},
		{
			name:        "load default template by display name",
			templateName: "example",
			wantError:   false,
			description: "should load example.default.md when searching for 'example'",
		},
		{
			name:        "load default template by full stem",
			templateName: "example.default",
			wantError:   false,
			description: "should also load by full stem name",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := processor.LoadTemplate(tt.templateName)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if tmpl == nil {
				t.Errorf("expected template but got nil")
			}
		})
	}
}

func TestProcessor_LoadTemplate_RealStrictTemplate(t *testing.T) {
	// Test with the real strict.default.md file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory")
		return
	}
	
	promptsPath := filepath.Join(homeDir, "prompts")
	if _, err := os.Stat(promptsPath); os.IsNotExist(err) {
		t.Skip("Real prompts directory doesn't exist")
		return
	}
	
	// Check if strict.default.md exists
	strictPath := filepath.Join(promptsPath, "post", "strict.default.md")
	if _, err := os.Stat(strictPath); os.IsNotExist(err) {
		t.Skip("strict.default.md doesn't exist")
		return
	}
	
	processor := NewProcessor(promptsPath)
	
	// Try to load the template by display name "strict"
	tmpl, err := processor.LoadTemplate("strict")
	if err != nil {
		t.Errorf("Failed to load strict template: %v", err)
		return
	}
	
	if tmpl == nil {
		t.Error("Expected template but got nil")
	}
	
	t.Logf("✓ Successfully loaded 'strict' template from strict.default.md")
}

func TestProcessor_Execute(t *testing.T) {
	processor := NewProcessor("")
	
	// Create a simple template
	templateContent := "Hello {{.Prompt}}! Current time: {{.Now.Format \"2006-01-02\"}}"
	tmpl := processor.createTestTemplate(t, templateContent)
	
	// Test data
	testTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)
	data := interfaces.TemplateData{
		Prompt: "World",
		Now:    testTime,
		CWD:    "/test/dir",
		Files:  []interfaces.FileInfo{},
		Git:    interfaces.GitInfo{},
		Config: make(map[string]interface{}),
		Env:    make(map[string]string),
		Fix:    interfaces.FixInfo{},
	}
	
	result, err := processor.Execute(tmpl, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	expected := "Hello World! Current time: 2023-12-25"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCustomHelperFunctions(t *testing.T) {
	processor := NewProcessor("")
	
	tests := []struct {
		name     string
		template string
		data     interfaces.TemplateData
		expected string
	}{
		{
			name:     "truncate function",
			template: `{{truncate 10 "This is a very long string"}}`,
			data:     interfaces.TemplateData{},
			expected: "This is...",
		},
		{
			name:     "mdFence function with language",
			template: `{{mdFence "go" "fmt.Println(\"hello\")"}}`,
			data:     interfaces.TemplateData{},
			expected: "```go\nfmt.Println(\"hello\")\n```",
		},
		{
			name:     "mdFence function without language",
			template: `{{mdFence "" "some code"}}`,
			data:     interfaces.TemplateData{},
			expected: "```\nsome code\n```",
		},
		{
			name:     "indent function",
			template: `{{indent 4 "line1\nline2\n\nline4"}}`,
			data:     interfaces.TemplateData{},
			expected: "    line1\n    line2\n\n    line4",
		},
		{
			name:     "dedent function",
			template: `{{dedent "    line1\n    line2\n        line3"}}`,
			data:     interfaces.TemplateData{},
			expected: "line1\nline2\n    line3",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := processor.createTestTemplate(t, tt.template)
			
			result, err := processor.Execute(tmpl, tt.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Helper method to create test templates
func (p *Processor) createTestTemplate(t *testing.T, content string) *template.Template {
	tmpl := template.New("test")
	
	if err := p.registerHelpersToTemplate(tmpl); err != nil {
		t.Fatalf("failed to register helpers: %v", err)
	}
	
	tmpl, err := tmpl.Parse(content)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}
	
	return tmpl
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{
			name: "truncateFunc short text",
			function: func() string {
				return truncateFunc(10, "short")
			},
			expected: "short",
		},
		{
			name: "truncateFunc long text",
			function: func() string {
				return truncateFunc(10, "this is a very long text")
			},
			expected: "this is...",
		},
		{
			name: "mdFenceFunc with language",
			function: func() string {
				return mdFenceFunc("python", "print('hello')")
			},
			expected: "```python\nprint('hello')\n```",
		},
		{
			name: "indentFunc",
			function: func() string {
				return indentFunc(2, "line1\nline2")
			},
			expected: "  line1\n  line2",
		},
		{
			name: "dedentFunc",
			function: func() string {
				return dedentFunc("  line1\n  line2\n    line3")
			},
			expected: "line1\nline2\n  line3",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}