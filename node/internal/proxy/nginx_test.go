package proxy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/BlueBeard63/archon-node/internal/config"
	"github.com/BlueBeard63/archon-node/internal/models"
)

func TestDefaultBotUserAgents(t *testing.T) {
	bots := DefaultBotUserAgents()

	// Test that default list contains expected bots
	expectedBots := []string{
		"Googlebot",
		"Bingbot",
		"Twitterbot",
		"LinkedInBot",
		"facebot",
		"Slackbot",
		"Discordbot",
	}

	for _, expected := range expectedBots {
		found := false
		for _, bot := range bots {
			if bot == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected bot %s not found in default list", expected)
	}

	// Verify list is not empty
	assert.NotEmpty(t, bots)
}

func TestBuildNginxBotMapEntries(t *testing.T) {
	tests := []struct {
		name       string
		userAgents []string
		wantCount  int
	}{
		{
			name:       "empty list uses defaults",
			userAgents: []string{},
			wantCount:  len(DefaultBotUserAgents()),
		},
		{
			name:       "nil list uses defaults",
			userAgents: nil,
			wantCount:  len(DefaultBotUserAgents()),
		},
		{
			name:       "custom list",
			userAgents: []string{"CustomBot", "AnotherBot"},
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildNginxBotMapEntries(tt.userAgents)
			lines := strings.Split(strings.TrimSpace(result), "\n")
			assert.Len(t, lines, tt.wantCount)

			// Verify format of each line
			for _, line := range lines {
				assert.True(t, strings.HasPrefix(strings.TrimSpace(line), "~*"), "Line should start with ~*")
				assert.True(t, strings.HasSuffix(strings.TrimSpace(line), "1;"), "Line should end with 1;")
			}
		})
	}
}

func TestNginxConfig_WithoutBotRedirect(t *testing.T) {
	// Create temp directory for nginx configs
	tempDir, err := os.MkdirTemp("", "nginx-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create manager
	proxyCfg := &config.ProxyConfig{
		Type:          config.ProxyTypeNginx,
		ConfigDir:     tempDir,
		ReloadCommand: "echo reload",
	}
	sslCfg := &config.SSLConfig{
		Mode: "manual",
	}
	manager := NewNginxManager(proxyCfg, sslCfg)

	// Create deploy request without bot redirect
	site := &models.DeployRequest{
		ID:   uuid.New(),
		Name: "test-site",
		DomainMappings: []models.DomainMapping{
			{Domain: "example.com", Port: 8080},
		},
		SSLEnabled:         false,
		BotRedirectEnabled: false,
	}

	// Generate config (this won't actually test nginx, just template generation)
	err = manager.generateNginxConfig(site, "", "")
	require.NoError(t, err)

	// Read generated config
	configPath := filepath.Join(tempDir, "example.com.conf")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Verify no bot-related directives
	assert.NotContains(t, string(content), "map $http_user_agent $is_bot")
	assert.NotContains(t, string(content), "if ($is_bot)")
}

func TestNginxConfig_WithBotRedirect(t *testing.T) {
	// Create temp directory for nginx configs
	tempDir, err := os.MkdirTemp("", "nginx-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create manager
	proxyCfg := &config.ProxyConfig{
		Type:          config.ProxyTypeNginx,
		ConfigDir:     tempDir,
		ReloadCommand: "echo reload",
	}
	sslCfg := &config.SSLConfig{
		Mode: "manual",
	}
	manager := NewNginxManager(proxyCfg, sslCfg)

	// Create deploy request with bot redirect
	site := &models.DeployRequest{
		ID:   uuid.New(),
		Name: "test-site",
		DomainMappings: []models.DomainMapping{
			{Domain: "example.com", Port: 8080},
		},
		SSLEnabled:         false,
		BotRedirectEnabled: true,
		BotRedirectURL:     "https://prerender.example.com/render",
	}

	// Generate config
	err = manager.generateNginxConfig(site, "", "")
	require.NoError(t, err)

	// Read generated config
	configPath := filepath.Join(tempDir, "example.com.conf")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configStr := string(content)

	// Verify bot redirect directives exist
	assert.Contains(t, configStr, "map $http_user_agent $is_bot")
	assert.Contains(t, configStr, "~*Googlebot")
	assert.Contains(t, configStr, "if ($is_bot)")
	assert.Contains(t, configStr, "https://prerender.example.com/render")
}

func TestNginxConfig_CustomBotUserAgents(t *testing.T) {
	// Create temp directory for nginx configs
	tempDir, err := os.MkdirTemp("", "nginx-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create manager
	proxyCfg := &config.ProxyConfig{
		Type:          config.ProxyTypeNginx,
		ConfigDir:     tempDir,
		ReloadCommand: "echo reload",
	}
	sslCfg := &config.SSLConfig{
		Mode: "manual",
	}
	manager := NewNginxManager(proxyCfg, sslCfg)

	// Create deploy request with custom bot user agents
	site := &models.DeployRequest{
		ID:   uuid.New(),
		Name: "test-site",
		DomainMappings: []models.DomainMapping{
			{Domain: "example.com", Port: 8080},
		},
		SSLEnabled:         false,
		BotRedirectEnabled: true,
		BotRedirectURL:     "https://prerender.example.com",
		BotUserAgents:      []string{"CustomBot", "MySpecialBot"},
	}

	// Generate config
	err = manager.generateNginxConfig(site, "", "")
	require.NoError(t, err)

	// Read generated config
	configPath := filepath.Join(tempDir, "example.com.conf")
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)

	configStr := string(content)

	// Verify custom bots are included
	assert.Contains(t, configStr, "~*CustomBot")
	assert.Contains(t, configStr, "~*MySpecialBot")

	// Verify default bots are NOT included (since custom list was provided)
	assert.NotContains(t, configStr, "~*Googlebot")
}

func TestNginxConfig_BotUserAgentEscaping(t *testing.T) {
	// Test that special characters in user agent patterns don't break config
	pattern := "Bot/1.0 (compatible)"
	escaped := EscapeNginxRegex(pattern)

	// Should not crash or produce invalid config
	assert.NotEmpty(t, escaped)
}
