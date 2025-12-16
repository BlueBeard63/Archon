package app

import (
	"fmt"

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
	// Create config loader
	loader := config.NewFileConfigLoader()

	// Load config from path
	cfg, err := loader.Load(configPath)
	if err != nil {
		return nil, err
	}

	// Initialize AppState from config
	appState := state.NewAppState()
	appState.Sites = cfg.Sites
	appState.Domains = cfg.Domains
	appState.Nodes = cfg.Nodes
	appState.ConfigPath = configPath
	appState.AutoSave = cfg.Settings.AutoSave

	return &Model{
		state:        appState,
		nodeClient:   api.NewHTTPNodeClient(),
		configLoader: loader,
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
		// Let table handle navigation on list screens
		switch m.state.CurrentScreen {
		case state.ScreenSitesList:
			if m.state.SitesTable != nil {
				switch msg.String() {
				case "up", "k", "down", "j", "pgup", "pgdown", "home", "end":
					cmd := m.state.SitesTable.Update(msg)
					m.state.SitesListIndex = m.state.SitesTable.GetCursor()
					return m, cmd
				}
			}
		case state.ScreenDomainsList:
			if m.state.DomainsTable != nil {
				switch msg.String() {
				case "up", "k", "down", "j", "pgup", "pgdown", "home", "end":
					cmd := m.state.DomainsTable.Update(msg)
					m.state.DomainsListIndex = m.state.DomainsTable.GetCursor()
					return m, cmd
				}
			}
		case state.ScreenNodesList:
			if m.state.NodesTable != nil {
				switch msg.String() {
				case "up", "k", "down", "j", "pgup", "pgdown", "home", "end":
					cmd := m.state.NodesTable.Update(msg)
					m.state.NodesListIndex = m.state.NodesTable.GetCursor()
					return m, cmd
				}
			}
		}

		// Fallback to handlers for other keys
		return m.handleKeyPress(msg)

	case tea.MouseMsg:
		// Handle bubblezone clicks
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			// Check if click was on a tab
			if m.zone.Get("tab:dashboard").InBounds(msg) {
				m.state.NavigateTo(state.ScreenDashboard)
				return m, nil
			}
			if m.zone.Get("tab:sites").InBounds(msg) {
				m.state.NavigateTo(state.ScreenSitesList)
				return m, nil
			}
			if m.zone.Get("tab:domains").InBounds(msg) {
				m.state.NavigateTo(state.ScreenDomainsList)
				return m, nil
			}
			if m.zone.Get("tab:nodes").InBounds(msg) {
				m.state.NavigateTo(state.ScreenNodesList)
				return m, nil
			}
			if m.zone.Get("tab:help").InBounds(msg) {
				m.state.NavigateTo(state.ScreenHelp)
				return m, nil
			}

			// Check if click was on a button
			if m.zone.Get("button:create-site").InBounds(msg) {
				m.state.NavigateTo(state.ScreenSiteCreate)
				return m, nil
			}
			if m.zone.Get("button:create-domain").InBounds(msg) {
				m.state.NavigateTo(state.ScreenDomainCreate)
				return m, nil
			}
			if m.zone.Get("button:create-node").InBounds(msg) {
				m.state.NavigateTo(state.ScreenNodeCreate)
				return m, nil
			}

			// Check for row action buttons (edit/delete/view)
			// Sites
			for i, site := range m.state.Sites {
				editID := "button:edit-site-" + site.ID.String()
				deleteID := "button:delete-site-" + site.ID.String()

				if m.zone.Get(editID).InBounds(msg) {
					// Sync table cursor
					m.state.SitesListIndex = i
					if m.state.SitesTable != nil {
						m.state.SitesTable.SetCursor(i)
					}

					// TODO: Navigate to site edit screen
					m.state.AddNotification("Edit site: "+site.Name+" (not yet implemented)", "info")
					return m, nil
				}
				if m.zone.Get(deleteID).InBounds(msg) {
					// Sync table cursor
					m.state.SitesListIndex = i
					if m.state.SitesTable != nil {
						m.state.SitesTable.SetCursor(i)
					}

					// Delete site
					return m.handleDeleteSite(site.ID)
				}
			}

			// Domains
			for i, domain := range m.state.Domains {
				editID := "button:edit-domain-" + domain.ID.String()
				deleteID := "button:delete-domain-" + domain.ID.String()

				if m.zone.Get(editID).InBounds(msg) {
					// Sync table cursor
					m.state.DomainsListIndex = i
					if m.state.DomainsTable != nil {
						m.state.DomainsTable.SetCursor(i)
					}

					// Navigate to domain edit screen
					m.state.SelectedDomainID = domain.ID
					m.state.NavigateTo(state.ScreenDomainEdit)
					return m, nil
				}
				if m.zone.Get(deleteID).InBounds(msg) {
					// Sync table cursor
					m.state.DomainsListIndex = i
					if m.state.DomainsTable != nil {
						m.state.DomainsTable.SetCursor(i)
					}

					// Delete domain
					return m.handleDeleteDomain(domain.ID)
				}
			}

			// Nodes
			for i, node := range m.state.Nodes {
				viewID := "button:view-node-" + node.ID.String()
				editID := "button:edit-node-" + node.ID.String()
				deleteID := "button:delete-node-" + node.ID.String()

				if m.zone.Get(viewID).InBounds(msg) {
					// Sync table cursor
					m.state.NodesListIndex = i
					if m.state.NodesTable != nil {
						m.state.NodesTable.SetCursor(i)
					}

					// View node config
					m.state.SelectedNodeID = node.ID
					m.state.NavigateTo(state.ScreenNodeConfig)
					return m, nil
				}
				if m.zone.Get(editID).InBounds(msg) {
					// Sync table cursor
					m.state.NodesListIndex = i
					if m.state.NodesTable != nil {
						m.state.NodesTable.SetCursor(i)
					}

					// TODO: Navigate to node edit screen
					m.state.AddNotification("Edit node: "+node.Name+" (not yet implemented)", "info")
					return m, nil
				}
				if m.zone.Get(deleteID).InBounds(msg) {
					// Sync table cursor
					m.state.NodesListIndex = i
					if m.state.NodesTable != nil {
						m.state.NodesTable.SetCursor(i)
					}

					// Delete node
					return m.handleDeleteNode(node.ID)
				}
			}

			// Check if click was on a form field
			for i := 0; i < 10; i++ { // Check up to 10 fields
				zoneID := fmt.Sprintf("field:%d", i)
				if m.zone.Get(zoneID).InBounds(msg) {
					if i < len(m.state.FormFields) {
						m.state.CurrentFieldIndex = i
						return m, nil
					}
				}
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
