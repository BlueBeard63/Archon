package ssl

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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

// EnsureCertificate ensures SSL certificate exists for one or more domains
// Returns paths to cert and key files
// If email is empty for Let's Encrypt mode, falls back to manager's default email
// For multiple domains, uses a Subject Alternative Name (SAN) certificate
func (m *Manager) EnsureCertificate(ctx context.Context, siteID uuid.UUID, domain string, certB64, keyB64, email string) (string, string, error) {
	return m.EnsureCertificateMulti(ctx, siteID, []string{domain}, certB64, keyB64, email)
}

// EnsureCertificateMulti ensures SSL certificate exists for one or more domains
// Returns paths to cert and key files
// If email is empty for Let's Encrypt mode, falls back to manager's default email
// For multiple domains, uses a Subject Alternative Name (SAN) certificate
func (m *Manager) EnsureCertificateMulti(ctx context.Context, siteID uuid.UUID, domains []string, certB64, keyB64, email string) (string, string, error) {
	if len(domains) == 0 {
		return "", "", fmt.Errorf("at least one domain is required")
	}

	switch m.mode {
	case config.SSLModeManual:
		return m.handleManualCert(siteID, domains[0], certB64, keyB64)
	case config.SSLModeLetsEncrypt:
		// Use provided email, or fall back to manager's default
		emailToUse := email
		if emailToUse == "" {
			emailToUse = m.email
		}
		return m.handleLetsEncryptMulti(ctx, domains, emailToUse)
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

// handleLetsEncryptMulti obtains a certificate from Let's Encrypt for multiple domains using SAN (Subject Alternative Name)
// Uses the primary domain as the certificate name, with other domains as SANs
func (m *Manager) handleLetsEncryptMulti(ctx context.Context, domains []string, email string) (string, string, error) {
	if email == "" {
		return "", "", fmt.Errorf("email is required for Let's Encrypt")
	}

	if len(domains) == 0 {
		return "", "", fmt.Errorf("at least one domain is required")
	}

	// Check if certbot is installed
	if _, err := exec.LookPath("certbot"); err != nil {
		return "", "", fmt.Errorf("certbot is not installed: %w", err)
	}

	// Use FindCertificates to check for existing cert (handles -0001 suffix patterns)
	primaryDomain := domains[0]
	if certPath, keyPath, err := FindCertificates(primaryDomain); err == nil {
		// Certificate exists, check if it covers all requested domains
		coveredDomains, parseErr := ExtractSANsFromCert(certPath)
		if parseErr == nil {
			allCovered := true
			for _, d := range domains {
				found := false
				for _, covered := range coveredDomains {
					if covered == d {
						found = true
						break
					}
				}
				if !found {
					allCovered = false
					break
				}
			}
			if allCovered {
				// Existing cert covers all domains, return it
				return certPath, keyPath, nil
			}
			// Cert exists but doesn't cover all domains - proceed to request new cert
			// Certbot will handle renewal/expansion automatically
		}
	}

	// No valid certificate found, proceed to request one from Let's Encrypt

	// Build certbot command based on proxy type
	// Use the first domain as the primary domain (certificate will be stored under this name)
	cmdArgs := []string{"certbot", "certonly", "--non-interactive", "--agree-tos", "--email", email}

	// Add proxy-specific arguments
	switch m.proxyType {
	case config.ProxyTypeNginx:
		// Use webroot method with certbot directory
		cmdArgs = append(cmdArgs, "--webroot", "-w", "/var/www/certbot")
	case config.ProxyTypeApache:
		cmdArgs = append(cmdArgs, "--apache")
	case config.ProxyTypeTraefik:
		return "", "", fmt.Errorf("Traefik should use SSLModeTraefikAuto, not Let's Encrypt mode")
	default:
		// Fallback to standalone mode
		cmdArgs = append(cmdArgs, "--standalone", "--http-01-port", "80")
	}

	// Add all domains (certbot will create a SAN certificate)
	for _, domain := range domains {
		cmdArgs = append(cmdArgs, "-d", domain)
	}

	cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("certbot failed: %s", string(output))
	}

	// Find the actual certificate location (certbot may create numbered directories like domain-0001)
	certPath, keyPath, err := FindCertificates(primaryDomain)
	if err != nil {
		return "", "", err
	}

	return certPath, keyPath, nil
}

// FindCertificates searches for certificate files in /etc/letsencrypt/live/ for the given domain
// Returns the paths to the certificate and key files, handling numbered directories like domain-0001
func FindCertificates(domain string) (string, string, error) {
	liveDir := "/etc/letsencrypt/live"
	entries, err := os.ReadDir(liveDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to read letsencrypt live directory: %w", err)
	}

	// First try exact match
	certPath := filepath.Join(liveDir, domain, "fullchain.pem")
	keyPath := filepath.Join(liveDir, domain, "privkey.pem")
	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			return certPath, keyPath, nil
		}
	}

	// If exact match not found, search for domain-* patterns (numbered versions)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirName := entry.Name()
		// Match domain or domain-* (e.g., api.jack-morrison.dev-0001)
		if dirName == domain || (len(dirName) > len(domain) && dirName[:len(domain)] == domain && dirName[len(domain)] == '-') {
			certPath := filepath.Join(liveDir, dirName, "fullchain.pem")
			keyPath := filepath.Join(liveDir, dirName, "privkey.pem")
			if _, err := os.Stat(certPath); err == nil {
				if _, err := os.Stat(keyPath); err == nil {
					return certPath, keyPath, nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("certificate not found for domain %s in %s", domain, liveDir)
}

// ExtractSANsFromCert reads a PEM certificate file and returns all Subject Alternative Names (DNS names)
func ExtractSANsFromCert(certPath string) ([]string, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse X.509 certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Return DNS names (SANs)
	return cert.DNSNames, nil
}

// FindCertificateForDomains checks if a valid certificate exists that covers all the given domains
// Returns (certPath, keyPath, coveredDomains, nil) if found, or ("", "", nil, err) if not
func FindCertificateForDomains(domains []string) (string, string, []string, error) {
	if len(domains) == 0 {
		return "", "", nil, fmt.Errorf("at least one domain is required")
	}

	// For SAN certs, the cert is stored under the primary domain name
	primaryDomain := domains[0]

	// Use FindCertificates to handle -0001 suffix patterns
	certPath, keyPath, err := FindCertificates(primaryDomain)
	if err != nil {
		return "", "", nil, err
	}

	// Parse the certificate to get SANs
	coveredDomains, err := ExtractSANsFromCert(certPath)
	if err != nil {
		return "", "", nil, err
	}

	return certPath, keyPath, coveredDomains, nil
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
