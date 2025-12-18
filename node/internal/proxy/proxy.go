package proxy

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/config"
	"github.com/BlueBeard63/archon-node/internal/models"
)

// ProxyManager defines the interface for reverse proxy management
type ProxyManager interface {
	// ConfigureForValidation sets up minimal proxy configuration for Let's Encrypt validation
	// This is called before certificate generation for Let's Encrypt mode
	ConfigureForValidation(ctx context.Context, site *models.DeployRequest) error

	// Configure sets up the proxy configuration for a site
	Configure(ctx context.Context, site *models.DeployRequest, certPath, keyPath string) error

	// Remove removes the proxy configuration for a site
	Remove(ctx context.Context, siteID uuid.UUID, domain string) error

	// Reload reloads the proxy configuration
	Reload(ctx context.Context) error

	// GetInfo returns information about the proxy (for health checks)
	GetInfo(ctx context.Context) (*models.TraefikInfo, error)
}

// NewProxyManager creates a new proxy manager based on the configuration
func NewProxyManager(cfg *config.ProxyConfig, sslCfg *config.SSLConfig) (ProxyManager, error) {
	switch cfg.Type {
	case config.ProxyTypeNginx:
		return NewNginxManager(cfg, sslCfg), nil
	case config.ProxyTypeApache:
		return NewApacheManager(cfg, sslCfg), nil
	case config.ProxyTypeTraefik:
		return NewTraefikManager(cfg, sslCfg), nil
	default:
		return nil, fmt.Errorf("unsupported proxy type: %s", cfg.Type)
	}
}
