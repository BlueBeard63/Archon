package stages

import (
	"context"
	"fmt"
	"log"

	"github.com/BlueBeard63/archon-node/internal/pipeline"
	"github.com/BlueBeard63/archon-node/internal/proxy"
)

// ProxyStage configures the reverse proxy
type ProxyStage struct {
	name         string
	proxyManager proxy.ProxyManager
}

// NewProxyStage creates a new proxy configuration stage
func NewProxyStage(proxyManager proxy.ProxyManager) *ProxyStage {
	return &ProxyStage{
		name:         "proxy-config",
		proxyManager: proxyManager,
	}
}

// Name returns the stage name
func (s *ProxyStage) Name() string {
	return s.name
}

// Execute configures the reverse proxy for the deployed site
func (s *ProxyStage) Execute(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	log.Printf("[PROXY] Configuring proxy for %d domains", len(req.DomainMappings))

	// Configure proxy with SSL paths (may be empty if SSL not enabled)
	if err := s.proxyManager.Configure(ctx, req, state.CertPath, state.KeyPath); err != nil {
		return fmt.Errorf("failed to configure proxy: %w", err)
	}

	// Reload proxy to apply configuration
	if err := s.proxyManager.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload proxy: %w", err)
	}

	log.Printf("[PROXY] Configuration complete")
	return nil
}

// Rollback removes the proxy configuration
func (s *ProxyStage) Rollback(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	if len(req.DomainMappings) == 0 {
		return nil
	}

	// Remove proxy config for primary domain
	primaryDomain := req.DomainMappings[0].Domain
	log.Printf("[ROLLBACK] Removing proxy config for: %s", primaryDomain)

	if err := s.proxyManager.Remove(ctx, req.ID, primaryDomain); err != nil {
		log.Printf("[ROLLBACK] Warning: failed to remove proxy config: %v", err)
	}

	// Reload proxy (ignore errors during rollback)
	s.proxyManager.Reload(ctx)

	return nil
}
