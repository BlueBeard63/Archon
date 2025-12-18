package proxy

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/config"
	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/BlueBeard63/archon-node/internal/ssl"
)

type NginxManager struct {
	configDir     string
	reloadCommand string
	sslMode       config.SSLMode
}

func NewNginxManager(cfg *config.ProxyConfig, sslCfg *config.SSLConfig) *NginxManager {
	return &NginxManager{
		configDir:     cfg.ConfigDir,
		reloadCommand: cfg.ReloadCommand,
		sslMode:       sslCfg.Mode,
	}
}

const nginxConfigTemplate = `{{ range $domain, $mapping := .Domains -}}
server {
    listen 80;
    server_name {{ $domain }};

    {{- if $.SSLEnabled }}
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name {{ $domain }};

    # SSL Configuration
    ssl_certificate {{ $mapping.CertPath }};
    ssl_certificate_key {{ $mapping.KeyPath }};
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    {{- end }}

    # Proxy configuration
    location / {
        proxy_pass http://127.0.0.1:{{ $mapping.Port }};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }

    # Access and error logs
    access_log /var/log/nginx/{{ $domain }}_access.log;
    error_log /var/log/nginx/{{ $domain }}_error.log;
}
{{ end }}
`

type DomainMappingPair struct {
	Port     int
	CertPath string
	KeyPath  string
}

type nginxTemplateData struct {
	Domains    map[string]DomainMappingPair
	SSLEnabled bool
}

// ConfigureForValidation configures Nginx with temporary HTTP vhosts for Let's Encrypt validation
func (n *NginxManager) ConfigureForValidation(ctx context.Context, site *models.DeployRequest) error {
	// Skip if not using Let's Encrypt mode
	if n.sslMode != "letsencrypt" {
		return nil
	}

	// Skip if SSL is not enabled
	if !site.SSLEnabled {
		return nil
	}

	// Get domain-port mappings
	domainMappings := getDomainMappings(site)
	if len(domainMappings) == 0 {
		return nil
	}

	// Create validation server block for port 80
	const validationTemplate = `{{ range $domain, $mapping := .Domains -}}
server {
    listen 80;
    server_name {{ $domain }};

    # Allow certbot validation
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    # Proxy other requests to the backend
    location / {
        proxy_pass http://127.0.0.1:{{ $mapping.Port }};
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
{{ end }}`

	// Create certbot webroot directory
	webrootPath := "/var/www/certbot"
	if err := os.MkdirAll(webrootPath, 0755); err != nil {
		return fmt.Errorf("failed to create certbot webroot directory: %w", err)
	}

	// Convert to template data
	domainsMap := make(map[string]DomainMappingPair)
	for _, mapping := range domainMappings {
		domainsMap[mapping.Domain] = DomainMappingPair{
			Port:     mapping.Port,
			CertPath: "", // Not used in validation template
			KeyPath:  "",
		}
	}

	// Prepare template data
	data := nginxTemplateData{
		Domains:    domainsMap,
		SSLEnabled: false, // Validation is HTTP only
	}

	// Parse template
	tmpl, err := template.New("nginx-validation").Parse(validationTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse nginx validation template: %w", err)
	}

	// Use primary domain for config filename
	// Use primary domain for config filename
	primaryDomain := domainMappings[0].Domain

	// ADD THIS: Remove existing SSL-enabled config file before creating validation config
	// This prevents nginx from trying to load configs that reference non-existent certificates
	existingConfigPath := filepath.Join(n.configDir, fmt.Sprintf("%s.conf", primaryDomain))
	if err := os.Remove(existingConfigPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing nginx config: %w", err)
	}

	// Create validation config file
	configPath := filepath.Join(n.configDir, fmt.Sprintf("%s-validation.conf", primaryDomain))
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create nginx validation config file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute nginx validation template: %w", err)
	}

	// Skip nginx -t test during validation since old config files might reference non-existent certificates
	// Just reload nginx with the validation config
	reloadCmd := exec.CommandContext(ctx, "sh", "-c", n.reloadCommand)
	if output, err := reloadCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nginx reload failed for validation: %s", string(output))
	}

	return nil
}

