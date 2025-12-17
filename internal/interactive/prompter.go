package interactive

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"prompter-cli/pkg/models"
)

// Prompter handles interactive user input collection
type Prompter struct {
	promptsLocation string
}

// NewPrompter creates a new interactive prompter
func NewPrompter(promptsLocation string) *Prompter {
	return &Prompter{
		promptsLocation: promptsLocation,
	}
}

// CollectMissingInputs prompts the user for any missing required inputs
func (p *Prompter) CollectMissingInputs(request *models.PromptRequest) error {
	if !request.Interactive {
		return nil // Skip interactive prompts in noninteractive mode
	}

	// Collect base prompt if missing and not in fix mode
	if request.BasePrompt == "" && !request.FixMode {
		if err := p.promptForBasePrompt(request); err != nil {
			return fmt.Errorf("failed to collect base prompt: %w", err)
		}
	}

	// Collect pre template if not specified
	if request.PreTemplate == "" {
		if err := p.promptForPreTemplate(request); err != nil {
			return fmt.Errorf("failed to collect pre template: %w", err)
		}
	}

	// Collect post template if not specified
	if request.PostTemplate == "" {
		if err := p.promptForPostTemplate(request); err != nil {
			return fmt.Errorf("failed to collect post template: %w", err)
		}
	}

	// Collect directory inclusion if not specified
	if request.Directory == "" && len(request.Files) == 0 {
		if err := p.promptForDirectoryInclusion(request); err != nil {
			return fmt.Errorf("failed to collect directory inclusion: %w", err)
		}
	}

	// Show confirmation summary
	if err := p.showConfirmationSummary(request); err != nil {
		return fmt.Errorf("user cancelled operation: %w", err)
	}

	return nil
}

// promptForBasePrompt asks the user to enter a base prompt
func (p *Prompter) promptForBasePrompt(request *models.PromptRequest) error {
	prompt := &survey.Input{
		Message: "Enter your base prompt:",
		Help:    "This is the main prompt text that will be sent to the AI",
	}

	var basePrompt string
	if err := survey.AskOne(prompt, &basePrompt, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	request.BasePrompt = strings.TrimSpace(basePrompt)
	return nil
}

// promptForPreTemplate asks the user to select a pre-template
func (p *Prompter) promptForPreTemplate(request *models.PromptRequest) error {
	templates, err := p.findTemplates("pre")
	if err != nil {
		return fmt.Errorf("failed to find pre templates: %w", err)
	}

	// Add "None" option
	options := append([]string{"None"}, templates...)

	prompt := &survey.Select{
		Message: "Select a pre-template (prepended to prompt):",
		Options: options,
		Help:    "Pre-templates are added before your base prompt",
	}

	var selected string
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}

	if selected != "None" {
		request.PreTemplate = selected
	}

	return nil
}

// promptForPostTemplate asks the user to select a post-template
func (p *Prompter) promptForPostTemplate(request *models.PromptRequest) error {
	templates, err := p.findTemplates("post")
	if err != nil {
		return fmt.Errorf("failed to find post templates: %w", err)
	}

	// Add "None" option
	options := append([]string{"None"}, templates...)

	prompt := &survey.Select{
		Message: "Select a post-template (appended to prompt):",
		Options: options,
		Help:    "Post-templates are added after your base prompt",
	}

	var selected string
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}

	if selected != "None" {
		request.PostTemplate = selected
	}

	return nil
}

// promptForDirectoryInclusion asks whether to include directory context
func (p *Prompter) promptForDirectoryInclusion(request *models.PromptRequest) error {
	prompt := &survey.Confirm{
		Message: "Include current directory context in the prompt?",
		Help:    "This will include relevant files from the current directory",
		Default: false,
	}

	var includeDirectory bool
	if err := survey.AskOne(prompt, &includeDirectory); err != nil {
		return err
	}

	if includeDirectory {
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		request.Directory = cwd
	}

	return nil
}

// showConfirmationSummary displays a summary and asks for confirmation
func (p *Prompter) showConfirmationSummary(request *models.PromptRequest) error {
	fmt.Println("\n=== Prompt Generation Summary ===")
	
	if request.FixMode {
		fmt.Println("Mode: Fix mode (processing captured command output)")
	} else {
		fmt.Printf("Base prompt: %s\n", truncateString(request.BasePrompt, 60))
	}
	
	if request.PreTemplate != "" {
		fmt.Printf("Pre-template: %s\n", request.PreTemplate)
	}
	
	if request.PostTemplate != "" {
		fmt.Printf("Post-template: %s\n", request.PostTemplate)
	}
	
	if len(request.Files) > 0 {
		fmt.Printf("Files: %s\n", strings.Join(request.Files, ", "))
	}
	
	if request.Directory != "" {
		fmt.Printf("Directory: %s\n", request.Directory)
	}
	
	if request.Target != "" {
		fmt.Printf("Output target: %s\n", request.Target)
	}
	
	if request.Editor != "" {
		fmt.Printf("Editor: %s\n", request.Editor)
	}

	prompt := &survey.Confirm{
		Message: "Generate prompt with these settings?",
		Default: true,
	}

	var confirmed bool
	if err := survey.AskOne(prompt, &confirmed); err != nil {
		return err
	}

	if !confirmed {
		return fmt.Errorf("operation cancelled by user")
	}

	return nil
}

// findTemplates discovers available templates in the specified subdirectory
func (p *Prompter) findTemplates(subdir string) ([]string, error) {
	templateDir := filepath.Join(p.promptsLocation, subdir)
	
	// Check if directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		return []string{}, nil // Return empty list if directory doesn't exist
	}

	entries, err := os.ReadDir(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory %s: %w", templateDir, err)
	}

	var templates []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Remove .md extension for display
			name := strings.TrimSuffix(entry.Name(), ".md")
			templates = append(templates, name)
		}
	}

	return templates, nil
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}