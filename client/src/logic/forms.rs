use floem::reactive::{RwSignal, SignalGet, SignalUpdate};
use uuid::Uuid;

use crate::config::{DockerCredential, Settings};
use crate::models::{Domain, Node, Site};

/// A key-value pair for environment variables.
#[derive(Debug, Clone, Default)]
pub struct EnvVarPair {
    pub key: String,
    pub value: String,
}

/// A domain mapping entry for multi-domain sites.
#[derive(Debug, Clone, Default)]
pub struct DomainMappingPair {
    pub subdomain: String,
    pub domain_name: String,
    pub domain_id: String,
    pub port: String,
}

/// A volume bind mount pair.
#[derive(Debug, Clone, Default)]
pub struct VolumePair {
    pub host_path: String,
    pub container_path: String,
}

/// Reactive form state shared across create/edit screens.
pub struct FormState {
    pub fields: RwSignal<Vec<String>>,
    pub current_field_index: RwSignal<usize>,
    pub dropdown_open: RwSignal<bool>,
    pub dropdown_index: RwSignal<usize>,
    pub edit_form_initialized: RwSignal<bool>,

    // --- Site fields ---
    pub site_name: RwSignal<String>,
    pub site_type: RwSignal<String>,
    pub site_node_id: RwSignal<String>,
    pub site_docker_image: RwSignal<String>,
    pub site_docker_credential_id: RwSignal<String>,
    pub site_ssl_email: RwSignal<String>,
    pub site_compose_content: RwSignal<String>,

    // --- Domain fields ---
    pub domain_name: RwSignal<String>,
    pub domain_provider: RwSignal<String>,
    pub domain_zone_id: RwSignal<String>,
    pub domain_access_key: RwSignal<String>,
    pub domain_secret_key: RwSignal<String>,

    // --- Node fields ---
    pub node_name: RwSignal<String>,
    pub node_api_endpoint: RwSignal<String>,
    pub node_proxy_type: RwSignal<String>,
    pub node_api_key: RwSignal<String>,
    pub node_ip_address: RwSignal<String>,

    // --- Docker Credential fields ---
    pub docker_cred_name: RwSignal<String>,
    pub docker_cred_registry: RwSignal<String>,
    pub docker_cred_username: RwSignal<String>,
    pub docker_cred_token: RwSignal<String>,

    // --- Settings fields ---
    pub settings_cloudflare_token: RwSignal<String>,
    pub settings_route53_access_key: RwSignal<String>,
    pub settings_route53_secret_key: RwSignal<String>,

    // Environment variables
    pub env_var_pairs: RwSignal<Vec<EnvVarPair>>,
    pub env_var_focused_pair: RwSignal<usize>,
    pub env_var_focused_field: RwSignal<usize>, // 0=key, 1=value

    // Domain mappings
    pub domain_mapping_pairs: RwSignal<Vec<DomainMappingPair>>,
    pub domain_mapping_focused_pair: RwSignal<usize>,
    pub domain_mapping_focused_field: RwSignal<usize>, // 0=subdomain, 1=domain, 2=port

    // Volumes
    pub volume_pairs: RwSignal<Vec<VolumePair>>,
    pub volume_focused_pair: RwSignal<usize>,
    pub volume_focused_field: RwSignal<usize>, // 0=host_path, 1=container_path

    // Deletion confirmation
    pub deletion_confirm_pending: RwSignal<bool>,
    pub deletion_confirm_input: RwSignal<String>,
    pub deletion_target_id: RwSignal<Option<Uuid>>,
    pub deletion_target_name: RwSignal<String>,
    pub deletion_target_type: RwSignal<String>,
}

impl FormState {
    pub fn new() -> Self {
        Self {
            fields: RwSignal::new(Vec::new()),
            current_field_index: RwSignal::new(0),
            dropdown_open: RwSignal::new(false),
            dropdown_index: RwSignal::new(0),
            edit_form_initialized: RwSignal::new(false),

            // Site
            site_name: RwSignal::new(String::new()),
            site_type: RwSignal::new("container".to_string()),
            site_node_id: RwSignal::new(String::new()),
            site_docker_image: RwSignal::new(String::new()),
            site_docker_credential_id: RwSignal::new(String::new()),
            site_ssl_email: RwSignal::new(String::new()),
            site_compose_content: RwSignal::new(String::new()),

            // Domain
            domain_name: RwSignal::new(String::new()),
            domain_provider: RwSignal::new("manual".to_string()),
            domain_zone_id: RwSignal::new(String::new()),
            domain_access_key: RwSignal::new(String::new()),
            domain_secret_key: RwSignal::new(String::new()),

            // Node
            node_name: RwSignal::new(String::new()),
            node_api_endpoint: RwSignal::new(String::new()),
            node_proxy_type: RwSignal::new("traefik".to_string()),
            node_api_key: RwSignal::new(String::new()),
            node_ip_address: RwSignal::new(String::new()),

            // Docker Credential
            docker_cred_name: RwSignal::new(String::new()),
            docker_cred_registry: RwSignal::new(String::new()),
            docker_cred_username: RwSignal::new(String::new()),
            docker_cred_token: RwSignal::new(String::new()),

            // Settings
            settings_cloudflare_token: RwSignal::new(String::new()),
            settings_route53_access_key: RwSignal::new(String::new()),
            settings_route53_secret_key: RwSignal::new(String::new()),

            env_var_pairs: RwSignal::new(vec![EnvVarPair::default()]),
            env_var_focused_pair: RwSignal::new(0),
            env_var_focused_field: RwSignal::new(0),

            domain_mapping_pairs: RwSignal::new(vec![DomainMappingPair::default()]),
            domain_mapping_focused_pair: RwSignal::new(0),
            domain_mapping_focused_field: RwSignal::new(0),

            volume_pairs: RwSignal::new(Vec::new()),
            volume_focused_pair: RwSignal::new(0),
            volume_focused_field: RwSignal::new(0),

            deletion_confirm_pending: RwSignal::new(false),
            deletion_confirm_input: RwSignal::new(String::new()),
            deletion_target_id: RwSignal::new(None),
            deletion_target_name: RwSignal::new(String::new()),
            deletion_target_type: RwSignal::new(String::new()),
        }
    }

