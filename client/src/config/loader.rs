use chrono::{DateTime, Utc};
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::{Path, PathBuf};
use thiserror::Error;
use uuid::Uuid;

use crate::models::{Domain, Node, Site};

// ---------------------------------------------------------------------------
// Error type
// ---------------------------------------------------------------------------

#[derive(Debug, Error)]
pub enum ConfigError {
    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),
    #[error("TOML deserialize error: {0}")]
    TomlDeserialize(#[from] toml::de::Error),
    #[error("TOML serialize error: {0}")]
    TomlSerialize(#[from] toml::ser::Error),
    #[error("{0}")]
    Validation(String),
    #[error("migration v{version} ({description}) failed: {source}")]
    Migration {
        version: i32,
        description: String,
        source: Box<ConfigError>,
    },
    #[error("credential with ID {0} not found")]
    CredentialNotFound(Uuid),
}

// ---------------------------------------------------------------------------
// DockerCredential
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DockerCredential {
    pub id: Uuid,
    pub name: String,
    pub registry: String,
    pub username: String,
    pub token: String,
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Settings {
    pub auto_save: bool,
    pub health_check_interval_secs: i32,
    pub default_dns_ttl: i32,
    pub theme: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub cloudflare_api_token: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub route53_access_key: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub route53_secret_key: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub docker_credentials: Vec<DockerCredential>,
}

impl Default for Settings {
    fn default() -> Self {
        Self {
            auto_save: true,
            health_check_interval_secs: 300,
            default_dns_ttl: 300,
            theme: "default".to_string(),
            cloudflare_api_token: String::new(),
            route53_access_key: String::new(),
            route53_secret_key: String::new(),
            docker_credentials: Vec::new(),
        }
    }
}

impl Settings {
    pub fn add_docker_credential(&mut self, mut cred: DockerCredential) -> Result<(), ConfigError> {
        if cred.name.is_empty() {
            return Err(ConfigError::Validation(
                "credential name is required".into(),
            ));
        }
        if cred.registry.is_empty() {
            return Err(ConfigError::Validation(
                "credential registry is required".into(),
            ));
        }
        if cred.id.is_nil() {
            cred.id = Uuid::new_v4();
        }
        self.docker_credentials.push(cred);
        Ok(())
    }

    pub fn get_docker_credential_by_id(&self, id: Uuid) -> Option<&DockerCredential> {
        if id.is_nil() {
            return None;
        }
        self.docker_credentials.iter().find(|c| c.id == id)
    }

    pub fn update_docker_credential(
        &mut self,
        id: Uuid,
        mut updated: DockerCredential,
    ) -> Result<(), ConfigError> {
        for cred in &mut self.docker_credentials {
            if cred.id == id {
                updated.id = id;
                *cred = updated;
                return Ok(());
            }
        }
        Err(ConfigError::CredentialNotFound(id))
    }

    pub fn delete_docker_credential(&mut self, id: Uuid) -> Result<(), ConfigError> {
        let len_before = self.docker_credentials.len();
        self.docker_credentials.retain(|c| c.id != id);
        if self.docker_credentials.len() == len_before {
            return Err(ConfigError::CredentialNotFound(id));
        }
        Ok(())
    }

    pub fn list_docker_credentials(&self) -> &[DockerCredential] {
        &self.docker_credentials
    }
}

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    pub version: String,
    #[serde(default)]
    pub migration_version: i32,
    #[serde(default)]
    pub sites: Vec<Site>,
    #[serde(default)]
    pub domains: Vec<Domain>,
    #[serde(default)]
    pub nodes: Vec<Node>,
    pub settings: Settings,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            version: "1.0.0".to_string(),
            migration_version: 0,
            sites: Vec::new(),
            domains: Vec::new(),
            nodes: Vec::new(),
            settings: Settings::default(),
        }
    }
}

