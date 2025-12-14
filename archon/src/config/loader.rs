use anyhow::{Context, Result};
use std::fs;
use std::path::Path;

use super::schema::ArchonConfig;

/// Load configuration from a TOML file
pub fn load_config(path: &Path) -> Result<ArchonConfig> {
    if !path.exists() {
        // If config doesn't exist, create a default one
        let config = ArchonConfig::default();
        save_config(path, &config)?;
        return Ok(config);
    }

    let content = fs::read_to_string(path)
        .with_context(|| format!("Failed to read config file: {}", path.display()))?;

    let config: ArchonConfig = toml::from_str(&content)
        .with_context(|| format!("Failed to parse config file: {}", path.display()))?;

    Ok(config)
}

/// Save configuration to a TOML file
pub fn save_config(path: &Path, config: &ArchonConfig) -> Result<()> {
    // Create parent directories if they don't exist
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)
            .with_context(|| format!("Failed to create config directory: {}", parent.display()))?;
    }

    let content = toml::to_string_pretty(config)
        .context("Failed to serialize config to TOML")?;

    fs::write(path, content)
        .with_context(|| format!("Failed to write config file: {}", path.display()))?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::path::PathBuf;
    use tempfile::TempDir;

    #[test]
    fn test_load_nonexistent_creates_default() {
        let temp_dir = TempDir::new().unwrap();
        let config_path = temp_dir.path().join("archon.toml");

        let config = load_config(&config_path).unwrap();
        assert_eq!(config.version, env!("CARGO_PKG_VERSION"));
        assert!(config.sites.is_empty());
        assert!(config_path.exists());
    }

    #[test]
    fn test_save_and_load() {
        let temp_dir = TempDir::new().unwrap();
        let config_path = temp_dir.path().join("test.toml");

        let mut config = ArchonConfig::default();
        config.settings.auto_save = false;

        save_config(&config_path, &config).unwrap();
        let loaded = load_config(&config_path).unwrap();

        assert_eq!(loaded.settings.auto_save, false);
    }
}
