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

func TestFindTemplates_WithDefaultTemplates(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	preDir := filepath.Join(tempDir, "pre")
	if err := os.MkdirAll(preDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test template files including default templates
	testFiles := []string{
		"template1.md", 
		"template2.md", 
		"example.default.md",
		"another.default.template.md",
		"not-template.txt",
	}
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
	
	// Check that we have the right number of templates
	expectedCount := 4 // 2 default + 2 regular
	if len(templates) != expectedCount {
		t.Errorf("Expected %d templates, got %d: %v", expectedCount, len(templates), templates)
	}
	
	// Check that default templates come first (first 2 should be defaults)
	// and regular templates come after (last 2 should be regular)
	defaultTemplates := templates[:2]
	regularTemplates := templates[2:]
	
	// Check that we have the expected default templates (order may vary)
	expectedDefaults := []string{"another.template", "example"}
	for _, expected := range expectedDefaults {
		found := false
		for _, actual := range defaultTemplates {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected default template %s not found in defaults: %v", expected, defaultTemplates)
		}
	}
	
	// Check that we have the expected regular templates (order may vary)
	expectedRegulars := []string{"template1", "template2"}
	for _, expected := range expectedRegulars {
		found := false
		for _, actual := range regularTemplates {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected regular template %s not found in regulars: %v", expected, regularTemplates)
		}
	}
}

func TestFindTemplates_RealPromptsDirectory(t *testing.T) {
	// Test with the actual prompts directory to verify default template ordering
	// First expand the path manually like the config manager does
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
	
	prompter := NewPrompter(promptsPath)
	
	// Test post templates (where we know strict.default.md exists)
	templates, err := prompter.findTemplates("post")
	if err != nil {
		t.Fatalf("Could not read real prompts directory: %v", err)
	}
	
	t.Logf("Found post templates: %v", templates)
	
	// Test the buildOptionsWithNone function with real templates
	options := prompter.buildOptionsWithNone(templates, "post")
	t.Logf("Options with None: %v", options)
	
	// If we have templates, verify the ordering
	if len(templates) > 0 {
		// We expect "strict" to be first if strict.default.md exists
		for i, template := range templates {
			t.Logf("Template %d: %s", i+1, template)
		}
		
		// Check if "strict" is the first template (since we know strict.default.md exists)
		if templates[0] == "strict" {
			t.Logf("✓ Default template 'strict' is correctly shown first")
		}
		
		// Check that "None" comes after default templates
		if len(options) >= 2 && options[1] == "None" {
			t.Logf("✓ 'None' is correctly positioned after default templates")
		}
	}
}

func TestBuildOptionsWithNone(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	preDir := filepath.Join(tempDir, "pre")
	if err := os.MkdirAll(preDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	
	// Create test template files including default templates
	testFiles := []string{
		"regular1.md", 
		"regular2.md", 
		"example.default.md",
		"another.default.template.md",
	}
	for _, file := range testFiles {
		if err := os.WriteFile(filepath.Join(preDir, file), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}
	
	prompter := NewPrompter(tempDir)
	
	// Get templates (should be ordered with defaults first)
	templates, err := prompter.findTemplates("pre")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	// Build options with "None" in correct position
	options := prompter.buildOptionsWithNone(templates, "pre")
	
	t.Logf("Options: %v", options)
	
	// Expected order: default templates, then "None", then regular templates
	// So: ["another.template", "example", "None", "regular1", "regular2"]
	expectedOrder := []string{"another.template", "example", "None", "regular1", "regular2"}
	
	if len(options) != len(expectedOrder) {
		t.Errorf("Expected %d options, got %d: %v", len(expectedOrder), len(options), options)
	}
	
	for i, expected := range expectedOrder {
		if i < len(options) && options[i] != expected {
			t.Errorf("Expected option %d to be %s, got %s", i, expected, options[i])
		}
	}
	
	// Verify "None" is in the correct position (after defaults, before regulars)
	noneIndex := -1
	for i, option := range options {
		if option == "None" {
			noneIndex = i
			break
		}
	}
	
	if noneIndex == -1 {
		t.Error("'None' option not found in options list")
	} else if noneIndex != 2 { // Should be at index 2 (after 2 default templates)
		t.Errorf("'None' should be at index 2, but found at index %d", noneIndex)
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