// ---------------------------------------------------------------------------
// DeletedSite
// ---------------------------------------------------------------------------

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeletedSite {
    pub site: Site,
    pub domain_name: String,
    pub archive_path: String,
    pub deleted_at: String,
}

// ---------------------------------------------------------------------------
// BackupInfo
// ---------------------------------------------------------------------------

#[derive(Debug, Clone)]
pub struct BackupInfo {
    pub name: String,
    pub path: PathBuf,
    pub size: u64,
    pub created_at: DateTime<Utc>,
}

// ---------------------------------------------------------------------------
// Migration system
// ---------------------------------------------------------------------------

const CURRENT_MIGRATION_VERSION: i32 = 1;

struct Migration {
    version: i32,
    description: &'static str,
    migrate: fn(&mut Config) -> Result<(), ConfigError>,
}

fn get_migrations() -> Vec<Migration> {
    vec![Migration {
        version: 1,
        description: "Migrate per-site Docker credentials to global credentials",
        migrate: migrate_docker_credentials,
    }]
}

fn migrate_docker_credentials(cfg: &mut Config) -> Result<(), ConfigError> {
    let mut credential_map: HashMap<String, Uuid> = HashMap::new();

    for site in &mut cfg.sites {
        if site.docker_username.is_empty() && site.docker_token.is_empty() {
            continue;
        }
        if let Some(ref cred_id) = site.docker_credential_id {
            if !cred_id.is_nil() {
                continue;
            }
        }

        let registry = "docker.io".to_string();
        let cred_key = format!("{}:{}", registry, site.docker_username);

        let cred_id = if let Some(&existing_id) = credential_map.get(&cred_key) {
            existing_id
        } else {
            let new_id = Uuid::new_v4();
            let new_cred = DockerCredential {
                id: new_id,
                name: format!("Migrated - {}", site.docker_username),
                registry: registry.clone(),
                username: site.docker_username.clone(),
                token: site.docker_token.clone(),
            };
            cfg.settings.docker_credentials.push(new_cred);
            credential_map.insert(cred_key, new_id);
            new_id
        };

        site.docker_credential_id = Some(cred_id);
        site.docker_username.clear();
        site.docker_token.clear();
    }

    Ok(())
}

fn migrate_config(cfg: &mut Config, config_path: &Path) -> Result<bool, ConfigError> {
    let current_version = cfg.migration_version;
    if current_version >= CURRENT_MIGRATION_VERSION {
        return Ok(false);
    }

    let migrations = get_migrations();
    let mut applied_any = false;

    for m in &migrations {
        if m.version > current_version {
            if !applied_any {
                backup_config(config_path)?;
            }
            (m.migrate)(cfg).map_err(|e| ConfigError::Migration {
                version: m.version,
                description: m.description.to_string(),
                source: Box::new(e),
            })?;
            applied_any = true;
        }
    }

    cfg.migration_version = CURRENT_MIGRATION_VERSION;
    Ok(applied_any)
}

fn backup_config(config_path: &Path) -> Result<(), ConfigError> {
    if !config_path.exists() {
        return Ok(());
    }

    let data = fs::read(config_path)?;
    let config_dir = config_path.parent().unwrap_or(Path::new("."));
    let backup_dir = config_dir.join("config_archives");
    fs::create_dir_all(&backup_dir)?;

    let timestamp = Utc::now().format("%Y-%m-%d_%H-%M-%S");
    let backup_name = format!("config_backup_{}.toml", timestamp);
    let backup_path = backup_dir.join(backup_name);

    fs::write(backup_path, data)?;
    Ok(())
}

