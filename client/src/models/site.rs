use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum SiteStatus {
    Inactive,
    Deploying,
    Running,
    Failed,
    Stopped,
}

impl std::fmt::Display for SiteStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Inactive => write!(f, "inactive"),
            Self::Deploying => write!(f, "deploying"),
            Self::Running => write!(f, "running"),
            Self::Failed => write!(f, "failed"),
            Self::Stopped => write!(f, "stopped"),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum SiteType {
    Container,
    Compose,
}

impl std::fmt::Display for SiteType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Container => write!(f, "container"),
            Self::Compose => write!(f, "compose"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigFile {
    pub name: String,
    pub content: String,
    pub container_path: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Volume {
    pub host_path: String,
    pub container_path: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DomainMapping {
    pub domain_id: Uuid,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub subdomain: String,
    pub port: i32,
    #[serde(default, skip_serializing_if = "is_zero")]
    pub host_port: i32,
}

fn is_zero(v: &i32) -> bool {
    *v == 0
}

impl DomainMapping {
    pub fn get_effective_host_port(&self) -> i32 {
        if self.host_port > 0 {
            self.host_port
        } else {
            self.port
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Site {
    pub id: Uuid,
    pub name: String,
    pub site_type: SiteType,
    pub domain_id: Uuid,
    pub node_id: Uuid,
    pub docker_image: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub docker_credential_id: Option<Uuid>,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub docker_username: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub docker_token: String,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub compose_content: String,
    pub environment_vars: HashMap<String, String>,
    pub port: i32,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub domain_mappings: Vec<DomainMapping>,
    pub ssl_enabled: bool,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub ssl_email: String,
    pub config_files: Vec<ConfigFile>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub volumes: Vec<Volume>,
    #[serde(default)]
    pub bot_redirect_enabled: bool,
    #[serde(default, skip_serializing_if = "String::is_empty")]
    pub bot_redirect_url: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    pub bot_user_agents: Vec<String>,
    pub status: SiteStatus,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
}

impl Site {
    pub fn new(
        name: String,
        domain_id: Uuid,
        node_id: Uuid,
        docker_image: String,
        port: i32,
    ) -> Self {
        let now = Utc::now();
        Self {
            id: Uuid::new_v4(),
            name,
            site_type: SiteType::Container,
            domain_id,
            node_id,
            docker_image,
            docker_credential_id: None,
            docker_username: String::new(),
            docker_token: String::new(),
            compose_content: String::new(),
            environment_vars: HashMap::new(),
            port,
            domain_mappings: Vec::new(),
            ssl_enabled: false,
            ssl_email: String::new(),
            config_files: Vec::new(),
            volumes: Vec::new(),
            bot_redirect_enabled: false,
            bot_redirect_url: String::new(),
            bot_user_agents: Vec::new(),
            status: SiteStatus::Inactive,
            created_at: now,
            updated_at: now,
        }
    }

    pub fn get_domain_mappings(&self) -> &[DomainMapping] {
        &self.domain_mappings
    }

    pub fn add_domain_mapping(&mut self, domain_id: Uuid, port: i32) {
        self.domain_mappings.push(DomainMapping {
            domain_id,
            subdomain: String::new(),
            port,
            host_port: 0,
        });
    }

    pub fn remove_domain_mapping(&mut self, index: usize) {
        if index < self.domain_mappings.len() {
            self.domain_mappings.remove(index);
        }
    }

    pub fn is_compose(&self) -> bool {
        self.site_type == SiteType::Compose
    }

    pub fn get_site_type(&self) -> &SiteType {
        &self.site_type
    }

    /// Generate Traefik labels for the site's domain mappings.
    pub fn generate_traefik_labels(&self, domain_name: &str) -> HashMap<String, String> {
        let mut labels = HashMap::new();

        labels.insert(
            "traefik.enable".to_string(),
            "true".to_string(),
        );

        for (i, mapping) in self.domain_mappings.iter().enumerate() {
            let full_domain = get_full_domain(domain_name, &mapping.subdomain);
            let router_name = if i == 0 {
                self.name.clone()
            } else {
                format!("{}-{}", self.name, i)
            };

            labels.insert(
                format!("traefik.http.routers.{}.rule", router_name),
                format!("Host(`{}`)", full_domain),
            );

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

            labels.insert(
                format!(
                    "traefik.http.services.{}.loadbalancer.server.port",
                    router_name
                ),
                mapping.port.to_string(),
            );
        }

        labels
    }
}

pub fn get_full_domain(domain_name: &str, subdomain: &str) -> String {
    if subdomain.is_empty() {
        domain_name.to_string()
    } else {
        format!("{}.{}", subdomain, domain_name)
    }
}

pub fn parse_port_mapping(port_str: &str) -> Result<(i32, i32), String> {
    let parts: Vec<&str> = port_str.split(':').collect();
    match parts.len() {
        1 => {
            let port: i32 = parts[0]
                .parse()
                .map_err(|_| format!("invalid port: {}", parts[0]))?;
            Ok((port, 0))
        }
        2 => {
            let host_port: i32 = parts[0]
                .parse()
                .map_err(|_| format!("invalid host port: {}", parts[0]))?;
            let container_port: i32 = parts[1]
                .parse()
                .map_err(|_| format!("invalid container port: {}", parts[1]))?;
            Ok((container_port, host_port))
        }
        _ => Err(format!("invalid port mapping format: {}", port_str)),
    }
}

pub fn format_port_mapping(container_port: i32, host_port: i32) -> String {
    if host_port > 0 && host_port != container_port {
        format!("{}:{}", host_port, container_port)
    } else {
        container_port.to_string()
    }
}
