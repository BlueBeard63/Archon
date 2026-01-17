package models

import (
	"fmt"
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

type SiteType string

const (
	SiteTypeContainer SiteType = "container"
	SiteTypeCompose   SiteType = "compose"
)

// DeployRequest is the request to deploy a site
type DeployRequest struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	SiteType        SiteType          `json:"site_type"`                   // container or compose (defaults to container)
	Docker          Docker            `json:"docker"`                      // Used for container deployments
	ComposeContent  string            `json:"compose_content,omitempty"`   // Docker Compose YAML content (for compose deployments)
	EnvironmentVars map[string]string `json:"environment_vars"`
	DomainMappings  []DomainMapping   `json:"domain_mappings"`
	SSLEnabled      bool              `json:"ssl_enabled"`
	SSLEmail        string            `json:"ssl_email,omitempty"` // Email for Let's Encrypt
	SSLCert         string            `json:"ssl_cert,omitempty"`  // Base64 encoded cert
	SSLKey          string            `json:"ssl_key,omitempty"`   // Base64 encoded key
	ConfigFiles     []ConfigFile      `json:"config_files"`
	TraefikLabels   map[string]string `json:"traefik_labels,omitempty"`
}

// IsCompose returns true if this is a compose deployment
func (r *DeployRequest) IsCompose() bool {
	return r.SiteType == SiteTypeCompose
}

// Validate checks that required fields are present based on site type
func (r *DeployRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.DomainMappings) == 0 {
		return fmt.Errorf("at least one domain mapping is required")
	}

	if r.IsCompose() {
		if r.ComposeContent == "" {
			return fmt.Errorf("compose content is required for compose deployments")
		}
	} else {
		if r.Docker.Image == "" {
			return fmt.Errorf("docker image is required for container deployments")
		}
	}
	return nil
}

// DomainMapping represents a domain-to-port mapping for multi-domain sites
type DomainMapping struct {
	Domain string `json:"domain"` // Full domain (e.g., "api.example.com")
	Port   int    `json:"port"`   // Container port for this domain
}

type Docker struct {
	Credentials DockerCredentials `json:"credentials"`
	Image       string            `json:"image"`
}

type DockerCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ConfigFile struct {
	Name          string `json:"name"`
	Content       string `json:"content"`
	ContainerPath string `json:"container_path"`
}

// DeployResponse is the response from deploying a site
type DeployResponse struct {
	SiteID      uuid.UUID  `json:"site_id"`
	Status      SiteStatus `json:"status"`
	ContainerID string     `json:"container_id,omitempty"`
	Message     string     `json:"message,omitempty"`
}

// SiteStatusResponse returns the current status of a site
type SiteStatusResponse struct {
	SiteID      uuid.UUID  `json:"site_id"`
	Status      SiteStatus `json:"status"`
	ContainerID string     `json:"container_id,omitempty"`
	IsRunning   bool       `json:"is_running"`
	Message     string     `json:"message,omitempty"`
}

// HealthResponse returns the health status of the node
type HealthResponse struct {
	Status  string       `json:"status"`
	Docker  *DockerInfo  `json:"docker,omitempty"`
	Traefik *TraefikInfo `json:"traefik,omitempty"`
}

type DockerInfo struct {
	Version           string `json:"version"`
	ContainersRunning int    `json:"containers_running"`
	ImagesCount       int    `json:"images_count"`
}

type TraefikInfo struct {
	Version       string `json:"version"`
	RoutersCount  int    `json:"routers_count"`
	ServicesCount int    `json:"services_count"`
}

// ContainerMetrics contains metrics for a container
type ContainerMetrics struct {
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryUsage    int64   `json:"memory_usage"`
	MemoryLimit    int64   `json:"memory_limit"`
	NetworkRxBytes int64   `json:"network_rx_bytes"`
	NetworkTxBytes int64   `json:"network_tx_bytes"`
}

// Site represents a deployed site on this node
type Site struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	SiteType        SiteType          `json:"site_type"` // container or compose
	Domain          string            `json:"domain"`
	DockerImage     string            `json:"docker_image"`
	ContainerID     string            `json:"container_id"`
	Port            int               `json:"port"`
	SSLEnabled      bool              `json:"ssl_enabled"`
	Status          SiteStatus        `json:"status"`
	EnvironmentVars map[string]string `json:"environment_vars"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
