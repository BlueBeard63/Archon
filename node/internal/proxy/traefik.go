package proxy

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/config"
	"github.com/BlueBeard63/archon-node/internal/models"
)

type TraefikManager struct {
	sslMode config.SSLMode
}

func NewTraefikManager(cfg *config.ProxyConfig, sslCfg *config.SSLConfig) *TraefikManager {
	return &TraefikManager{
		sslMode: sslCfg.Mode,
	}
}

// ConfigureForValidation is a no-op for Traefik as it handles SSL automatically
func (t *TraefikManager) ConfigureForValidation(ctx context.Context, site *models.DeployRequest) error {
	// Traefik handles Let's Encrypt automatically
	return nil
}

// Configure for Traefik is a no-op because Traefik uses Docker labels
// The labels are already set on the container when it's deployed
func (t *TraefikManager) Configure(ctx context.Context, site *models.DeployRequest, _, _ string) error {
	// Traefik configuration is done via Docker labels
	// which are already set in the DeploySite request
	return nil
}

func (t *TraefikManager) Remove(ctx context.Context, siteID uuid.UUID, domain string) error {
	// Traefik automatically removes routes when containers are removed
	return nil
}

func (t *TraefikManager) Reload(ctx context.Context) error {
	// Traefik automatically reloads when container labels change
	return nil
}

func (t *TraefikManager) GetInfo(ctx context.Context) (*models.TraefikInfo, error) {
	// This would typically query Traefik's API
	// For now, return placeholder info
	return &models.TraefikInfo{
		Version:       "2.x (auto-configured)",
		RoutersCount:  0,
		ServicesCount: 0,
	}, nil
}

// GenerateTraefikLabels generates Traefik labels for a site supporting multiple domains
// Creates routers and services for each domain-port mapping
func GenerateTraefikLabels(site *models.DeployRequest) map[string]string {
	labels := map[string]string{
		"traefik.enable": "true",
	}

	// Get domain-port mappings
	domainMappings := getDomainMappings(site)

	// Process each domain-port mapping
	for i, mapping := range domainMappings {
		// Create a unique identifier for this router (use index to make it unique)
		// Format: sitename-0, sitename-1, etc. or for first domain: sitename
		var routerName string
		if i == 0 {
			routerName = site.Name
		} else {
			routerName = fmt.Sprintf("%s-%d", site.Name, i)
		}

		// HTTP router
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] = fmt.Sprintf("Host(`%s`)", mapping.Domain)
		labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", routerName)] = "web"

		// Service for this router
		labels[fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", routerName)] = fmt.Sprintf("%d", mapping.Port)

		// HTTPS configuration if SSL is enabled
		if site.SSLEnabled {
			secureRouterName := fmt.Sprintf("%s-secure", routerName)
			labels[fmt.Sprintf("traefik.http.routers.%s.rule", secureRouterName)] = fmt.Sprintf("Host(`%s`)", mapping.Domain)
			labels[fmt.Sprintf("traefik.http.routers.%s.entrypoints", secureRouterName)] = "websecure"
			labels[fmt.Sprintf("traefik.http.routers.%s.tls", secureRouterName)] = "true"
			labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", secureRouterName)] = "letsencrypt"

			// Add redirect middleware from HTTP to HTTPS for this domain
			labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", routerName)] = fmt.Sprintf("redirect-%s", routerName)
			labels[fmt.Sprintf("traefik.http.middlewares.redirect-%s.redirectscheme.scheme", routerName)] = "https"
			labels[fmt.Sprintf("traefik.http.middlewares.redirect-%s.redirectscheme.permanent", routerName)] = "true"
		}
	}

	return labels
}
