use anyhow::Result;
use async_trait::async_trait;

use crate::models::dns::DnsRecord;
use crate::models::domain::DnsProvider as DnsProviderConfig;

use super::cloudflare::CloudflareProvider;

#[async_trait]
pub trait DnsProvider: Send + Sync {
    /// List all DNS records for a domain
    async fn list_records(&self, domain: &str) -> Result<Vec<DnsRecord>>;

    /// Create a new DNS record
    async fn create_record(&self, domain: &str, record: &DnsRecord) -> Result<DnsRecord>;

    /// Update an existing DNS record
    async fn update_record(&self, domain: &str, record: &DnsRecord) -> Result<DnsRecord>;

    /// Delete a DNS record
    async fn delete_record(&self, domain: &str, record_id: &str) -> Result<()>;
}

/// Factory function to create a DNS provider from configuration
pub fn create_provider(config: &DnsProviderConfig) -> Result<Box<dyn DnsProvider>> {
    match config {
        DnsProviderConfig::Cloudflare { api_token, zone_id } => {
            Ok(Box::new(CloudflareProvider::new(api_token, zone_id)))
        }
        DnsProviderConfig::Route53 { .. } => {
            anyhow::bail!("Route53 provider not yet implemented")
        }
        DnsProviderConfig::Manual => {
            anyhow::bail!("Manual DNS provider does not support API operations")
        }
    }
}
