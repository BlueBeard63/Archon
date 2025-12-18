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

// DeployRequest is the request to deploy a site
type DeployRequest struct {
	ID              uuid.UUID         `json:"id"`
	Name            string            `json:"name"`
	Domain          string            `json:"domain"`
	Docker          Docker            `json:"docker"`
	EnvironmentVars map[string]string `json:"environment_vars"`
	Port            int               `json:"port"`
	SSLEnabled      bool              `json:"ssl_enabled"`
	SSLEmail        string            `json:"ssl_email,omitempty"` // Email for Let's Encrypt
	SSLCert         string            `json:"ssl_cert,omitempty"`  // Base64 encoded cert
	SSLKey          string            `json:"ssl_key,omitempty"`   // Base64 encoded key
	ConfigFiles     []ConfigFile      `json:"config_files"`
	TraefikLabels   map[string]string `json:"traefik_labels,omitempty"`
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
