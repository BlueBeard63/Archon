use floem::reactive::{RwSignal, SignalGet, SignalUpdate, SignalWith};
use uuid::Uuid;

use crate::config::{Config, FileConfigLoader, Settings};
use crate::models::{Domain, Node, Site};

use super::forms::FormState;
use super::navigation::NavigationState;
use super::notifications::NotificationState;

/// Central application state, composed of reactive sub-states.
pub struct AppState {
    // --- Data ---
    pub sites: RwSignal<Vec<Site>>,
    pub domains: RwSignal<Vec<Domain>>,
    pub nodes: RwSignal<Vec<Node>>,

    // --- Selection indices ---
    pub sites_list_index: RwSignal<usize>,
    pub domains_list_index: RwSignal<usize>,
    pub nodes_list_index: RwSignal<usize>,
    pub docker_credentials_list_index: RwSignal<usize>,

    // --- Selected entity IDs (for edit screens) ---
    pub selected_site_id: RwSignal<Option<Uuid>>,
    pub selected_domain_id: RwSignal<Option<Uuid>>,
    pub selected_node_id: RwSignal<Option<Uuid>>,
    pub selected_docker_credential_id: RwSignal<Option<Uuid>>,

    // --- Sub-states ---
    pub navigation: NavigationState,
    pub notifications: NotificationState,
    pub form: FormState,

    // --- Config ---
    pub config_path: RwSignal<String>,
    pub settings: RwSignal<Settings>,

    // --- Async status ---
    pub force_refresh_in_progress: RwSignal<bool>,
    pub force_refresh_total: RwSignal<usize>,
    pub force_refresh_completed: RwSignal<usize>,

    // --- Quick Config ---
    pub quick_config_url: RwSignal<String>,
    pub quick_config_expires_at: RwSignal<String>,
    pub quick_config_node_id: RwSignal<Option<Uuid>>,
}

impl AppState {
    /// Create a new AppState, loading config from disk.
    pub fn new() -> Self {
        let loader = FileConfigLoader::new();
        let config_path = FileConfigLoader::default_config_path()
            .unwrap_or_default()
            .to_string_lossy()
            .into_owned();

        let config = loader
            .load(std::path::Path::new(&config_path))
            .unwrap_or_default();

        Self::from_config(config, config_path)
    }

    fn from_config(config: Config, config_path: String) -> Self {
        Self {
            sites: RwSignal::new(config.sites),
            domains: RwSignal::new(config.domains),
            nodes: RwSignal::new(config.nodes),

            sites_list_index: RwSignal::new(0),
            domains_list_index: RwSignal::new(0),
            nodes_list_index: RwSignal::new(0),
            docker_credentials_list_index: RwSignal::new(0),

            selected_site_id: RwSignal::new(None),
            selected_domain_id: RwSignal::new(None),
            selected_node_id: RwSignal::new(None),
            selected_docker_credential_id: RwSignal::new(None),

            navigation: NavigationState::new(),
            notifications: NotificationState::new(),
            form: FormState::new(),

            config_path: RwSignal::new(config_path),
            settings: RwSignal::new(config.settings),

            force_refresh_in_progress: RwSignal::new(false),
            force_refresh_total: RwSignal::new(0),
            force_refresh_completed: RwSignal::new(0),

            quick_config_url: RwSignal::new(String::new()),
            quick_config_expires_at: RwSignal::new(String::new()),
            quick_config_node_id: RwSignal::new(None),
        }
    }

    /// Build the current Config from reactive state for saving.
    pub fn to_config(&self) -> Config {
        Config {
            version: "1.0.0".to_string(),
            migration_version: 1,
            sites: self.sites.get_untracked(),
            domains: self.domains.get_untracked(),
            nodes: self.nodes.get_untracked(),
            settings: self.settings.get_untracked(),
        }
    }

    /// Save current state to disk.
    pub fn save(&self) -> Result<(), crate::config::ConfigError> {
        let loader = FileConfigLoader::new();
        let config = self.to_config();
        let path = self.config_path.get_untracked();
        loader.save(std::path::Path::new(&path), &config)
    }

    // --- Lookup helpers ---

    pub fn find_site(&self, id: Uuid) -> Option<Site> {
        self.sites
            .with_untracked(|sites| sites.iter().find(|s| s.id == id).cloned())
    }

    pub fn find_domain(&self, id: Uuid) -> Option<Domain> {
        self.domains
            .with_untracked(|domains| domains.iter().find(|d| d.id == id).cloned())
    }

    pub fn find_node(&self, id: Uuid) -> Option<Node> {
        self.nodes
            .with_untracked(|nodes| nodes.iter().find(|n| n.id == id).cloned())
    }

    pub fn domain_name_for_site(&self, site: &Site) -> String {
        self.domains.with_untracked(|domains| {
            // Try to find domain by the first domain mapping
            if let Some(mapping) = site.domain_mappings.first() {
                if let Some(d) = domains.iter().find(|d| d.id == mapping.domain_id) {
                    return d.name.clone();
                }
            }
            // Fall back to legacy domain_id
            domains
                .iter()
                .find(|d| d.id == site.domain_id)
                .map(|d| d.name.clone())
                .unwrap_or_else(|| "unknown".to_string())
        })
    }

    pub fn node_for_site(&self, site: &Site) -> Option<Node> {
        self.find_node(site.node_id)
    }
}

#[cfg(test)]
impl AppState {
    /// Create a test AppState with empty data and no disk I/O.
    pub fn new_test() -> Self {
        Self::from_config(Config::default(), String::new())
    }
}
