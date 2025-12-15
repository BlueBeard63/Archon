package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/api"
	"github.com/BlueBeard63/archon/internal/config"
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui"
)

// Used to suppress unused variable warnings
var _ = fmt.Sprint

// Model implements tea.Model for Bubbletea's Elm Architecture
type Model struct {
	state        *state.AppState
	nodeClient   api.NodeClient
	configLoader config.ConfigLoader
	configPath   string
	zone         *zone.Manager
}

// NewModel creates a new application model with initial state
func NewModel(configPath string) (*Model, error) {
	// TODO: Implement model initialization
	// Steps:
	// 1. Create config loader
	// 2. Load config from path
	// 3. Initialize AppState from config
	// 4. Create node client
	// 5. Return model

	// Example structure:
	// loader := config.NewFileConfigLoader()
	// cfg, err := loader.Load(configPath)
	// if err != nil {
	//     return nil, err
	// }
	//
	// appState := state.NewAppState()
	// appState.Sites = cfg.Sites
	// appState.Domains = cfg.Domains
	// appState.Nodes = cfg.Nodes
	// appState.ConfigPath = configPath
	// appState.AutoSave = cfg.Settings.AutoSave
	//
	// return &Model{
	//     state:        appState,
	//     nodeClient:   api.NewHTTPNodeClient(),
	//     configLoader: loader,
	//     configPath:   configPath,
	// }, nil

	return &Model{
		state:        state.NewAppState(),
		nodeClient:   api.NewHTTPNodeClient(),
		configLoader: config.NewFileConfigLoader(),
		configPath:   configPath,
		zone:         zone.New(),
	}, nil
}

// Init is called once when the program starts (TEA pattern)
// Returns initial commands to run
func (m Model) Init() tea.Cmd {
	// Return batch of initialization commands
	return tea.Batch(
		tea.EnterAltScreen,        // Enable alternate screen buffer
		tea.EnableMouseCellMotion, // MOUSE SUPPORT: Enable mouse events
	)
}

// Update handles incoming messages and updates the model (TEA pattern)
// This is the core of the application's state management
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ========================================================================
	// System Events
	// ========================================================================

	case tea.KeyMsg:
		// Delegate to key handler (implemented in handlers.go)
		return m.handleKeyPress(msg)

	case tea.MouseMsg:
		// Handle bubblezone clicks first
		if msg.Action == tea.MouseActionRelease {
			// Get zone ID at click position
			// bubblezone tracks zones during Scan() and Get() returns zone ID as string
			prefix, zoneID := m.zone.GetPrefix(msg.X, msg.Y)

			// Check if it's a menu item
			if prefix == "menu" {
				return m.handleMenuClick("menu:" + zoneID)
			}

			// Check if it's a form field
			if prefix == "field" {
				return m.handleFieldClick("field:" + zoneID)
			}
		}
		// Fallback to traditional mouse handler
		return m.handleMouseClick(msg)

	case tea.WindowSizeMsg:
		// Update window dimensions
		m.state.WindowWidth = msg.Width
		m.state.WindowHeight = msg.Height
		return m, nil

	// ========================================================================
	// Navigation
	// ========================================================================

	case NavigateToMsg:
		m.state.NavigateTo(msg.Screen)
		return m, nil

	case NavigateBackMsg:
		m.state.NavigateBack()
		return m, nil

	// ========================================================================
	// Site Operations
	// ========================================================================

	case CreateSiteMsg:
		// TODO: Add site to state, optionally trigger auto-save
		// m.state.Sites = append(m.state.Sites, *msg.Site)
		// if m.state.AutoSave {
		//     return m, m.saveConfig()
		// }
		return m, nil

	case DeploySiteMsg:
		// Spawn async deployment operation
		return m, m.spawnDeploySite(msg.SiteID)

	case SiteDeployedMsg:
		// TODO: Update site status based on result
		// if msg.Error != nil {
		//     m.state.AddNotification("Deployment failed: "+msg.Error.Error(), "error")
		// } else {
		//     m.state.AddNotification("Site deployed successfully", "success")
		// }
		return m, nil

	// ========================================================================
	// Domain Operations
	// ========================================================================

	case CreateDomainMsg:
		// TODO: Add domain to state, optionally trigger auto-save
		return m, nil

	case SyncDnsMsg:
		// Spawn async DNS sync operation
		return m, m.spawnSyncDns(msg.DomainID)

	case DnsSyncedMsg:
		// TODO: Update domain's DNS records with synced data
		return m, nil

	// ========================================================================
	// Node Operations
	// ========================================================================

	case CreateNodeMsg:
		// TODO: Add node to state, optionally trigger auto-save
		return m, nil

	case NodeHealthCheckMsg:
		// Spawn async health check
		return m, m.spawnNodeHealthCheck(msg.NodeID)

	case NodeHealthCheckResultMsg:
		// TODO: Update node status with health check result
		return m, nil

	// ========================================================================
	// Form Handling
	// ========================================================================

	case FormInputMsg:
		// TODO: Append character to current form field
		// if m.state.CurrentFieldIndex < len(m.state.FormFields) {
		//     m.state.FormFields[m.state.CurrentFieldIndex] += string(msg.Char)
		// }
		return m, nil

	case FormBackspaceMsg:
		// TODO: Remove last character from current field
		return m, nil

	case FormSubmitMsg:
		// Delegate to form submit handler
		return m.handleFormSubmit()

	// ========================================================================
	// Configuration
	// ========================================================================

	case SaveConfigMsg:
		return m, m.saveConfig()

	case ConfigSavedMsg:
		if msg.Error != nil {
			m.state.AddNotification("Failed to save config: "+msg.Error.Error(), "error")
		}
		return m, nil

	// ========================================================================
	// System
	// ========================================================================

	case NotificationMsg:
		m.state.AddNotification(msg.Message, msg.Level)
		return m, nil

	case QuitMsg:
		m.state.ShouldQuit = true
		return m, tea.Quit
	}

	return m, nil
}

