package interfaces

// CustomTemplate represents a custom template configuration
type CustomTemplate struct {
	Location    string `toml:"location"`
	Interactive bool   `toml:"interactive"`
	Flag        string `toml:"flag"`
	Shorthand   string `toml:"shorthand"`
	Type        string `toml:"type"`        // "pre" or "post", defaults to "pre"
	Description string `toml:"description"` // Custom help description
}

// Config represents the application configuration
type Config struct {
	PromptsLocation      string                     `toml:"prompts_location"`
	LocalPromptsLocation string                     `toml:"local_prompts_location"`
	Editor               string                     `toml:"editor"`
	DefaultPre           string                     `toml:"default_pre"`
	DefaultPost          string                     `toml:"default_post"`
	FixFile              string                     `toml:"fix_file"`
	DirectoryStrategy    string                     `toml:"directory_strategy"`
	Target               string                     `toml:"target"`
	InteractiveDefault   bool                       `toml:"interactive_default"`
	CustomTemplates      map[string]CustomTemplate `toml:"custom_template"`
}

// ConfigManager handles configuration loading and resolution
type ConfigManager interface {
	// Load loads configuration from the specified path
	Load(path string) (*Config, error)
	
	// Resolve applies precedence rules (flags > env > config > defaults)
	Resolve() (*Config, error)
	
	// Validate validates the configuration values
	Validate(config *Config) error
}