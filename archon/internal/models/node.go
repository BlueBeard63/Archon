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
