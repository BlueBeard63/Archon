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

// Load reads configuration from a TOML file and aggregates directory-based storage
func (f *FileConfigLoader) Load(path string) (*Config, error) {
	var config Config

	// Check if legacy config file exists
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with default config
			config = *DefaultConfig()
		} else {
			return nil, err
		}
	} else {
		// Unmarshal legacy TOML data
		if err := toml.Unmarshal(data, &config); err != nil {
			return nil, err
		}
	}

	// Load sites from directory structure (overrides legacy sites)
	sites, err := f.LoadAllSites()
	if err == nil && len(sites) > 0 {
		config.Sites = sites
	}

	// Load nodes from directory structure (overrides legacy nodes)
	nodes, err := f.LoadAllNodes()
	if err == nil && len(nodes) > 0 {
		config.Nodes = nodes
	}

	// If config is completely empty, initialize with defaults
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	if config.Settings.AutoSave == false && config.Settings.HealthCheckIntervalSecs == 0 {
		config.Settings = DefaultSettings()
	}

	return &config, nil
}

// Save writes configuration using new directory structure for sites/nodes
// and legacy config file for domains/settings
func (f *FileConfigLoader) Save(path string, config *Config) error {
	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Create a copy of config without sites and nodes (stored separately)
	legacyConfig := Config{
		Version:  config.Version,
		Sites:    []models.Site{},    // Empty - stored in directories
		Domains:  config.Domains,      // Keep in main config
		Nodes:    []models.Node{},     // Empty - stored in directories
		Settings: config.Settings,
	}

	// Save main config file (domains and settings only)
	data, err := toml.Marshal(legacyConfig)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Save each site to its directory
	for _, site := range config.Sites {
		// Get domain name for directory structure
		domainName := "unknown"
		for _, domain := range config.Domains {
			if domain.ID == site.DomainID {
				domainName = domain.Name
				break
			}
		}

		if err := f.SaveSite(&site, domainName); err != nil {
			// Log error but continue saving other sites
			// In a production system, you might want to collect errors and return them
			continue
		}
	}

	// Save each node to its directory
	for _, node := range config.Nodes {
		if err := f.SaveNode(&node); err != nil {
			// Log error but continue saving other nodes
			continue
		}
	}

	return nil
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

// GetArchonConfigDir returns the base archon config directory
func GetArchonConfigDir() (string, error) {
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
		return filepath.Join(appData, "archon"), nil
	}

	// On Unix-like systems, use XDG_CONFIG_HOME or ~/.config
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".config")
	}

	return filepath.Join(configDir, "archon"), nil
}

// SaveSite saves a single site to its directory structure
// Path: ~/.config/archon/sites/[domain]/[subdomain]/[siteName]/config.toml
func (f *FileConfigLoader) SaveSite(site *models.Site, domainName string) error {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return err
	}

	// Parse domain name to extract subdomain and root domain
	// For simplicity, treat the entire domain as the directory name
	// TODO: Could be enhanced to parse subdomain.domain.tld properly
	sitePath := filepath.Join(baseDir, "sites", domainName, site.Name, "config.toml")

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(sitePath), 0755); err != nil {
		return err
	}

	// Marshal site to TOML
	data, err := toml.Marshal(site)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(sitePath, data, 0644)
}

// LoadAllSites loads all sites from the directory structure
func (f *FileConfigLoader) LoadAllSites() ([]models.Site, error) {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return nil, err
	}

	sitesDir := filepath.Join(baseDir, "sites")

	// Check if sites directory exists
	if _, err := os.Stat(sitesDir); os.IsNotExist(err) {
		return []models.Site{}, nil
	}

	var sites []models.Site

	// Walk through sites/[domain]/[siteName]/config.toml
	err = filepath.Walk(sitesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for config.toml files
		if !info.IsDir() && info.Name() == "config.toml" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var site models.Site
			if err := toml.Unmarshal(data, &site); err != nil {
				return err
			}

			sites = append(sites, site)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return sites, nil
}

// SaveNode saves a single node to its directory structure
// Path: ~/.config/archon/nodes/[nodeName]/config.toml
func (f *FileConfigLoader) SaveNode(node *models.Node) error {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return err
	}

	nodePath := filepath.Join(baseDir, "nodes", node.Name, "config.toml")

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(nodePath), 0755); err != nil {
		return err
	}

	// Marshal node to TOML
	data, err := toml.Marshal(node)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(nodePath, data, 0644)
}

// LoadAllNodes loads all nodes from the directory structure
func (f *FileConfigLoader) LoadAllNodes() ([]models.Node, error) {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return nil, err
	}

	nodesDir := filepath.Join(baseDir, "nodes")

	// Check if nodes directory exists
	if _, err := os.Stat(nodesDir); os.IsNotExist(err) {
		return []models.Node{}, nil
	}

	var nodes []models.Node

	// Walk through nodes/[nodeName]/config.toml
	err = filepath.Walk(nodesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for config.toml files
		if !info.IsDir() && info.Name() == "config.toml" {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var node models.Node
			if err := toml.Unmarshal(data, &node); err != nil {
				return err
			}

			nodes = append(nodes, node)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// DeleteSite removes a site's directory
func (f *FileConfigLoader) DeleteSite(siteName, domainName string) error {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return err
	}

	sitePath := filepath.Join(baseDir, "sites", domainName, siteName)
	return os.RemoveAll(sitePath)
}

// DeleteNode removes a node's directory
func (f *FileConfigLoader) DeleteNode(nodeName string) error {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return err
	}

	nodePath := filepath.Join(baseDir, "nodes", nodeName)
	return os.RemoveAll(nodePath)
}
