package config

import (
	"github.com/BlueBeard63/archon/internal/models"
)

type Config struct {
	Version  string          `toml:"version"`
	Sites    []models.Site   `toml:"sites"`
	Domains  []models.Domain `toml:"domains"`
	Nodes    []models.Node   `toml:"nodes"`
	Settings Settings        `toml:"settings"`
}

type Settings struct {
	AutoSave                bool   `toml:"auto_save"`
	HealthCheckIntervalSecs int    `toml:"health_check_interval_secs"`
	DefaultDnsTTL           int    `toml:"default_dns_ttl"`
	Theme                   string `toml:"theme"`
}

// DefaultSettings returns default configuration settings
func DefaultSettings() Settings {
	return Settings{
		AutoSave:                true,
		HealthCheckIntervalSecs: 300, // 5 minutes
		DefaultDnsTTL:           300, // 5 minutes
		Theme:                   "default",
	}
}

// ConfigLoader interface for loading and saving configuration
type ConfigLoader interface {
	Load(path string) (*Config, error)
	Save(path string, config *Config) error
}
