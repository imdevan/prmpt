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

	// Collect pre-template if not specified
	if request.PreTemplate == "" && !request.FixMode {
		if err := p.promptForPreTemplate(request); err != nil {
			return fmt.Errorf("failed to collect pre-template: %w", err)
		}
	}

	// Collect post-template if not specified
	if request.PostTemplate == "" && !request.FixMode {
		if err := p.promptForPostTemplate(request); err != nil {
			return fmt.Errorf("failed to collect post-template: %w", err)
		}
	}

	// Collect directory inclusion if not specified
	if request.Directory == "" && len(request.Files) == 0 && !request.FixMode {
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

	// Build options with proper ordering: defaults first, then "None", then regulars
	options := p.buildOptionsWithNone(templates, "pre")

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

	// Build options with proper ordering: defaults first, then "None", then regulars
	options := p.buildOptionsWithNone(templates, "post")

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
	// Skip confirmation prompt entirely
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

	var defaultTemplates []string
	var regularTemplates []string
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Remove .md extension for processing
			name := strings.TrimSuffix(entry.Name(), ".md")
			
			// Check if this is a default template
			if strings.Contains(name, ".default.") {
				// Strip the .default. part for display
				displayName := strings.ReplaceAll(name, ".default.", ".")
				// Remove any leading/trailing dots that might result
				displayName = strings.Trim(displayName, ".")
				defaultTemplates = append(defaultTemplates, displayName)
			} else if strings.HasSuffix(name, ".default") {
				// Handle case where .default is at the end
				displayName := strings.TrimSuffix(name, ".default")
				defaultTemplates = append(defaultTemplates, displayName)
			} else {
				regularTemplates = append(regularTemplates, name)
			}
		}
	}

	// Combine lists with defaults first
	var templates []string
	templates = append(templates, defaultTemplates...)
	templates = append(templates, regularTemplates...)

	return templates, nil
}



// buildOptionsWithNone constructs the options list with proper ordering:
// default templates first, then "None", then regular templates
func (p *Prompter) buildOptionsWithNone(templates []string, subdir string) []string {
	// We need to separate default templates from regular templates
	// to insert "None" in the right place
	templateDir := filepath.Join(p.promptsLocation, subdir)
	
	var defaultTemplates []string
	var regularTemplates []string
	
	// Check if directory exists
	if entries, err := os.ReadDir(templateDir); err == nil {
		// Build a map of which templates are defaults
		defaultNames := make(map[string]bool)
		
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				name := strings.TrimSuffix(entry.Name(), ".md")
				
				// Check if this is a default template
				if strings.Contains(name, ".default.") || strings.HasSuffix(name, ".default") {
					var displayName string
					if strings.Contains(name, ".default.") {
						displayName = strings.ReplaceAll(name, ".default.", ".")
						displayName = strings.Trim(displayName, ".")
					} else {
						displayName = strings.TrimSuffix(name, ".default")
					}
					defaultNames[displayName] = true
				}
			}
		}
		
		// Separate templates based on whether they're defaults
		for _, template := range templates {
			if defaultNames[template] {
				defaultTemplates = append(defaultTemplates, template)
			} else {
				regularTemplates = append(regularTemplates, template)
			}
		}
	} else {
		// Fallback: if we can't read the directory, treat all as regular
		regularTemplates = templates
	}
	
	// Build final options list: defaults first, then "None", then regulars
	var options []string
	options = append(options, defaultTemplates...)
	options = append(options, "None")
	options = append(options, regularTemplates...)
	
	return options
}

// truncateString truncates a string to the specified length with ellipsis
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}