// View renders the UI (TEA pattern)
// Returns a string to be printed to the terminal
func (m Model) View() string {
	// Check if we should quit
	if m.state.ShouldQuit {
		return "Goodbye!\n"
	}

	// Delegate to UI package for rendering with zone manager
	return m.zone.Scan(ui.RenderWithZones(m.state, m.zone))
}

// ============================================================================
// Async Operation Spawners
// ============================================================================
// These functions return tea.Cmd that run async operations and return messages

func (m Model) spawnDeploySite(siteID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement deployment logic
		// 1. Get site from state by ID
		// 2. Get node from state by site.NodeID
		// 3. Get domain from state by site.DomainID
		// 4. Call nodeClient.DeploySite()
		// 5. Return SiteDeployedMsg with result

		// site := m.state.GetSiteByID(siteID)
		// node := m.state.GetNodeByID(site.NodeID)
		// domain := m.state.GetDomainByID(site.DomainID)
		//
		// err := m.nodeClient.DeploySite(
		//     node.APIEndpoint,
		//     node.APIKey,
		//     site,
		//     domain.Name,
		// )
		//
		// return SiteDeployedMsg{
		//     SiteID: siteID,
		//     Error:  err,
		// }

		return SiteDeployedMsg{SiteID: siteID, Error: nil}
	}
}

func (m Model) spawnSyncDns(domainID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement DNS sync logic
		// 1. Get domain from state by ID
		// 2. Create DNS provider from domain.DnsProvider
		// 3. Call provider.ListRecords()
		// 4. Return DnsSyncedMsg with records

		return DnsSyncedMsg{DomainID: domainID, Error: nil}
	}
}

func (m Model) spawnNodeHealthCheck(nodeID uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement health check logic
		// 1. Get node from state by ID
		// 2. Call nodeClient.HealthCheck()
		// 3. Return NodeHealthCheckResultMsg

		return NodeHealthCheckResultMsg{NodeID: nodeID, Error: nil}
	}
}

func (m Model) saveConfig() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement config save
		// 1. Build config.Config from m.state
		// 2. Call m.configLoader.Save()
		// 3. Return ConfigSavedMsg

		// cfg := &config.Config{
		//     Version:  "1.0.0",
		//     Sites:    m.state.Sites,
		//     Domains:  m.state.Domains,
		//     Nodes:    m.state.Nodes,
		//     Settings: ...,
		// }
		//
		// err := m.configLoader.Save(m.configPath, cfg)
		// return ConfigSavedMsg{Error: err}

		return ConfigSavedMsg{Error: nil}
	}
}
