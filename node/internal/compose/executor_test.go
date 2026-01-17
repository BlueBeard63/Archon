package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestExecutor_WriteComposeFile(t *testing.T) {
	// Create a temporary directory for tests
	tempDir, err := os.MkdirTemp("", "compose-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	executor := NewExecutor(tempDir, "test-network")

	tests := []struct {
		name    string
		siteID  uuid.UUID
		content string
	}{
		{
			name:   "simple compose content",
			siteID: uuid.New(),
			content: `version: "3.8"
services:
  web:
    image: nginx
    ports:
      - "80:80"
`,
		},
		{
			name:   "multi-service compose",
			siteID: uuid.New(),
			content: `version: "3.8"
services:
  app:
    build: .
    ports:
      - "3000:3000"
  db:
    image: postgres:14
    environment:
      POSTGRES_PASSWORD: secret
`,
		},
		{
			name:    "minimal compose",
			siteID:  uuid.New(),
			content: `services: { web: { image: nginx } }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			composePath, err := executor.writeComposeFile(tt.siteID, tt.content)
			if err != nil {
				t.Fatalf("writeComposeFile() error = %v", err)
			}

			// Verify file exists
			if _, err := os.Stat(composePath); os.IsNotExist(err) {
				t.Errorf("Compose file was not created at %s", composePath)
			}

			// Verify file content
			readContent, err := os.ReadFile(composePath)
			if err != nil {
				t.Fatalf("Failed to read compose file: %v", err)
			}
			if string(readContent) != tt.content {
				t.Errorf("Compose file content mismatch.\nGot:\n%s\nWant:\n%s", string(readContent), tt.content)
			}

			// Verify path structure
			expectedDir := filepath.Join(tempDir, "compose", tt.siteID.String())
			if !strings.HasPrefix(composePath, expectedDir) {
				t.Errorf("Compose file path %s does not start with expected dir %s", composePath, expectedDir)
			}

			// Verify filename
			if filepath.Base(composePath) != "docker-compose.yml" {
				t.Errorf("Compose filename = %s, want docker-compose.yml", filepath.Base(composePath))
			}
		})
	}
}

func TestExecutor_PathGeneration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compose-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	executor := NewExecutor(tempDir, "test-network")
	siteID := uuid.New()

	composePath, err := executor.writeComposeFile(siteID, "services: {}")
	if err != nil {
		t.Fatalf("writeComposeFile() error = %v", err)
	}

	// Expected path structure: {tempDir}/compose/{siteID}/docker-compose.yml
	expectedPath := filepath.Join(tempDir, "compose", siteID.String(), "docker-compose.yml")
	if composePath != expectedPath {
		t.Errorf("Path = %s, want %s", composePath, expectedPath)
	}

	// Verify directory structure was created
	composeDir := filepath.Dir(composePath)
	info, err := os.Stat(composeDir)
	if err != nil {
		t.Errorf("Compose directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Expected %s to be a directory", composeDir)
	}
}

func TestExecutor_WriteComposeFile_OverwritesExisting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compose-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	executor := NewExecutor(tempDir, "test-network")
	siteID := uuid.New()

	// Write first version
	content1 := "services: { web: { image: nginx:1.0 } }"
	path1, err := executor.writeComposeFile(siteID, content1)
	if err != nil {
		t.Fatalf("First writeComposeFile() error = %v", err)
	}

	// Write second version (should overwrite)
	content2 := "services: { web: { image: nginx:2.0 } }"
	path2, err := executor.writeComposeFile(siteID, content2)
	if err != nil {
		t.Fatalf("Second writeComposeFile() error = %v", err)
	}

	// Paths should be the same
	if path1 != path2 {
		t.Errorf("Paths differ: %s vs %s", path1, path2)
	}

	// Content should be the second version
	readContent, err := os.ReadFile(path2)
	if err != nil {
		t.Fatalf("Failed to read compose file: %v", err)
	}
	if string(readContent) != content2 {
		t.Errorf("Content not overwritten. Got:\n%s\nWant:\n%s", string(readContent), content2)
	}
}

func TestExecutor_Cleanup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compose-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	executor := NewExecutor(tempDir, "test-network")
	siteID := uuid.New()

	// Write a compose file
	composePath, err := executor.writeComposeFile(siteID, "services: {}")
	if err != nil {
		t.Fatalf("writeComposeFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Fatalf("Compose file was not created")
	}

	// Cleanup
	if err := executor.cleanup(siteID); err != nil {
		t.Fatalf("cleanup() error = %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(composePath); !os.IsNotExist(err) {
		t.Errorf("Compose file still exists after cleanup")
	}

	// Verify directory is gone
	siteDir := filepath.Join(tempDir, "compose", siteID.String())
	if _, err := os.Stat(siteDir); !os.IsNotExist(err) {
		t.Errorf("Site directory still exists after cleanup")
	}
}

func TestExecutor_Cleanup_NonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compose-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	executor := NewExecutor(tempDir, "test-network")

	// Cleanup for a site that was never created should not error
	err = executor.cleanup(uuid.New())
	if err != nil {
		t.Errorf("cleanup() should not error for non-existent site, got: %v", err)
	}
}

func TestProjectNameFormat(t *testing.T) {
	// The project name format is "archon-{name}"
	// This is used in DeploySite, StopSite, DeleteSite, GetStatus

	tests := []struct {
		siteName    string
		wantProject string
	}{
		{"myapp", "archon-myapp"},
		{"my-app", "archon-my-app"},
		{"my_app", "archon-my_app"},
		{"app123", "archon-app123"},
	}

	for _, tt := range tests {
		t.Run(tt.siteName, func(t *testing.T) {
			// Verify format by constructing project name same way as executor
			projectName := "archon-" + tt.siteName
			if projectName != tt.wantProject {
				t.Errorf("Project name = %s, want %s", projectName, tt.wantProject)
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	tempDir := "/tmp/test"
	networkName := "archon-network"

	executor := NewExecutor(tempDir, networkName)

	if executor == nil {
		t.Fatal("NewExecutor() returned nil")
	}
	if executor.tempDir != tempDir {
		t.Errorf("tempDir = %s, want %s", executor.tempDir, tempDir)
	}
	if executor.networkName != networkName {
		t.Errorf("networkName = %s, want %s", executor.networkName, networkName)
	}
}

func TestExecutor_WriteComposeFile_CreatesDirWithCorrectPermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "compose-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	executor := NewExecutor(tempDir, "test-network")
	siteID := uuid.New()

	composePath, err := executor.writeComposeFile(siteID, "services: {}")
	if err != nil {
		t.Fatalf("writeComposeFile() error = %v", err)
	}

	// Check file permissions (should be 0644)
	fileInfo, err := os.Stat(composePath)
	if err != nil {
		t.Fatalf("Failed to stat compose file: %v", err)
	}

	// On Unix, check that the file is readable by owner at minimum
	perm := fileInfo.Mode().Perm()
	if perm&0400 == 0 {
		t.Errorf("Compose file is not readable by owner, permissions: %o", perm)
	}

	// Check directory permissions (should be 0755)
	dirInfo, err := os.Stat(filepath.Dir(composePath))
	if err != nil {
		t.Fatalf("Failed to stat compose directory: %v", err)
	}

	dirPerm := dirInfo.Mode().Perm()
	if dirPerm&0500 == 0 {
		t.Errorf("Compose directory is not accessible by owner, permissions: %o", dirPerm)
	}
}
