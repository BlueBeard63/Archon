use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::models::dns::DnsRecord;
use crate::models::node::{DockerInfo, NodeStatus, TraefikInfo};

#[derive(Debug, Clone)]
pub struct AsyncOperation {
    pub id: Uuid,
    pub operation_type: OperationType,
    pub status: AsyncStatus,
    pub started_at: DateTime<Utc>,
}

#[derive(Debug, Clone)]
pub enum OperationType {
    DeploySite(Uuid),
    UpdateSite(Uuid),
    DeleteSite(Uuid),
    UpdateDns(Uuid),
    NodeHealthCheck(Uuid),
    FetchNodeStats(Uuid),
    FetchLogs(Uuid),
    FetchMetrics(Uuid),
}

#[derive(Debug, Clone)]
pub enum AsyncStatus {
    InProgress,
    Completed,
    Failed(String),
}

#[derive(Debug, Clone)]
pub enum AsyncOperationResult {
    SiteDeployed(Uuid),
    SiteUpdated(Uuid),
    SiteDeleted(Uuid),
    DnsSynced(Uuid, Vec<DnsRecord>),
    NodeHealth(Uuid, NodeStatus, Option<DockerInfo>, Option<TraefikInfo>),
    NodeStats(Uuid, DockerInfo, TraefikInfo),
    Logs(Uuid, Vec<String>),
    Metrics(Uuid, ContainerMetrics),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ContainerMetrics {
    pub cpu_usage_percent: f64,
    pub memory_usage_mb: u64,
    pub memory_limit_mb: u64,
    pub network_rx_bytes: u64,
    pub network_tx_bytes: u64,
}

impl AsyncOperation {
    pub fn new(operation_type: OperationType) -> Self {
        Self {
            id: Uuid::new_v4(),
            operation_type,
            status: AsyncStatus::InProgress,
            started_at: Utc::now(),
        }
    }

    pub fn complete(&mut self) {
        self.status = AsyncStatus::Completed;
    }

    pub fn fail(&mut self, error: String) {
        self.status = AsyncStatus::Failed(error);
    }

    pub fn is_completed(&self) -> bool {
        matches!(self.status, AsyncStatus::Completed)
    }

    pub fn is_failed(&self) -> bool {
        matches!(self.status, AsyncStatus::Failed(_))
    }
}
