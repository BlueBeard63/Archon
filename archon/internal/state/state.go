package state

import (
	"github.com/google/uuid"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
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
	ScreenDashboard              Screen = "dashboard"
	ScreenSitesList              Screen = "sites_list"
	ScreenSiteCreate             Screen = "site_create"
	ScreenSiteEdit               Screen = "site_edit"
	ScreenSiteEnvVars            Screen = "site_env_vars"
	ScreenDomainsList            Screen = "domains_list"
	ScreenDomainCreate           Screen = "domain_create"
	ScreenDomainEdit             Screen = "domain_edit"
	ScreenDomainDnsRecords       Screen = "domain_dns_records"
	ScreenNodesList              Screen = "nodes_list"
	ScreenNodeCreate             Screen = "node_create"
	ScreenNodeEdit               Screen = "node_edit"
	ScreenNodeConfig             Screen = "node_config"
	ScreenNodeConfigSave         Screen = "node_config_save"
	ScreenNodeQuickConfig        Screen = "node_quick_config"
	ScreenSettings               Screen = "settings"
	ScreenDockerCredentialsList  Screen = "docker_credentials_list"
	ScreenDockerCredentialCreate Screen = "docker_credential_create"
	ScreenDockerCredentialEdit   Screen = "docker_credential_edit"
	ScreenHelp                   Screen = "help"
	ScreenSiteDeleteConfirm      Screen = "site_delete_confirm"
)

// AppState holds all application state for the TUI
type AppState struct {
	// Data
	Sites             []models.Site      `json:"sites"`
	Domains           []models.Domain    `json:"domains"`
	Nodes             []models.Node      `json:"nodes"`
	DockerCredentials []DockerCredential `json:"docker_credentials"`

	// UI State
	CurrentScreen   Screen   `json:"current_screen"`
	PreviousScreens []Screen `json:"previous_screens"` // Navigation stack for back button

	// Selection state (for table lists)
	SitesListIndex            int       `json:"sites_list_index"`
	DomainsListIndex          int       `json:"domains_list_index"`
	NodesListIndex            int       `json:"nodes_list_index"`
	DockerCredentialsListIndex int      `json:"docker_credentials_list_index"`
	SelectedSiteID            uuid.UUID `json:"selected_site_id"`             // For editing site
	SelectedDomainID          uuid.UUID `json:"selected_domain_id"`           // For editing domain
	SelectedNodeID            uuid.UUID `json:"selected_node_id"`             // For viewing/editing node config
	SelectedDockerCredentialID uuid.UUID `json:"selected_docker_credential_id"` // For editing Docker credential

	// Table component instances (runtime only, not serialized)
	SitesTable             TableComponent `json:"-"`
	DomainsTable           TableComponent `json:"-"`
	NodesTable             TableComponent `json:"-"`
	DockerCredentialsTable TableComponent `json:"-"`

	// Viewport for scrollable content (runtime only, not serialized)
	NodeConfigViewport viewport.Model `json:"-"`

	// Form state (for create/edit screens)
	FormFields        []string    `json:"form_fields"`        // Current values of form fields
	CurrentFieldIndex int         `json:"current_field_index"` // Which field has focus
	CursorPosition    int         `json:"cursor_position"`     // Cursor position within current field
	DropdownOpen      bool        `json:"dropdown_open"`       // Is a dropdown currently expanded
	DropdownIndex     int         `json:"dropdown_index"`      // Currently highlighted option in dropdown
	EnvVarPairs       []EnvVarPair `json:"env_var_pairs"`       // Environment variable key-value pairs
	EnvVarFocusedPair int         `json:"env_var_focused_pair"` // Which ENV pair is currently focused
	EnvVarFocusedField int        `json:"env_var_focused_field"` // 0=key, 1=value

	// Domain mappings for multi-domain sites
	DomainMappingPairs       []DomainMappingPair `json:"domain_mapping_pairs"`       // Domain mapping entries
	DomainMappingFocusedPair int                 `json:"domain_mapping_focused_pair"` // Which mapping is currently focused
	DomainMappingFocusedField int               `json:"domain_mapping_focused_field"` // 0=subdomain, 1=domain, 2=port

	// Volume bind mounts for container sites
	VolumePairs        []VolumePair `json:"volume_pairs"`
	VolumeFocusedPair  int          `json:"volume_focused_pair"`
	VolumeFocusedField int          `json:"volume_focused_field"` // 0=host_path, 1=container_path

	// Edit form initialization tracking
	EditFormInitialized bool `json:"edit_form_initialized"` // Track if edit form data has been loaded

	// Deletion confirmation state
	DeletionConfirmPending bool      `json:"deletion_confirm_pending"` // Whether a deletion confirmation is in progress
	DeletionConfirmInput   string    `json:"deletion_confirm_input"`   // User's typed confirmation input
	DeletionTargetID       uuid.UUID `json:"deletion_target_id"`       // ID of item pending deletion
	DeletionTargetName     string    `json:"deletion_target_name"`     // Name of item pending deletion (for comparison)
	DeletionTargetType     string    `json:"deletion_target_type"`     // Type of item: "site", "domain", "node"

	// Quick config state (dpaste.org based)
	QuickConfigURL           string    `json:"quick_config_url"`             // dpaste URL for sharing
	QuickConfigExpiresAt     string    `json:"quick_config_expires_at"`      // Expiration time
	QuickConfigNodeID        uuid.UUID `json:"quick_config_node_id"`         // Node being configured
	QuickConfigHealthConfirmed bool    `json:"quick_config_health_confirmed"` // Whether health check confirmed

	// Compose deployment state (for create/edit screens)
	SiteTypeSelection  string `json:"site_type_selection"`  // "container" or "compose"
	ComposeInputMethod string `json:"compose_input_method"` // "file" or "paste"
	ComposeFilePath    string `json:"compose_file_path"`    // Path to compose file (when input method is "file")
	ComposeContent     string `json:"compose_content"`      // Pasted compose YAML content (when input method is "paste")

	// Async operations tracking
	PendingOperations       []AsyncOperation `json:"pending_operations"`
	Notifications           []Notification   `json:"notifications"`
	ForceRefreshInProgress  bool             `json:"force_refresh_in_progress"`
	ForceRefreshTotal       int              `json:"force_refresh_total"`
	ForceRefreshCompleted   int              `json:"force_refresh_completed"`

	// Window dimensions (updated on resize)
	WindowWidth  int `json:"window_width"`
	WindowHeight int `json:"window_height"`

	// Configuration
	ConfigPath              string `json:"config_path"`
	AutoSave                bool   `json:"auto_save"`
	ShouldQuit              bool   `json:"should_quit"`
	CloudflareAPIToken      string `json:"cloudflare_api_token"`       // Global default, can be overridden per-domain
	Route53AccessKey        string `json:"route53_access_key"`         // Global default, can be overridden per-domain
	Route53SecretKey        string `json:"route53_secret_key"`         // Global default, can be overridden per-domain
	HealthCheckIntervalSecs int    `json:"health_check_interval_secs"` // Interval for automatic health checks
}

