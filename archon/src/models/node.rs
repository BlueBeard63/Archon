use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::net::IpAddr;
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Node {
    pub id: Uuid,
    pub name: String,
    pub api_endpoint: String,
    pub api_key: String,
    pub ip_address: IpAddr,
    pub status: NodeStatus,
    pub docker_info: Option<DockerInfo>,
    pub traefik_info: Option<TraefikInfo>,
    pub last_health_check: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DockerInfo {
    pub version: String,
    pub containers_running: u32,
    pub images_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TraefikInfo {
    pub version: String,
    pub routers_count: u32,
    pub services_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum NodeStatus {
    Unknown,
    Online,
    Offline,
    Degraded,
}

impl Node {
    pub fn new(name: String, api_endpoint: String, api_key: String, ip_address: IpAddr) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            api_endpoint,
            api_key,
            ip_address,
            status: NodeStatus::Unknown,
            docker_info: None,
            traefik_info: None,
            last_health_check: None,
        }
    }

    pub fn update_health(
        &mut self,
        status: NodeStatus,
        docker_info: Option<DockerInfo>,
        traefik_info: Option<TraefikInfo>,
    ) {
        self.status = status;
        self.docker_info = docker_info;
        self.traefik_info = traefik_info;
        self.last_health_check = Some(Utc::now());
    }
}

impl std::fmt::Display for NodeStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            NodeStatus::Unknown => write!(f, "Unknown"),
            NodeStatus::Online => write!(f, "Online"),
            NodeStatus::Offline => write!(f, "Offline"),
            NodeStatus::Degraded => write!(f, "Degraded"),
        }
    }
}

impl DockerInfo {
    pub fn new(version: String, containers_running: u32, images_count: u32) -> Self {
        Self {
            version,
            containers_running,
            images_count,
        }
    }
}

impl TraefikInfo {
    pub fn new(version: String, routers_count: u32, services_count: u32) -> Self {
        Self {
            version,
            routers_count,
            services_count,
        }
    }
}
