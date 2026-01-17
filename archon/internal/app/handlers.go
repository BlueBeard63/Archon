package app

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/BlueBeard63/archon/internal/compose"
	"github.com/BlueBeard63/archon/internal/config"
	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
)

// ============================================================================
// Keyboard Event Handlers
// ============================================================================

// handleTextInput handles text input for a form field with cursor support
func (m *Model) handleTextInput(msg tea.KeyMsg, fieldIndex int) bool {
	if fieldIndex >= len(m.state.FormFields) {
		return false
	}

	value := m.state.FormFields[fieldIndex]
	cursor := m.state.CursorPosition

	// Ensure cursor is within bounds
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(value) {
		cursor = len(value)
	}

	switch msg.Type {
	case tea.KeyLeft:
		// Move cursor left
		if cursor > 0 {
			m.state.CursorPosition--
		}
		return true

	case tea.KeyRight:
		// Move cursor right
		if cursor < len(value) {
			m.state.CursorPosition++
		}
		return true

	case tea.KeyHome, tea.KeyCtrlA:
		// Move cursor to start
		m.state.CursorPosition = 0
		return true

	case tea.KeyEnd, tea.KeyCtrlE:
		// Move cursor to end
		m.state.CursorPosition = len(value)
		return true

	case tea.KeyBackspace:
		// Delete character before cursor
		if cursor > 0 {
			m.state.FormFields[fieldIndex] = value[:cursor-1] + value[cursor:]
			m.state.CursorPosition--
		}
		return true

	case tea.KeyDelete:
		// Delete character at cursor
		if cursor < len(value) {
			m.state.FormFields[fieldIndex] = value[:cursor] + value[cursor+1:]
		}
		return true

	case tea.KeySpace:
		// Insert space at cursor
		m.state.FormFields[fieldIndex] = value[:cursor] + " " + value[cursor:]
		m.state.CursorPosition++
		return true

	case tea.KeyRunes:
		// Insert character at cursor
		m.state.FormFields[fieldIndex] = value[:cursor] + string(msg.Runes) + value[cursor:]
		m.state.CursorPosition++
		return true
	}

	return false
}

// setFieldAndResetCursor sets the current field index and moves cursor to end
func (m *Model) setFieldAndResetCursor(fieldIndex int) {
	m.state.CurrentFieldIndex = fieldIndex
	if fieldIndex < len(m.state.FormFields) {
		m.state.CursorPosition = len(m.state.FormFields[fieldIndex])
	} else {
		m.state.CursorPosition = 0
	}
}

// handleKeyPress routes keyboard events to screen-specific handlers
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on a form screen (prioritize form input)
	isFormScreen := m.state.CurrentScreen == state.ScreenSiteCreate ||
		m.state.CurrentScreen == state.ScreenSiteEdit ||
		m.state.CurrentScreen == state.ScreenDomainCreate ||
		m.state.CurrentScreen == state.ScreenDomainEdit ||
		m.state.CurrentScreen == state.ScreenNodeCreate ||
		m.state.CurrentScreen == state.ScreenNodeEdit ||
		m.state.CurrentScreen == state.ScreenNodeConfigSave ||
		m.state.CurrentScreen == state.ScreenSettings

	// Critical global key bindings (work on all screens)
	switch msg.String() {
	case "ctrl+c":
		// Quit application (always available)
		return m, func() tea.Msg { return QuitMsg{} }

	case "esc":
		// Go back to previous screen (always available)
		m.state.NavigateBack()
		return m, nil
	}

	// Non-form global key bindings (skip on form screens to allow text input)
	if !isFormScreen {
		switch msg.String() {
		case "q":
			// Quit application
			return m, func() tea.Msg { return QuitMsg{} }

		case "0":
			// Go to dashboard
			m.state.NavigateTo(state.ScreenDashboard)
			return m, nil

		case "?":
			// Show help screen
			m.state.NavigateTo(state.ScreenHelp)
			return m, nil

		case "ctrl+s":
			// Manual save
			return m, func() tea.Msg { return SaveConfigMsg{} }
		}
	}

	// Screen-specific key bindings
	switch m.state.CurrentScreen {
	case state.ScreenDashboard:
		return m.handleDashboardKeys(msg)
	case state.ScreenSitesList:
		return m.handleSitesListKeys(msg)
	case state.ScreenSiteCreate:
		return m.handleSiteCreateKeys(msg)
	case state.ScreenSiteEdit:
		return m.handleSiteEditKeys(msg)
	case state.ScreenSiteEnvVars:
		return m.handleSiteEnvVarsKeys(msg)
	case state.ScreenDomainsList:
		return m.handleDomainsListKeys(msg)
	case state.ScreenDomainCreate:
		return m.handleDomainCreateKeys(msg)
	case state.ScreenDomainEdit:
		return m.handleDomainEditKeys(msg)
	case state.ScreenNodesList:
		return m.handleNodesListKeys(msg)
	case state.ScreenNodeCreate:
		return m.handleNodeCreateKeys(msg)
	case state.ScreenNodeEdit:
		return m.handleNodeEditKeys(msg)
	case state.ScreenSettings:
		return m.handleSettingsKeys(msg)
	case state.ScreenNodeConfig:
		return m.handleNodeConfigKeys(msg)
	case state.ScreenNodeConfigSave:
		return m.handleNodeConfigSaveKeys(msg)
	case state.ScreenHelp:
		return m.handleHelpKeys(msg)
	}

	return m, nil
}

// handleDashboardKeys handles keys on the dashboard screen
func (m Model) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "1", "s":
		m.state.NavigateTo(state.ScreenSitesList)
		return m, nil
	case "2", "d":
		m.state.NavigateTo(state.ScreenDomainsList)
		return m, nil
	case "3", "n":
		m.state.NavigateTo(state.ScreenNodesList)
		return m, nil
	case "4", "c":
		m.state.NavigateTo(state.ScreenSettings)
		return m, nil
	}

	return m, nil
}

// handleSitesListKeys handles keys on the sites list screen
func (m Model) handleSitesListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "c":
		m.state.NavigateTo(state.ScreenSiteCreate)
		return m, nil

	case " ", "enter":
		// Deploy selected site
		if len(m.state.Sites) > 0 && m.state.SitesListIndex >= 0 && m.state.SitesListIndex < len(m.state.Sites) {
			site := m.state.Sites[m.state.SitesListIndex]
			m.state.AddNotification("Deploying site: "+site.Name, "info")
			return m, m.spawnDeploySite(site.ID)
		}
		return m, nil

	case "e":
		// Edit selected site
		if len(m.state.Sites) > 0 && m.state.SitesListIndex >= 0 && m.state.SitesListIndex < len(m.state.Sites) {
			site := m.state.Sites[m.state.SitesListIndex]
			m.state.SelectedSiteID = site.ID
			m.state.CurrentFieldIndex = 0 // Reset to first field
			m.state.NavigateTo(state.ScreenSiteEdit)
		}
		return m, nil

	case "d":
		// Delete selected site
		if len(m.state.Sites) > 0 && m.state.SitesListIndex >= 0 && m.state.SitesListIndex < len(m.state.Sites) {
			site := m.state.Sites[m.state.SitesListIndex]
			return m.handleDeleteSite(site.ID)
		}
		return m, nil

	case "s":
		// Toggle start/stop for selected site
		if len(m.state.Sites) > 0 && m.state.SitesListIndex >= 0 && m.state.SitesListIndex < len(m.state.Sites) {
			site := m.state.Sites[m.state.SitesListIndex]

			// If running or deploying, stop it
			if site.Status == models.SiteStatusRunning || site.Status == models.SiteStatusDeploying {
				m.state.AddNotification("Stopping site: "+site.Name, "info")
				return m, m.spawnStopSite(site.ID)
			}

			// If stopped or failed, restart it
			if site.Status == models.SiteStatusStopped || site.Status == models.SiteStatusFailed {
				m.state.AddNotification("Restarting site: "+site.Name, "info")
				return m, m.spawnRestartSite(site.ID)
			}

			// If inactive, do nothing (user should press space/enter to deploy)
			if site.Status == models.SiteStatusInactive || site.Status == "" {
				m.state.AddNotification("Site is inactive. Press Space/Enter to deploy", "info")
			}
		}
		return m, nil

	case "r":
		// Setup DNS records for selected site
		if len(m.state.Sites) > 0 && m.state.SitesListIndex >= 0 && m.state.SitesListIndex < len(m.state.Sites) {
			site := m.state.Sites[m.state.SitesListIndex]
			m.state.AddNotification("Setting up DNS records for: "+site.Name, "info")
			return m, m.spawnSetupDNS(site.ID)
		}
		return m, nil
	}

	return m, nil
}

