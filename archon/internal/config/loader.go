package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/pelletier/go-toml/v2"
)

// FileConfigLoader implements ConfigLoader using file-based TOML storage
type FileConfigLoader struct{}

// NewFileConfigLoader creates a new file-based config loader
func NewFileConfigLoader() *FileConfigLoader {
	return &FileConfigLoader{}
}

// Load reads configuration from a TOML file
func (f *FileConfigLoader) Load(path string) (*Config, error) {
	// Check if file exists
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create default config and save it
			defaultCfg := DefaultConfig()
			if saveErr := f.Save(path, defaultCfg); saveErr != nil {
				// If we can't save, still return default config
				return defaultCfg, nil
			}
			return defaultCfg, nil
		}
		return nil, err
	}

	// Unmarshal TOML data
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save writes configuration to a TOML file
func (f *FileConfigLoader) Save(path string, config *Config) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Marshal config to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}

// DefaultConfigPath returns the platform-specific default config path
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// On Windows, use AppData
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appData, "archon", "config.toml"), nil
	}

	// On Unix-like systems, use XDG_CONFIG_HOME or ~/.config
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "archon", "config.toml"), nil
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Version:  "1.0.0",
		Sites:    []models.Site{},
		Domains:  []models.Domain{},
		Nodes:    []models.Node{},
		Settings: DefaultSettings(),
	}
}
