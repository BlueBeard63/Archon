package models

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type NodeStatus string

const (
	NodeStatusUnknown  NodeStatus = "unknown"
	NodeStatusOnline   NodeStatus = "online"
	NodeStatusOffline  NodeStatus = "offline"
	NodeStatusDegraded NodeStatus = "degraded"
)

type ProxyType string

const (
	ProxyTypeNginx   ProxyType = "nginx"
	ProxyTypeApache  ProxyType = "apache"
	ProxyTypeTraefik ProxyType = "traefik"
)

type Node struct {
	ID              uuid.UUID    `json:"id" toml:"id"`
	Name            string       `json:"name" toml:"name"`
	APIEndpoint     string       `json:"api_endpoint" toml:"api_endpoint"`
	APIKey          string       `json:"api_key" toml:"api_key"`
	IPAddress       net.IP       `json:"ip_address" toml:"ip_address"`
	ProxyType       ProxyType    `json:"proxy_type" toml:"proxy_type"`
	Status          NodeStatus   `json:"status" toml:"status"`
	DockerInfo      *DockerInfo  `json:"docker_info,omitempty" toml:"docker_info,omitempty"`
	TraefikInfo     *TraefikInfo `json:"traefik_info,omitempty" toml:"traefik_info,omitempty"`
	LastHealthCheck *time.Time   `json:"last_health_check,omitempty" toml:"last_health_check,omitempty"`
}

type DockerInfo struct {
	Version           string `json:"version" toml:"version"`
	ContainersRunning int    `json:"containers_running" toml:"containers_running"`
	ImagesCount       int    `json:"images_count" toml:"images_count"`
}

type TraefikInfo struct {
	Version       string `json:"version" toml:"version"`
	RoutersCount  int    `json:"routers_count" toml:"routers_count"`
	ServicesCount int    `json:"services_count" toml:"services_count"`
}

// NewNode creates a new Node with default values
func NewNode(name, apiEndpoint, apiKey string, ipAddress net.IP, proxyType ProxyType) *Node {
	// Default to nginx if proxy type is empty
	if proxyType == "" {
		proxyType = ProxyTypeNginx
	}

	return &Node{
		ID:          uuid.New(),
		Name:        name,
		APIEndpoint: apiEndpoint,
		APIKey:      apiKey,
		IPAddress:   ipAddress,
		ProxyType:   proxyType,
		Status:      NodeStatusUnknown,
	}
}

// generateSSLConfig generates SSL configuration based on proxy type
func (n *Node) generateSSLConfig() string {
	if n.ProxyType == ProxyTypeTraefik {
		return `# Traefik handles SSL automatically
[ssl]
mode = "traefik-auto"
cert_dir = ""
email = "admin@example.com"  # CHANGE THIS to your email`
	}

	// For Nginx and Apache
	return `# Let's Encrypt automatic SSL
[ssl]
mode = "letsencrypt"
cert_dir = "/etc/archon/ssl"
email = "admin@example.com"  # CHANGE THIS to your email

[letsencrypt]
enabled = true
email = "admin@example.com"  # CHANGE THIS to your email
staging_mode = false  # Set to true for testing`
}

// generateInstallInstructions generates installation instructions based on proxy type
func (n *Node) generateInstallInstructions() string {
	var packages string
	switch n.ProxyType {
	case ProxyTypeNginx:
		packages = "docker.io nginx certbot"
	case ProxyTypeApache:
		packages = "docker.io apache2 certbot"
	case ProxyTypeTraefik:
		packages = "docker.io"
	default:
		packages = "docker.io nginx certbot"
	}

	return `# 1. Install prerequisites on your server (` + n.IPAddress.String() + `):
#    Ubuntu/Debian:
#      sudo apt update
#      sudo apt install ` + packages + `
#      sudo systemctl enable docker
#      sudo systemctl start docker`
}

// GenerateNodeConfigTOML generates the TOML configuration file content for this node
func (n *Node) GenerateNodeConfigTOML() string {
	// Determine proxy configuration based on selected type
	var proxyConfig string
	switch n.ProxyType {
	case ProxyTypeNginx:
		proxyConfig = `[proxy]
type = "nginx"
config_dir = "/etc/nginx/sites-enabled"
reload_command = "nginx -s reload"`
	case ProxyTypeApache:
		proxyConfig = `[proxy]
type = "apache"
config_dir = "/etc/apache2/sites-enabled"
reload_command = "apache2ctl graceful"`
	case ProxyTypeTraefik:
		proxyConfig = `[proxy]
type = "traefik"
config_dir = ""
reload_command = ""`
	default:
		proxyConfig = `[proxy]
type = "nginx"
config_dir = "/etc/nginx/sites-enabled"
reload_command = "nginx -s reload"`
	}

	return `# Archon Node Server Configuration
# Place this file at: /etc/archon/node-config.toml
# Generated for node: ` + n.Name + `
# Proxy Type: ` + string(n.ProxyType) + `

[server]
host = "0.0.0.0"
port = 8080
# IMPORTANT: This API key must match the one in your Archon TUI config
api_key = "` + n.APIKey + `"
data_dir = "/var/lib/archon"

# ==============================================================================
# Proxy Configuration
# ==============================================================================

` + proxyConfig + `

# ==============================================================================
# Docker Configuration
# ==============================================================================

[docker]
host = "unix:///var/run/docker.sock"
network = "archon-net"

# ==============================================================================
# SSL Configuration
# ==============================================================================

` + n.generateSSLConfig() + `

# ==============================================================================
# Installation Instructions
# ==============================================================================
` + n.generateInstallInstructions() + `
#
# 2. Create directories:
#      sudo mkdir -p /etc/archon
#      sudo mkdir -p /var/lib/archon
#      sudo mkdir -p /etc/archon/ssl
#
# 3. Save this config to /etc/archon/node-config.toml:
#      sudo nano /etc/archon/node-config.toml
#      # Paste this content and save
#
# 4. Download and install archon-node:
#      # Build from source or download binary
#      sudo cp archon-node /usr/local/bin/
#      sudo chmod +x /usr/local/bin/archon-node
#
# 5. Create systemd service at /etc/systemd/system/archon-node.service:
#      [Unit]
#      Description=Archon Node Server
#      After=network.target docker.service
#
#      [Service]
#      Type=simple
#      User=root
#      ExecStart=/usr/local/bin/archon-node --config /etc/archon/node-config.toml
#      Restart=on-failure
#      RestartSec=5s
#
#      [Install]
#      WantedBy=multi-user.target
#
# 6. Start the service:
#      sudo systemctl daemon-reload
#      sudo systemctl enable archon-node
#      sudo systemctl start archon-node
#      sudo systemctl status archon-node
#
# 7. Verify it's running:
#      curl http://localhost:8080/health
#
# 8. Open firewall (if needed):
#      sudo ufw allow 8080/tcp
#      sudo ufw allow 80/tcp   # For Let's Encrypt
#      sudo ufw allow 443/tcp  # For HTTPS
#
# ==============================================================================
# Node Details
# ==============================================================================
# Node Name:     ` + n.Name + `
# API Endpoint:  ` + n.APIEndpoint + `
# IP Address:    ` + n.IPAddress.String() + `
# API Key:       ` + n.APIKey + `
#
# Keep the API key secure! It's used for authentication between Archon TUI
# and this node server.
# ==============================================================================
`
}