// handleSiteCreateKeys handles keys on the site creation form
func (m Model) handleSiteCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on a dropdown field (SiteType=-1, Node=1)
	isSiteTypeField := m.state.CurrentFieldIndex == -1
	isNodeField := m.state.CurrentFieldIndex == 1
	isDropdownField := isSiteTypeField || isNodeField

	// Helper to get next visible field index (skips hidden fields based on site type)
	getNextVisibleField := func(current int) int {
		isCompose := m.state.SiteTypeSelection == "compose"
		next := current + 1

		// For compose mode: skip fields 3 (docker username), 4 (docker token), 6 (config file)
		// For container mode: all fields are visible
		for next < len(m.state.FormFields) {
			if isCompose && (next == 3 || next == 4 || next == 6) {
				next++
				continue
			}
			break
		}

		if next >= len(m.state.FormFields) {
			// Move to domain mapping section
			return 200
		}
		return next
	}

	// Helper to get previous visible field index
	getPrevVisibleField := func(current int) int {
		isCompose := m.state.SiteTypeSelection == "compose"
		prev := current - 1

		// For compose mode: skip fields 6, 4, 3
		for prev >= 0 {
			if isCompose && (prev == 3 || prev == 4 || prev == 6) {
				prev--
				continue
			}
			break
		}

		if prev < 0 {
			return -1 // Site type selector
		}
		return prev
	}

	// Handle dropdown-specific keys when dropdown is open
	if m.state.DropdownOpen && isDropdownField {
		switch msg.Type {
		case tea.KeyUp:
			// Navigate up in dropdown
			if m.state.DropdownIndex > 0 {
				m.state.DropdownIndex--
			}
			return m, nil

		case tea.KeyDown:
			// Navigate down in dropdown
			var maxIndex int
			if isSiteTypeField {
				maxIndex = 1 // Container, Compose
			} else {
				maxIndex = len(m.state.Nodes) - 1
			}
			if m.state.DropdownIndex < maxIndex {
				m.state.DropdownIndex++
			}
			return m, nil

		case tea.KeyEnter, tea.KeyTab:
			// Confirm selection and close dropdown
			if isSiteTypeField {
				// Site type selection
				if m.state.DropdownIndex == 0 {
					m.state.SiteTypeSelection = "container"
				} else {
					m.state.SiteTypeSelection = "compose"
				}
			} else if len(m.state.Nodes) > 0 {
				m.state.FormFields[1] = m.state.Nodes[m.state.DropdownIndex].Name
			}
			m.state.DropdownOpen = false

			// If Tab, move to next field
			if msg.Type == tea.KeyTab {
				m.state.CurrentFieldIndex = getNextVisibleField(m.state.CurrentFieldIndex)
				if m.state.CurrentFieldIndex == 200 {
					m.state.DomainMappingFocusedPair = 0
					m.state.DomainMappingFocusedField = 0
					if len(m.state.DomainMappingPairs) > 0 {
						m.state.CursorPosition = len(m.state.DomainMappingPairs[0].Subdomain)
					}
				}
			}
			return m, nil

		case tea.KeyEsc:
			// Close dropdown without selecting
			m.state.DropdownOpen = false
			return m, nil

		case tea.KeyBackspace, tea.KeyRunes, tea.KeySpace:
			// Close dropdown and allow manual input (not for site type - it's a fixed dropdown)
			if !isSiteTypeField {
				m.state.DropdownOpen = false
			}
			// Fall through to normal input handling for node field
			if isSiteTypeField {
				return m, nil
			}
		}
	}

	// Handle domain mapping input if focused on domain mapping section
	if m.state.CurrentFieldIndex == 200 {
		return m.handleDomainMappingInput(msg)
	}

	// Normal field input handling
	switch msg.Type {
	case tea.KeyUp:
		// Open dropdown on up arrow if on dropdown field and not open
		if isDropdownField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}

	case tea.KeyDown:
		// Open dropdown on down arrow if on dropdown field and not open
		if isDropdownField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}

	case tea.KeySpace:
		// Add space to current field (not site type)
		if m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to current field (not site type)
		if m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from current field (not site type)
		if m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			value := m.state.FormFields[m.state.CurrentFieldIndex]
			if len(value) > 0 {
				m.state.FormFields[m.state.CurrentFieldIndex] = value[:len(value)-1]
			}
		}
		return m, nil

	case tea.KeyTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}

		// Auto-detect ports when leaving compose file path field (index 2)
		previousField := m.state.CurrentFieldIndex
		if previousField == 2 && m.state.SiteTypeSelection == "compose" && m.state.FormFields[2] != "" {
			m.tryDetectComposePorts()
		}

		// Move to next visible field
		m.state.CurrentFieldIndex = getNextVisibleField(m.state.CurrentFieldIndex)
		if m.state.CurrentFieldIndex == 200 {
			m.state.DomainMappingFocusedPair = 0
			m.state.DomainMappingFocusedField = 0
			if len(m.state.DomainMappingPairs) > 0 {
				m.state.CursorPosition = len(m.state.DomainMappingPairs[0].Subdomain)
			}
		}
		return m, nil

	case tea.KeyShiftTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}
		// Move to previous visible field
		m.state.CurrentFieldIndex = getPrevVisibleField(m.state.CurrentFieldIndex)
		return m, nil

	case tea.KeyEnter:
		// If on dropdown field and not open, open it
		if isDropdownField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}
		// Otherwise submit form
		return m.handleSiteCreateSubmit()
	}

	return m, nil
}

// handleSiteEditKeys handles keys on the site edit form
func (m Model) handleSiteEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on a dropdown field (Node=1 in new layout)
	isDropdownField := m.state.CurrentFieldIndex == 1

	// Handle dropdown-specific keys when dropdown is open
	if m.state.DropdownOpen && isDropdownField {
		switch msg.Type {
		case tea.KeyUp:
			// Navigate up in dropdown
			if m.state.DropdownIndex > 0 {
				m.state.DropdownIndex--
			}
			return m, nil

		case tea.KeyDown:
			// Navigate down in dropdown
			maxIndex := len(m.state.Nodes) - 1
			if m.state.DropdownIndex < maxIndex {
				m.state.DropdownIndex++
			}
			return m, nil

		case tea.KeyEnter, tea.KeyTab:
			// Confirm selection and close dropdown
			if len(m.state.Nodes) > 0 {
				m.state.FormFields[1] = m.state.Nodes[m.state.DropdownIndex].Name
			}
			m.state.DropdownOpen = false

			// If Tab, move to next field
			if msg.Type == tea.KeyTab {
				m.state.CurrentFieldIndex++
			}
			return m, nil

		case tea.KeyEsc:
			// Close dropdown without selecting
			m.state.DropdownOpen = false
			return m, nil

		case tea.KeyBackspace, tea.KeyRunes, tea.KeySpace:
			// Close dropdown and allow manual input
			m.state.DropdownOpen = false
			// Fall through to normal input handling
		}
	}

	// Handle domain mapping input if focused on domain mapping section
	if m.state.CurrentFieldIndex == 200 {
		return m.handleDomainMappingInput(msg)
	}

	// Normal field input handling
	switch msg.Type {
	case tea.KeyUp:
		// Open dropdown on up arrow if on dropdown field and not open
		if isDropdownField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}

	case tea.KeyDown:
		// Open dropdown on down arrow if on dropdown field and not open
		if isDropdownField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}

	case tea.KeySpace:
		// Add space to current field
		if m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Handle 'v' key to navigate to ENV vars screen (container sites only)
		if string(msg.Runes) == "v" {
			site := m.state.GetSiteByID(m.state.SelectedSiteID)
			if site != nil && site.GetSiteType() != models.SiteTypeCompose {
				m.state.EnvVarPairs = []state.EnvVarPair{} // Clear to trigger reload
				m.state.NavigateTo(state.ScreenSiteEnvVars)
				return m, nil
			}
		}
		// Add character to current field
		if m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from current field
		if m.state.CurrentFieldIndex < len(m.state.FormFields) {
			value := m.state.FormFields[m.state.CurrentFieldIndex]
			if len(value) > 0 {
				m.state.FormFields[m.state.CurrentFieldIndex] = value[:len(value)-1]
			}
		}
		return m, nil

	case tea.KeyTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}
		// Move to next field or domain mapping section
		m.state.CurrentFieldIndex++
		if m.state.CurrentFieldIndex >= len(m.state.FormFields) {
			// Move to domain mapping section
			m.state.CurrentFieldIndex = 200
			m.state.DomainMappingFocusedPair = 0
			m.state.DomainMappingFocusedField = 0
			m.state.CursorPosition = len(m.state.DomainMappingPairs[0].Subdomain)
		}
		return m, nil

	case tea.KeyShiftTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}
		// Move to previous field
		m.state.CurrentFieldIndex--
		if m.state.CurrentFieldIndex < 0 {
			m.state.CurrentFieldIndex = len(m.state.FormFields) - 1
		}
		return m, nil

	case tea.KeyEnter:
		// If on dropdown field and not open, open it
		if isDropdownField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}
		// Otherwise submit form
		return m.handleSiteEditSubmit()
	}

	return m, nil
}

// handleSiteEnvVarsKeys handles keys on the dedicated ENV vars screen
func (m Model) handleSiteEnvVarsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Go back to edit screen
		m.state.NavigateBack()
		return m, nil
	case tea.KeyEnter:
		// Save and go back (ENV vars already in state, will be saved with site)
		m.state.NavigateBack()
		return m, nil
	default:
		// Delegate to existing ENV input handler
		return m.handleEnvVarInput(msg)
	}
}

// handleDomainMappingInput handles keyboard input for domain mapping fields
func (m Model) handleDomainMappingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.state.DomainMappingPairs) == 0 {
		return m, nil
	}

	pairIdx := m.state.DomainMappingFocusedPair
	if pairIdx >= len(m.state.DomainMappingPairs) {
		pairIdx = 0
		m.state.DomainMappingFocusedPair = 0
	}

	// Handle dropdown for domain field (field 1)
	if m.state.DomainMappingFocusedField == 1 && m.state.DropdownOpen {
		switch msg.Type {
		case tea.KeyUp:
			if m.state.DropdownIndex > 0 {
				m.state.DropdownIndex--
			}
			return m, nil
		case tea.KeyDown:
			if m.state.DropdownIndex < len(m.state.Domains)-1 {
				m.state.DropdownIndex++
			}
			return m, nil
		case tea.KeyEnter, tea.KeyTab:
			// Select domain
			if m.state.DropdownIndex >= 0 && m.state.DropdownIndex < len(m.state.Domains) {
				selectedDomain := m.state.Domains[m.state.DropdownIndex]
				m.state.DomainMappingPairs[pairIdx].DomainName = selectedDomain.Name
				m.state.DomainMappingPairs[pairIdx].DomainID = selectedDomain.ID.String()
			}
			m.state.DropdownOpen = false
			if msg.Type == tea.KeyTab {
				// Move to port field
				m.state.DomainMappingFocusedField = 2
				m.state.CursorPosition = len(m.state.DomainMappingPairs[pairIdx].Port)
			}
			return m, nil
		case tea.KeyEsc:
			m.state.DropdownOpen = false
			return m, nil
		}
	}

	switch msg.Type {
	case tea.KeyTab:
		// Switch between subdomain/domain/port, or move to next pair
		if m.state.DomainMappingFocusedField == 0 {
			// Move from subdomain to domain
			m.state.DomainMappingFocusedField = 1
			m.state.DropdownOpen = false
		} else if m.state.DomainMappingFocusedField == 1 {
			// Move from domain to port
			m.state.DomainMappingFocusedField = 2
			m.state.CursorPosition = len(m.state.DomainMappingPairs[pairIdx].Port)
		} else {
			// Move from port to next pair's subdomain, or wrap to first pair
			if pairIdx < len(m.state.DomainMappingPairs)-1 {
				m.state.DomainMappingFocusedPair++
				m.state.DomainMappingFocusedField = 0
				m.state.CursorPosition = len(m.state.DomainMappingPairs[m.state.DomainMappingFocusedPair].Subdomain)
			} else {
				// Wrap to first pair (ENV vars now on separate screen)
				m.state.DomainMappingFocusedPair = 0
				m.state.DomainMappingFocusedField = 0
				m.state.CursorPosition = len(m.state.DomainMappingPairs[0].Subdomain)
			}
		}
		return m, nil

	case tea.KeyEnter:
		// Open domain dropdown when on domain field
		if m.state.DomainMappingFocusedField == 1 && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}
		// Submit form based on current screen
		if m.state.CurrentScreen == state.ScreenSiteEdit {
			return m.handleSiteEditSubmit()
		}
		return m.handleSiteCreateSubmit()

	case tea.KeyDown:
		// Open domain dropdown when on domain field
		if m.state.DomainMappingFocusedField == 1 && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			m.state.DropdownIndex = 0
			return m, nil
		}
		return m, nil

	case tea.KeyEsc:
		// Go back to main form
		m.state.CurrentFieldIndex = len(m.state.FormFields) - 1
		return m, nil

	case tea.KeyLeft:
		if m.state.CursorPosition > 0 {
			m.state.CursorPosition--
		}
		return m, nil

	case tea.KeyRight:
		currentValue := ""
		if m.state.DomainMappingFocusedField == 0 {
			currentValue = m.state.DomainMappingPairs[pairIdx].Subdomain
		} else if m.state.DomainMappingFocusedField == 2 {
			currentValue = m.state.DomainMappingPairs[pairIdx].Port
		}
		if m.state.CursorPosition < len(currentValue) {
			m.state.CursorPosition++
		}
		return m, nil

	case tea.KeyHome:
		m.state.CursorPosition = 0
		return m, nil

	case tea.KeyEnd:
		if m.state.DomainMappingFocusedField == 0 {
			m.state.CursorPosition = len(m.state.DomainMappingPairs[pairIdx].Subdomain)
		} else if m.state.DomainMappingFocusedField == 2 {
			m.state.CursorPosition = len(m.state.DomainMappingPairs[pairIdx].Port)
		}
		return m, nil

	case tea.KeyBackspace:
		cursor := m.state.CursorPosition
		if cursor > 0 {
			if m.state.DomainMappingFocusedField == 0 {
				// Editing subdomain
				value := m.state.DomainMappingPairs[pairIdx].Subdomain
				m.state.DomainMappingPairs[pairIdx].Subdomain = value[:cursor-1] + value[cursor:]
			} else if m.state.DomainMappingFocusedField == 2 {
				// Editing port
				value := m.state.DomainMappingPairs[pairIdx].Port
				m.state.DomainMappingPairs[pairIdx].Port = value[:cursor-1] + value[cursor:]
			}
			m.state.CursorPosition--
		}
		return m, nil

	case tea.KeyDelete:
		cursor := m.state.CursorPosition
		if m.state.DomainMappingFocusedField == 0 {
			// Editing subdomain
			value := m.state.DomainMappingPairs[pairIdx].Subdomain
			if cursor < len(value) {
				m.state.DomainMappingPairs[pairIdx].Subdomain = value[:cursor] + value[cursor+1:]
			}
		} else if m.state.DomainMappingFocusedField == 2 {
			// Editing port
			value := m.state.DomainMappingPairs[pairIdx].Port
			if cursor < len(value) {
				m.state.DomainMappingPairs[pairIdx].Port = value[:cursor] + value[cursor+1:]
			}
		}
		return m, nil

	case tea.KeySpace:
		cursor := m.state.CursorPosition
		if m.state.DomainMappingFocusedField == 0 {
			// Editing subdomain
			value := m.state.DomainMappingPairs[pairIdx].Subdomain
			m.state.DomainMappingPairs[pairIdx].Subdomain = value[:cursor] + " " + value[cursor:]
		} else if m.state.DomainMappingFocusedField == 2 {
			// Editing port (spaces not typically allowed in ports, but handle anyway)
			value := m.state.DomainMappingPairs[pairIdx].Port
			m.state.DomainMappingPairs[pairIdx].Port = value[:cursor] + " " + value[cursor:]
		}
		m.state.CursorPosition++
		return m, nil

	case tea.KeyRunes:
		cursor := m.state.CursorPosition
		if m.state.DomainMappingFocusedField == 0 {
			// Editing subdomain
			value := m.state.DomainMappingPairs[pairIdx].Subdomain
			m.state.DomainMappingPairs[pairIdx].Subdomain = value[:cursor] + string(msg.Runes) + value[cursor:]
		} else if m.state.DomainMappingFocusedField == 2 {
			// Editing port
			value := m.state.DomainMappingPairs[pairIdx].Port
			m.state.DomainMappingPairs[pairIdx].Port = value[:cursor] + string(msg.Runes) + value[cursor:]
		}
		m.state.CursorPosition++
		return m, nil
	}

	return m, nil
}

