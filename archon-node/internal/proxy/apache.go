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

type ApacheManager struct {
	configDir     string
	reloadCommand string
	sslMode       config.SSLMode
}

func NewApacheManager(cfg *config.ProxyConfig, sslCfg *config.SSLConfig) *ApacheManager {
	return &ApacheManager{
		configDir:     cfg.ConfigDir,
		reloadCommand: cfg.ReloadCommand,
		sslMode:       sslCfg.Mode,
	}
}

const apacheConfigTemplate = `<VirtualHost *:80>
    ServerName {{ .Domain }}

    {{- if .SSLEnabled }}
    # Redirect HTTP to HTTPS
    RewriteEngine On
    RewriteCond %{HTTPS} off
    RewriteRule ^ https://%{HTTP_HOST}%{REQUEST_URI} [R=301,L]
</VirtualHost>

<VirtualHost *:443>
    ServerName {{ .Domain }}

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile {{ .CertPath }}
    SSLCertificateKeyFile {{ .KeyPath }}
    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLCipherSuite HIGH:!aNULL:!MD5
    SSLHonorCipherOrder on

    # Security headers
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
    Header always set X-Frame-Options "SAMEORIGIN"
    Header always set X-Content-Type-Options "nosniff"
    Header always set X-XSS-Protection "1; mode=block"
    {{- end }}

    # Proxy configuration
    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:{{ .Port }}/
    ProxyPassReverse / http://127.0.0.1:{{ .Port }}/

    # WebSocket support
    RewriteEngine On
    RewriteCond %{HTTP:Upgrade} =websocket [NC]
    RewriteRule /(.*)           ws://127.0.0.1:{{ .Port }}/$1 [P,L]

    # Request headers
    RequestHeader set X-Forwarded-Proto "https"
    RequestHeader set X-Forwarded-Port "443"

    # Access and error logs
    ErrorLog ${APACHE_LOG_DIR}/{{ .Domain }}_error.log
    CustomLog ${APACHE_LOG_DIR}/{{ .Domain }}_access.log combined
</VirtualHost>
`

type apacheTemplateData struct {
	Domain     string
	Port       int
	SSLEnabled bool
	CertPath   string
	KeyPath    string
}

func (a *ApacheManager) Configure(ctx context.Context, site *models.DeployRequest, certPath, keyPath string) error {
	// Prepare template data
	data := apacheTemplateData{
		Domain:     site.Domain,
		Port:       site.Port,
		SSLEnabled: site.SSLEnabled,
		CertPath:   certPath,
		KeyPath:    keyPath,
	}

	// Parse template
	tmpl, err := template.New("apache").Parse(apacheConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse apache template: %w", err)
	}

	// Create config file
	configPath := filepath.Join(a.configDir, fmt.Sprintf("%s.conf", site.Domain))
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create apache config file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute apache template: %w", err)
	}

	// Test apache configuration
	cmd := exec.CommandContext(ctx, "apache2ctl", "configtest")
	if output, err := cmd.CombinedOutput(); err != nil {
		// Apache configtest returns error even on success sometimes, check output
		if !contains(string(output), "Syntax OK") {
			return fmt.Errorf("apache config test failed: %s", string(output))
		}
	}

	return nil
}

func (a *ApacheManager) Remove(ctx context.Context, siteID uuid.UUID, domain string) error {
	configPath := filepath.Join(a.configDir, fmt.Sprintf("%s.conf", domain))

	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove apache config: %w", err)
	}

	return nil
}

func (a *ApacheManager) Reload(ctx context.Context) error {
	// Execute reload command
	cmd := exec.CommandContext(ctx, "sh", "-c", a.reloadCommand)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apache reload failed: %s", string(output))
	}

	return nil
}

func (a *ApacheManager) GetInfo(ctx context.Context) (*models.TraefikInfo, error) {
	// Get apache version
	cmd := exec.CommandContext(ctx, "apache2", "-v")
	output, _ := cmd.CombinedOutput()

	version := string(output)

	// Count config files as "routers"
	files, err := os.ReadDir(a.configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read apache config dir: %w", err)
	}

	return &models.TraefikInfo{
		Version:       version,
		RoutersCount:  len(files),
		ServicesCount: len(files),
	}, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
