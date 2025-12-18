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
func (t *TraefikManager) Configure(ctx context.Context, site *models.DeployRequest, certPath, keyPath string) error {
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

// GenerateTraefikLabels generates Traefik labels for a site
func GenerateTraefikLabels(site *models.DeployRequest) map[string]string {
	labels := map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.rule", site.Name):                      fmt.Sprintf("Host(`%s`)", site.Domain),
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", site.Name):               "web",
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", site.Name): fmt.Sprintf("%d", site.Port),
	}

	if site.SSLEnabled {
		// Add HTTPS configuration
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.rule", site.Name)] = fmt.Sprintf("Host(`%s`)", site.Domain)
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.entrypoints", site.Name)] = "websecure"
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.tls", site.Name)] = "true"
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.tls.certresolver", site.Name)] = "letsencrypt"

		// Add redirect middleware from HTTP to HTTPS
		labels[fmt.Sprintf("traefik.http.routers.%s.middlewares", site.Name)] = "redirect-to-https"
		labels["traefik.http.middlewares.redirect-to-https.redirectscheme.scheme"] = "https"
		labels["traefik.http.middlewares.redirect-to-https.redirectscheme.permanent"] = "true"
	}

	return labels
}
