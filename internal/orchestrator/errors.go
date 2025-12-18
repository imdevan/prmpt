package orchestrator

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Error types for different categories of failures
var (
	ErrConfigurationInvalid = errors.New("configuration error")
	ErrTemplateNotFound     = errors.New("template not found")
	ErrTemplateInvalid      = errors.New("template error")
	ErrContentCollection    = errors.New("content collection error")
	ErrFixModeInvalid       = errors.New("fix mode error")
	ErrOutputFailed         = errors.New("output error")
	ErrValidationFailed     = errors.New("validation error")
)

// PrompterError represents a structured error with actionable guidance
type PrompterError struct {
	Type     error
	Message  string
	Guidance string
	Cause    error
}

func (e *PrompterError) Error() string {
	if e.Guidance != "" {
		return fmt.Sprintf("%s: %s\n\n%s", e.Type, e.Message, e.Guidance)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *PrompterError) Unwrap() error {
	return e.Cause
}

// Error constructors with actionable guidance

func NewConfigurationError(message string, cause error) *PrompterError {
	guidance := "Run 'prompter --help' for usage information and configuration options."
	
	if strings.Contains(message, "permission") {
		guidance = "Check file permissions for your configuration directory. " +
			"Run 'prompter --help' for more information."
	} else if strings.Contains(message, "not found") || strings.Contains(message, "does not exist") {
		guidance = "Configuration file not found. Run 'prompter --help' to see default locations and options."
	}
	
	return &PrompterError{
		Type:     ErrConfigurationInvalid,
		Message:  message,
		Guidance: guidance,
		Cause:    cause,
	}
}

func NewTemplateError(templateName string, cause error) *PrompterError {
	message := fmt.Sprintf("failed to process template '%s'", templateName)
	guidance := "Run 'prompter --help' for template usage and configuration."
	
	if strings.Contains(cause.Error(), "not found") {
		guidance = fmt.Sprintf("Template '%s' not found. Run 'prompter --help' for template setup.", templateName)
	} else if strings.Contains(cause.Error(), "parse") || strings.Contains(cause.Error(), "syntax") {
		guidance = fmt.Sprintf("Template '%s' has syntax errors. Run 'prompter --help' for template format.", templateName)
	}
	
	return &PrompterError{
		Type:     ErrTemplateInvalid,
		Message:  message,
		Guidance: guidance,
		Cause:    cause,
	}
}

func NewContentCollectionError(path string, cause error) *PrompterError {
	message := fmt.Sprintf("failed to collect content from '%s'", path)
	guidance := "Run 'prompter --help' for file and directory usage options."
	
	if strings.Contains(cause.Error(), "permission") {
		guidance = fmt.Sprintf("Permission denied accessing '%s'. Run 'prompter --help' for usage.", path)
	} else if strings.Contains(cause.Error(), "not found") || strings.Contains(cause.Error(), "does not exist") {
		guidance = fmt.Sprintf("Path '%s' not found. Run 'prompter --help' for usage.", path)
	}
	
	return &PrompterError{
		Type:     ErrContentCollection,
		Message:  message,
		Guidance: guidance,
		Cause:    cause,
	}
}

func NewFixModeError(fixFile string, cause error) *PrompterError {
	message := fmt.Sprintf("fix mode failed with file '%s'", fixFile)
	guidance := "Run 'prompter --help' for fix mode usage and examples."
	
	if strings.Contains(cause.Error(), "not found") || strings.Contains(cause.Error(), "does not exist") {
		guidance = "Fix file not found. Run 'prompter --help' for fix mode setup."
	} else if strings.Contains(cause.Error(), "empty") {
		guidance = "Fix file is empty. Run 'prompter --help' for fix mode usage."
	}
	
	return &PrompterError{
		Type:     ErrFixModeInvalid,
		Message:  message,
		Guidance: guidance,
		Cause:    cause,
	}
}

func NewOutputError(target string, cause error) *PrompterError {
	message := fmt.Sprintf("failed to output to target '%s'", target)
	guidance := "Run 'prompter --help' for output target options."
	
	if target == "clipboard" {
		guidance = "Clipboard access failed. Try --target stdout or run 'prompter --help' for options."
	} else if strings.HasPrefix(target, "file:") {
		guidance = "File write failed. Run 'prompter --help' for output options."
	} else if strings.Contains(cause.Error(), "editor") {
		guidance = "Editor launch failed. Run 'prompter --help' for editor configuration."
	}
	
	return &PrompterError{
		Type:     ErrOutputFailed,
		Message:  message,
		Guidance: guidance,
		Cause:    cause,
	}
}

func NewValidationError(field string, value interface{}, reason string) *PrompterError {
	message := fmt.Sprintf("validation failed for %s: %v (%s)", field, value, reason)
	guidance := "Run 'prompter --help' for usage information."
	
	switch field {
	case "base_prompt":
		guidance = "Base prompt required in non-interactive mode. Run 'prompter --help' for options."
	case "target":
		guidance = "Invalid target. Run 'prompter --help' for valid output targets."
	case "config_path":
		guidance = "Invalid config path. Run 'prompter --help' for configuration options."
	case "template_name":
		guidance = "Invalid template name. Run 'prompter --help' for template usage."
	}
	
	return &PrompterError{
		Type:     ErrValidationFailed,
		Message:  message,
		Guidance: guidance,
		Cause:    nil,
	}
}

// Recovery strategies

// RecoverFromError attempts to recover from common errors with fallback strategies
func RecoverFromError(err error) error {
	if err == nil {
		return nil
	}
	
	var prompterErr *PrompterError
	if !errors.As(err, &prompterErr) {
		// Wrap unknown errors
		return &PrompterError{
			Type:     errors.New("unknown error"),
			Message:  err.Error(),
			Guidance: "Run 'prompter --help' for usage information.",
			Cause:    err,
		}
	}
	
	// Apply recovery strategies based on error type
	switch prompterErr.Type {
	case ErrConfigurationInvalid:
		return recoverFromConfigError(prompterErr)
	case ErrTemplateNotFound:
		return recoverFromTemplateError(prompterErr)
	case ErrOutputFailed:
		return recoverFromOutputError(prompterErr)
	default:
		return prompterErr
	}
}

func recoverFromConfigError(err *PrompterError) error {
	// Try to create default config directory if it doesn't exist
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return err // Can't recover
	}
	
	configDir := fmt.Sprintf("%s/.config/prompter", homeDir)
	if _, statErr := os.Stat(configDir); os.IsNotExist(statErr) {
		if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
			// Add recovery attempt info to guidance
			err.Guidance = "Run 'prompter --help' for configuration help."
			return err
		}
		
		// Successfully created directory
		err.Guidance = "Config directory created. Run 'prompter --help' for configuration options."
	}
	
	return err
}

func recoverFromTemplateError(err *PrompterError) error {
	// For template not found errors, we can suggest continuing without the template
	if strings.Contains(err.Message, "not found") {
		err.Guidance = "Template not found. Run 'prompter --help' for template setup or omit template flags."
	}
	return err
}

func recoverFromOutputError(err *PrompterError) error {
	// For clipboard errors, suggest stdout fallback
	if strings.Contains(err.Message, "clipboard") {
		err.Guidance = "Clipboard failed. Try --target stdout or run 'prompter --help' for options."
	}
	return err
}

// IsRecoverableError checks if an error can be recovered from
func IsRecoverableError(err error) bool {
	var prompterErr *PrompterError
	if !errors.As(err, &prompterErr) {
		return false
	}
	
	// Some errors are recoverable with user intervention
	switch prompterErr.Type {
	case ErrTemplateNotFound:
		return true // Can continue without template
	case ErrOutputFailed:
		return strings.Contains(prompterErr.Message, "clipboard") // Can fallback to stdout
	default:
		return false
	}
}