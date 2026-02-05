package config

import (
	"fmt"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/google/uuid"
)

type Config struct {
	Version  string          `toml:"version"`
	Sites    []models.Site   `toml:"sites"`
	Domains  []models.Domain `toml:"domains"`
	Nodes    []models.Node   `toml:"nodes"`
	Settings Settings        `toml:"settings"`
}

// DockerCredential represents stored Docker registry credentials
type DockerCredential struct {
	ID       uuid.UUID `toml:"id"`
	Name     string    `toml:"name"`     // Display name (e.g., "GitHub Container Registry", "DockerHub")
	Registry string    `toml:"registry"` // Registry URL (e.g., "ghcr.io", "docker.io")
	Username string    `toml:"username"`
	Token    string    `toml:"token"` // Will be encrypted at rest in future
}

type Settings struct {
	AutoSave                bool               `toml:"auto_save"`
	HealthCheckIntervalSecs int                `toml:"health_check_interval_secs"`
	DefaultDnsTTL           int                `toml:"default_dns_ttl"`
	Theme                   string             `toml:"theme"`
	CloudflareAPIToken      string             `toml:"cloudflare_api_token,omitempty"`  // Global default
	Route53AccessKey        string             `toml:"route53_access_key,omitempty"`    // Global default
	Route53SecretKey        string             `toml:"route53_secret_key,omitempty"`    // Global default
	DockerCredentials       []DockerCredential `toml:"docker_credentials,omitempty"`
}

// AddDockerCredential adds a new Docker credential to settings
func (s *Settings) AddDockerCredential(cred DockerCredential) error {
	if cred.Name == "" {
		return fmt.Errorf("credential name is required")
	}
	if cred.Registry == "" {
		return fmt.Errorf("credential registry is required")
	}

	// Generate new ID if not set
	if cred.ID == uuid.Nil {
		cred.ID = uuid.New()
	}

	s.DockerCredentials = append(s.DockerCredentials, cred)
	return nil
}

// GetDockerCredentialByID returns a credential by ID, or nil if not found
func (s *Settings) GetDockerCredentialByID(id uuid.UUID) *DockerCredential {
	if id == uuid.Nil {
		return nil
	}
	for i := range s.DockerCredentials {
		if s.DockerCredentials[i].ID == id {
			return &s.DockerCredentials[i]
		}
	}
	return nil
}

// UpdateDockerCredential updates an existing credential
func (s *Settings) UpdateDockerCredential(id uuid.UUID, updated DockerCredential) error {
	for i := range s.DockerCredentials {
		if s.DockerCredentials[i].ID == id {
			// Preserve the ID
			updated.ID = id
			s.DockerCredentials[i] = updated
			return nil
		}
	}
	return fmt.Errorf("credential with ID %s not found", id)
}

// DeleteDockerCredential removes a credential by ID
func (s *Settings) DeleteDockerCredential(id uuid.UUID) error {
	for i := range s.DockerCredentials {
		if s.DockerCredentials[i].ID == id {
			s.DockerCredentials = append(s.DockerCredentials[:i], s.DockerCredentials[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("credential with ID %s not found", id)
}

// ListDockerCredentials returns all stored credentials
func (s *Settings) ListDockerCredentials() []DockerCredential {
	if s.DockerCredentials == nil {
		return []DockerCredential{}
	}
	return s.DockerCredentials
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
	DeleteSite(siteName, domainName string) error
	DeleteNode(nodeName string) error
}
