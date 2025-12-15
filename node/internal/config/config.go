package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type ProxyType string

const (
	ProxyTypeNginx   ProxyType = "nginx"
	ProxyTypeApache  ProxyType = "apache"
	ProxyTypeTraefik ProxyType = "traefik"
)

type SSLMode string

const (
	SSLModeManual       SSLMode = "manual"       // User provides cert/key files
	SSLModeLetsEncrypt  SSLMode = "letsencrypt"  // Auto LetsEncrypt
	SSLModeTraefikAuto  SSLMode = "traefik-auto" // Traefik handles it
)

type Config struct {
	Server     ServerConfig     `toml:"server"`
	Proxy      ProxyConfig      `toml:"proxy"`
	Docker     DockerConfig     `toml:"docker"`
	SSL        SSLConfig        `toml:"ssl"`
	LetsEncrypt LetsEncryptConfig `toml:"letsencrypt"`
}

type ServerConfig struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	APIKey     string `toml:"api_key"`
	DataDir    string `toml:"data_dir"`
}

type ProxyConfig struct {
	Type          ProxyType `toml:"type"`
	ConfigDir     string    `toml:"config_dir"`
	ReloadCommand string    `toml:"reload_command"`
}

type DockerConfig struct {
	Host    string `toml:"host"`
	Network string `toml:"network"`
}

type SSLConfig struct {
	Mode        SSLMode `toml:"mode"`
	CertDir     string  `toml:"cert_dir"`
	Email       string  `toml:"email"`
}

type LetsEncryptConfig struct {
	Enabled     bool   `toml:"enabled"`
	Email       string `toml:"email"`
	StagingMode bool   `toml:"staging_mode"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:    "0.0.0.0",
			Port:    8080,
			APIKey:  "",
			DataDir: "/var/lib/archon",
		},
		Proxy: ProxyConfig{
			Type:          ProxyTypeNginx,
			ConfigDir:     "/etc/nginx/sites-enabled",
			ReloadCommand: "nginx -s reload",
		},
		Docker: DockerConfig{
			Host:    "unix:///var/run/docker.sock",
			Network: "archon-net",
		},
		SSL: SSLConfig{
			Mode:    SSLModeLetsEncrypt,
			CertDir: "/etc/archon/ssl",
			Email:   "",
		},
		LetsEncrypt: LetsEncryptConfig{
			Enabled:     true,
			Email:       "",
			StagingMode: false,
		},
	}
}

// Load reads config from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			config := DefaultConfig()
			if err := Save(path, config); err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// Save writes config to file
func Save(path string, config *Config) error {
	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := toml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