pub fn list_backups(config_path: &Path) -> Result<Vec<BackupInfo>, ConfigError> {
    let config_dir = config_path.parent().unwrap_or(Path::new("."));
    let backup_dir = config_dir.join("config_archives");

    if !backup_dir.exists() {
        return Ok(Vec::new());
    }

    let mut backups = Vec::new();
    for entry in fs::read_dir(&backup_dir)? {
        let entry = entry?;
        let path = entry.path();
        if path.is_dir() || path.extension().and_then(|e| e.to_str()) != Some("toml") {
            continue;
        }
        let metadata = entry.metadata()?;
        let modified: DateTime<Utc> = metadata.modified()?.into();
        backups.push(BackupInfo {
            name: entry.file_name().to_string_lossy().into_owned(),
            path,
            size: metadata.len(),
            created_at: modified,
        });
    }
    Ok(backups)
}

pub fn restore_backup(backup_path: &Path, config_path: &Path) -> Result<(), ConfigError> {
    backup_config(config_path)?;
    let data = fs::read(backup_path)?;
    fs::write(config_path, data)?;
    Ok(())
}

// ---------------------------------------------------------------------------
// FileConfigLoader
// ---------------------------------------------------------------------------

pub struct FileConfigLoader;

impl FileConfigLoader {
    pub fn new() -> Self {
        Self
    }

    /// Returns the platform-specific default config path.
    pub fn default_config_path() -> Result<PathBuf, ConfigError> {
        let config_dir = get_archon_config_dir()?;
        Ok(config_dir.join("config.toml"))
    }

    /// Returns the base archon config directory.
    pub fn load(&self, path: &Path) -> Result<Config, ConfigError> {
        let mut config = if path.exists() {
            let data = fs::read_to_string(path)?;
            toml::from_str::<Config>(&data)?
        } else {
            Config::default()
        };

        // Load sites from directory structure
        if let Ok(sites) = self.load_all_sites() {
            if !sites.is_empty() {
                config.sites = sites;
            }
        }

        // Load nodes from directory structure
        if let Ok(nodes) = self.load_all_nodes() {
            if !nodes.is_empty() {
                config.nodes = nodes;
            }
        }

        // Ensure defaults
        if config.version.is_empty() {
            config.version = "1.0.0".to_string();
        }
        if !config.settings.auto_save && config.settings.health_check_interval_secs == 0 {
            config.settings = Settings::default();
        }

        // Run migrations
        let migrated = migrate_config(&mut config, path)?;
        if migrated {
            self.save(path, &config)?;
        }

        Ok(config)
    }

