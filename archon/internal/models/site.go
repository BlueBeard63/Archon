package models

import (
	"fmt"
	"strconv"
	"strings"
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
	DomainID        uuid.UUID         `json:"domain_id" toml:"domain_id"` // Legacy: single domain (kept for backward compatibility)
	NodeID          uuid.UUID         `json:"node_id" toml:"node_id"`
	DockerImage     string            `json:"docker_image" toml:"docker_image"`
	DockerUsername  string            `json:"docker_username,omitempty" toml:"docker_username,omitempty"`
	DockerToken     string            `json:"docker_token,omitempty" toml:"docker_token,omitempty"`
	EnvironmentVars map[string]string `json:"environment_vars" toml:"environment_vars"`
	Port            int               `json:"port" toml:"port"`                                           // Legacy: single port (kept for backward compatibility)
	DomainMappings  []DomainMapping   `json:"domain_mappings,omitempty" toml:"domain_mappings,omitempty"` // New: multiple domain-port mappings
	SSLEnabled      bool              `json:"ssl_enabled" toml:"ssl_enabled"`
	SSLEmail        string            `json:"ssl_email,omitempty" toml:"ssl_email,omitempty"` // Email for Let's Encrypt certificate registration
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

// DomainMapping represents a domain-to-port mapping for multi-domain sites
type DomainMapping struct {
	DomainID  uuid.UUID `json:"domain_id" toml:"domain_id"`
	Subdomain string    `json:"subdomain,omitempty" toml:"subdomain,omitempty"` // Optional subdomain (e.g., "www", "api", "app"). Empty = root domain
	Port      int       `json:"port" toml:"port"`                               // Container port
	HostPort  int       `json:"host_port,omitempty" toml:"host_port,omitempty"` // Host port (optional, defaults to Port if not specified)
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
	site := &Site{
		ID:              uuid.New(),
		Name:            name,
		DomainID:        domainID, // Set legacy field for backward compatibility
		NodeID:          nodeID,
		DockerImage:     dockerImage,
		DockerUsername:  "",
		DockerToken:     "",
		EnvironmentVars: make(map[string]string),
		Port:            port, // Set legacy field for backward compatibility
		SSLEnabled:      true, // Default to SSL enabled
		ConfigFiles:     []ConfigFile{},
		Status:          SiteStatusInactive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Initialize DomainMappings with the primary domain-port pair
	site.DomainMappings = []DomainMapping{
		{
			DomainID: domainID,
			Port:     port,
		},
	}

	return site
}

// GetDomainMappings returns all domain-port mappings for this site
// Falls back to legacy DomainID/Port fields if DomainMappings is empty
func (s *Site) GetDomainMappings() []DomainMapping {
	if len(s.DomainMappings) > 0 {
		return s.DomainMappings
	}

	// Fallback to legacy fields
	if s.DomainID != uuid.Nil && s.Port > 0 {
		return []DomainMapping{
			{
				DomainID: s.DomainID,
				Port:     s.Port,
			},
		}
	}

	return []DomainMapping{}
}

// AddDomainMapping adds a new domain-port mapping to the site
func (s *Site) AddDomainMapping(domainID uuid.UUID, port int) {
	s.DomainMappings = append(s.DomainMappings, DomainMapping{
		DomainID: domainID,
		Port:     port,
	})
	s.UpdatedAt = time.Now()
}

// RemoveDomainMapping removes a domain-port mapping by index
func (s *Site) RemoveDomainMapping(index int) {
	if index >= 0 && index < len(s.DomainMappings) {
		s.DomainMappings = append(s.DomainMappings[:index], s.DomainMappings[index+1:]...)
		s.UpdatedAt = time.Now()
	}
}

// GetFullDomain returns the full domain name for a mapping (subdomain.domain or just domain)
func GetFullDomain(domainName, subdomain string) string {
	if subdomain == "" {
		return domainName
	}
	return subdomain + "." + domainName
}

// ParsePortMapping parses port notation from a string
// Accepts formats:
//   - "3000" - single port (container and host use same port)
//   - "3000:3001" - container port 3000 mapped to host port 3001
//
// Returns (containerPort, hostPort, error)
func ParsePortMapping(portStr string) (int, int, error) {
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		return 0, 0, fmt.Errorf("port string cannot be empty")
	}

	parts := strings.Split(portStr, ":")
	if len(parts) == 1 {
		// Single port - use same for both container and host
		port, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid port: %s", portStr)
		}
		if port < 1 || port > 65535 {
			return 0, 0, fmt.Errorf("port out of range (1-65535): %d", port)
		}
		return port, port, nil
	} else if len(parts) == 2 {
		// Container:Host format
		containerPort, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid container port: %s", parts[0])
		}
		if containerPort < 1 || containerPort > 65535 {
			return 0, 0, fmt.Errorf("container port out of range (1-65535): %d", containerPort)
		}

		hostPort, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return 0, 0, fmt.Errorf("invalid host port: %s", parts[1])
		}
		if hostPort < 1 || hostPort > 65535 {
			return 0, 0, fmt.Errorf("host port out of range (1-65535): %d", hostPort)
		}

		return containerPort, hostPort, nil
	}

	return 0, 0, fmt.Errorf("invalid port format: %s (use '3000' or '3000:3001')", portStr)
}

// FormatPortMapping formats port mapping for display
// If hostPort is 0 or same as containerPort, displays just the port number
// Otherwise displays in "container:host" format
func FormatPortMapping(containerPort, hostPort int) string {
	if hostPort == 0 || hostPort == containerPort {
		return fmt.Sprintf("%d", containerPort)
	}
	return fmt.Sprintf("%d:%d", containerPort, hostPort)
}

// GetEffectiveHostPort returns the host port for a domain mapping
// If HostPort is not set (0), returns Port as the default
func (dm *DomainMapping) GetEffectiveHostPort() int {
	if dm.HostPort > 0 {
		return dm.HostPort
	}
	return dm.Port
}
