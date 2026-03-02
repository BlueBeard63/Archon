use floem::reactive::{RwSignal, SignalGet, SignalUpdate};
use uuid::Uuid;

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