    /// Reset form state (called when navigating away from forms).
    pub fn reset(&self) {
        self.fields.set(Vec::new());
        self.current_field_index.set(0);
        self.dropdown_open.set(false);
        self.dropdown_index.set(0);
        self.edit_form_initialized.set(false);

        // Site
        self.site_name.set(String::new());
        self.site_type.set("container".to_string());
        self.site_node_id.set(String::new());
        self.site_docker_image.set(String::new());
        self.site_docker_credential_id.set(String::new());
        self.site_ssl_email.set(String::new());
        self.site_compose_content.set(String::new());

        // Domain
        self.domain_name.set(String::new());
        self.domain_provider.set("manual".to_string());
        self.domain_zone_id.set(String::new());
        self.domain_access_key.set(String::new());
        self.domain_secret_key.set(String::new());

        // Node
        self.node_name.set(String::new());
        self.node_api_endpoint.set(String::new());
        self.node_proxy_type.set("traefik".to_string());
        self.node_api_key.set(String::new());
        self.node_ip_address.set(String::new());

        // Docker Credential
        self.docker_cred_name.set(String::new());
        self.docker_cred_registry.set(String::new());
        self.docker_cred_username.set(String::new());
        self.docker_cred_token.set(String::new());

        // Settings
        self.settings_cloudflare_token.set(String::new());
        self.settings_route53_access_key.set(String::new());
        self.settings_route53_secret_key.set(String::new());

        self.env_var_pairs.set(vec![EnvVarPair::default()]);
        self.env_var_focused_pair.set(0);
        self.env_var_focused_field.set(0);

        self.domain_mapping_pairs
            .set(vec![DomainMappingPair::default()]);
        self.domain_mapping_focused_pair.set(0);
        self.domain_mapping_focused_field.set(0);

        self.volume_pairs.set(Vec::new());
        self.volume_focused_pair.set(0);
        self.volume_focused_field.set(0);

        self.reset_deletion();
    }

    pub fn reset_deletion(&self) {
        self.deletion_confirm_pending.set(false);
        self.deletion_confirm_input.set(String::new());
        self.deletion_target_id.set(None);
        self.deletion_target_name.set(String::new());
        self.deletion_target_type.set(String::new());
    }

    /// Start a deletion confirmation flow.
    pub fn begin_deletion(&self, id: Uuid, name: &str, target_type: &str) {
        self.deletion_confirm_pending.set(true);
        self.deletion_confirm_input.set(String::new());
        self.deletion_target_id.set(Some(id));
        self.deletion_target_name.set(name.to_string());
        self.deletion_target_type.set(target_type.to_string());
    }

    /// Check if the user's deletion input matches the target name.
    pub fn is_deletion_confirmed(&self) -> bool {
        let input = self.deletion_confirm_input.get_untracked();
        let target = self.deletion_target_name.get_untracked();
        !input.is_empty() && input == target
    }

    // --- Init helpers ---

