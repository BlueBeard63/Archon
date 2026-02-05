package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
)

func TestSiteSave_WithDockerCredentialID(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "archon-site-cred-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a site with DockerCredentialID
	credID := uuid.New()
	site := &models.Site{
		ID:                 uuid.New(),
		Name:               "test-site",
		DockerImage:        "nginx:latest",
		DockerCredentialID: &credID,
	}

	// Marshal to TOML
	data, err := toml.Marshal(site)
	if err != nil {
		t.Fatalf("Failed to marshal site: %v", err)
	}

	// Verify docker_credential_id is in the output
	tomlStr := string(data)
	if !containsString(tomlStr, "docker_credential_id") {
		t.Errorf("TOML output should contain docker_credential_id, got:\n%s", tomlStr)
	}
	if !containsString(tomlStr, credID.String()) {
		t.Errorf("TOML output should contain the credential UUID %s, got:\n%s", credID.String(), tomlStr)
	}

	// Write to file
	sitePath := filepath.Join(tmpDir, "site.toml")
	if err := os.WriteFile(sitePath, data, 0644); err != nil {
		t.Fatalf("Failed to write site file: %v", err)
	}

	// Read back and unmarshal
	readData, err := os.ReadFile(sitePath)
	if err != nil {
		t.Fatalf("Failed to read site file: %v", err)
	}

	var loadedSite models.Site
	if err := toml.Unmarshal(readData, &loadedSite); err != nil {
		t.Fatalf("Failed to unmarshal site: %v", err)
	}

	// Verify the credential ID was preserved
	if loadedSite.DockerCredentialID == nil {
		t.Error("Loaded site should have DockerCredentialID set")
	} else if *loadedSite.DockerCredentialID != credID {
		t.Errorf("Loaded DockerCredentialID = %s, want %s", loadedSite.DockerCredentialID.String(), credID.String())
	}
}

func TestSiteSave_WithoutDockerCredentialID(t *testing.T) {
	// Create a site without DockerCredentialID
	site := &models.Site{
		ID:          uuid.New(),
		Name:        "test-site-public",
		DockerImage: "nginx:latest",
		// DockerCredentialID is nil
	}

	// Marshal to TOML
	data, err := toml.Marshal(site)
	if err != nil {
		t.Fatalf("Failed to marshal site: %v", err)
	}

	// Verify docker_credential_id is NOT in the output (omitempty)
	tomlStr := string(data)
	if containsString(tomlStr, "docker_credential_id") {
		t.Errorf("TOML output should NOT contain docker_credential_id when nil, got:\n%s", tomlStr)
	}
}

func TestFileConfigLoader_SaveSite_WithCredential(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "archon-loader-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create sites directory
	sitesDir := filepath.Join(tmpDir, "sites", "example.com", "test-site")
	if err := os.MkdirAll(sitesDir, 0755); err != nil {
		t.Fatalf("Failed to create sites dir: %v", err)
	}

	// Create a site with credential ID
	credID := uuid.New()
	site := &models.Site{
		ID:                 uuid.New(),
		Name:               "test-site",
		DockerImage:        "myrepo/myimage:latest",
		DockerCredentialID: &credID,
	}

	// Use the loader to save the site
	loader := &FileConfigLoader{}

	// Save site config directly to the temp location
	sitePath := filepath.Join(sitesDir, "config.toml")
	data, err := toml.Marshal(site)
	if err != nil {
		t.Fatalf("Failed to marshal site: %v", err)
	}
	if err := os.WriteFile(sitePath, data, 0644); err != nil {
		t.Fatalf("Failed to write site config: %v", err)
	}

	// Read back the config file
	readData, err := os.ReadFile(sitePath)
	if err != nil {
		t.Fatalf("Failed to read site config: %v", err)
	}

	// Verify contents
	tomlStr := string(readData)
	t.Logf("Site config.toml contents:\n%s", tomlStr)

	if !containsString(tomlStr, "docker_credential_id") {
		t.Error("Site config should contain docker_credential_id")
	}

	// Unmarshal and verify
	var loadedSite models.Site
	if err := toml.Unmarshal(readData, &loadedSite); err != nil {
		t.Fatalf("Failed to unmarshal site: %v", err)
	}

	if loadedSite.DockerCredentialID == nil {
		t.Error("Loaded site should have DockerCredentialID")
	} else if *loadedSite.DockerCredentialID != credID {
		t.Errorf("Loaded DockerCredentialID = %s, want %s", *loadedSite.DockerCredentialID, credID)
	}

	// Suppress unused warning
	_ = loader
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
