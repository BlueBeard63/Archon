use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::net::IpAddr;
use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum NodeStatus {
    Unknown,
    Online,
    Offline,
    Degraded,
}

impl std::fmt::Display for NodeStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Unknown => write!(f, "unknown"),
            Self::Online => write!(f, "online"),
            Self::Offline => write!(f, "offline"),
            Self::Degraded => write!(f, "degraded"),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum ProxyType {
    Nginx,
    Apache,
    Traefik,
}

impl std::fmt::Display for ProxyType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Nginx => write!(f, "nginx"),
            Self::Apache => write!(f, "apache"),
            Self::Traefik => write!(f, "traefik"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DockerInfo {
    pub version: String,
    pub containers_running: i32,
    pub images_count: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TraefikInfo {
    pub version: String,
    pub routers_count: i32,
    pub services_count: i32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Node {
    pub id: Uuid,
    pub name: String,
    pub api_endpoint: String,
    pub api_key: String,
    pub ip_address: IpAddr,
    pub proxy_type: ProxyType,
    pub status: NodeStatus,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub docker_info: Option<DockerInfo>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub traefik_info: Option<TraefikInfo>,
    pub last_health_check: DateTime<Utc>,
}

impl Node {
    pub fn new(
        name: String,
        api_endpoint: String,
        api_key: String,
        ip_address: IpAddr,
        proxy_type: ProxyType,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            api_endpoint,
            api_key,
            ip_address,
            proxy_type,
            status: NodeStatus::Unknown,
            docker_info: None,
            traefik_info: None,
            last_health_check: Utc::now(),
        }
    }
}
