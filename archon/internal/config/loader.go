package config

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/pelletier/go-toml/v2"
)

// datetimeStringRe matches TOML datetime values that were incorrectly
// stored as quoted strings (e.g. last_health_check = '2026-01-01T00:00:00Z')
// and converts them to native TOML datetimes (unquoted).
var datetimeStringRe = regexp.MustCompile(`(last_health_check\s*=\s*)['"](\d{4}-\d{2}-\d{2}T[^'"]+)['"]`)

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

	// Run migrations if needed
	migrated, err := MigrateConfig(&config, path)
	if err != nil {
		return nil, err
	}

	// If migrations were applied, save the config
	if migrated {
		if err := f.Save(path, &config); err != nil {
			return nil, err
		}
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
		Version:          config.Version,
		MigrationVersion: config.MigrationVersion,
		Sites:            []models.Site{},   // Empty - stored in directories
		Domains:          config.Domains,    // Keep in main config
		Nodes:            []models.Node{},   // Empty - stored in directories
		Settings:         config.Settings,
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

			// Fix legacy files where *time.Time was marshaled as a quoted string
			// instead of a native TOML datetime
			data = datetimeStringRe.ReplaceAll(data, []byte("${1}${2}"))

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

// DeletedSite represents an archived site with metadata
type DeletedSite struct {
	Site        models.Site `json:"site" toml:"site"`
	DomainName  string      `json:"domain_name" toml:"domain_name"`
	ArchivePath string      `json:"archive_path" toml:"archive_path"`
	DeletedAt   string      `json:"deleted_at" toml:"deleted_at"` // Timestamp from directory name
}

// ArchiveSite moves a site to the deleted archive before deletion
func (f *FileConfigLoader) ArchiveSite(siteName, domainName string, site models.Site) (string, error) {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return "", err
	}
	return f.ArchiveSiteWithBaseDir(baseDir, siteName, domainName, site)
}

// ArchiveSiteWithBaseDir archives a site with a configurable base directory (for testing)
func (f *FileConfigLoader) ArchiveSiteWithBaseDir(baseDir, siteName, domainName string, site models.Site) (string, error) {
	// Source path
	srcPath := filepath.Join(baseDir, "sites", domainName, siteName)

	// Check if source exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", err
	}

	// Create archive path with timestamp
	timestamp := site.UpdatedAt.Format("20060102-150405")
	if timestamp == "00010101-000000" {
		// Fallback to current time if UpdatedAt is not set
		timestamp = filepath.Base(srcPath) + "-" + string(rune(os.Getpid()))
	}
	archivePath := filepath.Join(baseDir, "deleted", domainName, siteName, timestamp)

	// Create archive directory
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		return "", err
	}

	// Read original config
	configData, err := os.ReadFile(filepath.Join(srcPath, "config.toml"))
	if err != nil {
		return "", err
	}

	// Write to archive
	if err := os.WriteFile(filepath.Join(archivePath, "config.toml"), configData, 0644); err != nil {
		return "", err
	}

	// Remove original directory
	if err := os.RemoveAll(srcPath); err != nil {
		return "", err
	}

	return archivePath, nil
}

// LoadDeletedSites returns all archived (deleted) sites
func (f *FileConfigLoader) LoadDeletedSites() ([]DeletedSite, error) {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return nil, err
	}
	return f.LoadDeletedSitesWithBaseDir(baseDir)
}

// LoadDeletedSitesWithBaseDir loads deleted sites with a configurable base directory (for testing)
func (f *FileConfigLoader) LoadDeletedSitesWithBaseDir(baseDir string) ([]DeletedSite, error) {
	var deletedSites []DeletedSite

	deletedDir := filepath.Join(baseDir, "deleted")

	// Check if deleted directory exists
	if _, err := os.Stat(deletedDir); os.IsNotExist(err) {
		return deletedSites, nil // Return empty list, not an error
	}

	// Iterate over domains
	domainDirs, err := os.ReadDir(deletedDir)
	if err != nil {
		return nil, err
	}

	for _, domainDir := range domainDirs {
		if !domainDir.IsDir() {
			continue
		}
		domainName := domainDir.Name()
		domainPath := filepath.Join(deletedDir, domainName)

		// Iterate over sites within domain
		siteDirs, err := os.ReadDir(domainPath)
		if err != nil {
			continue
		}

		for _, siteDir := range siteDirs {
			if !siteDir.IsDir() {
				continue
			}
			siteName := siteDir.Name()
			sitePath := filepath.Join(domainPath, siteName)

			// Iterate over timestamps (archived versions)
			timestampDirs, err := os.ReadDir(sitePath)
			if err != nil {
				continue
			}

			for _, tsDir := range timestampDirs {
				if !tsDir.IsDir() {
					continue
				}
				archivePath := filepath.Join(sitePath, tsDir.Name())
				configPath := filepath.Join(archivePath, "config.toml")

				// Read config file
				data, err := os.ReadFile(configPath)
				if err != nil {
					continue
				}

				var site models.Site
				if err := toml.Unmarshal(data, &site); err != nil {
					continue
				}

				deletedSites = append(deletedSites, DeletedSite{
					Site:        site,
					DomainName:  domainName,
					ArchivePath: archivePath,
					DeletedAt:   tsDir.Name(),
				})
			}
		}
	}

	return deletedSites, nil
}

// RestoreDeletedSite moves an archived site back to active sites
func (f *FileConfigLoader) RestoreDeletedSite(archivePath, siteName, domainName string) error {
	baseDir, err := GetArchonConfigDir()
	if err != nil {
		return err
	}
	return f.RestoreDeletedSiteWithBaseDir(baseDir, archivePath, siteName, domainName)
}

// RestoreDeletedSiteWithBaseDir restores a site with a configurable base directory (for testing)
func (f *FileConfigLoader) RestoreDeletedSiteWithBaseDir(baseDir, archivePath, siteName, domainName string) error {
	// Check if archive exists
	configPath := filepath.Join(archivePath, "config.toml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return err
	}

	// Destination path
	destPath := filepath.Join(baseDir, "sites", domainName, siteName)

	// Create destination directory
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}

	// Read archived config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Write to active location
	if err := os.WriteFile(filepath.Join(destPath, "config.toml"), data, 0644); err != nil {
		return err
	}

	// Remove archive directory
	if err := os.RemoveAll(archivePath); err != nil {
		return err
	}

	return nil
}