    pub fn save(&self, path: &Path, config: &Config) -> Result<(), ConfigError> {
        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)?;
        }

        // Save main config without sites/nodes (they're stored in directories)
        let legacy_config = Config {
            version: config.version.clone(),
            migration_version: config.migration_version,
            sites: Vec::new(),
            domains: config.domains.clone(),
            nodes: Vec::new(),
            settings: config.settings.clone(),
        };

        let data = toml::to_string(&legacy_config)?;
        fs::write(path, data)?;

        // Save each site to its directory
        for site in &config.sites {
            let domain_name = config
                .domains
                .iter()
                .find(|d| d.id == site.domain_id)
                .map(|d| d.name.as_str())
                .unwrap_or("unknown");

            let _ = self.save_site(site, domain_name);
        }

        // Save each node to its directory
        for node in &config.nodes {
            let _ = self.save_node(node);
        }

        Ok(())
    }

    pub fn save_site(&self, site: &Site, domain_name: &str) -> Result<(), ConfigError> {
        let base_dir = get_archon_config_dir()?;
        let site_path = base_dir
            .join("sites")
            .join(domain_name)
            .join(&site.name)
            .join("config.toml");

        if let Some(parent) = site_path.parent() {
            fs::create_dir_all(parent)?;
        }

        let data = toml::to_string(site)?;
        fs::write(site_path, data)?;
        Ok(())
    }

    pub fn load_all_sites(&self) -> Result<Vec<Site>, ConfigError> {
        let base_dir = get_archon_config_dir()?;
        let sites_dir = base_dir.join("sites");

        if !sites_dir.exists() {
            return Ok(Vec::new());
        }

        let mut sites = Vec::new();
        walk_config_files(&sites_dir, &mut |data| {
            if let Ok(site) = toml::from_str::<Site>(&data) {
                sites.push(site);
            }
        })?;

        Ok(sites)
    }

    pub fn save_node(&self, node: &Node) -> Result<(), ConfigError> {
        let base_dir = get_archon_config_dir()?;
        let node_path = base_dir
            .join("nodes")
            .join(&node.name)
            .join("config.toml");

        if let Some(parent) = node_path.parent() {
            fs::create_dir_all(parent)?;
        }

        let data = toml::to_string(node)?;
        fs::write(node_path, data)?;
        Ok(())
    }

    pub fn load_all_nodes(&self) -> Result<Vec<Node>, ConfigError> {
        let base_dir = get_archon_config_dir()?;
        let nodes_dir = base_dir.join("nodes");

        if !nodes_dir.exists() {
            return Ok(Vec::new());
        }

        // Regex to fix legacy datetime strings in TOML
        let datetime_re =
            Regex::new(r#"(last_health_check\s*=\s*)['"](\d{4}-\d{2}-\d{2}T[^'"]+)['"]"#)
                .unwrap();

        let mut nodes = Vec::new();
        walk_config_files(&nodes_dir, &mut |data| {
            let fixed = datetime_re.replace_all(&data, "${1}${2}").to_string();
            if let Ok(node) = toml::from_str::<Node>(&fixed) {
                nodes.push(node);
            }
        })?;

        Ok(nodes)
    }

    pub fn delete_site(&self, site_name: &str, domain_name: &str) -> Result<(), ConfigError> {
        let base_dir = get_archon_config_dir()?;
        let site_path = base_dir.join("sites").join(domain_name).join(site_name);
        if site_path.exists() {
            fs::remove_dir_all(site_path)?;
        }
        Ok(())
    }

    pub fn delete_node(&self, node_name: &str) -> Result<(), ConfigError> {
        let base_dir = get_archon_config_dir()?;
        let node_path = base_dir.join("nodes").join(node_name);
        if node_path.exists() {
            fs::remove_dir_all(node_path)?;
        }
        Ok(())
    }

    pub fn archive_site(
        &self,
        site_name: &str,
        domain_name: &str,
        site: &Site,
    ) -> Result<PathBuf, ConfigError> {
        let base_dir = get_archon_config_dir()?;
        self.archive_site_with_base_dir(&base_dir, site_name, domain_name, site)
    }

    pub fn archive_site_with_base_dir(
        &self,
        base_dir: &Path,
        site_name: &str,
        domain_name: &str,
        _site: &Site,
    ) -> Result<PathBuf, ConfigError> {
        let src_path = base_dir.join("sites").join(domain_name).join(site_name);
        if !src_path.exists() {
            return Err(ConfigError::Io(std::io::Error::new(
                std::io::ErrorKind::NotFound,
                format!("site directory not found: {}", src_path.display()),
            )));
        }

        let timestamp = Utc::now().format("%Y%m%d-%H%M%S").to_string();
        let archive_path = base_dir
            .join("deleted")
            .join(domain_name)
            .join(site_name)
            .join(&timestamp);

        fs::create_dir_all(&archive_path)?;

        let config_data = fs::read(src_path.join("config.toml"))?;
        fs::write(archive_path.join("config.toml"), config_data)?;

        fs::remove_dir_all(src_path)?;

        Ok(archive_path)
    }

    pub fn load_deleted_sites(&self) -> Result<Vec<DeletedSite>, ConfigError> {
        let base_dir = get_archon_config_dir()?;
        self.load_deleted_sites_with_base_dir(&base_dir)
    }

    pub fn load_deleted_sites_with_base_dir(
        &self,
        base_dir: &Path,
    ) -> Result<Vec<DeletedSite>, ConfigError> {
        let deleted_dir = base_dir.join("deleted");
        if !deleted_dir.exists() {
            return Ok(Vec::new());
        }

        let mut deleted_sites = Vec::new();

        for domain_entry in fs::read_dir(&deleted_dir)? {
            let domain_entry = domain_entry?;
            if !domain_entry.path().is_dir() {
                continue;
            }
            let domain_name = domain_entry.file_name().to_string_lossy().into_owned();

            for site_entry in fs::read_dir(domain_entry.path())? {
                let site_entry = site_entry?;
                if !site_entry.path().is_dir() {
                    continue;
                }

                for ts_entry in fs::read_dir(site_entry.path())? {
                    let ts_entry = ts_entry?;
                    if !ts_entry.path().is_dir() {
                        continue;
                    }

                    let archive_path = ts_entry.path();
                    let config_path = archive_path.join("config.toml");

                    if let Ok(data) = fs::read_to_string(&config_path) {
                        if let Ok(site) = toml::from_str::<Site>(&data) {
                            deleted_sites.push(DeletedSite {
                                site,
                                domain_name: domain_name.clone(),
                                archive_path: archive_path.to_string_lossy().into_owned(),
                                deleted_at: ts_entry
                                    .file_name()
                                    .to_string_lossy()
                                    .into_owned(),
                            });
                        }
                    }
                }
            }
        }

        Ok(deleted_sites)
    }

    pub fn restore_deleted_site(
        &self,
        archive_path: &str,
        site_name: &str,
        domain_name: &str,
    ) -> Result<(), ConfigError> {
        let base_dir = get_archon_config_dir()?;
        self.restore_deleted_site_with_base_dir(&base_dir, archive_path, site_name, domain_name)
    }

    pub fn restore_deleted_site_with_base_dir(
        &self,
        base_dir: &Path,
        archive_path: &str,
        site_name: &str,
        domain_name: &str,
    ) -> Result<(), ConfigError> {
        let archive = Path::new(archive_path);
        let config_path = archive.join("config.toml");
        if !config_path.exists() {
            return Err(ConfigError::Io(std::io::Error::new(
                std::io::ErrorKind::NotFound,
                "archived config.toml not found",
            )));
        }

        let dest_path = base_dir.join("sites").join(domain_name).join(site_name);
        fs::create_dir_all(&dest_path)?;

        let data = fs::read(&config_path)?;
        fs::write(dest_path.join("config.toml"), data)?;

        fs::remove_dir_all(archive)?;
        Ok(())
    }
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

pub fn get_archon_config_dir() -> Result<PathBuf, ConfigError> {
    if cfg!(target_os = "windows") {
        let appdata = std::env::var("APPDATA").unwrap_or_else(|_| {
            let home = dirs::home_dir().unwrap_or_default();
            home.join("AppData")
                .join("Roaming")
                .to_string_lossy()
                .into_owned()
        });
        Ok(PathBuf::from(appdata).join("archon"))
    } else {
        let config_dir = std::env::var("XDG_CONFIG_HOME").ok().map(PathBuf::from);
        let config_dir = config_dir.unwrap_or_else(|| {
            dirs::home_dir()
                .unwrap_or_default()
                .join(".config")
        });
        Ok(config_dir.join("archon"))
    }
}

/// Walks a directory tree looking for `config.toml` files and calls the callback
/// with the file contents.
fn walk_config_files(
    dir: &Path,
    callback: &mut dyn FnMut(String),
) -> Result<(), ConfigError> {
    if !dir.exists() {
        return Ok(());
    }

    for entry in fs::read_dir(dir)? {
        let entry = entry?;
        let path = entry.path();

        if path.is_dir() {
            walk_config_files(&path, callback)?;
        } else if path.file_name().and_then(|n| n.to_str()) == Some("config.toml") {
            if let Ok(data) = fs::read_to_string(&path) {
                callback(data);
            }
        }
    }

    Ok(())
}
