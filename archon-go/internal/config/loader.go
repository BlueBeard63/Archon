package config

import (
	"os"
	"path/filepath"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/pelletier/go-toml/v2"
)

// FileConfigLoader implements ConfigLoader using file-based TOML storage
type FileConfigLoader struct{}

// NewFileConfigLoader creates a new file-based config loader
func NewFileConfigLoader() *FileConfigLoader {
	// Suppress unused import warnings for skeleton code
	_ = os.ReadFile
	_ = filepath.Join
	_ = toml.Marshal

	return &FileConfigLoader{}
}

// Load reads configuration from a TOML file
func (f *FileConfigLoader) Load(path string) (*Config, error) {
	// TODO: Implement TOML loading
	// Steps:
	// 1. Check if file exists at path
	// 2. If not exists, return DefaultConfig() and create the file
	// 3. Read file contents
	// 4. Unmarshal with toml.Unmarshal()
	// 5. Return config or error

	// Example implementation structure:
	// data, err := os.ReadFile(path)
	// if err != nil {
	//     if os.IsNotExist(err) {
	//         return DefaultConfig(), nil
	//     }
	//     return nil, err
	// }
	//
	// var config Config
	// if err := toml.Unmarshal(data, &config); err != nil {
	//     return nil, err
	// }
	// return &config, nil

	return nil, nil
}

// Save writes configuration to a TOML file
func (f *FileConfigLoader) Save(path string, config *Config) error {
	// TODO: Implement TOML saving
	// Steps:
	// 1. Create parent directories if they don't exist
	// 2. Marshal config to TOML bytes
	// 3. Write to file with appropriate permissions (0644)
	// 4. Return error if any step fails

	// Example implementation structure:
	// if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
	//     return err
	// }
	//
	// data, err := toml.Marshal(config)
	// if err != nil {
	//     return err
	// }
	//
	// return os.WriteFile(path, data, 0644)

	return nil
}

// DefaultConfigPath returns the platform-specific default config path
func DefaultConfigPath() (string, error) {
	// TODO: Implement platform-specific config path detection
	// Linux/Mac: ~/.config/archon/config.toml
	// Windows: %APPDATA%\archon\config.toml

	// Example implementation:
	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	//     return "", err
	// }
	//
	// // On Windows, use AppData
	// if runtime.GOOS == "windows" {
	//     return filepath.Join(os.Getenv("APPDATA"), "archon", "config.toml"), nil
	// }
	//
	// // On Unix-like systems, use XDG_CONFIG_HOME or ~/.config
	// configDir := os.Getenv("XDG_CONFIG_HOME")
	// if configDir == "" {
	//     configDir = filepath.Join(homeDir, ".config")
	// }
	//
	// return filepath.Join(configDir, "archon", "config.toml"), nil

	return "", nil
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
