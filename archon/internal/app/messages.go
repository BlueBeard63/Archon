package app

import (
	"github.com/google/uuid"
	"github.com/BlueBeard63/archon/internal/api"
	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
)

// Message types for the Bubbletea Update function
// These replace the Rust Action enum with individual message structs

// ============================================================================
// Navigation Messages
// ============================================================================

// NavigateToMsg switches to a different screen
type NavigateToMsg struct {
	Screen state.Screen
}

// NavigateBackMsg goes back to the previous screen
type NavigateBackMsg struct{}

// ============================================================================
// Site Messages
// ============================================================================

// CreateSiteMsg creates a new site in the state
type CreateSiteMsg struct {
	Site *models.Site
}

// DeleteSiteMsg removes a site from the state
type DeleteSiteMsg struct {
	SiteID uuid.UUID
}

// DeploySiteMsg triggers site deployment to its node
type DeploySiteMsg struct {
	SiteID uuid.UUID
}

// SiteDeployedMsg is returned after deployment completes
type SiteDeployedMsg struct {
	SiteID uuid.UUID
	Error  error
}

// SiteDeployProgressMsg is sent during deployment with progress updates
type SiteDeployProgressMsg struct {
	SiteID  uuid.UUID
	Message string
	Step    string // "init", "ssl", "docker", "proxy", "complete"
}

// SetupDNSMsg triggers DNS record setup for a site
type SetupDNSMsg struct {
	SiteID uuid.UUID
}

// DNSSetupResultMsg is returned after DNS setup completes
type DNSSetupResultMsg struct {
	SiteID  uuid.UUID
	Message string // Success message with details
	Error   error
}

// StopSiteMsg stops a running site
type StopSiteMsg struct {
	SiteID uuid.UUID
}

// RestartSiteMsg restarts a site
type RestartSiteMsg struct {
	SiteID uuid.UUID
}

// SiteOperationResultMsg is returned after stop/restart operations
type SiteOperationResultMsg struct {
	SiteID    uuid.UUID
	Operation string // "stop" or "restart"
	Error     error
}

// ============================================================================
// Domain Messages
// ============================================================================

// CreateDomainMsg creates a new domain in the state
type CreateDomainMsg struct {
	Domain *models.Domain
}

// DeleteDomainMsg removes a domain from the state
type DeleteDomainMsg struct {
	DomainID uuid.UUID
}

// SyncDnsMsg triggers DNS record synchronization with provider
type SyncDnsMsg struct {
	DomainID uuid.UUID
}

// DnsSyncedMsg is returned after DNS sync completes
type DnsSyncedMsg struct {
	DomainID uuid.UUID
	Records  []models.DnsRecord
	Error    error
}

// CreateDnsRecordMsg adds a new DNS record to a domain
type CreateDnsRecordMsg struct {
	DomainID uuid.UUID
	Record   *models.DnsRecord
}

// UpdateDnsRecordMsg updates an existing DNS record
type UpdateDnsRecordMsg struct {
	DomainID uuid.UUID
	Record   *models.DnsRecord
}

// DeleteDnsRecordMsg removes a DNS record
type DeleteDnsRecordMsg struct {
	DomainID uuid.UUID
	RecordID string
}

// DnsRecordOperationResultMsg is returned after DNS record operations
type DnsRecordOperationResultMsg struct {
	DomainID  uuid.UUID
	Operation string // "create", "update", "delete"
	Error     error
}

// ============================================================================
// Node Messages
// ============================================================================

// CreateNodeMsg creates a new node in the state
type CreateNodeMsg struct {
	Node *models.Node
}

// DeleteNodeMsg removes a node from the state
type DeleteNodeMsg struct {
	NodeID uuid.UUID
}

// NodeHealthCheckMsg triggers a health check on a node
type NodeHealthCheckMsg struct {
	NodeID uuid.UUID
}

// NodeHealthCheckResultMsg is returned after health check completes
type NodeHealthCheckResultMsg struct {
	NodeID uuid.UUID
	Result *api.HealthResponse
	Error  error
}

// FetchNodeLogsMsg retrieves logs from a site on a node
type FetchNodeLogsMsg struct {
	SiteID uuid.UUID
	Lines  int
}

// NodeLogsResultMsg is returned with log lines
type NodeLogsResultMsg struct {
	SiteID uuid.UUID
	Logs   []string
	Error  error
}

// FetchNodeMetricsMsg retrieves resource metrics for a site
type FetchNodeMetricsMsg struct {
	SiteID uuid.UUID
}

// NodeMetricsResultMsg is returned with metrics data
type NodeMetricsResultMsg struct {
	SiteID  uuid.UUID
	Metrics *api.ContainerMetrics
	Error   error
}

// ============================================================================
// Form Messages
// ============================================================================

// FormInputMsg handles character input in forms
type FormInputMsg struct {
	Char rune
}

// FormBackspaceMsg removes the last character from current field
type FormBackspaceMsg struct{}

// NextFormFieldMsg moves to the next form field (Tab)
type NextFormFieldMsg struct{}

// PrevFormFieldMsg moves to the previous form field (Shift+Tab)
type PrevFormFieldMsg struct{}

// FormSubmitMsg submits the current form
type FormSubmitMsg struct{}

// FormCancelMsg cancels the current form and navigates back
type FormCancelMsg struct{}

// ============================================================================
// Selection Messages (for lists/tables)
// ============================================================================

// SelectNextMsg moves selection down in a list
type SelectNextMsg struct{}

// SelectPreviousMsg moves selection up in a list
type SelectPreviousMsg struct{}

// SelectItemMsg selects a specific item by index (for mouse clicks)
type SelectItemMsg struct {
	Index int
}

// ============================================================================
// System Messages
// ============================================================================

// SaveConfigMsg triggers configuration save to disk
type SaveConfigMsg struct{}

// ConfigSavedMsg is returned after config save completes
type ConfigSavedMsg struct {
	Error error
}

// QuitMsg signals the application should exit
type QuitMsg struct{}

// NotificationMsg displays a message to the user
type NotificationMsg struct {
	Message string
	Level   string // "success", "error", "warning", "info"
}

// ClearNotificationsMsg removes all notifications
type ClearNotificationsMsg struct{}

// ============================================================================
// Async Operation Messages
// ============================================================================

// AsyncResultMsg is a generic async operation result
type AsyncResultMsg struct {
	OperationID uuid.UUID
	Result      interface{}
	Error       error
}

// TickMsg is sent periodically for background tasks
type TickMsg struct{}
