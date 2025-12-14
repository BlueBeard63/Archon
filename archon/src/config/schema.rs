use serde::{Deserialize, Serialize};

use crate::models::{Domain, Node, Site};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ArchonConfig {
    pub version: String,
    pub sites: Vec<Site>,
    pub domains: Vec<Domain>,
    pub nodes: Vec<Node>,
    pub settings: Settings,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Settings {
    pub auto_save: bool,
    pub health_check_interval_seconds: u64,
    pub default_dns_ttl: u32,
    pub theme: String,
}

impl Default for ArchonConfig {
    fn default() -> Self {
        Self {
            version: env!("CARGO_PKG_VERSION").to_string(),
            sites: Vec::new(),
            domains: Vec::new(),
            nodes: Vec::new(),
            settings: Settings::default(),
        }
    }
}

impl Default for Settings {
    fn default() -> Self {
        Self {
            auto_save: true,
            health_check_interval_seconds: 300, // 5 minutes
            default_dns_ttl: 300,               // 5 minutes
            theme: "default".to_string(),
        }
    }
}

impl ArchonConfig {
    pub fn new() -> Self {
        Self::default()
    }
}
