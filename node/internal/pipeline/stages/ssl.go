package stages

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/BlueBeard63/archon-node/internal/pipeline"
	"github.com/BlueBeard63/archon-node/internal/proxy"
	"github.com/BlueBeard63/archon-node/internal/ssl"
)

// SSLStage handles SSL certificate acquisition
type SSLStage struct {
	name         string
	proxyManager proxy.ProxyManager
	sslManager   *ssl.Manager
}

// NewSSLStage creates a new SSL stage
func NewSSLStage(proxyManager proxy.ProxyManager, sslManager *ssl.Manager) *SSLStage {
	return &SSLStage{
		name:         "ssl-setup",
		proxyManager: proxyManager,
		sslManager:   sslManager,
	}
}

// Name returns the stage name
func (s *SSLStage) Name() string {
	return s.name
}

// Execute sets up SSL certificates if enabled
func (s *SSLStage) Execute(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	// Skip if SSL not enabled
	if !req.SSLEnabled {
		return nil
	}

	// Collect all domains
	domains := make([]string, 0, len(req.DomainMappings))
	for _, mapping := range req.DomainMappings {
		domains = append(domains, mapping.Domain)
	}

	log.Printf("[SSL] Setting up certificate for %d domains: %v", len(domains), domains)

	// Configure proxy for Let's Encrypt validation
	if err := s.proxyManager.ConfigureForValidation(ctx, req); err != nil {
		return fmt.Errorf("failed to configure proxy for validation: %w", err)
	}

	// Reload proxy
	if err := s.proxyManager.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload proxy for validation: %w", err)
	}

	// Wait for DNS propagation
	for _, domain := range domains {
		if err := s.waitForDNS(domain, 60*time.Second); err != nil {
			return fmt.Errorf("DNS propagation timeout for %s: %w", domain, err)
		}
	}

	// Request certificate
	certPath, keyPath, err := s.sslManager.EnsureCertificateMulti(ctx, req.ID, domains, req.SSLCert, req.SSLKey, req.SSLEmail)
	if err != nil {
		return fmt.Errorf("failed to obtain SSL certificate: %w", err)
	}

	// Store paths in state for proxy stage
	state.CertPath = certPath
	state.KeyPath = keyPath

	log.Printf("[SSL] Certificate obtained successfully")
	return nil
}

// Rollback removes any partial SSL setup
func (s *SSLStage) Rollback(ctx context.Context, state *pipeline.DeploymentState) error {
	if state.CertPath != "" || state.KeyPath != "" {
		// Attempt to remove certificate (don't fail if it doesn't exist)
		if err := s.sslManager.RemoveCertificate(state.SiteID()); err != nil {
			log.Printf("[SSL ROLLBACK] Warning: failed to remove certificate: %v", err)
		}
	}
	return nil
}

// waitForDNS waits for DNS propagation with timeout
func (s *SSLStage) waitForDNS(domain string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		_, err := net.LookupHost(domain)
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for DNS resolution")
}
