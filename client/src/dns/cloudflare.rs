use reqwest::header::{HeaderMap, HeaderValue, AUTHORIZATION, CONTENT_TYPE};
use serde::{Deserialize, Serialize};
use std::time::Duration;
use thiserror::Error;

use crate::models::{DnsRecord, DnsRecordType};

const CLOUDFLARE_API_BASE: &str = "https://api.cloudflare.com/client/v4";

// ---------------------------------------------------------------------------
// Error type
// ---------------------------------------------------------------------------

#[derive(Debug, Error)]
pub enum DnsError {
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),
    #[error("Cloudflare API error: {0}")]
    CloudflareApi(String),
    #[error("record ID is required for {0}")]
    MissingRecordId(&'static str),
    #[error("{0}")]
    Other(String),
}

// ---------------------------------------------------------------------------
// Cloudflare wire types
// ---------------------------------------------------------------------------

#[derive(Debug, Deserialize)]
struct CloudflareResponse {
    success: bool,
    errors: Vec<CloudflareError>,
    result: serde_json::Value,
}

#[derive(Debug, Deserialize)]
struct CloudflareError {
    #[allow(dead_code)]
    code: i32,
    message: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct CloudflareRecord {
    #[serde(skip_serializing_if = "Option::is_none")]
    id: Option<String>,
    #[serde(rename = "type")]
    record_type: String,
    name: String,
    content: String,
    ttl: i32,
    proxied: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    comment: Option<String>,
}

fn to_cloudflare_record(record: &DnsRecord, tags: &[String]) -> CloudflareRecord {
    let mut cf = CloudflareRecord {
        id: record.id.clone(),
        record_type: record.record_type.to_string(),
        name: record.name.clone(),
        content: record.value.clone(),
        ttl: if record.ttl == 0 { 300 } else { record.ttl },
        proxied: record.proxied,
        comment: tags.first().cloned(),
    };
    // Don't send ID on create
    if cf.id.as_deref() == Some("") {
        cf.id = None;
    }
    cf
}

fn from_cloudflare_record(cf: CloudflareRecord) -> DnsRecord {
    let record_type = match cf.record_type.as_str() {
        "A" => DnsRecordType::A,
        "AAAA" => DnsRecordType::AAAA,
        "CNAME" => DnsRecordType::CNAME,
        "MX" => DnsRecordType::MX,
        "TXT" => DnsRecordType::TXT,
        "SRV" => DnsRecordType::SRV,
        _ => DnsRecordType::A,
    };

    DnsRecord {
        id: cf.id,
        record_type,
        name: cf.name,
        value: cf.content,
        ttl: cf.ttl,
        proxied: cf.proxied,
    }
}

// ---------------------------------------------------------------------------
// DnsProvider trait
// ---------------------------------------------------------------------------

#[allow(async_fn_in_trait)]
pub trait DnsProvider: Send + Sync {
    async fn list_records(&self, domain: &str) -> Result<Vec<DnsRecord>, DnsError>;
    async fn create_record(
        &self,
        domain: &str,
        record: &DnsRecord,
        tags: &[String],
    ) -> Result<DnsRecord, DnsError>;
    async fn update_record(
        &self,
        domain: &str,
        record: &DnsRecord,
        tags: &[String],
    ) -> Result<DnsRecord, DnsError>;
    async fn delete_record(&self, domain: &str, record_id: &str) -> Result<(), DnsError>;
}

// ---------------------------------------------------------------------------
// CloudflareProvider
// ---------------------------------------------------------------------------

pub struct CloudflareProvider {
    api_token: String,
    zone_id: String,
    client: reqwest::Client,
}

impl CloudflareProvider {
    pub fn new(api_token: String, zone_id: String) -> Self {
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(15))
            .build()
            .expect("failed to build HTTP client");

        Self {
            api_token,
            zone_id,
            client,
        }
    }

    fn headers(&self) -> HeaderMap {
        let mut headers = HeaderMap::new();
        headers.insert(
            AUTHORIZATION,
            HeaderValue::from_str(&format!("Bearer {}", self.api_token))
                .expect("invalid API token"),
        );
        headers.insert(CONTENT_TYPE, HeaderValue::from_static("application/json"));
        headers
    }

    fn check_response(resp: &CloudflareResponse) -> Result<(), DnsError> {
        if !resp.success {
            let msg = resp
                .errors
                .first()
                .map(|e| e.message.clone())
                .unwrap_or_else(|| "cloudflare API request failed".into());
            return Err(DnsError::CloudflareApi(msg));
        }
        Ok(())
    }
}