// handleEnvVarInput handles keyboard input for ENV var fields
func (m Model) handleEnvVarInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if len(m.state.EnvVarPairs) == 0 {
		return m, nil
	}

	pairIdx := m.state.EnvVarFocusedPair
	if pairIdx >= len(m.state.EnvVarPairs) {
		pairIdx = 0
		m.state.EnvVarFocusedPair = 0
	}

	switch msg.Type {
	case tea.KeyTab:
		// Switch between key and value, or move to next pair
		if m.state.EnvVarFocusedField == 0 {
			// Move from key to value
			m.state.EnvVarFocusedField = 1
			m.state.CursorPosition = len(m.state.EnvVarPairs[pairIdx].Value)
		} else {
			// Move from value to next pair's key, or back to regular fields
			if pairIdx < len(m.state.EnvVarPairs)-1 {
				m.state.EnvVarFocusedPair++
				m.state.EnvVarFocusedField = 0
				m.state.CursorPosition = len(m.state.EnvVarPairs[m.state.EnvVarFocusedPair].Key)
			} else {
				// Done with ENV vars, move back to regular fields
				m.state.CurrentFieldIndex = 0
			}
		}
		return m, nil

	case tea.KeyShiftTab:
		// Move backwards: value to key, or to previous pair's value
		if m.state.EnvVarFocusedField == 1 {
			// Move from value to key
			m.state.EnvVarFocusedField = 0
			m.state.CursorPosition = len(m.state.EnvVarPairs[pairIdx].Key)
		} else {
			// Move from key to previous pair's value, or back to domain mappings
			if pairIdx > 0 {
				m.state.EnvVarFocusedPair--
				m.state.EnvVarFocusedField = 1
				m.state.CursorPosition = len(m.state.EnvVarPairs[m.state.EnvVarFocusedPair].Value)
			} else {
				// Back to domain mappings section
				m.state.CurrentFieldIndex = 200
				if len(m.state.DomainMappingPairs) > 0 {
					m.state.DomainMappingFocusedPair = len(m.state.DomainMappingPairs) - 1
					m.state.DomainMappingFocusedField = 2 // Port field
					m.state.CursorPosition = len(m.state.DomainMappingPairs[m.state.DomainMappingFocusedPair].Port)
				}
			}
		}
		return m, nil

	case tea.KeyUp:
		// Move to previous ENV pair
		if pairIdx > 0 {
			m.state.EnvVarFocusedPair--
			// Keep same field (key or value)
			if m.state.EnvVarFocusedField == 0 {
				m.state.CursorPosition = len(m.state.EnvVarPairs[m.state.EnvVarFocusedPair].Key)
			} else {
				m.state.CursorPosition = len(m.state.EnvVarPairs[m.state.EnvVarFocusedPair].Value)
			}
		}
		return m, nil

	case tea.KeyDown:
		// Move to next ENV pair
		if pairIdx < len(m.state.EnvVarPairs)-1 {
			m.state.EnvVarFocusedPair++
			// Keep same field (key or value)
			if m.state.EnvVarFocusedField == 0 {
				m.state.CursorPosition = len(m.state.EnvVarPairs[m.state.EnvVarFocusedPair].Key)
			} else {
				m.state.CursorPosition = len(m.state.EnvVarPairs[m.state.EnvVarFocusedPair].Value)
			}
		}
		return m, nil

	case tea.KeyEnter:
		// Submit form based on current screen
		if m.state.CurrentScreen == state.ScreenSiteEdit {
			return m.handleSiteEditSubmit()
		}
		return m.handleSiteCreateSubmit()

	case tea.KeyEsc:
		// Go back to main form
		m.state.CurrentFieldIndex = len(m.state.FormFields) - 1
		return m, nil

	case tea.KeyLeft:
		if m.state.CursorPosition > 0 {
			m.state.CursorPosition--
		}
		return m, nil

	case tea.KeyRight:
		currentValue := ""
		if m.state.EnvVarFocusedField == 0 {
			currentValue = m.state.EnvVarPairs[pairIdx].Key
		} else {
			currentValue = m.state.EnvVarPairs[pairIdx].Value
		}
		if m.state.CursorPosition < len(currentValue) {
			m.state.CursorPosition++
		}
		return m, nil

	case tea.KeyHome:
		m.state.CursorPosition = 0
		return m, nil

	case tea.KeyEnd:
		if m.state.EnvVarFocusedField == 0 {
			m.state.CursorPosition = len(m.state.EnvVarPairs[pairIdx].Key)
		} else {
			m.state.CursorPosition = len(m.state.EnvVarPairs[pairIdx].Value)
		}
		return m, nil

	case tea.KeyBackspace:
		cursor := m.state.CursorPosition
		if cursor > 0 {
			if m.state.EnvVarFocusedField == 0 {
				// Editing key
				value := m.state.EnvVarPairs[pairIdx].Key
				m.state.EnvVarPairs[pairIdx].Key = value[:cursor-1] + value[cursor:]
			} else {
				// Editing value
				value := m.state.EnvVarPairs[pairIdx].Value
				m.state.EnvVarPairs[pairIdx].Value = value[:cursor-1] + value[cursor:]
			}
			m.state.CursorPosition--
		}
		return m, nil

	case tea.KeyDelete:
		cursor := m.state.CursorPosition
		if m.state.EnvVarFocusedField == 0 {
			// Editing key
			value := m.state.EnvVarPairs[pairIdx].Key
			if cursor < len(value) {
				m.state.EnvVarPairs[pairIdx].Key = value[:cursor] + value[cursor+1:]
			}
		} else {
			// Editing value
			value := m.state.EnvVarPairs[pairIdx].Value
			if cursor < len(value) {
				m.state.EnvVarPairs[pairIdx].Value = value[:cursor] + value[cursor+1:]
			}
		}
		return m, nil

	case tea.KeySpace:
		cursor := m.state.CursorPosition
		if m.state.EnvVarFocusedField == 0 {
			// Editing key
			value := m.state.EnvVarPairs[pairIdx].Key
			m.state.EnvVarPairs[pairIdx].Key = value[:cursor] + " " + value[cursor:]
		} else {
			// Editing value
			value := m.state.EnvVarPairs[pairIdx].Value
			m.state.EnvVarPairs[pairIdx].Value = value[:cursor] + " " + value[cursor:]
		}
		m.state.CursorPosition++
		return m, nil

	case tea.KeyRunes:
		cursor := m.state.CursorPosition
		if m.state.EnvVarFocusedField == 0 {
			// Editing key
			value := m.state.EnvVarPairs[pairIdx].Key
			m.state.EnvVarPairs[pairIdx].Key = value[:cursor] + string(msg.Runes) + value[cursor:]
		} else {
			// Editing value
			value := m.state.EnvVarPairs[pairIdx].Value
			m.state.EnvVarPairs[pairIdx].Value = value[:cursor] + string(msg.Runes) + value[cursor:]
		}
		m.state.CursorPosition++
		return m, nil
	}

	return m, nil
}

// handleDomainsListKeys handles keys on the domains list screen
func (m Model) handleDomainsListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "c":
		m.state.NavigateTo(state.ScreenDomainCreate)
		return m, nil

	case "e":
		// Edit selected domain
		if len(m.state.Domains) > 0 && m.state.DomainsListIndex >= 0 && m.state.DomainsListIndex < len(m.state.Domains) {
			domain := m.state.Domains[m.state.DomainsListIndex]
			m.state.SelectedDomainID = domain.ID
			m.state.NavigateTo(state.ScreenDomainEdit)
		}
		return m, nil

	case "d":
		// Delete selected domain
		if len(m.state.Domains) > 0 && m.state.DomainsListIndex >= 0 && m.state.DomainsListIndex < len(m.state.Domains) {
			domain := m.state.Domains[m.state.DomainsListIndex]
			return m.handleDeleteDomain(domain.ID)
		}
		return m, nil

	case "enter":
		// View DNS records for selected domain
		if len(m.state.Domains) > 0 && m.state.DomainsListIndex >= 0 && m.state.DomainsListIndex < len(m.state.Domains) {
			m.state.NavigateTo(state.ScreenDomainDnsRecords)
		}
		return m, nil
	}

	return m, nil
}

