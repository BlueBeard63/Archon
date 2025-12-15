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

type Node struct {
	ID              uuid.UUID    `json:"id" toml:"id"`
	Name            string       `json:"name" toml:"name"`
	APIEndpoint     string       `json:"api_endpoint" toml:"api_endpoint"`
	APIKey          string       `json:"api_key" toml:"api_key"`
	IPAddress       net.IP       `json:"ip_address" toml:"ip_address"`
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
func NewNode(name, apiEndpoint, apiKey string, ipAddress net.IP) *Node {
	return &Node{
		ID:          uuid.New(),
		Name:        name,
		APIEndpoint: apiEndpoint,
		APIKey:      apiKey,
		IPAddress:   ipAddress,
		Status:      NodeStatusUnknown,
	}
}

// GenerateNodeConfigTOML generates the TOML configuration file content for this node
func (n *Node) GenerateNodeConfigTOML() string {
	return `# Archon Node Server Configuration
# Place this file at: /etc/archon/node-config.toml
# Generated for node: ` + n.Name + `

[server]
host = "0.0.0.0"
port = 8080
# IMPORTANT: This API key must match the one in your Archon TUI config
api_key = "` + n.APIKey + `"
data_dir = "/var/lib/archon"

# ==============================================================================
# Proxy Configuration - Choose ONE of the following proxy types
# ==============================================================================

# --- OPTION 1: Nginx with Let's Encrypt (Recommended for most users) ---
[proxy]
type = "nginx"
config_dir = "/etc/nginx/sites-enabled"
reload_command = "nginx -s reload"

# --- OPTION 2: Apache with Let's Encrypt ---
# [proxy]
# type = "apache"
# config_dir = "/etc/apache2/sites-enabled"
# reload_command = "apache2ctl graceful"

# --- OPTION 3: Traefik (Best for auto-SSL and dynamic routing) ---
# [proxy]
# type = "traefik"
# config_dir = ""
# reload_command = ""

# ==============================================================================
# Docker Configuration
# ==============================================================================

[docker]
host = "unix:///var/run/docker.sock"
network = "archon-net"

# ==============================================================================
# SSL Configuration - Choose mode based on your proxy type
# ==============================================================================

# --- For Nginx/Apache: Let's Encrypt (Automatic SSL) ---
[ssl]
mode = "letsencrypt"
cert_dir = "/etc/archon/ssl"
email = "admin@example.com"  # CHANGE THIS to your email

[letsencrypt]
enabled = true
email = "admin@example.com"  # CHANGE THIS to your email
staging_mode = false  # Set to true for testing

# --- For Nginx/Apache: Manual SSL (Upload your own certificates) ---
# [ssl]
# mode = "manual"
# cert_dir = "/etc/archon/ssl"
# email = ""

# --- For Traefik: Auto SSL (Traefik handles everything) ---
# [ssl]
# mode = "traefik-auto"
# cert_dir = ""
# email = "admin@example.com"  # CHANGE THIS to your email

# ==============================================================================
# Installation Instructions
# ==============================================================================
# 1. Install prerequisites on your server (` + n.IPAddress.String() + `):
#    Ubuntu/Debian:
#      sudo apt update
#      sudo apt install docker.io nginx certbot
#      sudo systemctl enable docker
#      sudo systemctl start docker
#
#    For Traefik instead of nginx:
#      sudo apt update
#      sudo apt install docker.io
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