    pub fn init_from_site(&self, site: &Site, domains: &[Domain]) {
        self.site_name.set(site.name.clone());
        self.site_type.set(site.site_type.to_string());
        self.site_node_id.set(site.node_id.to_string());
        self.site_docker_image.set(site.docker_image.clone());
        self.site_docker_credential_id.set(
            site.docker_credential_id
                .map(|id| id.to_string())
                .unwrap_or_default(),
        );
        self.site_ssl_email.set(site.ssl_email.clone());
        self.site_compose_content.set(site.compose_content.clone());

        // Env vars
        let env_pairs: Vec<EnvVarPair> = site
            .environment_vars
            .iter()
            .map(|(k, v)| EnvVarPair {
                key: k.clone(),
                value: v.clone(),
            })
            .collect();
        self.env_var_pairs.set(if env_pairs.is_empty() {
            vec![EnvVarPair::default()]
        } else {
            env_pairs
        });

        // Domain mappings
        let mapping_pairs: Vec<DomainMappingPair> = site
            .domain_mappings
            .iter()
            .map(|m| {
                let domain_name = domains
                    .iter()
                    .find(|d| d.id == m.domain_id)
                    .map(|d| d.name.clone())
                    .unwrap_or_default();
                DomainMappingPair {
                    subdomain: m.subdomain.clone(),
                    domain_name,
                    domain_id: m.domain_id.to_string(),
                    port: m.port.to_string(),
                }
            })
            .collect();
        self.domain_mapping_pairs.set(if mapping_pairs.is_empty() {
            vec![DomainMappingPair::default()]
        } else {
            mapping_pairs
        });

        // Volumes
        let vol_pairs: Vec<VolumePair> = site
            .volumes
            .iter()
            .map(|v| VolumePair {
                host_path: v.host_path.clone(),
                container_path: v.container_path.clone(),
            })
            .collect();
        self.volume_pairs.set(vol_pairs);
    }

    pub fn init_from_domain(&self, domain: &Domain) {
        self.domain_name.set(domain.name.clone());
        self.domain_provider.set(domain.dns_provider.provider_type.to_string());
        self.domain_zone_id.set(
            domain.dns_provider.zone_id.clone().unwrap_or_default(),
        );
        self.domain_access_key.set(
            domain.dns_provider.access_key.clone().unwrap_or_default(),
        );
        self.domain_secret_key.set(
            domain.dns_provider.secret_key.clone().unwrap_or_default(),
        );
    }

    pub fn init_from_node(&self, node: &Node) {
        self.node_name.set(node.name.clone());
        self.node_api_endpoint.set(node.api_endpoint.clone());
        self.node_proxy_type.set(node.proxy_type.to_string());
        self.node_api_key.set(node.api_key.clone());
        self.node_ip_address.set(node.ip_address.to_string());
    }

    pub fn init_from_docker_credential(&self, cred: &DockerCredential) {
        self.docker_cred_name.set(cred.name.clone());
        self.docker_cred_registry.set(cred.registry.clone());
        self.docker_cred_username.set(cred.username.clone());
        self.docker_cred_token.set(cred.token.clone());
    }

    pub fn init_from_settings(&self, settings: &Settings) {
        self.settings_cloudflare_token
            .set(settings.cloudflare_api_token.clone());
        self.settings_route53_access_key
            .set(settings.route53_access_key.clone());
        self.settings_route53_secret_key
            .set(settings.route53_secret_key.clone());
    }

    // --- Env var helpers ---

    pub fn add_env_var(&self) {
        self.env_var_pairs
            .update(|pairs| pairs.push(EnvVarPair::default()));
    }

    pub fn remove_env_var(&self, index: usize) {
        self.env_var_pairs.update(|pairs| {
            if index < pairs.len() && pairs.len() > 1 {
                pairs.remove(index);
            }
        });
    }

    // --- Domain mapping helpers ---

    pub fn add_domain_mapping(&self) {
        self.domain_mapping_pairs
            .update(|pairs| pairs.push(DomainMappingPair::default()));
    }

    pub fn remove_domain_mapping(&self, index: usize) {
        self.domain_mapping_pairs.update(|pairs| {
            if index < pairs.len() && pairs.len() > 1 {
                pairs.remove(index);
            }
        });
    }

    // --- Volume helpers ---

    pub fn add_volume(&self) {
        self.volume_pairs
            .update(|pairs| pairs.push(VolumePair::default()));
    }

    pub fn remove_volume(&self, index: usize) {
        self.volume_pairs.update(|pairs| {
            if index < pairs.len() {
                pairs.remove(index);
            }
        });
    }
}

/// Validate that a required string field is non-empty.
pub fn validate_required(value: &str, field_name: &str) -> Result<(), String> {
    if value.trim().is_empty() {
        Err(format!("{} is required", field_name))
    } else {
        Ok(())
    }
}

/// Validate that a port number is in valid range.
pub fn validate_port(value: &str, field_name: &str) -> Result<i32, String> {
    let port: i32 = value
        .trim()
        .parse()
        .map_err(|_| format!("{} must be a number", field_name))?;
    if !(1..=65535).contains(&port) {
        return Err(format!("{} must be between 1 and 65535", field_name));
    }
    Ok(port)
}

/// Validate a URL format.
pub fn validate_url(value: &str, field_name: &str) -> Result<(), String> {
    if value.trim().is_empty() {
        return Err(format!("{} is required", field_name));
    }
    url::Url::parse(value.trim()).map_err(|_| format!("{} must be a valid URL", field_name))?;
    Ok(())
}

/// Validate an IP address.
pub fn validate_ip(value: &str, field_name: &str) -> Result<std::net::IpAddr, String> {
    value
        .trim()
        .parse()
        .map_err(|_| format!("{} must be a valid IP address", field_name))
}