// handleDomainCreateKeys handles keys on the domain creation form
func (m Model) handleDomainCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on provider field (index 1)
	isProviderField := m.state.CurrentFieldIndex == 1

	// Handle dropdown-specific keys when dropdown is open
	if m.state.DropdownOpen && isProviderField {
		providers := []string{"manual", "cloudflare", "route53"}

		switch msg.Type {
		case tea.KeyUp:
			// Navigate up in dropdown
			if m.state.DropdownIndex > 0 {
				m.state.DropdownIndex--
			}
			return m, nil

		case tea.KeyDown:
			// Navigate down in dropdown
			if m.state.DropdownIndex < len(providers)-1 {
				m.state.DropdownIndex++
			}
			return m, nil

		case tea.KeyEnter, tea.KeyTab:
			// Confirm selection and close dropdown
			m.state.FormFields[1] = providers[m.state.DropdownIndex]
			m.state.DropdownOpen = false

			// If Tab, move to next field or wrap around
			if msg.Type == tea.KeyTab {
				m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % len(m.state.FormFields)
			}
			return m, nil

		case tea.KeyEsc:
			// Close dropdown without selecting
			m.state.DropdownOpen = false
			return m, nil

		case tea.KeyBackspace, tea.KeyRunes, tea.KeySpace:
			// Close dropdown and allow manual input
			m.state.DropdownOpen = false
			// Fall through to normal input handling
		}
	}

	// Normal field input handling
	switch msg.Type {
	case tea.KeyUp:
		// Open dropdown on up arrow if on provider field and not open
		if isProviderField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			// Find current provider in list
			providers := []string{"manual", "cloudflare", "route53"}
			for i, p := range providers {
				if p == m.state.FormFields[1] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		}

	case tea.KeyDown:
		// Open dropdown on down arrow if on provider field and not open
		if isProviderField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			// Find current provider in list
			providers := []string{"manual", "cloudflare", "route53"}
			for i, p := range providers {
				if p == m.state.FormFields[1] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		}

	case tea.KeySpace:
		// Add space to current field (except provider dropdown)
		if m.state.CurrentFieldIndex != 1 && m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to current field (except provider dropdown)
		if m.state.CurrentFieldIndex != 1 && m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from current field (except provider dropdown)
		if m.state.CurrentFieldIndex != 1 && m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			if len(m.state.FormFields[m.state.CurrentFieldIndex]) > 0 {
				m.state.FormFields[m.state.CurrentFieldIndex] = m.state.FormFields[m.state.CurrentFieldIndex][:len(m.state.FormFields[m.state.CurrentFieldIndex])-1]
			}
		}
		return m, nil

	case tea.KeyTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}
		// Move to next field
		m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % len(m.state.FormFields)
		return m, nil

	case tea.KeyShiftTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}
		// Move to previous field
		m.state.CurrentFieldIndex--
		if m.state.CurrentFieldIndex < 0 {
			m.state.CurrentFieldIndex = len(m.state.FormFields) - 1
		}
		return m, nil

	case tea.KeyEnter:
		// If on provider field and not open, open it
		if isProviderField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			// Find current provider in list
			providers := []string{"manual", "cloudflare", "route53"}
			for i, p := range providers {
				if p == m.state.FormFields[1] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		}
		// Otherwise submit form
		return m.handleDomainCreateSubmit()
	}

	return m, nil
}

// handleDomainEditKeys handles keys on the domain edit form
func (m Model) handleDomainEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on provider field (index 1)
	isProviderField := m.state.CurrentFieldIndex == 1

	// Handle dropdown-specific keys when dropdown is open
	if m.state.DropdownOpen && isProviderField {
		providers := []string{"manual", "cloudflare", "route53"}

		switch msg.Type {
		case tea.KeyUp:
			// Navigate up in dropdown
			if m.state.DropdownIndex > 0 {
				m.state.DropdownIndex--
			}
			return m, nil

		case tea.KeyDown:
			// Navigate down in dropdown
			if m.state.DropdownIndex < len(providers)-1 {
				m.state.DropdownIndex++
			}
			return m, nil

		case tea.KeyEnter, tea.KeyTab:
			// Confirm selection and close dropdown
			m.state.FormFields[1] = providers[m.state.DropdownIndex]
			m.state.DropdownOpen = false

			// If Tab, move to next field or wrap around
			if msg.Type == tea.KeyTab {
				m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % len(m.state.FormFields)
			}
			return m, nil

		case tea.KeyEsc:
			// Close dropdown without selecting
			m.state.DropdownOpen = false
			return m, nil

		case tea.KeyBackspace, tea.KeyRunes, tea.KeySpace:
			// Close dropdown and allow manual input
			m.state.DropdownOpen = false
			// Fall through to normal input handling
		}
	}

	// Normal field input handling
	switch msg.Type {
	case tea.KeyUp:
		// Open dropdown on up arrow if on provider field and not open
		if isProviderField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			// Find current provider in list
			providers := []string{"manual", "cloudflare", "route53"}
			for i, p := range providers {
				if p == m.state.FormFields[1] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		}

	case tea.KeyDown:
		// Open dropdown on down arrow if on provider field and not open
		if isProviderField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			// Find current provider in list
			providers := []string{"manual", "cloudflare", "route53"}
			for i, p := range providers {
				if p == m.state.FormFields[1] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		}

	case tea.KeySpace:
		// Add space to current field (except provider dropdown)
		if m.state.CurrentFieldIndex != 1 && m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to current field (except provider dropdown)
		if m.state.CurrentFieldIndex != 1 && m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from current field (except provider dropdown)
		if m.state.CurrentFieldIndex != 1 && m.state.CurrentFieldIndex >= 0 && m.state.CurrentFieldIndex < len(m.state.FormFields) {
			if len(m.state.FormFields[m.state.CurrentFieldIndex]) > 0 {
				m.state.FormFields[m.state.CurrentFieldIndex] = m.state.FormFields[m.state.CurrentFieldIndex][:len(m.state.FormFields[m.state.CurrentFieldIndex])-1]
			}
		}
		return m, nil

	case tea.KeyTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}
		// Move to next field
		m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % len(m.state.FormFields)
		return m, nil

	case tea.KeyShiftTab:
		// Close dropdown if open
		if m.state.DropdownOpen {
			m.state.DropdownOpen = false
		}
		// Move to previous field
		m.state.CurrentFieldIndex--
		if m.state.CurrentFieldIndex < 0 {
			m.state.CurrentFieldIndex = len(m.state.FormFields) - 1
		}
		return m, nil

	case tea.KeyEnter:
		// If on provider field and not open, open it
		if isProviderField && !m.state.DropdownOpen {
			m.state.DropdownOpen = true
			// Find current provider in list
			providers := []string{"manual", "cloudflare", "route53"}
			for i, p := range providers {
				if p == m.state.FormFields[1] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		}
		// Otherwise submit form
		return m.handleDomainEditSubmit()
	}

	return m, nil
}

// handleNodesListKeys handles keys on the nodes list screen
func (m Model) handleNodesListKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n", "c":
		m.state.NavigateTo(state.ScreenNodeCreate)
		return m, nil

	case "v", "enter":
		// View config for selected node
		if len(m.state.Nodes) > 0 && m.state.NodesListIndex >= 0 && m.state.NodesListIndex < len(m.state.Nodes) {
			m.state.SelectedNodeID = m.state.Nodes[m.state.NodesListIndex].ID
			m.state.NavigateTo(state.ScreenNodeConfig)
		}
		return m, nil

	case "e":
		// Edit selected node
		if len(m.state.Nodes) > 0 && m.state.NodesListIndex >= 0 && m.state.NodesListIndex < len(m.state.Nodes) {
			node := m.state.Nodes[m.state.NodesListIndex]
			m.state.AddNotification("Edit node: "+node.Name+" (not yet implemented)", "info")
		}
		return m, nil

	case "d":
		// Delete selected node
		if len(m.state.Nodes) > 0 && m.state.NodesListIndex >= 0 && m.state.NodesListIndex < len(m.state.Nodes) {
			node := m.state.Nodes[m.state.NodesListIndex]
			return m.handleDeleteNode(node.ID)
		}
		return m, nil
	}

	return m, nil
}

