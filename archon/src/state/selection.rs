use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SelectionState {
    pub sites_list_index: usize,
    pub domains_list_index: usize,
    pub nodes_list_index: usize,
    pub form_field_index: usize,
    pub dns_records_index: usize,
}

impl Default for SelectionState {
    fn default() -> Self {
        Self {
            sites_list_index: 0,
            domains_list_index: 0,
            nodes_list_index: 0,
            form_field_index: 0,
            dns_records_index: 0,
        }
    }
}

impl SelectionState {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn reset(&mut self) {
        *self = Self::default();
    }
}
