use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use super::dns::DnsRecord;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Domain {
    pub id: Uuid,
    pub name: String,
    pub dns_provider: DnsProvider,
    pub dns_records: Vec<DnsRecord>,
    pub traefik_enabled: bool,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum DnsProvider {
    Cloudflare {
        api_token: String,
        zone_id: String,
    },
    Route53 {
        access_key: String,
        secret_key: String,
        hosted_zone_id: String,
    },
    Manual,
}

impl Domain {
    pub fn new(name: String, dns_provider: DnsProvider) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            dns_provider,
            dns_records: Vec::new(),
            traefik_enabled: true,
            created_at: Utc::now(),
        }
    }

    pub fn is_manual_dns(&self) -> bool {
        matches!(self.dns_provider, DnsProvider::Manual)
    }

    pub fn provider_name(&self) -> &'static str {
        match self.dns_provider {
            DnsProvider::Cloudflare { .. } => "Cloudflare",
            DnsProvider::Route53 { .. } => "Route53",
            DnsProvider::Manual => "Manual",
        }
    }
}

impl DnsProvider {
    pub fn is_manual(&self) -> bool {
        matches!(self, DnsProvider::Manual)
    }
}