// handleNodeCreateKeys handles keys on the node creation form
func (m Model) handleNodeCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on proxy field (index 2)
	isProxyField := m.state.CurrentFieldIndex == 2

	// Handle dropdown-specific keys when dropdown is open
	if m.state.DropdownOpen && isProxyField {
		proxies := []string{"nginx", "apache", "traefik"}

		switch msg.Type {
		case tea.KeyUp:
			if m.state.DropdownIndex > 0 {
				m.state.DropdownIndex--
			}
			return m, nil
		case tea.KeyDown:
			if m.state.DropdownIndex < len(proxies)-1 {
				m.state.DropdownIndex++
			}
			return m, nil
		case tea.KeyEnter, tea.KeyTab:
			// Confirm selection
			m.state.FormFields[2] = proxies[m.state.DropdownIndex]
			m.state.DropdownOpen = false
			if msg.Type == tea.KeyTab {
				// Move to next field (cycle through 0, 1, 2)
				m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % 3
			}
			return m, nil
		case tea.KeyEsc:
			// Close dropdown without changing selection
			m.state.DropdownOpen = false
			return m, nil
		}
		return m, nil
	}

	// If on proxy field but dropdown not open
	if isProxyField && !m.state.DropdownOpen {
		switch msg.Type {
		case tea.KeyEnter, tea.KeyDown:
			// Open dropdown
			m.state.DropdownOpen = true
			// Set dropdown index based on current selection
			proxies := []string{"nginx", "apache", "traefik"}
			for i, p := range proxies {
				if p == m.state.FormFields[2] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		case tea.KeyTab:
			// Move to next field without opening dropdown
			m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % 3
			return m, nil
		case tea.KeyShiftTab:
			// Move to previous field
			m.state.CurrentFieldIndex--
			if m.state.CurrentFieldIndex < 0 {
				m.state.CurrentFieldIndex = 2
			}
			return m, nil
		}
	}

	// Handle regular text input for Name and API Endpoint fields (0, 1)
	switch msg.Type {
	case tea.KeySpace:
		// Add space to current field (only editable text fields - first 2)
		if m.state.CurrentFieldIndex < 2 {
			m.state.FormFields[m.state.CurrentFieldIndex] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to current field (only editable text fields - first 2)
		if m.state.CurrentFieldIndex < 2 {
			m.state.FormFields[m.state.CurrentFieldIndex] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from current field (only editable text fields - first 2)
		if m.state.CurrentFieldIndex < 2 {
			value := m.state.FormFields[m.state.CurrentFieldIndex]
			if len(value) > 0 {
				m.state.FormFields[m.state.CurrentFieldIndex] = value[:len(value)-1]
			}
		}
		return m, nil

	case tea.KeyTab:
		// Move to next field (cycle through editable fields: 0, 1, 2)
		m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % 3
		return m, nil

	case tea.KeyShiftTab:
		// Move to previous field
		m.state.CurrentFieldIndex--
		if m.state.CurrentFieldIndex < 0 {
			m.state.CurrentFieldIndex = 2
		}
		return m, nil

	case tea.KeyEnter:
		// Submit form (only if not on proxy field or dropdown is not open)
		if !isProxyField {
			return m.handleNodeCreateSubmit()
		}
	}

	return m, nil
}

// handleNodeEditKeys handles keys on the node edit form
func (m Model) handleNodeEditKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on proxy field (index 2)
	isProxyField := m.state.CurrentFieldIndex == 2

	// Handle dropdown-specific keys when dropdown is open
	if m.state.DropdownOpen && isProxyField {
		proxies := []string{"nginx", "apache", "traefik"}

		switch msg.Type {
		case tea.KeyUp:
			if m.state.DropdownIndex > 0 {
				m.state.DropdownIndex--
			}
			return m, nil
		case tea.KeyDown:
			if m.state.DropdownIndex < len(proxies)-1 {
				m.state.DropdownIndex++
			}
			return m, nil
		case tea.KeyEnter, tea.KeyTab:
			// Confirm selection
			m.state.FormFields[2] = proxies[m.state.DropdownIndex]
			m.state.DropdownOpen = false
			if msg.Type == tea.KeyTab {
				// Move to next field (cycle through 0, 1, 2)
				m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % 3
			}
			return m, nil
		case tea.KeyEsc:
			// Close dropdown without changing selection
			m.state.DropdownOpen = false
			return m, nil
		}
		return m, nil
	}

	// If on proxy field but dropdown not open
	if isProxyField && !m.state.DropdownOpen {
		switch msg.Type {
		case tea.KeyEnter, tea.KeyDown:
			// Open dropdown
			m.state.DropdownOpen = true
			// Set dropdown index based on current selection
			proxies := []string{"nginx", "apache", "traefik"}
			for i, p := range proxies {
				if p == m.state.FormFields[2] {
					m.state.DropdownIndex = i
					break
				}
			}
			return m, nil
		case tea.KeyTab:
			// Move to next field without opening dropdown
			m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % 3
			return m, nil
		case tea.KeyShiftTab:
			// Move to previous field
			m.state.CurrentFieldIndex--
			if m.state.CurrentFieldIndex < 0 {
				m.state.CurrentFieldIndex = 2
			}
			return m, nil
		}
	}

	// Handle regular text input for Name and API Endpoint fields (0, 1)
	switch msg.Type {
	case tea.KeySpace:
		// Add space to current field (only editable text fields - first 2)
		if m.state.CurrentFieldIndex < 2 {
			m.state.FormFields[m.state.CurrentFieldIndex] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to current field (only editable text fields - first 2)
		if m.state.CurrentFieldIndex < 2 {
			m.state.FormFields[m.state.CurrentFieldIndex] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from current field (only editable text fields - first 2)
		if m.state.CurrentFieldIndex < 2 {
			value := m.state.FormFields[m.state.CurrentFieldIndex]
			if len(value) > 0 {
				m.state.FormFields[m.state.CurrentFieldIndex] = value[:len(value)-1]
			}
		}
		return m, nil

	case tea.KeyTab:
		// Move to next field (cycle through editable fields: 0, 1, 2)
		m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % 3
		return m, nil

	case tea.KeyShiftTab:
		// Move to previous field
		m.state.CurrentFieldIndex--
		if m.state.CurrentFieldIndex < 0 {
			m.state.CurrentFieldIndex = 2
		}
		return m, nil

	case tea.KeyEnter:
		// Submit form (only if not on proxy field or dropdown is not open)
		if !isProxyField {
			return m.handleNodeEditSubmit()
		}
	}

	return m, nil
}

// handleNodeConfigKeys handles keys on the node config screen (scrollable viewport)
func (m Model) handleNodeConfigKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "down", "j":
		// Scroll down one line
		m.state.NodeConfigViewport.LineDown(1)
		return m, nil

	case "up", "k":
		// Scroll up one line
		m.state.NodeConfigViewport.LineUp(1)
		return m, nil

	case "pgdown", "f":
		// Scroll down one page
		m.state.NodeConfigViewport, cmd = m.state.NodeConfigViewport.Update(msg)
		return m, cmd

	case "pgup", "b":
		// Scroll up one page
		m.state.NodeConfigViewport, cmd = m.state.NodeConfigViewport.Update(msg)
		return m, cmd

	case "home", "g":
		// Jump to top
		m.state.NodeConfigViewport.GotoTop()
		return m, nil

	case "end", "G":
		// Jump to bottom
		m.state.NodeConfigViewport.GotoBottom()
		return m, nil

	case "s":
		// Save config to file
		return m.handleSaveNodeConfig()
	}

	return m, nil
}

// handleNodeConfigSaveKeys handles keys on the save dialog screen
func (m Model) handleNodeConfigSaveKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Try text input with cursor support first
	if m.handleTextInput(msg, 0) {
		return m, nil
	}

	// Handle non-text-input keys
	switch msg.Type {
	case tea.KeyEnter:
		// Save to the specified path
		return m.handleSaveNodeConfigSubmit()
	}

	return m, nil
}

// handleSaveNodeConfig navigates to the save dialog
func (m Model) handleSaveNodeConfig() (tea.Model, tea.Cmd) {
	// Navigate to save dialog screen
	m.state.NavigateTo(state.ScreenNodeConfigSave)
	return m, nil
}

// handleSaveNodeConfigSubmit saves the node config to the user-specified path
func (m Model) handleSaveNodeConfigSubmit() (tea.Model, tea.Cmd) {
	// Find the selected node
	var node *models.Node
	for i := range m.state.Nodes {
		if m.state.Nodes[i].ID == m.state.SelectedNodeID {
			node = &m.state.Nodes[i]
			break
		}
	}

	if node == nil {
		m.state.AddNotification("Node not found", "error")
		m.state.NavigateBack()
		return m, nil
	}

	// Get the file path from form field
	savePath := m.state.FormFields[0]
	if savePath == "" {
		m.state.AddNotification("File path is required", "error")
		return m, nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(savePath, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			m.state.AddNotification("Failed to get home directory: "+err.Error(), "error")
			return m, nil
		}
		savePath = strings.Replace(savePath, "~", homeDir, 1)
	}

	// Generate config content
	configContent := node.GenerateNodeConfigTOML()

	// Create parent directories if they don't exist
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		m.state.AddNotification("Failed to create directory: "+err.Error(), "error")
		return m, nil
	}

	// Write to file
	err := os.WriteFile(savePath, []byte(configContent), 0644)
	if err != nil {
		m.state.AddNotification("Failed to save config: "+err.Error(), "error")
		return m, nil
	}

	m.state.AddNotification(fmt.Sprintf("Config saved to %s", savePath), "success")

	// Navigate back to config screen
	m.state.NavigateBack()

	return m, nil
}

// handleSettingsKeys handles keys on the settings form
func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Try text input with cursor support first
	if m.handleTextInput(msg, m.state.CurrentFieldIndex) {
		return m, nil
	}

	// Handle non-text-input keys
	switch msg.Type {
	case tea.KeyTab:
		// Move to next field
		nextField := (m.state.CurrentFieldIndex + 1) % len(m.state.FormFields)
		m.setFieldAndResetCursor(nextField)
		return m, nil

	case tea.KeyShiftTab:
		// Move to previous field
		prevField := m.state.CurrentFieldIndex - 1
		if prevField < 0 {
			prevField = len(m.state.FormFields) - 1
		}
		m.setFieldAndResetCursor(prevField)
		return m, nil

	case tea.KeyEnter:
		// Submit form
		return m.handleSettingsSave()
	}

	return m, nil
}

// handleHelpKeys handles keys on the help screen
func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Help screen is mostly static, just allow navigation back
	return m, nil
}

// ============================================================================
// Mouse Event Handlers
// ============================================================================

// handleMouseClick routes mouse click events to screen-specific handlers
func (m Model) handleMouseClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Only handle mouse button release (not press/drag)
	if msg.Action != tea.MouseActionRelease {
		return m, nil
	}

	// TODO: Implement mouse handling per screen
	// Each screen will need to calculate what was clicked based on msg.X and msg.Y
	// Example:
	// - Table rows: Calculate which row based on Y coordinate
	// - Buttons: Check if X,Y falls within button bounds
	// - Form fields: Check if Y coordinate matches a field's row

	switch m.state.CurrentScreen {
	case state.ScreenSitesList:
		return m.handleSitesListClick(msg)
	case state.ScreenDomainsList:
		return m.handleDomainsListClick(msg)
	case state.ScreenNodesList:
		return m.handleNodesListClick(msg)
	case state.ScreenSiteCreate:
		return m.handleSiteCreateClick(msg)
	case state.ScreenDomainCreate:
		return m.handleDomainCreateClick(msg)
	case state.ScreenNodeCreate:
		return m.handleNodeCreateClick(msg)
	}

	return m, nil
}

// handleSitesListClick handles mouse clicks on the sites list
func (m Model) handleSitesListClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement click-to-select for sites table
	// Calculate which table row was clicked based on msg.Y
	// Account for header, menu, and table header rows

	// Example structure:
	// const (
	//     headerHeight = 1
	//     menuHeight = 1
	//     tableHeaderHeight = 2
	// )
	// tableStartY := headerHeight + menuHeight + tableHeaderHeight
	// if msg.Y >= tableStartY {
	//     rowIndex := msg.Y - tableStartY
	//     if rowIndex >= 0 && rowIndex < len(m.state.Sites) {
	//         m.state.SitesListIndex = rowIndex
	//     }
	// }

	return m, nil
}

// handleDomainsListClick handles mouse clicks on the domains list
func (m Model) handleDomainsListClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement click-to-select for domains table
	return m, nil
}

// handleNodesListClick handles mouse clicks on the nodes list
func (m Model) handleNodesListClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement click-to-select for nodes table
	return m, nil
}

// handleSiteCreateClick handles mouse clicks on the site creation form
func (m Model) handleSiteCreateClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement click-to-focus for form fields
	// Calculate which form field was clicked based on msg.Y
	// Each field is rendered at a specific Y coordinate

	return m, nil
}

// handleDomainCreateClick handles mouse clicks on the domain creation form
func (m Model) handleDomainCreateClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement click-to-focus for domain form field
	return m, nil
}

// handleNodeCreateClick handles mouse clicks on the node creation form
func (m Model) handleNodeCreateClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// TODO: Implement click-to-focus for node form fields
	return m, nil
}

