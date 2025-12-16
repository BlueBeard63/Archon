package state

import (
	"github.com/google/uuid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/BlueBeard63/archon/internal/models"
)

// TableComponent interface to avoid circular import
type TableComponent interface {
	View() string
	SetCursor(int)
	GetCursor() int
	SetRows([]table.Row)
	Update(tea.Msg) tea.Cmd
	SetWidth(int)
	GetSelectedRow() table.Row
}

// Screen represents different screens in the TUI
type Screen string

const (
	ScreenDashboard         Screen = "dashboard"
	ScreenSitesList         Screen = "sites_list"
	ScreenSiteCreate        Screen = "site_create"
	ScreenSiteEdit          Screen = "site_edit"
	ScreenDomainsList       Screen = "domains_list"
	ScreenDomainCreate      Screen = "domain_create"
	ScreenDomainEdit        Screen = "domain_edit"
	ScreenDomainDnsRecords  Screen = "domain_dns_records"
	ScreenNodesList         Screen = "nodes_list"
	ScreenNodeCreate        Screen = "node_create"
	ScreenNodeEdit          Screen = "node_edit"
	ScreenNodeConfig        Screen = "node_config"
	ScreenHelp              Screen = "help"
)

// AppState holds all application state for the TUI
type AppState struct {
	// Data
	Sites   []models.Site   `json:"sites"`
	Domains []models.Domain `json:"domains"`
	Nodes   []models.Node   `json:"nodes"`

	// UI State
	CurrentScreen   Screen   `json:"current_screen"`
	PreviousScreens []Screen `json:"previous_screens"` // Navigation stack for back button

	// Selection state (for table lists)
	SitesListIndex   int       `json:"sites_list_index"`
	DomainsListIndex int       `json:"domains_list_index"`
	NodesListIndex   int       `json:"nodes_list_index"`
	SelectedSiteID   uuid.UUID `json:"selected_site_id"`   // For editing site
	SelectedDomainID uuid.UUID `json:"selected_domain_id"` // For editing domain
	SelectedNodeID   uuid.UUID `json:"selected_node_id"`   // For viewing/editing node config

	// Table component instances (runtime only, not serialized)
	SitesTable   TableComponent `json:"-"`
	DomainsTable TableComponent `json:"-"`
	NodesTable   TableComponent `json:"-"`

	// Form state (for create/edit screens)
	FormFields        []string `json:"form_fields"`        // Current values of form fields
	CurrentFieldIndex int      `json:"current_field_index"` // Which field has focus

	// Async operations tracking
	PendingOperations []AsyncOperation `json:"pending_operations"`
	Notifications     []Notification   `json:"notifications"`

	// Window dimensions (updated on resize)
	WindowWidth  int `json:"window_width"`
	WindowHeight int `json:"window_height"`

	// Configuration
	ConfigPath string `json:"config_path"`
	AutoSave   bool   `json:"auto_save"`
	ShouldQuit bool   `json:"should_quit"`
}

// AsyncOperation tracks background operations like deployments
type AsyncOperation struct {
	ID     uuid.UUID `json:"id"`
	OpType string    `json:"op_type"` // "deploy_site", "sync_dns", "health_check", etc.
	Status string    `json:"status"`  // "pending", "completed", "failed"
	Target string    `json:"target"`  // Description of what's being operated on
}

// Notification represents a message to display to the user
type Notification struct {
	Message string `json:"message"`
	Level   string `json:"level"` // "success", "error", "warning", "info"
}

// NewAppState creates a new AppState with default values
func NewAppState() *AppState {
	return &AppState{
		Sites:             []models.Site{},
		Domains:           []models.Domain{},
		Nodes:             []models.Node{},
		CurrentScreen:     ScreenDashboard,
		PreviousScreens:   []Screen{},
		SitesListIndex:    0,
		DomainsListIndex:  0,
		NodesListIndex:    0,
		FormFields:        []string{},
		CurrentFieldIndex: 0,
		PendingOperations: []AsyncOperation{},
		Notifications:     []Notification{},
		AutoSave:          true,
		ShouldQuit:        false,
	}
}

// NavigateTo switches to a new screen and adds current screen to history
func (s *AppState) NavigateTo(screen Screen) {
	// Push current screen to history
	s.PreviousScreens = append(s.PreviousScreens, s.CurrentScreen)

	// Switch to new screen
	s.CurrentScreen = screen

	// Reset form state when navigating
	s.FormFields = []string{}
	s.CurrentFieldIndex = 0
}

// NavigateBack goes back to the previous screen in history
func (s *AppState) NavigateBack() {
	if len(s.PreviousScreens) > 0 {
		// Pop from history stack
		lastIndex := len(s.PreviousScreens) - 1
		s.CurrentScreen = s.PreviousScreens[lastIndex]
		s.PreviousScreens = s.PreviousScreens[:lastIndex]
	}
}

// AddNotification adds a new notification to the queue
func (s *AppState) AddNotification(message string, level string) {
	s.Notifications = append(s.Notifications, Notification{
		Message: message,
		Level:   level,
	})

	// Keep only last 50 notifications
	if len(s.Notifications) > 50 {
		s.Notifications = s.Notifications[1:]
	}
}

// ClearNotifications removes all notifications
func (s *AppState) ClearNotifications() {
	s.Notifications = []Notification{}
}

// GetSiteByID finds a site by its UUID
func (s *AppState) GetSiteByID(id uuid.UUID) *models.Site {
	// TODO: Implement site lookup
	// Iterate through s.Sites and return pointer to matching site
	// Return nil if not found
	return nil
}

// GetDomainByID finds a domain by its UUID
func (s *AppState) GetDomainByID(id uuid.UUID) *models.Domain {
	// TODO: Implement domain lookup
	return nil
}

// GetNodeByID finds a node by its UUID
func (s *AppState) GetNodeByID(id uuid.UUID) *models.Node {
	// TODO: Implement node lookup
	return nil
}

// AddAsyncOperation adds a new async operation to track
func (s *AppState) AddAsyncOperation(opType, target string) uuid.UUID {
	// TODO: Implement async operation tracking
	// Create new AsyncOperation with UUID, add to slice, return UUID
	// Example:
	// id := uuid.New()
	// op := AsyncOperation{
	//     ID:     id,
	//     OpType: opType,
	//     Status: "pending",
	//     Target: target,
	// }
	// s.PendingOperations = append(s.PendingOperations, op)
	// return id
	return uuid.Nil
}

// CompleteAsyncOperation marks an operation as completed
func (s *AppState) CompleteAsyncOperation(id uuid.UUID, success bool) {
	// TODO: Implement operation completion
	// Find operation by ID and update status to "completed" or "failed"
	// Optionally remove from slice after completion
}
