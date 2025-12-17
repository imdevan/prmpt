package interactive

import (
	"os"
	"path/filepath"
	"testing"

	"prompter-cli/pkg/models"
)

func TestNewPrompter(t *testing.T) {
	promptsLocation := "/test/prompts"
	prompter := NewPrompter(promptsLocation)
	
	if prompter.promptsLocation != promptsLocation {
		t.Errorf("Expected prompts location %s, got %s", promptsLocation, prompter.promptsLocation)
	}
}

func TestCollectMissingInputs_NonInteractive(t *testing.T) {
	prompter := NewPrompter("/test/prompts")
	request := &models.PromptRequest{
		Interactive: false,
		BasePrompt:  "test prompt",
	}
	
	// Should not prompt in noninteractive mode
	err := prompter.CollectMissingInputs(request)
	if err != nil {
		t.Errorf("Expected no error in noninteractive mode, got: %v", err)
	}
}

func TestFindTemplates(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	preDir := filepath.Join(tempDir, "pre")
	if err := os.MkdirAll(preDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test template files
	testFiles := []string{"template1.md", "template2.md", "not-template.txt"}
	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(preDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	prompter := NewPrompter(tempDir)
	templates, err := prompter.findTemplates("pre")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	expected := []string{"template1", "template2"}
	if len(templates) != len(expected) {
		t.Errorf("Expected %d templates, got %d", len(expected), len(templates))
	}
	
	for i, template := range templates {
		if template != expected[i] {
			t.Errorf("Expected template %s, got %s", expected[i], template)
		}
	}
}

func TestFindTemplates_NonExistentDirectory(t *testing.T) {
	prompter := NewPrompter("/nonexistent")
	templates, err := prompter.findTemplates("pre")
	if err != nil {
		t.Errorf("Expected no error for nonexistent directory, got: %v", err)
	}
	
	if len(templates) != 0 {
		t.Errorf("Expected empty template list, got %v", templates)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"exactly10c", 10, "exactly10c"},
		{"", 5, ""},
	}
	
	for _, test := range tests {
		result := truncateString(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("truncateString(%q, %d) = %q, expected %q", 
				test.input, test.maxLen, result, test.expected)
		}
	}
}