// ============================================================================
// Form Submission Handlers
// ============================================================================

// handleFormSubmit routes form submissions to screen-specific handlers
func (m Model) handleFormSubmit() (tea.Model, tea.Cmd) {
	switch m.state.CurrentScreen {
	case state.ScreenSiteCreate:
		return m.handleSiteCreateSubmit()
	case state.ScreenDomainCreate:
		return m.handleDomainCreateSubmit()
	case state.ScreenNodeCreate:
		return m.handleNodeCreateSubmit()
	}
	return m, nil
}

// handleSiteCreateSubmit processes site creation form submission
func (m Model) handleSiteCreateSubmit() (tea.Model, tea.Cmd) {
	isCompose := m.state.SiteTypeSelection == "compose"

	// Validate required fields based on site type
	// Name(0), Node(1) always required
	// For container: Docker Image(2) required
	// For compose: Compose File Path(2) required
	if m.state.FormFields[0] == "" {
		m.state.AddNotification("Required field Name must be filled", "error")
		return m, nil
	}
	if m.state.FormFields[1] == "" {
		m.state.AddNotification("Required field Node must be filled", "error")
		return m, nil
	}
	if m.state.FormFields[2] == "" {
		if isCompose {
			m.state.AddNotification("Required field Compose File Path must be filled", "error")
		} else {
			m.state.AddNotification("Required field Docker Image must be filled", "error")
		}
		return m, nil
	}

	// For compose: load and validate compose file
	var composeContent string
	var detectedPorts []compose.DetectedPort
	if isCompose {
		content, ports, err := loadComposeFile(m.state.FormFields[2])
		if err != nil {
			m.state.AddNotification(err.Error(), "error")
			return m, nil
		}
		composeContent = content
		detectedPorts = ports
	}

	// Auto-populate domain mapping port from detected compose ports
	if isCompose && len(detectedPorts) > 0 {
		firstPort := compose.GetFirstPort(detectedPorts)
		if firstPort > 0 && len(m.state.DomainMappingPairs) > 0 {
			// Only update if the port is still the default
			if m.state.DomainMappingPairs[0].Port == "8080" || m.state.DomainMappingPairs[0].Port == "" {
				m.state.DomainMappingPairs[0].Port = fmt.Sprintf("%d", firstPort)
			}
		}
		// Notify user about detected ports
		if len(detectedPorts) == 1 {
			m.state.AddNotification(fmt.Sprintf("Auto-detected port %d from compose file", firstPort), "info")
		} else {
			m.state.AddNotification(fmt.Sprintf("Auto-detected %d ports from compose file, using port %d", len(detectedPorts), firstPort), "info")
		}
	}

	// Find node by name (index 1)
	var nodeID uuid.UUID
	nodeFound := false
	for _, node := range m.state.Nodes {
		if node.Name == m.state.FormFields[1] {
			nodeID = node.ID
			nodeFound = true
			break
		}
	}
	if !nodeFound {
		m.state.AddNotification("Node not found: "+m.state.FormFields[1], "error")
		return m, nil
	}

	// Parse and validate domain mappings
	var domainMappings []models.DomainMapping
	var firstDomainID uuid.UUID
	var firstPort int

	for i, pair := range m.state.DomainMappingPairs {
		// Skip empty mappings
		if pair.DomainName == "" && pair.Port == "" {
			continue
		}

		// Validate domain is selected
		if pair.DomainName == "" || pair.DomainID == "" {
			m.state.AddNotification(fmt.Sprintf("Domain mapping %d: domain must be selected", i+1), "error")
			return m, nil
		}

		// Parse port mapping (supports "3000" or "3000:3001" notation)
		containerPort, hostPort, err := models.ParsePortMapping(pair.Port)
		if err != nil {
			m.state.AddNotification(fmt.Sprintf("Domain mapping %d: %v", i+1, err), "error")
			return m, nil
		}

		// Parse domain ID
		domainID, err := uuid.Parse(pair.DomainID)
		if err != nil {
			m.state.AddNotification(fmt.Sprintf("Domain mapping %d: invalid domain ID", i+1), "error")
			return m, nil
		}

		// Store first domain and port for backwards compatibility
		if len(domainMappings) == 0 {
			firstDomainID = domainID
			firstPort = containerPort
		}

		domainMappings = append(domainMappings, models.DomainMapping{
			DomainID:  domainID,
			Subdomain: strings.TrimSpace(pair.Subdomain),
			Port:      containerPort,
			HostPort:  hostPort,
		})
	}

	// Validate at least one domain mapping
	if len(domainMappings) == 0 {
		m.state.AddNotification("At least one domain mapping is required", "error")
		return m, nil
	}

	// Create new site
	var site *models.Site
	if isCompose {
		// For compose: create site with empty docker image (compose handles containers)
		site = models.NewSite(m.state.FormFields[0], firstDomainID, nodeID, "", firstPort)
		site.SiteType = models.SiteTypeCompose
		site.ComposeContent = composeContent
	} else {
		// For container: use docker image from field 2
		site = models.NewSite(m.state.FormFields[0], firstDomainID, nodeID, m.state.FormFields[2], firstPort)
		site.SiteType = models.SiteTypeContainer
	}

	// Replace default domain mapping with all mappings from the form
	site.DomainMappings = domainMappings

	// Set SSL email (field 5) if provided
	if m.state.FormFields[5] != "" {
		site.SSLEmail = strings.TrimSpace(m.state.FormFields[5])
	}

	// For container deployments: parse environment variables and config files
	if !isCompose {
		// Set Docker credentials (fields 3, 4)
		if m.state.FormFields[3] != "" {
			site.DockerUsername = strings.TrimSpace(m.state.FormFields[3])
		}
		if m.state.FormFields[4] != "" {
			site.DockerToken = strings.TrimSpace(m.state.FormFields[4])
		}

		// Parse environment variables from EnvVarPairs
		for _, pair := range m.state.EnvVarPairs {
			key := strings.TrimSpace(pair.Key)
			value := strings.TrimSpace(pair.Value)
			if key != "" {
				site.EnvironmentVars[key] = value
			}
		}

		// Load config file (field 6) if provided
		if m.state.FormFields[6] != "" {
			configPath := strings.TrimSpace(m.state.FormFields[6])
			content, err := os.ReadFile(configPath)
			if err != nil {
				m.state.AddNotification("Failed to read config file: "+err.Error(), "warning")
			} else {
				// Extract filename from path
				filename := filepath.Base(configPath)
				site.ConfigFiles = append(site.ConfigFiles, models.ConfigFile{
					Name:          filename,
					Content:       string(content),
					ContainerPath: "/config/" + filename, // Default container path
				})
			}
		}
	}

	m.state.Sites = append(m.state.Sites, *site)

	siteTypeLabel := "Container"
	if isCompose {
		siteTypeLabel = "Compose"
	}
	m.state.AddNotification(fmt.Sprintf("%s site created: %s", siteTypeLabel, site.Name), "success")

	// Auto-save config if enabled
	if m.state.AutoSave {
		go func() {
			_ = m.saveConfigSync()
		}()
	}

	m.state.NavigateBack()

	return m, nil
}

// handleSiteEditSubmit processes site edit form submission
func (m Model) handleSiteEditSubmit() (tea.Model, tea.Cmd) {
	// Find the site being edited
	var siteIndex = -1
	for i := range m.state.Sites {
		if m.state.Sites[i].ID == m.state.SelectedSiteID {
			siteIndex = i
			break
		}
	}

	if siteIndex == -1 {
		m.state.AddNotification("Site not found", "error")
		m.state.NavigateBack()
		return m, nil
	}

	isCompose := m.state.Sites[siteIndex].GetSiteType() == models.SiteTypeCompose

	// Validate required fields based on site type
	// Name(0), Node(1) always required
	// For container: Docker Image(2) required
	// For compose: field 2 is read-only placeholder, skip validation
	if m.state.FormFields[0] == "" {
		m.state.AddNotification("Required field Name must be filled", "error")
		return m, nil
	}
	if m.state.FormFields[1] == "" {
		m.state.AddNotification("Required field Node must be filled", "error")
		return m, nil
	}
	if !isCompose && m.state.FormFields[2] == "" {
		m.state.AddNotification("Required field Docker Image must be filled", "error")
		return m, nil
	}

	// Find node by name (index 1)
	var nodeID uuid.UUID
	nodeFound := false
	for _, node := range m.state.Nodes {
		if node.Name == m.state.FormFields[1] {
			nodeID = node.ID
			nodeFound = true
			break
		}
	}
	if !nodeFound {
		m.state.AddNotification("Node not found: "+m.state.FormFields[1], "error")
		return m, nil
	}

	// Parse and validate domain mappings
	var domainMappings []models.DomainMapping
	var firstDomainID uuid.UUID
	var firstPort int

	for i, pair := range m.state.DomainMappingPairs {
		// Skip empty mappings
		if pair.DomainName == "" && pair.Port == "" {
			continue
		}

		// Validate domain is selected
		if pair.DomainName == "" || pair.DomainID == "" {
			m.state.AddNotification(fmt.Sprintf("Domain mapping %d: domain must be selected", i+1), "error")
			return m, nil
		}

		// Parse port mapping (supports "3000" or "3000:3001" notation)
		containerPort, hostPort, err := models.ParsePortMapping(pair.Port)
		if err != nil {
			m.state.AddNotification(fmt.Sprintf("Domain mapping %d: %v", i+1, err), "error")
			return m, nil
		}

		// Parse domain ID
		domainID, err := uuid.Parse(pair.DomainID)
		if err != nil {
			m.state.AddNotification(fmt.Sprintf("Domain mapping %d: invalid domain ID", i+1), "error")
			return m, nil
		}

		// Store first domain and port for backwards compatibility
		if len(domainMappings) == 0 {
			firstDomainID = domainID
			firstPort = containerPort
		}

		domainMappings = append(domainMappings, models.DomainMapping{
			DomainID:  domainID,
			Subdomain: strings.TrimSpace(pair.Subdomain),
			Port:      containerPort,
			HostPort:  hostPort,
		})
	}

	// Validate at least one domain mapping
	if len(domainMappings) == 0 {
		m.state.AddNotification("At least one domain mapping is required", "error")
		return m, nil
	}

	// Update common site fields
	oldName := m.state.Sites[siteIndex].Name
	m.state.Sites[siteIndex].Name = m.state.FormFields[0]
	m.state.Sites[siteIndex].DomainID = firstDomainID
	m.state.Sites[siteIndex].NodeID = nodeID
	m.state.Sites[siteIndex].Port = firstPort
	m.state.Sites[siteIndex].SSLEmail = strings.TrimSpace(m.state.FormFields[5]) // SSL Email at index 5

	// Update domain mappings with all mappings from form
	m.state.Sites[siteIndex].DomainMappings = domainMappings

	// For container deployments: update docker-specific fields
	if !isCompose {
		m.state.Sites[siteIndex].DockerImage = m.state.FormFields[2]                     // Docker Image at index 2
		m.state.Sites[siteIndex].DockerUsername = strings.TrimSpace(m.state.FormFields[3]) // Docker Username at index 3
		m.state.Sites[siteIndex].DockerToken = strings.TrimSpace(m.state.FormFields[4])    // Docker Token at index 4

		// Update environment variables from EnvVarPairs
		m.state.Sites[siteIndex].EnvironmentVars = make(map[string]string)
		for _, pair := range m.state.EnvVarPairs {
			key := strings.TrimSpace(pair.Key)
			value := strings.TrimSpace(pair.Value)
			if key != "" {
				m.state.Sites[siteIndex].EnvironmentVars[key] = value
			}
		}

		// Load config file (field 6) if provided
		if m.state.FormFields[6] != "" {
			configPath := strings.TrimSpace(m.state.FormFields[6])
			content, err := os.ReadFile(configPath)
			if err != nil {
				m.state.AddNotification("Failed to read config file: "+err.Error(), "warning")
			} else {
				// Extract filename from path
				filename := filepath.Base(configPath)
				// Replace existing config files
				m.state.Sites[siteIndex].ConfigFiles = []models.ConfigFile{
					{
						Name:          filename,
						Content:       string(content),
						ContainerPath: "/config/" + filename,
					},
				}
			}
		}
	}
	// Note: For compose sites, SiteType and ComposeContent are preserved (read-only in edit)

	// Update timestamp
	m.state.Sites[siteIndex].UpdatedAt = time.Now()

	// Build notification message
	siteTypeLabel := "Container"
	if isCompose {
		siteTypeLabel = "Compose"
	}
	var changes []string
	if oldName != m.state.Sites[siteIndex].Name {
		changes = append(changes, fmt.Sprintf("name: %s → %s", oldName, m.state.Sites[siteIndex].Name))
	}
	changes = append(changes, "updated site configuration")

	message := fmt.Sprintf("%s site updated: %s", siteTypeLabel, strings.Join(changes, ", "))
	m.state.AddNotification(message, "success")

	// Auto-save config if enabled
	if m.state.AutoSave {
		go func() {
			_ = m.saveConfigSync()
		}()
	}

	m.state.NavigateBack()

	return m, nil
}

