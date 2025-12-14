use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Site {
    pub id: Uuid,
    pub name: String,
    pub domain_id: Uuid,
    pub node_id: Uuid,
    pub docker_image: String,
    pub environment_vars: HashMap<String, String>,
    pub port: u16,
    pub ssl_enabled: bool,
    pub config_files: Vec<ConfigFile>,
    pub status: SiteStatus,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigFile {
    pub name: String,
    pub content: String,
    pub container_path: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum SiteStatus {
    Inactive,
    Deploying,
    Running,
    Failed,
    Stopped,
}

impl Site {
    pub fn new(
        name: String,
        domain_id: Uuid,
        node_id: Uuid,
        docker_image: String,
        port: u16,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            name,
            domain_id,
            node_id,
            docker_image,
            environment_vars: HashMap::new(),
            port,
            ssl_enabled: true, // Default to SSL enabled
            config_files: Vec::new(),
            status: SiteStatus::Inactive,
            created_at: now,
            updated_at: now,
        }
    }

    /// Generate Docker labels for Traefik reverse proxy configuration
    pub fn generate_traefik_labels(&self, domain_name: &str) -> HashMap<String, String> {
        let mut labels = HashMap::new();

        // Enable Traefik
        labels.insert("traefik.enable".to_string(), "true".to_string());

        // Router configuration
        let router_name = format!("site-{}", self.id);

        // Routing rule
        labels.insert(
            format!("traefik.http.routers.{}.rule", router_name),
            format!("Host(`{}`)", domain_name),
        );

        // Entry points
        let entrypoint = if self.ssl_enabled { "websecure" } else { "web" };
        labels.insert(
            format!("traefik.http.routers.{}.entrypoints", router_name),
            entrypoint.to_string(),
        );

        // SSL/TLS configuration
        if self.ssl_enabled {
            labels.insert(
                format!("traefik.http.routers.{}.tls", router_name),
                "true".to_string(),
            );
            labels.insert(
                format!("traefik.http.routers.{}.tls.certresolver", router_name),
                "letsencrypt".to_string(),
            );
        }

        // Service configuration
        labels.insert(
            format!("traefik.http.services.{}.loadbalancer.server.port", router_name),
            self.port.to_string(),
        );

        labels
    }
}

impl ConfigFile {
    pub fn new(name: String, container_path: String) -> Self {
        Self {
            name,
            content: String::new(),
            container_path,
        }
    }

    pub fn with_content(name: String, content: String, container_path: String) -> Self {
        Self {
            name,
            content,
            container_path,
        }
    }
}

impl std::fmt::Display for SiteStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            SiteStatus::Inactive => write!(f, "Inactive"),
            SiteStatus::Deploying => write!(f, "Deploying"),
            SiteStatus::Running => write!(f, "Running"),
            SiteStatus::Failed => write!(f, "Failed"),
            SiteStatus::Stopped => write!(f, "Stopped"),
        }
    }
}
