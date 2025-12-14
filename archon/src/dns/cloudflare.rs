use anyhow::{Context, Result};
use async_trait::async_trait;
use reqwest::Client;
use serde::{Deserialize, Serialize};

use crate::models::dns::{DnsRecord, DnsRecordType};

use super::provider::DnsProvider;

pub struct CloudflareProvider {
    client: Client,
    api_token: String,
    zone_id: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CloudflareResponse<T> {
    result: T,
    success: bool,
    errors: Vec<CloudflareError>,
}

#[derive(Debug, Serialize, Deserialize)]
struct CloudflareError {
    code: u32,
    message: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CloudflareDnsRecord {
    id: Option<String>,
    #[serde(rename = "type")]
    record_type: String,
    name: String,
    content: String,
    ttl: u32,
    proxied: bool,
}

impl CloudflareProvider {
    pub fn new(api_token: &str, zone_id: &str) -> Self {
        Self {
            client: Client::new(),
            api_token: api_token.to_string(),
            zone_id: zone_id.to_string(),
        }
    }

    fn records_url(&self) -> String {
        format!(
            "https://api.cloudflare.com/client/v4/zones/{}/dns_records",
            self.zone_id
        )
    }

    fn record_url(&self, record_id: &str) -> String {
        format!("{}/{}", self.records_url(), record_id)
    }
}

#[async_trait]
impl DnsProvider for CloudflareProvider {
    async fn list_records(&self, _domain: &str) -> Result<Vec<DnsRecord>> {
        let response = self
            .client
            .get(&self.records_url())
            .header("Authorization", format!("Bearer {}", self.api_token))
            .send()
            .await
            .context("Failed to fetch DNS records from Cloudflare")?;

        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
            anyhow::bail!("Cloudflare API error: {}", error_text);
        }

        let cf_response: CloudflareResponse<Vec<CloudflareDnsRecord>> = response
            .json()
            .await
            .context("Failed to parse Cloudflare response")?;

        if !cf_response.success {
            let errors: Vec<String> = cf_response
                .errors
                .iter()
                .map(|e| format!("{}: {}", e.code, e.message))
                .collect();
            anyhow::bail!("Cloudflare API errors: {}", errors.join(", "));
        }

        Ok(cf_response
            .result
            .into_iter()
            .filter_map(|cf_record| Self::from_cloudflare_record(cf_record).ok())
            .collect())
    }

    async fn create_record(&self, _domain: &str, record: &DnsRecord) -> Result<DnsRecord> {
        let cf_record = Self::to_cloudflare_record(record);

        let response = self
            .client
            .post(&self.records_url())
            .header("Authorization", format!("Bearer {}", self.api_token))
            .json(&cf_record)
            .send()
            .await
            .context("Failed to create DNS record in Cloudflare")?;

        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
            anyhow::bail!("Cloudflare API error: {}", error_text);
        }

        let cf_response: CloudflareResponse<CloudflareDnsRecord> = response
            .json()
            .await
            .context("Failed to parse Cloudflare response")?;

        if !cf_response.success {
            let errors: Vec<String> = cf_response
                .errors
                .iter()
                .map(|e| format!("{}: {}", e.code, e.message))
                .collect();
            anyhow::bail!("Cloudflare API errors: {}", errors.join(", "));
        }

        Self::from_cloudflare_record(cf_response.result)
    }

    async fn update_record(&self, _domain: &str, record: &DnsRecord) -> Result<DnsRecord> {
        let record_id = record
            .id
            .as_ref()
            .context("Cannot update record without ID")?;

        let cf_record = Self::to_cloudflare_record(record);

        let response = self
            .client
            .put(&self.record_url(record_id))
            .header("Authorization", format!("Bearer {}", self.api_token))
            .json(&cf_record)
            .send()
            .await
            .context("Failed to update DNS record in Cloudflare")?;

        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
            anyhow::bail!("Cloudflare API error: {}", error_text);
        }

        let cf_response: CloudflareResponse<CloudflareDnsRecord> = response
            .json()
            .await
            .context("Failed to parse Cloudflare response")?;

        if !cf_response.success {
            let errors: Vec<String> = cf_response
                .errors
                .iter()
                .map(|e| format!("{}: {}", e.code, e.message))
                .collect();
            anyhow::bail!("Cloudflare API errors: {}", errors.join(", "));
        }

        Self::from_cloudflare_record(cf_response.result)
    }

    async fn delete_record(&self, _domain: &str, record_id: &str) -> Result<()> {
        let response = self
            .client
            .delete(&self.record_url(record_id))
            .header("Authorization", format!("Bearer {}", self.api_token))
            .send()
            .await
            .context("Failed to delete DNS record from Cloudflare")?;

        if !response.status().is_success() {
            let error_text = response.text().await.unwrap_or_else(|_| "Unknown error".to_string());
            anyhow::bail!("Cloudflare API error: {}", error_text);
        }

        let cf_response: CloudflareResponse<serde_json::Value> = response
            .json()
            .await
            .context("Failed to parse Cloudflare response")?;

        if !cf_response.success {
            let errors: Vec<String> = cf_response
                .errors
                .iter()
                .map(|e| format!("{}: {}", e.code, e.message))
                .collect();
            anyhow::bail!("Cloudflare API errors: {}", errors.join(", "));
        }

        Ok(())
    }
}

impl CloudflareProvider {
    fn from_cloudflare_record(cf_record: CloudflareDnsRecord) -> Result<DnsRecord> {
        let record_type = cf_record.record_type.parse::<DnsRecordType>()
            .map_err(|e| anyhow::anyhow!("Invalid DNS record type from Cloudflare: {}", e))?;

        Ok(DnsRecord {
            id: cf_record.id,
            record_type,
            name: cf_record.name,
            value: cf_record.content,
            ttl: cf_record.ttl,
            proxied: cf_record.proxied,
        })
    }

    fn to_cloudflare_record(record: &DnsRecord) -> CloudflareDnsRecord {
        CloudflareDnsRecord {
            id: record.id.clone(),
            record_type: record.record_type.to_string(),
            name: record.name.clone(),
            content: record.value.clone(),
            ttl: record.ttl,
            proxied: record.proxied,
        }
    }
}