// handleDomainCreateSubmit processes domain creation form submission
func (m Model) handleDomainCreateSubmit() (tea.Model, tea.Cmd) {
	// Fields: 0=domain name, 1=provider, 2=zone/hosted zone ID, 3=API token/access key, 4=secret key (route53 only)
	domainName := m.state.FormFields[0]
	providerType := m.state.FormFields[1]

	// Validate domain name is not empty
	if domainName == "" {
		m.state.AddNotification("Domain name is required", "error")
		return m, nil
	}

	// Check for duplicates
	for _, domain := range m.state.Domains {
		if domain.Name == domainName {
			m.state.AddNotification("Domain already exists: "+domainName, "error")
			return m, nil
		}
	}

	// Create DNS provider based on type with configuration
	var provider models.DnsProvider
	switch providerType {
	case "cloudflare":
		zoneID := m.state.FormFields[2]

		// Validate required fields for Cloudflare (Zone ID only, API token is global)
		if zoneID == "" {
			m.state.AddNotification("Cloudflare Zone ID is required", "error")
			return m, nil
		}

		provider = models.DnsProvider{
			Type:   models.DnsProviderCloudflare,
			ZoneID: zoneID,
			// APIToken is stored globally in settings, not per-domain
		}
	case "route53":
		hostedZoneID := m.state.FormFields[2]
		accessKey := m.state.FormFields[3]
		secretKey := m.state.FormFields[4]

		// Validate required fields for Route53
		if hostedZoneID == "" || accessKey == "" || secretKey == "" {
			m.state.AddNotification("Route53 Hosted Zone ID, Access Key, and Secret Key are required", "error")
			return m, nil
		}

		provider = models.DnsProvider{
			Type:         models.DnsProviderRoute53,
			HostedZoneID: hostedZoneID,
			AccessKey:    accessKey,
			SecretKey:    secretKey,
		}
	default:
		provider = models.DnsProvider{Type: models.DnsProviderManual}
	}

	domain := models.NewDomain(domainName, provider)
	m.state.Domains = append(m.state.Domains, *domain)

	m.state.AddNotification("Domain created: "+domainName+" ("+domain.ProviderName()+")", "success")

	// Auto-save config if enabled
	if m.state.AutoSave {
		go func() {
			_ = m.saveConfigSync()
		}()
	}

	m.state.NavigateBack()

	return m, nil
}

// handleDomainEditSubmit processes domain edit form submission
func (m Model) handleDomainEditSubmit() (tea.Model, tea.Cmd) {
	// Fields: 0=domain name, 1=provider, 2=zone/hosted zone ID, 3=API token/access key, 4=secret key (route53 only)
	newDomainName := m.state.FormFields[0]
	providerType := m.state.FormFields[1]

	// Validate domain name is not empty
	if newDomainName == "" {
		m.state.AddNotification("Domain name is required", "error")
		return m, nil
	}

	// Find the domain being edited
	var domainIndex = -1
	for i := range m.state.Domains {
		if m.state.Domains[i].ID == m.state.SelectedDomainID {
			domainIndex = i
			break
		}
	}

	if domainIndex == -1 {
		m.state.AddNotification("Domain not found", "error")
		m.state.NavigateBack()
		return m, nil
	}

	oldName := m.state.Domains[domainIndex].Name
	oldProviderType := string(m.state.Domains[domainIndex].DnsProvider.Type)

	// Check for duplicates (excluding current domain)
	for i, domain := range m.state.Domains {
		if i != domainIndex && domain.Name == newDomainName {
			m.state.AddNotification("Domain already exists: "+newDomainName, "error")
			return m, nil
		}
	}

	// Update domain name
	m.state.Domains[domainIndex].Name = newDomainName

	// Update DNS provider based on type with configuration
	var provider models.DnsProvider
	switch providerType {
	case "cloudflare":
		zoneID := m.state.FormFields[2]

		// Validate required fields for Cloudflare (Zone ID only, API token is global)
		if zoneID == "" {
			m.state.AddNotification("Cloudflare Zone ID is required", "error")
			return m, nil
		}

		provider = models.DnsProvider{
			Type:   models.DnsProviderCloudflare,
			ZoneID: zoneID,
			// APIToken is stored globally in settings, not per-domain
		}
	case "route53":
		hostedZoneID := m.state.FormFields[2]
		accessKey := m.state.FormFields[3]
		secretKey := m.state.FormFields[4]

		// Validate required fields for Route53
		if hostedZoneID == "" || accessKey == "" || secretKey == "" {
			m.state.AddNotification("Route53 Hosted Zone ID, Access Key, and Secret Key are required", "error")
			return m, nil
		}

		provider = models.DnsProvider{
			Type:         models.DnsProviderRoute53,
			HostedZoneID: hostedZoneID,
			AccessKey:    accessKey,
			SecretKey:    secretKey,
		}
	default:
		provider = models.DnsProvider{Type: models.DnsProviderManual}
	}
	m.state.Domains[domainIndex].DnsProvider = provider

	// Build notification message
	var changes []string
	if oldName != newDomainName {
		changes = append(changes, fmt.Sprintf("name: %s → %s", oldName, newDomainName))
	}
	if oldProviderType != providerType {
		providerLabels := map[string]string{
			"manual":     "Manual DNS",
			"cloudflare": "Cloudflare",
			"route53":    "AWS Route53",
		}
		oldLabel := providerLabels[oldProviderType]
		if oldLabel == "" {
			oldLabel = oldProviderType
		}
		newLabel := providerLabels[providerType]
		if newLabel == "" {
			newLabel = providerType
		}
		changes = append(changes, fmt.Sprintf("provider: %s → %s", oldLabel, newLabel))
	}

	var message string
	if len(changes) > 0 {
		message = "Domain updated: " + strings.Join(changes, ", ")
	} else {
		message = "Domain updated (no changes)"
	}
	m.state.AddNotification(message, "success")

	// Auto-save config if enabled
	if m.state.AutoSave {
		go func() {
			_ = m.saveConfigSync()
		}()
	}

	m.state.NavigateBack()

	return m, nil
}

// handleSettingsSave processes settings form submission
func (m Model) handleSettingsSave() (tea.Model, tea.Cmd) {
	// Update state with new API keys (Zone ID is now per-domain)
	m.state.CloudflareAPIToken = m.state.FormFields[0]
	m.state.Route53AccessKey = m.state.FormFields[1]
	m.state.Route53SecretKey = m.state.FormFields[2]

	m.state.AddNotification("Settings saved successfully", "success")

	// Auto-save config if enabled
	if m.state.AutoSave {
		go func() {
			_ = m.saveConfigSync()
		}()
	}

	m.state.NavigateBack()

	return m, nil
}

// handleNodeCreateSubmit processes node creation form submission
func (m Model) handleNodeCreateSubmit() (tea.Model, tea.Cmd) {
	// Validate editable fields are filled (first 2)
	if m.state.FormFields[0] == "" {
		m.state.AddNotification("Name is required", "error")
		return m, nil
	}
	if m.state.FormFields[1] == "" {
		m.state.AddNotification("API Endpoint is required", "error")
		return m, nil
	}

	// Try to extract IP from API endpoint, or use placeholder
	endpoint := m.state.FormFields[1]
	var ip net.IP

	// Try to extract hostname from URL
	if strings.Contains(endpoint, "://") {
		// Parse as URL
		parts := strings.Split(endpoint, "://")
		if len(parts) > 1 {
			host := strings.Split(parts[1], ":")[0] // Remove port if present
			ip = net.ParseIP(host)
		}
	} else {
		// Try direct IP parse
		ip = net.ParseIP(endpoint)
	}

	// If still no valid IP, use placeholder
	if ip == nil {
		ip = net.ParseIP("0.0.0.0")
	}

	// Check for duplicate name
	for _, node := range m.state.Nodes {
		if node.Name == m.state.FormFields[0] {
			m.state.AddNotification("Node already exists: "+m.state.FormFields[0], "error")
			return m, nil
		}
	}

	// Parse proxy type from field 2
	proxyTypeStr := m.state.FormFields[2]
	var proxyType models.ProxyType
	switch proxyTypeStr {
	case "nginx":
		proxyType = models.ProxyTypeNginx
	case "apache":
		proxyType = models.ProxyTypeApache
	case "traefik":
		proxyType = models.ProxyTypeTraefik
	default:
		proxyType = models.ProxyTypeNginx
	}

	// Create new node with proxy type and generated API key
	node := models.NewNode(
		m.state.FormFields[0], // name
		m.state.FormFields[1], // endpoint
		m.state.FormFields[3], // generated api key (now field 3)
		ip,
		proxyType, // proxy type
	)

	m.state.Nodes = append(m.state.Nodes, *node)

	// Get proxy label for notification
	proxyLabels := map[string]string{
		"nginx":   "Nginx",
		"apache":  "Apache2",
		"traefik": "Traefik",
	}
	proxyLabel := proxyLabels[proxyTypeStr]
	if proxyLabel == "" {
		proxyLabel = proxyTypeStr
	}

	m.state.AddNotification(fmt.Sprintf("Node created: %s (%s, API Key: %s)", node.Name, proxyLabel, node.APIKey), "success")

	// Auto-save config if enabled
	if m.state.AutoSave {
		go func() {
			_ = m.saveConfigSync()
		}()
	}

	// Set selected node and navigate to config screen
	m.state.SelectedNodeID = node.ID
	m.state.NavigateTo(state.ScreenNodeConfig)

	return m, nil
}

