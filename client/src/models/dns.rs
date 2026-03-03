use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum DnsRecordType {
    A,
    AAAA,
    CNAME,
    MX,
    TXT,
    SRV,
}

impl fmt::Display for DnsRecordType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::A => write!(f, "A"),
            Self::AAAA => write!(f, "AAAA"),
            Self::CNAME => write!(f, "CNAME"),
            Self::MX => write!(f, "MX"),
            Self::TXT => write!(f, "TXT"),
            Self::SRV => write!(f, "SRV"),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DnsRecord {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub id: Option<String>,
    pub record_type: DnsRecordType,
    pub name: String,
    pub value: String,
    pub ttl: i32,
    pub proxied: bool,
}

impl DnsRecord {
    pub fn new(record_type: DnsRecordType, name: String, value: String, ttl: i32) -> Self {
        Self {
            id: None,
            record_type,
            name,
            value,
            ttl,
            proxied: false,
        }
    }
}
