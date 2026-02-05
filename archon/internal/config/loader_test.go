package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "archon-test-*")
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

func createTestSite(t *testing.T, baseDir, domainName, siteName string) models.Site {
	t.Helper()
	site := models.Site{
		ID:       uuid.New(),
		Name:     siteName,
		Status:   models.SiteStatusInactive,
		SiteType: models.SiteTypeContainer,
		DomainID: uuid.New(),
		NodeID:   uuid.New(),
	}

	sitePath := filepath.Join(baseDir, "sites", domainName, siteName)
	err := os.MkdirAll(sitePath, 0755)
	require.NoError(t, err)

	data, err := toml.Marshal(site)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sitePath, "config.toml"), data, 0644)
	require.NoError(t, err)

	return site
}

func TestArchiveSite_CreatesCorrectPath(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a test site
	domainName := "example.com"
	siteName := "test-site"
	site := createTestSite(t, baseDir, domainName, siteName)

	loader := &FileConfigLoader{}

	// Archive the site
	archivePath, err := loader.ArchiveSiteWithBaseDir(baseDir, siteName, domainName, site)
	require.NoError(t, err)

	// Verify the archive path structure
	assert.Contains(t, archivePath, filepath.Join(baseDir, "deleted", domainName, siteName))

	// Verify the archived file exists
	configPath := filepath.Join(archivePath, "config.toml")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Verify the original site directory is removed
	originalPath := filepath.Join(baseDir, "sites", domainName, siteName)
	_, err = os.Stat(originalPath)
	assert.True(t, os.IsNotExist(err))
}

func TestArchiveSite_PreservesAllFields(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a site with all fields populated
	domainName := "example.com"
	siteName := "full-site"
	site := models.Site{
		ID:             uuid.New(),
		Name:           siteName,
		Status:         models.SiteStatusRunning,
		SiteType:       models.SiteTypeContainer,
		DomainID:       uuid.New(),
		NodeID:         uuid.New(),
		DockerImage:    "nginx:latest",
		DockerUsername: "testuser",
		Port:           8080,
		EnvironmentVars: map[string]string{
			"ENV_VAR": "value",
		},
	}

	// Create the site directory manually
	sitePath := filepath.Join(baseDir, "sites", domainName, siteName)
	err := os.MkdirAll(sitePath, 0755)
	require.NoError(t, err)

	data, err := toml.Marshal(site)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(sitePath, "config.toml"), data, 0644)
	require.NoError(t, err)

	loader := &FileConfigLoader{}

	// Archive the site
	archivePath, err := loader.ArchiveSiteWithBaseDir(baseDir, siteName, domainName, site)
	require.NoError(t, err)

	// Read back and verify all fields match
	configPath := filepath.Join(archivePath, "config.toml")
	archivedData, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var archivedSite models.Site
	err = toml.Unmarshal(archivedData, &archivedSite)
	require.NoError(t, err)

	assert.Equal(t, site.ID, archivedSite.ID)
	assert.Equal(t, site.Name, archivedSite.Name)
	assert.Equal(t, site.Status, archivedSite.Status)
	assert.Equal(t, site.SiteType, archivedSite.SiteType)
	assert.Equal(t, site.DockerImage, archivedSite.DockerImage)
	assert.Equal(t, site.DockerUsername, archivedSite.DockerUsername)
	assert.Equal(t, site.Port, archivedSite.Port)
	assert.Equal(t, site.EnvironmentVars, archivedSite.EnvironmentVars)
}

func TestLoadDeletedSites_ReturnsAllArchived(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create multiple archived sites
	domainName := "example.com"
	loader := &FileConfigLoader{}

	// Archive site 1
	site1 := createTestSite(t, baseDir, domainName, "site1")
	_, err := loader.ArchiveSiteWithBaseDir(baseDir, "site1", domainName, site1)
	require.NoError(t, err)

	// Wait a bit to ensure different timestamps
	time.Sleep(10 * time.Millisecond)

	// Archive site 2
	site2 := createTestSite(t, baseDir, domainName, "site2")
	_, err = loader.ArchiveSiteWithBaseDir(baseDir, "site2", domainName, site2)
	require.NoError(t, err)

	// Load deleted sites
	deletedSites, err := loader.LoadDeletedSitesWithBaseDir(baseDir)
	require.NoError(t, err)

	// Verify both sites are returned
	assert.Len(t, deletedSites, 2)

	// Verify site names
	siteNames := make([]string, len(deletedSites))
	for i, ds := range deletedSites {
		siteNames[i] = ds.Site.Name
	}
	assert.Contains(t, siteNames, "site1")
	assert.Contains(t, siteNames, "site2")
}

func TestLoadDeletedSites_EmptyWhenNoArchives(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	loader := &FileConfigLoader{}

	// Load deleted sites from empty directory
	deletedSites, err := loader.LoadDeletedSitesWithBaseDir(baseDir)
	require.NoError(t, err)

	// Verify empty list returned, no error
	assert.Len(t, deletedSites, 0)
}

func TestRestoreDeletedSite_MovesToActiveSites(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create and archive a site
	domainName := "example.com"
	siteName := "restore-test"
	site := createTestSite(t, baseDir, domainName, siteName)

	loader := &FileConfigLoader{}
	archivePath, err := loader.ArchiveSiteWithBaseDir(baseDir, siteName, domainName, site)
	require.NoError(t, err)

	// Restore the site
	err = loader.RestoreDeletedSiteWithBaseDir(baseDir, archivePath, siteName, domainName)
	require.NoError(t, err)

	// Verify site exists in active location
	activePath := filepath.Join(baseDir, "sites", domainName, siteName, "config.toml")
	_, err = os.Stat(activePath)
	assert.NoError(t, err)

	// Verify removed from archive
	_, err = os.Stat(filepath.Join(archivePath, "config.toml"))
	assert.True(t, os.IsNotExist(err))
}

func TestRestoreDeletedSite_InvalidPath_ReturnsError(t *testing.T) {
	baseDir, cleanup := setupTestDir(t)
	defer cleanup()

	loader := &FileConfigLoader{}

	// Try to restore non-existent archive
	err := loader.RestoreDeletedSiteWithBaseDir(baseDir, "/non/existent/path", "test", "example.com")
	assert.Error(t, err)
}