// handleNodeEditSubmit processes node edit form submission
func (m Model) handleNodeEditSubmit() (tea.Model, tea.Cmd) {
	// Validate fields
	if m.state.FormFields[0] == "" {
		m.state.AddNotification("Name is required", "error")
		return m, nil
	}
	if m.state.FormFields[1] == "" {
		m.state.AddNotification("API Endpoint is required", "error")
		return m, nil
	}

	// Find the node being edited
	var nodeIndex = -1
	for i := range m.state.Nodes {
		if m.state.Nodes[i].ID == m.state.SelectedNodeID {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		m.state.AddNotification("Node not found", "error")
		m.state.NavigateBack()
		return m, nil
	}

	oldName := m.state.Nodes[nodeIndex].Name
	oldEndpoint := m.state.Nodes[nodeIndex].APIEndpoint
	oldProxyType := string(m.state.Nodes[nodeIndex].ProxyType)

	// Check for duplicate name (excluding current node)
	for i, node := range m.state.Nodes {
		if i != nodeIndex && node.Name == m.state.FormFields[0] {
			m.state.AddNotification("Node already exists: "+m.state.FormFields[0], "error")
			return m, nil
		}
	}

	// Update node fields
	m.state.Nodes[nodeIndex].Name = m.state.FormFields[0]
	m.state.Nodes[nodeIndex].APIEndpoint = m.state.FormFields[1]

	// Parse and update proxy type
	proxyTypeStr := m.state.FormFields[2]
	var proxyType models.ProxyType
	switch proxyTypeStr {
	case "nginx":
		proxyType = models.ProxyTypeNginx
	case "apache":
		proxyType = models.ProxyTypeApache
	case "traefik":
		proxyType = models.ProxyTypeTraefik
	default:
		proxyType = models.ProxyTypeNginx
	}
	m.state.Nodes[nodeIndex].ProxyType = proxyType

	// Try to extract IP from API endpoint
	endpoint := m.state.FormFields[1]
	var ip net.IP
	if strings.Contains(endpoint, "://") {
		parts := strings.Split(endpoint, "://")
		if len(parts) > 1 {
			host := strings.Split(parts[1], ":")[0]
			ip = net.ParseIP(host)
		}
	} else {
		ip = net.ParseIP(endpoint)
	}
	if ip != nil {
		m.state.Nodes[nodeIndex].IPAddress = ip
	}

	// Build notification message
	var changes []string
	if oldName != m.state.FormFields[0] {
		changes = append(changes, fmt.Sprintf("name: %s → %s", oldName, m.state.FormFields[0]))
	}
	if oldEndpoint != m.state.FormFields[1] {
		changes = append(changes, fmt.Sprintf("endpoint: %s → %s", oldEndpoint, m.state.FormFields[1]))
	}
	if oldProxyType != proxyTypeStr {
		proxyLabels := map[string]string{
			"nginx":   "Nginx",
			"apache":  "Apache2",
			"traefik": "Traefik",
		}
		oldLabel := proxyLabels[oldProxyType]
		if oldLabel == "" {
			oldLabel = oldProxyType
		}
		newLabel := proxyLabels[proxyTypeStr]
		if newLabel == "" {
			newLabel = proxyTypeStr
		}
		changes = append(changes, fmt.Sprintf("proxy: %s → %s", oldLabel, newLabel))
	}

	var message string
	if len(changes) > 0 {
		message = "Node updated: " + strings.Join(changes, ", ")
	} else {
		message = "Node updated (no changes)"
	}
	m.state.AddNotification(message, "success")

	// Auto-save config if enabled
	if m.state.AutoSave {
		go func() {
			_ = m.saveConfigSync()
		}()
	}

	m.state.NavigateBack()

	return m, nil
}

// ============================================================================
// Bubblezone Mouse Handlers
// ============================================================================

// handleTabClick handles clicks on tab items
func (m Model) handleTabClick(zoneID string) (tea.Model, tea.Cmd) {
	// Extract tab ID from zone ID (format: "tab:dashboard", "tab:sites", etc.)
	tabID := zoneID[4:] // Remove "tab:" prefix

	switch tabID {
	case "dashboard":
		m.state.NavigateTo(state.ScreenDashboard)
	case "sites":
		m.state.NavigateTo(state.ScreenSitesList)
	case "domains":
		m.state.NavigateTo(state.ScreenDomainsList)
	case "nodes":
		m.state.NavigateTo(state.ScreenNodesList)
	case "settings":
		m.state.NavigateTo(state.ScreenSettings)
	case "help":
		m.state.NavigateTo(state.ScreenHelp)
	}

	return m, nil
}

// handleFieldClick handles clicks on form fields
func (m Model) handleFieldClick(zoneID string) (tea.Model, tea.Cmd) {
	// Extract field index from zone ID (format: "field:0", "field:1", etc.)
	var fieldIndex int
	fmt.Sscanf(zoneID, "field:%d", &fieldIndex)

	// Update current field index if valid
	if fieldIndex >= 0 && fieldIndex < len(m.state.FormFields) {
		m.state.CurrentFieldIndex = fieldIndex
	}

	return m, nil
}

// ============================================================================
// Config Management Helpers
// ============================================================================

// saveConfigSync synchronously saves the current state to config file
func (m Model) saveConfigSync() error {
	cfg := &config.Config{
		Version:  "1.0.0",
		Sites:    m.state.Sites,
		Domains:  m.state.Domains,
		Nodes:    m.state.Nodes,
		Settings: config.Settings{
			AutoSave:                m.state.AutoSave,
			HealthCheckIntervalSecs: 60,
			DefaultDnsTTL:           3600,
			Theme:                   "default",
			CloudflareAPIToken:      m.state.CloudflareAPIToken,
			Route53AccessKey:        m.state.Route53AccessKey,
			Route53SecretKey:        m.state.Route53SecretKey,
		},
	}

	return m.configLoader.Save(m.configPath, cfg)
}

// ============================================================================
// Delete Handlers
// ============================================================================

// handleDeleteSite removes a site from the state and filesystem
func (m Model) handleDeleteSite(siteID uuid.UUID) (tea.Model, tea.Cmd) {
	// Find and remove site
	for i, site := range m.state.Sites {
		if site.ID == siteID {
			// Get domain name for site directory structure
			domain := m.state.GetDomainByID(site.DomainID)
			if domain != nil {
				// Delete site files from filesystem
				if err := m.configLoader.DeleteSite(site.Name, domain.Name); err != nil {
					m.state.AddNotification("Failed to delete site files: "+err.Error(), "error")
					// Continue with state removal anyway
				}
			}

			// Remove from slice
			m.state.Sites = append(m.state.Sites[:i], m.state.Sites[i+1:]...)
			m.state.AddNotification("Deleted site: "+site.Name, "success")

			// Auto-save config if enabled
			if m.state.AutoSave {
				go func() {
					_ = m.saveConfigSync()
				}()
			}

			return m, nil
		}
	}

	m.state.AddNotification("Site not found", "error")
	return m, nil
}

// handleDeleteDomain removes a domain from the state
func (m Model) handleDeleteDomain(domainID uuid.UUID) (tea.Model, tea.Cmd) {
	// Check if domain is used by any sites
	for _, site := range m.state.Sites {
		if site.DomainID == domainID {
			m.state.AddNotification("Cannot delete domain: used by site "+site.Name, "error")
			return m, nil
		}
	}

	// Find and remove domain
	for i, domain := range m.state.Domains {
		if domain.ID == domainID {
			// Remove from slice
			m.state.Domains = append(m.state.Domains[:i], m.state.Domains[i+1:]...)
			m.state.AddNotification("Deleted domain: "+domain.Name, "success")

			// Auto-save config if enabled
			if m.state.AutoSave {
				go func() {
					_ = m.saveConfigSync()
				}()
			}

			return m, nil
		}
	}

	m.state.AddNotification("Domain not found", "error")
	return m, nil
}

// handleDeleteNode removes a node from the state and filesystem
func (m Model) handleDeleteNode(nodeID uuid.UUID) (tea.Model, tea.Cmd) {
	// Check if node is used by any sites
	for _, site := range m.state.Sites {
		if site.NodeID == nodeID {
			m.state.AddNotification("Cannot delete node: used by site "+site.Name, "error")
			return m, nil
		}
	}

	// Find and remove node
	for i, node := range m.state.Nodes {
		if node.ID == nodeID {
			// Delete node files from filesystem
			if err := m.configLoader.DeleteNode(node.Name); err != nil {
				m.state.AddNotification("Failed to delete node files: "+err.Error(), "error")
				// Continue with state removal anyway
			}

			// Remove from slice
			m.state.Nodes = append(m.state.Nodes[:i], m.state.Nodes[i+1:]...)
			m.state.AddNotification("Deleted node: "+node.Name, "success")

			// Auto-save config if enabled
			if m.state.AutoSave {
				go func() {
					_ = m.saveConfigSync()
				}()
			}

			return m, nil
		}
	}

	m.state.AddNotification("Node not found", "error")
	return m, nil
}

// tryDetectComposePorts attempts to detect ports from the compose file and pre-populate domain mapping
func (m *Model) tryDetectComposePorts() {
	path := m.state.FormFields[2]
	if path == "" {
		return
	}

	_, ports, err := loadComposeFile(path)
	if err != nil {
		// Show error but don't block - user may fix the path
		m.state.AddNotification(err.Error(), "warning")
		return
	}

	if len(ports) > 0 {
		firstPort := compose.GetFirstPort(ports)
		if firstPort > 0 && len(m.state.DomainMappingPairs) > 0 {
			// Only update if port is still the default
			if m.state.DomainMappingPairs[0].Port == "8080" || m.state.DomainMappingPairs[0].Port == "" {
				m.state.DomainMappingPairs[0].Port = fmt.Sprintf("%d", firstPort)
				m.state.AddNotification(fmt.Sprintf("Detected port %d from compose file", firstPort), "info")
			}
		}
	}
}

// loadComposeFile reads and validates a Docker Compose file from the given path
func loadComposeFile(path string) (string, []compose.DetectedPort, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	contentStr := string(content)

	// Basic validation: check for required 'services:' key
	if !strings.Contains(contentStr, "services:") {
		return "", nil, fmt.Errorf("invalid compose file: missing 'services:' key")
	}

	// Parse ports from compose file
	ports, err := compose.ParsePorts(contentStr)
	if err != nil {
		// Log warning but don't fail - ports are optional enhancement
		ports = nil
	}

	return contentStr, ports, nil
}
