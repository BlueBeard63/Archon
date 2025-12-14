pub mod provider;
pub mod cloudflare;

pub use provider::{DnsProvider as DnsProviderTrait, create_provider};
pub use cloudflare::CloudflareProvider;
