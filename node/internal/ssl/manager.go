package ssl

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/config"
)

type Manager struct {
	mode      config.SSLMode
	certDir   string
	email     string
	proxyType config.ProxyType
}

func NewManager(sslCfg *config.SSLConfig, proxyType config.ProxyType) *Manager {
	return &Manager{
		mode:      sslCfg.Mode,
		certDir:   sslCfg.CertDir,
		email:     sslCfg.Email,
		proxyType: proxyType,
	}
}

// EnsureCertificate ensures SSL certificate exists for a domain
// Returns paths to cert and key files
// If email is empty for Let's Encrypt mode, falls back to manager's default email
func (m *Manager) EnsureCertificate(ctx context.Context, siteID uuid.UUID, domain string, certB64, keyB64, email string) (string, string, error) {
	switch m.mode {
	case config.SSLModeManual:
		return m.handleManualCert(siteID, domain, certB64, keyB64)
	case config.SSLModeLetsEncrypt:
		// Use provided email, or fall back to manager's default
		emailToUse := email
		if emailToUse == "" {
			emailToUse = m.email
		}
		return m.handleLetsEncrypt(ctx, domain, emailToUse)
	case config.SSLModeTraefikAuto:
		// Traefik handles certificates automatically
		return "", "", nil
	default:
		return "", "", fmt.Errorf("unsupported SSL mode: %s", m.mode)
	}
}

// handleManualCert handles user-provided certificates
func (m *Manager) handleManualCert(siteID uuid.UUID, domain string, certB64, keyB64 string) (string, string, error) {
	if certB64 == "" || keyB64 == "" {
		return "", "", fmt.Errorf("manual SSL mode requires certificate and key")
	}

	// Decode base64
	certData, err := base64.StdEncoding.DecodeString(certB64)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode certificate: %w", err)
	}

	keyData, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode key: %w", err)
	}

	// Create site-specific cert directory
	siteCertDir := filepath.Join(m.certDir, siteID.String())
	if err := os.MkdirAll(siteCertDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create cert directory: %w", err)
	}

	// Write certificate
	certPath := filepath.Join(siteCertDir, "cert.pem")
	if err := os.WriteFile(certPath, certData, 0644); err != nil {
		return "", "", fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write key
	keyPath := filepath.Join(siteCertDir, "key.pem")
	if err := os.WriteFile(keyPath, keyData, 0600); err != nil {
		return "", "", fmt.Errorf("failed to write key: %w", err)
	}

	return certPath, keyPath, nil
}

// handleLetsEncrypt obtains a certificate from Let's Encrypt using certbot
func (m *Manager) handleLetsEncrypt(ctx context.Context, domain string, email string) (string, string, error) {
	if email == "" {
		return "", "", fmt.Errorf("email is required for Let's Encrypt")
	}

	// Check if certbot is installed
	if _, err := exec.LookPath("certbot"); err != nil {
		return "", "", fmt.Errorf("certbot is not installed: %w", err)
	}

	// Build certbot command based on proxy type
	var cmd *exec.Cmd
	switch m.proxyType {
	case config.ProxyTypeNginx:
		// Use nginx plugin - it works with running nginx without modifying configs
		cmd = exec.CommandContext(ctx,
			"certbot", "certonly",
			"--nginx",
			"--non-interactive",
			"--agree-tos",
			"--force-renewal",
			"--email", email,
			"-d", domain,
		)
	case config.ProxyTypeApache:
		// Use apache plugin now that ConfigureForValidation has created the port 80 vhost
		cmd = exec.CommandContext(ctx,
			"certbot", "certonly",
			"--apache",
			"--non-interactive",
			"--agree-tos",
			"--force-renewal",
			"--email", email,
			"-d", domain,
		)
	case config.ProxyTypeTraefik:
		// Traefik should handle this automatically, but if we get here, try webroot
		return "", "", fmt.Errorf("Traefik should use SSLModeTraefikAuto, not Let's Encrypt mode")
	default:
		// Fallback to standalone mode (will fail if port 80 is in use)
		cmd = exec.CommandContext(ctx,
			"certbot", "certonly",
			"--standalone",
			"--non-interactive",
			"--agree-tos",
			"--force-renewal",
			"--email", email,
			"-d", domain,
			"--http-01-port", "80",
		)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("certbot failed: %s", string(output))
	}

	// Certbot stores certificates in /etc/letsencrypt/live/<domain>/
	certPath := filepath.Join("/etc/letsencrypt/live", domain, "fullchain.pem")
	keyPath := filepath.Join("/etc/letsencrypt/live", domain, "privkey.pem")

	// Verify files exist
	if _, err := os.Stat(certPath); err != nil {
		return "", "", fmt.Errorf("certificate not found at %s: %w", certPath, err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		return "", "", fmt.Errorf("key not found at %s: %w", keyPath, err)
	}

	return certPath, keyPath, nil
}

// RenewCertificates renews all Let's Encrypt certificates
func (m *Manager) RenewCertificates(ctx context.Context) error {
	if m.mode != config.SSLModeLetsEncrypt {
		return nil // Nothing to renew for other modes
	}

	// Check if certbot is installed
	if _, err := exec.LookPath("certbot"); err != nil {
		return fmt.Errorf("certbot is not installed: %w", err)
	}

	// Run certbot renew
	cmd := exec.CommandContext(ctx, "certbot", "renew", "--non-interactive")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("certbot renew failed: %s", string(output))
	}

	return nil
}

// RemoveCertificate removes certificate files for a site
func (m *Manager) RemoveCertificate(siteID uuid.UUID) error {
	siteCertDir := filepath.Join(m.certDir, siteID.String())

	if err := os.RemoveAll(siteCertDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove certificate directory: %w", err)
	}

	return nil
}
