use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use super::DnsRecord;

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum DnsProviderType {
    Cloudflare,
    Route53,
    Manual,
}

impl std::fmt::Display for DnsProviderType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Cloudflare => write!(f, "cloudflare"),
            Self::Route53 => write!(f, "route53"),
            Self::Manual => write!(f, "manual"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DnsProvider {
    #[serde(rename = "type")]
    pub provider_type: DnsProviderType,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub api_token: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub zone_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub access_key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub secret_key: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub hosted_zone_id: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Domain {
    pub id: Uuid,
    pub name: String,
    pub dns_provider: DnsProvider,
    pub dns_records: Vec<DnsRecord>,
    pub traefik_enabled: bool,
    pub created_at: DateTime<Utc>,
}

impl Domain {
    pub fn new(name: String, provider: DnsProvider) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            dns_provider: provider,
            dns_records: Vec::new(),
            traefik_enabled: false,
            created_at: Utc::now(),
        }
    }

    pub fn is_manual_dns(&self) -> bool {
        self.dns_provider.provider_type == DnsProviderType::Manual
    }

    pub fn provider_name(&self) -> &str {
        match self.dns_provider.provider_type {
            DnsProviderType::Cloudflare => "Cloudflare",
            DnsProviderType::Route53 => "Route53",
            DnsProviderType::Manual => "Manual",
        }
    }
}