impl DnsProvider for CloudflareProvider {
    async fn list_records(&self, _domain: &str) -> Result<Vec<DnsRecord>, DnsError> {
        let url = format!("{}/zones/{}/dns_records", CLOUDFLARE_API_BASE, self.zone_id);

        let resp = self
            .client
            .get(&url)
            .headers(self.headers())
            .send()
            .await?;

        let cf_resp: CloudflareResponse = resp.json().await?;
        Self::check_response(&cf_resp)?;

        let cf_records: Vec<CloudflareRecord> =
            serde_json::from_value(cf_resp.result).map_err(|e| DnsError::Other(e.to_string()))?;

        Ok(cf_records.into_iter().map(from_cloudflare_record).collect())
    }

    async fn create_record(
        &self,
        _domain: &str,
        record: &DnsRecord,
        tags: &[String],
    ) -> Result<DnsRecord, DnsError> {
        let url = format!("{}/zones/{}/dns_records", CLOUDFLARE_API_BASE, self.zone_id);
        let cf_record = to_cloudflare_record(record, tags);

        let resp = self
            .client
            .post(&url)
            .headers(self.headers())
            .json(&cf_record)
            .send()
            .await?;

        let cf_resp: CloudflareResponse = resp.json().await?;
        Self::check_response(&cf_resp)?;

        let created: CloudflareRecord =
            serde_json::from_value(cf_resp.result).map_err(|e| DnsError::Other(e.to_string()))?;

        Ok(from_cloudflare_record(created))
    }

    async fn update_record(
        &self,
        _domain: &str,
        record: &DnsRecord,
        tags: &[String],
    ) -> Result<DnsRecord, DnsError> {
        let record_id = record
            .id
            .as_deref()
            .filter(|id| !id.is_empty())
            .ok_or(DnsError::MissingRecordId("updates"))?;

        let url = format!(
            "{}/zones/{}/dns_records/{}",
            CLOUDFLARE_API_BASE, self.zone_id, record_id
        );
        let cf_record = to_cloudflare_record(record, tags);

        let resp = self
            .client
            .put(&url)
            .headers(self.headers())
            .json(&cf_record)
            .send()
            .await?;

        let cf_resp: CloudflareResponse = resp.json().await?;
        Self::check_response(&cf_resp)?;

        let updated: CloudflareRecord =
            serde_json::from_value(cf_resp.result).map_err(|e| DnsError::Other(e.to_string()))?;

        Ok(from_cloudflare_record(updated))
    }

    async fn delete_record(&self, _domain: &str, record_id: &str) -> Result<(), DnsError> {
        if record_id.is_empty() {
            return Err(DnsError::MissingRecordId("deletion"));
        }

        let url = format!(
            "{}/zones/{}/dns_records/{}",
            CLOUDFLARE_API_BASE, self.zone_id, record_id
        );

        let resp = self
            .client
            .delete(&url)
            .headers(self.headers())
            .send()
            .await?;

        let cf_resp: CloudflareResponse = resp.json().await?;
        Self::check_response(&cf_resp)?;

        Ok(())
    }
}

// ---------------------------------------------------------------------------
// Factory
// ---------------------------------------------------------------------------

use crate::models::{DnsProvider as DnsProviderModel, DnsProviderType};

/// Creates a CloudflareProvider from a DnsProvider model config.
/// Returns None for Manual DNS, Err for Route53 (unimplemented).
pub fn create_provider(
    provider: &DnsProviderModel,
) -> Result<Option<CloudflareProvider>, DnsError> {
    match provider.provider_type {
        DnsProviderType::Cloudflare => {
            let token = provider
                .api_token
                .as_deref()
                .filter(|t| !t.is_empty())
                .ok_or_else(|| {
                    DnsError::Other("Cloudflare provider requires api_token".into())
                })?;
            let zone_id = provider
                .zone_id
                .as_deref()
                .filter(|z| !z.is_empty())
                .ok_or_else(|| {
                    DnsError::Other("Cloudflare provider requires zone_id".into())
                })?;
            Ok(Some(CloudflareProvider::new(
                token.to_string(),
                zone_id.to_string(),
            )))
        }
        DnsProviderType::Route53 => Err(DnsError::Other(
            "Route53 provider not yet implemented".into(),
        )),
        DnsProviderType::Manual => Ok(None),
    }
}
