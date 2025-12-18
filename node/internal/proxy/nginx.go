package proxy

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/config"
	"github.com/BlueBeard63/archon-node/internal/models"
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

const nginxConfigTemplate = `server {
    listen 80;
    server_name {{ .Domain }};

    {{- if .SSLEnabled }}
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name {{ .Domain }};

    # SSL Configuration
    ssl_certificate {{ .CertPath }};
    ssl_certificate_key {{ .KeyPath }};
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
        proxy_pass http://127.0.0.1:{{ .Port }};
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
    access_log /var/log/nginx/{{ .Domain }}_access.log;
    error_log /var/log/nginx/{{ .Domain }}_error.log;
}
`

type nginxTemplateData struct {
	Domain     string
	Port       int
	SSLEnabled bool
	CertPath   string
	KeyPath    string
}

// ConfigureForValidation is a no-op for Nginx as it doesn't require pre-configuration for Let's Encrypt
func (n *NginxManager) ConfigureForValidation(ctx context.Context, site *models.DeployRequest) error {
	// Nginx plugin doesn't require pre-configuration
	return nil
}

func (n *NginxManager) Configure(ctx context.Context, site *models.DeployRequest, certPath, keyPath string) error {
	// Prepare template data
	data := nginxTemplateData{
		Domain:     site.Domain,
		Port:       site.Port,
		SSLEnabled: site.SSLEnabled,
		CertPath:   certPath,
		KeyPath:    keyPath,
	}

	// Parse template
	tmpl, err := template.New("nginx").Parse(nginxConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse nginx template: %w", err)
	}

	// Create config file
	configPath := filepath.Join(n.configDir, fmt.Sprintf("%s.conf", site.Domain))
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create nginx config file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute nginx template: %w", err)
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
