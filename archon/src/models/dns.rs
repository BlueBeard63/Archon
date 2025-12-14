use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DnsRecord {
    pub id: Option<String>, // Provider's record ID
    pub record_type: DnsRecordType,
    pub name: String,
    pub value: String,
    pub ttl: u32,
    pub proxied: bool, // Cloudflare specific
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
pub enum DnsRecordType {
    A,
    AAAA,
    CNAME,
    MX,
    TXT,
    SRV,
}

impl DnsRecord {
    pub fn new(
        record_type: DnsRecordType,
        name: String,
        value: String,
        ttl: u32,
    ) -> Self {
        Self {
            id: None,
            record_type,
            name,
            value,
            ttl,
            proxied: false,
        }
    }

    pub fn with_proxy(mut self, proxied: bool) -> Self {
        self.proxied = proxied;
        self
    }
}

impl std::fmt::Display for DnsRecordType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            DnsRecordType::A => write!(f, "A"),
            DnsRecordType::AAAA => write!(f, "AAAA"),
            DnsRecordType::CNAME => write!(f, "CNAME"),
            DnsRecordType::MX => write!(f, "MX"),
            DnsRecordType::TXT => write!(f, "TXT"),
            DnsRecordType::SRV => write!(f, "SRV"),
        }
    }
}

impl std::str::FromStr for DnsRecordType {
    type Err = String;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_uppercase().as_str() {
            "A" => Ok(DnsRecordType::A),
            "AAAA" => Ok(DnsRecordType::AAAA),
            "CNAME" => Ok(DnsRecordType::CNAME),
            "MX" => Ok(DnsRecordType::MX),
            "TXT" => Ok(DnsRecordType::TXT),
            "SRV" => Ok(DnsRecordType::SRV),
            _ => Err(format!("Invalid DNS record type: {}", s)),
        }
    }
}
