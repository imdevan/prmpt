package config

import (
	"testing"
	"prompter-cli/internal/interfaces"
)

// TestManagerImplementsInterface verifies that Manager implements ConfigManager interface
func TestManagerImplementsInterface(t *testing.T) {
	var _ interfaces.ConfigManager = (*Manager)(nil)
	
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
	
	// Test that all interface methods are callable
	config, err := manager.Load("")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	
	err = manager.Validate(config)
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}
	
	resolvedConfig, err := manager.Resolve()
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}
	
	if resolvedConfig == nil {
		t.Fatal("Resolve() returned nil config")
	}
}