// EnvVarPair represents a single environment variable key-value pair
type EnvVarPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DockerCredential represents stored Docker registry credentials (mirrors config.DockerCredential)
type DockerCredential struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`     // Display name (e.g., "GitHub Container Registry", "DockerHub")
	Registry string    `json:"registry"` // Registry URL (e.g., "ghcr.io", "docker.io")
	Username string    `json:"username"`
	Token    string    `json:"token"`
}

// VolumePair represents a volume bind mount entry in the UI
type VolumePair struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
}

// DomainMappingPair represents a port-to-domain mapping entry in the UI
type DomainMappingPair struct {
	Subdomain  string `json:"subdomain"`   // Optional subdomain (e.g., "api", "www")
	DomainName string `json:"domain_name"` // Domain name (for display in UI)
	DomainID   string `json:"domain_id"`   // UUID as string
	Port       string `json:"port"`        // Port number as string (for form input)
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

	// Reset edit form flag when leaving edit screen for non-ENV screen
	if s.CurrentScreen == ScreenSiteEdit && screen != ScreenSiteEnvVars {
		s.EditFormInitialized = false
	}

	// Switch to new screen
	s.CurrentScreen = screen

	// Reset form state when navigating (except when going to/from ENV screen)
	if screen != ScreenSiteEnvVars {
		s.FormFields = []string{}
		s.CurrentFieldIndex = 0
		s.CursorPosition = 0
		s.DropdownOpen = false
		s.DropdownIndex = 0
		s.EnvVarPairs = []EnvVarPair{}
		s.EnvVarFocusedPair = 0
		s.EnvVarFocusedField = 0
		s.DomainMappingPairs = []DomainMappingPair{}
		s.DomainMappingFocusedPair = 0
		s.DomainMappingFocusedField = 0
		s.VolumePairs = []VolumePair{}
		s.VolumeFocusedPair = 0
		s.VolumeFocusedField = 0
		s.SiteTypeSelection = "container" // Default to container
		s.ComposeInputMethod = "file"     // Default to file input
		s.ComposeFilePath = ""
		s.ComposeContent = ""
	}

	// Reset deletion confirmation state when leaving confirmation screen
	if screen != ScreenSiteDeleteConfirm {
		s.DeletionConfirmPending = false
		s.DeletionConfirmInput = ""
		s.DeletionTargetID = uuid.Nil
		s.DeletionTargetName = ""
		s.DeletionTargetType = ""
	}
}

// NavigateBack goes back to the previous screen in history
func (s *AppState) NavigateBack() {
	if len(s.PreviousScreens) > 0 {
		// Pop from history stack
		lastIndex := len(s.PreviousScreens) - 1
		targetScreen := s.PreviousScreens[lastIndex]
		s.PreviousScreens = s.PreviousScreens[:lastIndex]

		// Reset edit form flag when leaving edit screen for non-ENV screen
		if s.CurrentScreen == ScreenSiteEdit && targetScreen != ScreenSiteEnvVars {
			s.EditFormInitialized = false
		}

		s.CurrentScreen = targetScreen
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
	for i := range s.Sites {
		if s.Sites[i].ID == id {
			return &s.Sites[i]
		}
	}
	return nil
}

// GetDomainByID finds a domain by its UUID
func (s *AppState) GetDomainByID(id uuid.UUID) *models.Domain {
	for i := range s.Domains {
		if s.Domains[i].ID == id {
			return &s.Domains[i]
		}
	}
	return nil
}

// GetNodeByID finds a node by its UUID
func (s *AppState) GetNodeByID(id uuid.UUID) *models.Node {
	for i := range s.Nodes {
		if s.Nodes[i].ID == id {
			return &s.Nodes[i]
		}
	}
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