func (n *NginxManager) Configure(ctx context.Context, site *models.DeployRequest, certPath, keyPath string) error {
	// Get domain-port mappings
	domainMappings := getDomainMappings(site)

	// Use passed cert paths if provided, otherwise find them
	primaryDomain := domainMappings[0].Domain
	var primaryCertPath, primaryKeyPath string

	if certPath != "" && keyPath != "" {
		// Use the cert paths passed from the SSL manager
		primaryCertPath = certPath
		primaryKeyPath = keyPath
	} else if site.SSLEnabled {
		// Fallback: try to find certificates using FindCertificates (handles -0001 suffixes)
		if cert, key, err := ssl.FindCertificates(primaryDomain); err == nil {
			primaryCertPath = cert
			primaryKeyPath = key
		} else {
			return fmt.Errorf("SSL enabled but certificate not found for %s: %w", primaryDomain, err)
		}
	}

	log.Printf("nginx cert and key: %s, %s", primaryCertPath, primaryKeyPath)

	// Convert to template data with domain-specific cert paths using a map
	// All domains use the same certificate (SAN certificate)
	domainsMap := make(map[string]DomainMappingPair)
	for _, mapping := range domainMappings {
		domainsMap[mapping.Domain] = DomainMappingPair{
			Port:     mapping.Port,
			CertPath: primaryCertPath,
			KeyPath:  primaryKeyPath,
		}
	}

	// Prepare template data
	data := nginxTemplateData{
		Domains:    domainsMap,
		SSLEnabled: site.SSLEnabled,
	}

	// Parse template
	tmpl, err := template.New("nginx").Parse(nginxConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse nginx template: %w", err)
	}

	// Create config file (use primary domain for filename)
	configPath := filepath.Join(n.configDir, fmt.Sprintf("%s.conf", primaryDomain))
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create nginx config file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute nginx template: %w", err)
	}

	// Remove validation config file if it exists
	validationConfigPath := filepath.Join(n.configDir, fmt.Sprintf("%s-validation.conf", primaryDomain))
	if err := os.Remove(validationConfigPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove nginx validation config: %w", err)
	}

	// Test nginx configuration
	cmd := exec.CommandContext(ctx, "nginx", "-t")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Log the actual error for debugging
		return fmt.Errorf("nginx config test failed (exit code: %v): %s", err, string(output))
	}

	return nil
}

func (n *NginxManager) Remove(ctx context.Context, siteID uuid.UUID, domain string) error {
	configPath := filepath.Join(n.configDir, fmt.Sprintf("%s.conf", domain))

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove nginx config: %w", err)
	}

	return nil
}

func (n *NginxManager) Reload(ctx context.Context) error {
	// Test nginx configuration before reloading
	testCmd := exec.CommandContext(ctx, "nginx", "-t")
	testOutput, err := testCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx config test failed before reload (exit code: %v): %s", err, string(testOutput))
	}

	// Execute reload command only if test passed
	cmd := exec.CommandContext(ctx, "sh", "-c", n.reloadCommand)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("nginx reload failed: %s", string(output))
	}

	return nil
}

func (n *NginxManager) GetInfo(ctx context.Context) (*models.TraefikInfo, error) {
	// Get nginx version
	cmd := exec.CommandContext(ctx, "nginx", "-v")
	output, _ := cmd.CombinedOutput()

	// Nginx outputs version to stderr
	version := string(output)

	// Count config files as "routers"
	files, err := os.ReadDir(n.configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read nginx config dir: %w", err)
	}

	return &models.TraefikInfo{
		Version:       version,
		RoutersCount:  len(files),
		ServicesCount: len(files),
	}, nil
}

// getDomainMappings extracts domain-port mappings from a DeployRequest
func getDomainMappings(site *models.DeployRequest) []models.DomainMapping {
	return site.DomainMappings
}
