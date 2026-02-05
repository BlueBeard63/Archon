package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchConfigFromURL_Success(t *testing.T) {
	// Create a test server that returns valid config
	nodeID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"config":  "[server]\nhost = \"0.0.0.0\"\nport = 8080\napi_key = \"test-key\"\n",
			"node_id": nodeID.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create a temp directory for the config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Fetch and save the config
	cfg, err := FetchConfigFromURL(server.URL, configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "test-key", cfg.Server.APIKey)

	// Verify file was saved
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify file content
	content, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "[server]")
}

func TestFetchConfigFromURL_NetworkError(t *testing.T) {
	// Use an invalid URL that won't connect
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	_, err := FetchConfigFromURL("http://localhost:99999/invalid", configPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch config")

	// Verify file was not created
	_, err = os.Stat(configPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFetchConfigFromURL_InvalidTOML(t *testing.T) {
	// Create a test server that returns invalid TOML
	nodeID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"config":  "this is not valid toml {{{{",
			"node_id": nodeID.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	_, err := FetchConfigFromURL(server.URL, configPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

func TestFetchConfigFromURL_SavesCorrectly(t *testing.T) {
	// Create a test server that returns valid config
	configContent := `[server]
host = "0.0.0.0"
port = 9090
api_key = "my-secret-key"
data_dir = "/var/lib/archon"

[docker]
host = ""
network = "archon"

[proxy]
type = "traefik"
config_dir = "/etc/traefik"
api_url = "http://traefik:8080"

[ssl]
mode = "auto"
cert_dir = "/etc/certs"
`
	nodeID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"config":  configContent,
			"node_id": nodeID.String(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	cfg, err := FetchConfigFromURL(server.URL, configPath)

	require.NoError(t, err)

	// Verify all config fields
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "my-secret-key", cfg.Server.APIKey)
	assert.Equal(t, "/var/lib/archon", cfg.Server.DataDir)
	assert.Equal(t, "archon", cfg.Docker.Network)
	assert.Equal(t, ProxyTypeTraefik, cfg.Proxy.Type)
	assert.Equal(t, SSLMode("auto"), cfg.SSL.Mode)

	// Verify saved file matches
	savedContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, configContent, string(savedContent))

	// Verify file permissions (0600)
	info, err := os.Stat(configPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestFetchConfigFromURL_ExpiredOrFetched(t *testing.T) {
	// Create a test server that returns 410 Gone (already fetched)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
		json.NewEncoder(w).Encode(map[string]string{"error": "Config has already been fetched"})
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	_, err := FetchConfigFromURL(server.URL, configPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "410")
}

func TestFetchConfigFromURL_NotFound(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Config not found"})
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	_, err := FetchConfigFromURL(server.URL, configPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestFetchConfigFromURL_PlainTextTOML(t *testing.T) {
	// Create a test server that returns plain text TOML (like dpaste.org)
	configContent := `[server]
host = "0.0.0.0"
port = 8080
api_key = "test-key"
data_dir = "/var/lib/archon"

[docker]
host = ""
network = "archon"

[proxy]
type = "traefik"
config_dir = "/etc/traefik"
api_url = "http://traefik:8080"

[ssl]
mode = "auto"
cert_dir = "/etc/certs"
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return plain text (like dpaste.org/XXXX/raw)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(configContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	cfg, err := FetchConfigFromURL(server.URL, configPath)

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "test-key", cfg.Server.APIKey)
	assert.Equal(t, "/var/lib/archon", cfg.Server.DataDir)
	assert.Equal(t, "archon", cfg.Docker.Network)
	assert.Equal(t, ProxyTypeTraefik, cfg.Proxy.Type)

	// Verify saved file
	savedContent, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Equal(t, configContent, string(savedContent))
}

func TestFetchConfigFromURL_PlainTextTOML_NoContentType(t *testing.T) {
	// Test when server doesn't set Content-Type header (should still work)
	configContent := `[server]
host = "0.0.0.0"
port = 9000
api_key = "another-key"
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't set Content-Type, just return plain text
		w.Write([]byte(configContent))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	cfg, err := FetchConfigFromURL(server.URL, configPath)

	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 9000, cfg.Server.Port)
	assert.Equal(t, "another-key", cfg.Server.APIKey)
}

func TestFetchConfigFromURL_PlainTextTOML_InvalidTOML(t *testing.T) {
	// Test when plain text response is invalid TOML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("this is not valid toml {{{{"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	_, err := FetchConfigFromURL(server.URL, configPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}
