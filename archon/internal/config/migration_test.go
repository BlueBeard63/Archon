package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/google/uuid"
)

func TestMigrateDockerCredentials(t *testing.T) {
	// Create a config with sites using old-style credentials
	cfg := &Config{
		Version: "1.0.0",
		Sites: []models.Site{
			{
				ID:             uuid.New(),
				Name:           "site1",
				DockerImage:    "myrepo/image1:latest",
				DockerUsername: "user1",
				DockerToken:    "token1",
			},
			{
				ID:             uuid.New(),
				Name:           "site2",
				DockerImage:    "myrepo/image2:latest",
				DockerUsername: "user1", // Same user as site1
				DockerToken:    "token1",
			},
			{
				ID:             uuid.New(),
				Name:           "site3",
				DockerImage:    "myrepo/image3:latest",
				DockerUsername: "user2", // Different user
				DockerToken:    "token2",
			},
			{
				ID:          uuid.New(),
				Name:        "site4",
				DockerImage: "nginx:latest",
				// No credentials - public image
			},
		},
		Settings: Settings{},
	}

	// Run migration
	err := migrateDockerCredentials(cfg)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify credentials were created
	if len(cfg.Settings.DockerCredentials) != 2 {
		t.Errorf("Expected 2 credentials (deduplicated), got %d", len(cfg.Settings.DockerCredentials))
	}

	// Verify sites were updated
	for i, site := range cfg.Sites {
		if i < 3 { // Sites 0-2 had credentials
			if site.DockerCredentialID == nil {
				t.Errorf("Site %d should have DockerCredentialID set", i)
			}
			if site.DockerUsername != "" {
				t.Errorf("Site %d should have DockerUsername cleared, got %q", i, site.DockerUsername)
			}
			if site.DockerToken != "" {
				t.Errorf("Site %d should have DockerToken cleared", i)
			}
		} else { // Site 3 had no credentials
			if site.DockerCredentialID != nil {
				t.Errorf("Site %d should not have DockerCredentialID", i)
			}
		}
	}

	// Verify sites 0 and 1 share the same credential (deduplication)
	if cfg.Sites[0].DockerCredentialID == nil || cfg.Sites[1].DockerCredentialID == nil {
		t.Fatal("Sites 0 and 1 should have credential IDs")
	}
	if *cfg.Sites[0].DockerCredentialID != *cfg.Sites[1].DockerCredentialID {
		t.Error("Sites 0 and 1 should share the same credential ID")
	}

	// Verify site 2 has different credential
	if cfg.Sites[2].DockerCredentialID == nil {
		t.Fatal("Site 2 should have credential ID")
	}
	if *cfg.Sites[2].DockerCredentialID == *cfg.Sites[0].DockerCredentialID {
		t.Error("Site 2 should have different credential ID than sites 0/1")
	}
}

func TestMigrateDockerCredentials_AlreadyMigrated(t *testing.T) {
	credID := uuid.New()
	cfg := &Config{
		Version: "1.0.0",
		Sites: []models.Site{
			{
				ID:                 uuid.New(),
				Name:               "site1",
				DockerImage:        "myrepo/image1:latest",
				DockerCredentialID: &credID, // Already migrated
			},
		},
		Settings: Settings{
			DockerCredentials: []DockerCredential{
				{
					ID:       credID,
					Name:     "Existing",
					Registry: "docker.io",
					Username: "existinguser",
					Token:    "existingtoken",
				},
			},
		},
	}

	// Run migration
	err := migrateDockerCredentials(cfg)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Should not add new credentials
	if len(cfg.Settings.DockerCredentials) != 1 {
		t.Errorf("Expected 1 credential (no new ones), got %d", len(cfg.Settings.DockerCredentials))
	}

	// Credential ID should be unchanged
	if *cfg.Sites[0].DockerCredentialID != credID {
		t.Error("Credential ID should not have changed")
	}
}

func TestBackupConfig(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "archon-migration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	configPath := filepath.Join(tmpDir, "config.toml")
	testContent := []byte(`version = "1.0.0"

[settings]
auto_save = true
`)
	if err := os.WriteFile(configPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Run backup
	err = BackupConfig(configPath)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup directory was created
	backupDir := filepath.Join(tmpDir, "config_archives")
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Error("Backup directory should have been created")
	}

	// Verify backup file exists
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("Failed to read backup dir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 backup file, got %d", len(entries))
	}

	// Verify backup content matches original
	backupPath := filepath.Join(backupDir, entries[0].Name())
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}
	if string(backupContent) != string(testContent) {
		t.Error("Backup content does not match original")
	}
}

func TestListBackups(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "archon-backup-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")

	// Test with no backups
	backups, err := ListBackups(configPath)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups, got %d", len(backups))
	}

	// Create some backup files
	backupDir := filepath.Join(tmpDir, "config_archives")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}

	testBackups := []string{
		"config_backup_2024-01-01_12-00-00.toml",
		"config_backup_2024-01-02_12-00-00.toml",
	}
	for _, name := range testBackups {
		if err := os.WriteFile(filepath.Join(backupDir, name), []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create backup file: %v", err)
		}
	}

	// List backups
	backups, err = ListBackups(configPath)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}
	if len(backups) != 2 {
		t.Errorf("Expected 2 backups, got %d", len(backups))
	}
}

func TestMigrateConfig_VersionCheck(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "archon-version-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.toml")
	if err := os.WriteFile(configPath, []byte("version = \"1.0.0\""), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Config already at current version
	cfg := &Config{
		Version:          "1.0.0",
		MigrationVersion: CurrentMigrationVersion,
	}

	migrated, err := MigrateConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("MigrateConfig failed: %v", err)
	}
	if migrated {
		t.Error("Should not have migrated - already at current version")
	}

	// Config at older version should migrate
	cfg2 := &Config{
		Version:          "1.0.0",
		MigrationVersion: 0,
		Sites: []models.Site{
			{
				ID:             uuid.New(),
				DockerUsername: "user",
				DockerToken:    "token",
			},
		},
	}

	migrated, err = MigrateConfig(cfg2, configPath)
	if err != nil {
		t.Fatalf("MigrateConfig failed: %v", err)
	}
	if !migrated {
		t.Error("Should have migrated from version 0")
	}
	if cfg2.MigrationVersion != CurrentMigrationVersion {
		t.Errorf("Migration version should be %d, got %d", CurrentMigrationVersion, cfg2.MigrationVersion)
	}
}

func TestRestoreBackup(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "archon-restore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create current config
	configPath := filepath.Join(tmpDir, "config.toml")
	currentContent := []byte("version = \"2.0.0\"\n")
	if err := os.WriteFile(configPath, currentContent, 0644); err != nil {
		t.Fatalf("Failed to write current config: %v", err)
	}

	// Create backup
	backupDir := filepath.Join(tmpDir, "config_archives")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup dir: %v", err)
	}
	backupPath := filepath.Join(backupDir, "config_backup_old.toml")
	backupContent := []byte("version = \"1.0.0\"\n")
	if err := os.WriteFile(backupPath, backupContent, 0644); err != nil {
		t.Fatalf("Failed to write backup: %v", err)
	}

	// Restore
	err = RestoreBackup(backupPath, configPath)
	if err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	// Verify config was restored
	restoredContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}
	if string(restoredContent) != string(backupContent) {
		t.Error("Restored content does not match backup")
	}

	// Verify a backup of the current config was created
	entries, _ := os.ReadDir(backupDir)
	if len(entries) < 2 {
		t.Error("Should have created a backup of current config before restore")
	}
}
