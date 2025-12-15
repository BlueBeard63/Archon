package models

import (
	"time"

	"github.com/google/uuid"
)

type SiteStatus string

const (
	SiteStatusInactive  SiteStatus = "inactive"
	SiteStatusDeploying SiteStatus = "deploying"
	SiteStatusRunning   SiteStatus = "running"
	SiteStatusFailed    SiteStatus = "failed"
	SiteStatusStopped   SiteStatus = "stopped"
)

type Site struct {
	ID              uuid.UUID         `json:"id" toml:"id"`
	Name            string            `json:"name" toml:"name"`
	DomainID        uuid.UUID         `json:"domain_id" toml:"domain_id"`
	NodeID          uuid.UUID         `json:"node_id" toml:"node_id"`
	DockerImage     string            `json:"docker_image" toml:"docker_image"`
	EnvironmentVars map[string]string `json:"environment_vars" toml:"environment_vars"`
	Port            int               `json:"port" toml:"port"`
	SSLEnabled      bool              `json:"ssl_enabled" toml:"ssl_enabled"`
	ConfigFiles     []ConfigFile      `json:"config_files" toml:"config_files"`
	Status          SiteStatus        `json:"status" toml:"status"`
	CreatedAt       time.Time         `json:"created_at" toml:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" toml:"updated_at"`
}

type ConfigFile struct {
	Name          string `json:"name" toml:"name"`
	Content       string `json:"content" toml:"content"`
	ContainerPath string `json:"container_path" toml:"container_path"`
}

// GenerateTraefikLabels generates Docker labels for Traefik reverse proxy configuration
// TODO: Implement this method to generate appropriate Traefik labels for automatic routing
// Reference: https://doc.traefik.io/traefik/routing/providers/docker/
func (s *Site) GenerateTraefikLabels(domainName string) map[string]string {
	labels := make(map[string]string)

	// TODO: Add Traefik labels here
	// Example labels to implement:
	// - traefik.enable=true
	// - traefik.http.routers.<site-id>.rule=Host(`<domain>`)
	// - traefik.http.routers.<site-id>.entrypoints=web,websecure
	// - traefik.http.services.<site-id>.loadbalancer.server.port=<port>
	// If SSL enabled:
	//   - traefik.http.routers.<site-id>.tls=true
	//   - traefik.http.routers.<site-id>.tls.certresolver=letsencrypt

	return labels
}

// NewSite creates a new Site with default values
func NewSite(name string, domainID, nodeID uuid.UUID, dockerImage string, port int) *Site {
	now := time.Now()
	return &Site{
		ID:              uuid.New(),
		Name:            name,
		DomainID:        domainID,
		NodeID:          nodeID,
		DockerImage:     dockerImage,
		EnvironmentVars: make(map[string]string),
		Port:            port,
		SSLEnabled:      true, // Default to SSL enabled
		ConfigFiles:     []ConfigFile{},
		Status:          SiteStatusInactive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
