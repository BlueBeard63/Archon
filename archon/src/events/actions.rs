use uuid::Uuid;

use crate::models::dns::DnsRecord;
use crate::models::{Domain, Node, Site};
use crate::state::{AsyncOperationResult, Notification};
use crate::ui::Screen;

#[derive(Debug, Clone)]
pub enum Action {
    // Navigation
    NavigateTo(Screen),
    NavigateBack,

    // Site management
    CreateSite(Site),
    UpdateSite(Uuid, Site),
    DeleteSite(Uuid),
    DeploySite(Uuid),
    StopSite(Uuid),
    RestartSite(Uuid),

    // Domain management
    CreateDomain(Domain),
    DeleteDomain(Uuid),
    AddDnsRecord(Uuid, DnsRecord),
    UpdateDnsRecord(Uuid, usize, DnsRecord),
    DeleteDnsRecord(Uuid, usize),
    SyncDnsRecords(Uuid),

    // Node management
    AddNode(Node),
    UpdateNode(Uuid, Node),
    RemoveNode(Uuid),
    CheckNodeHealth(Uuid),
    FetchNodeStats(Uuid),

    // Site monitoring
    FetchLogs(Uuid),
    FetchMetrics(Uuid),

    // Selection/UI
    SelectNext,
    SelectPrevious,
    SelectItem(usize),
    NextFormField,
    PreviousFormField,

    // Async operation results
    AsyncOperationCompleted(Uuid, Result<AsyncOperationResult, String>),

    // Notifications
    ShowNotification(Notification),
    DismissNotification,

    // System
    SaveConfig,
    LoadConfig,
    Quit,

    // No-op
    None,
}
