use anyhow::Result;
use chrono::{DateTime, Utc};
use std::collections::VecDeque;
use std::path::PathBuf;
use uuid::Uuid;

use crate::config::{load_config, save_config, ArchonConfig};
use crate::models::{Domain, Node, Site};
use crate::ui::Screen;

use super::async_ops::AsyncOperation;
use super::selection::SelectionState;

#[derive(Debug, Clone)]
pub struct AppState {
    // Core data
    pub sites: Vec<Site>,
    pub domains: Vec<Domain>,
    pub nodes: Vec<Node>,

    // UI state
    pub current_screen: Screen,
    pub previous_screen: Vec<Screen>,
    pub selection_state: SelectionState,

    // Async operations tracking
    pub pending_operations: Vec<AsyncOperation>,
    pub notifications: VecDeque<Notification>,

    // Configuration
    pub config_path: PathBuf,
    pub auto_save: bool,

    // Application control
    pub should_quit: bool,
}

#[derive(Debug, Clone)]
pub struct Notification {
    pub message: String,
    pub level: NotificationLevel,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum NotificationLevel {
    Info,
    Success,
    Warning,
    Error,
}

impl AppState {
    pub fn new(config_path: PathBuf) -> Result<Self> {
        let config = load_config(&config_path)?;

        Ok(Self {
            sites: config.sites,
            domains: config.domains,
            nodes: config.nodes,
            current_screen: Screen::default(),
            previous_screen: Vec::new(),
            selection_state: SelectionState::default(),
            pending_operations: Vec::new(),
            notifications: VecDeque::new(),
            config_path,
            auto_save: config.settings.auto_save,
            should_quit: false,
        })
    }

    pub fn save(&self) -> Result<()> {
        let config = ArchonConfig {
            version: env!("CARGO_PKG_VERSION").to_string(),
            sites: self.sites.clone(),
            domains: self.domains.clone(),
            nodes: self.nodes.clone(),
            settings: crate::config::Settings {
                auto_save: self.auto_save,
                ..Default::default()
            },
        };

        save_config(&self.config_path, &config)
    }

    // Navigation methods
    pub fn navigate_to(&mut self, screen: Screen) {
        self.previous_screen.push(self.current_screen.clone());
        self.current_screen = screen;
    }

    pub fn navigate_back(&mut self) {
        if let Some(previous) = self.previous_screen.pop() {
            self.current_screen = previous;
        }
    }

    // Notification methods
    pub fn add_notification(&mut self, message: String, level: NotificationLevel) {
        self.notifications.push_back(Notification {
            message,
            level,
            timestamp: Utc::now(),
        });

        // Keep only the last 50 notifications
        while self.notifications.len() > 50 {
            self.notifications.pop_front();
        }
    }

    pub fn dismiss_notification(&mut self) {
        self.notifications.pop_front();
    }

    // Helper methods for finding entities
    pub fn get_site(&self, id: Uuid) -> Option<&Site> {
        self.sites.iter().find(|s| s.id == id)
    }

    pub fn get_site_mut(&mut self, id: Uuid) -> Option<&mut Site> {
        self.sites.iter_mut().find(|s| s.id == id)
    }

    pub fn get_domain(&self, id: Uuid) -> Option<&Domain> {
        self.domains.iter().find(|d| d.id == id)
    }

    pub fn get_domain_mut(&mut self, id: Uuid) -> Option<&mut Domain> {
        self.domains.iter_mut().find(|d| d.id == id)
    }

    pub fn get_node(&self, id: Uuid) -> Option<&Node> {
        self.nodes.iter().find(|n| n.id == id)
    }

    pub fn get_node_mut(&mut self, id: Uuid) -> Option<&mut Node> {
        self.nodes.iter_mut().find(|n| n.id == id)
    }

    pub fn get_node_for_site(&self, site_id: Uuid) -> Option<&Node> {
        let site = self.get_site(site_id)?;
        self.get_node(site.node_id)
    }

    pub fn get_domain_for_site(&self, site_id: Uuid) -> Option<&Domain> {
        let site = self.get_site(site_id)?;
        self.get_domain(site.domain_id)
    }

    // CRUD operations
    pub fn add_site(&mut self, site: Site) {
        self.sites.push(site);
    }

    pub fn remove_site(&mut self, id: Uuid) {
        self.sites.retain(|s| s.id != id);
    }

    pub fn add_domain(&mut self, domain: Domain) {
        self.domains.push(domain);
    }

    pub fn remove_domain(&mut self, id: Uuid) {
        self.domains.retain(|d| d.id != id);
    }

    pub fn add_node(&mut self, node: Node) {
        self.nodes.push(node);
    }

    pub fn remove_node(&mut self, id: Uuid) {
        self.nodes.retain(|n| n.id != id);
    }
}

impl Notification {
    pub fn info(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
            level: NotificationLevel::Info,
            timestamp: Utc::now(),
        }
    }

    pub fn success(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
            level: NotificationLevel::Success,
            timestamp: Utc::now(),
        }
    }

    pub fn warning(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
            level: NotificationLevel::Warning,
            timestamp: Utc::now(),
        }
    }

    pub fn error(message: impl Into<String>) -> Self {
        Self {
            message: message.into(),
            level: NotificationLevel::Error,
            timestamp: Utc::now(),
        }
    }
}
