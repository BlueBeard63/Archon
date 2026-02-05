package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// MigrationVersion tracks which migrations have been applied
// Increment this when adding new migrations
const CurrentMigrationVersion = 1

// Migration represents a single migration step
type Migration struct {
	Version     int
	Description string
	Migrate     func(cfg *Config) error
}

// GetMigrations returns all available migrations in order
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Migrate per-site Docker credentials to global credentials",
			Migrate:     migrateDockerCredentials,
		},
	}
}

// MigrateConfig runs all necessary migrations on a config
// Returns true if any migrations were applied
func MigrateConfig(cfg *Config, configPath string) (bool, error) {
	currentVersion := cfg.GetMigrationVersion()
	if currentVersion >= CurrentMigrationVersion {
		return false, nil
	}

	migrations := GetMigrations()
	appliedAny := false

	for _, m := range migrations {
		if m.Version > currentVersion {
			// Backup before first migration
			if !appliedAny {
				if err := BackupConfig(configPath); err != nil {
					return false, fmt.Errorf("failed to backup config before migration: %w", err)
				}
			}

			// Run migration
			if err := m.Migrate(cfg); err != nil {
				return appliedAny, fmt.Errorf("migration v%d (%s) failed: %w", m.Version, m.Description, err)
			}

			appliedAny = true
		}
	}

	// Update migration version
	cfg.MigrationVersion = CurrentMigrationVersion

	return appliedAny, nil
}

// BackupConfig creates a timestamped backup of the config file
func BackupConfig(configPath string) error {
	// Read original file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file to backup
		}
		return fmt.Errorf("failed to read config for backup: %w", err)
	}

	// Create backup directory
	configDir := filepath.Dir(configPath)
	backupDir := filepath.Join(configDir, "config_archives")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupName := fmt.Sprintf("config_backup_%s.toml", timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	return nil
}

// GetMigrationVersion returns the current migration version from config
func (c *Config) GetMigrationVersion() int {
	return c.MigrationVersion
}

// migrateDockerCredentials migrates per-site Docker credentials to global credentials
func migrateDockerCredentials(cfg *Config) error {
	// Track unique credentials we've seen (registry+username -> credential ID)
	credentialMap := make(map[string]uuid.UUID)

	for i := range cfg.Sites {
		site := &cfg.Sites[i]

		// Skip sites without credentials or already migrated
		if site.DockerUsername == "" && site.DockerToken == "" {
			continue
		}
		if site.DockerCredentialID != nil && *site.DockerCredentialID != uuid.Nil {
			continue
		}

		// Determine registry (default to docker.io if not specified)
		registry := "docker.io"

		// Create a key for deduplication
		credKey := fmt.Sprintf("%s:%s", registry, site.DockerUsername)

		// Check if we already created a credential for this combination
		credID, exists := credentialMap[credKey]
		if !exists {
			// Create new credential
			credID = uuid.New()
			newCred := DockerCredential{
				ID:       credID,
				Name:     fmt.Sprintf("Migrated - %s", site.DockerUsername),
				Registry: registry,
				Username: site.DockerUsername,
				Token:    site.DockerToken,
			}
			cfg.Settings.DockerCredentials = append(cfg.Settings.DockerCredentials, newCred)
			credentialMap[credKey] = credID
		}

		// Update site to use credential ID
		site.DockerCredentialID = &credID

		// Clear deprecated fields
		site.DockerUsername = ""
		site.DockerToken = ""
	}

	return nil
}

// ListBackups returns a list of available config backups
func ListBackups(configPath string) ([]BackupInfo, error) {
	configDir := filepath.Dir(configPath)
	backupDir := filepath.Join(configDir, "config_archives")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".toml" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Name:      entry.Name(),
			Path:      filepath.Join(backupDir, entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	return backups, nil
}

// BackupInfo represents information about a config backup
type BackupInfo struct {
	Name      string
	Path      string
	Size      int64
	CreatedAt time.Time
}

// RestoreBackup restores a config from a backup file
func RestoreBackup(backupPath, configPath string) error {
	// First backup the current config (so we don't lose it)
	if err := BackupConfig(configPath); err != nil {
		return fmt.Errorf("failed to backup current config before restore: %w", err)
	}

	// Read backup
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Write to config path
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	return nil
}
