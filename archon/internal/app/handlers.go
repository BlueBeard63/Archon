package app

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/BlueBeard63/archon/internal/config"
	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
)

// ============================================================================
// Keyboard Event Handlers
// ============================================================================

// handleKeyPress routes keyboard events to screen-specific handlers
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on a form screen (prioritize form input)
	isFormScreen := m.state.CurrentScreen == state.ScreenSiteCreate ||
		m.state.CurrentScreen == state.ScreenSiteEdit ||
		m.state.CurrentScreen == state.ScreenDomainCreate ||
		m.state.CurrentScreen == state.ScreenDomainEdit ||
		m.state.CurrentScreen == state.ScreenNodeCreate ||
		m.state.CurrentScreen == state.ScreenNodeEdit ||
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

	case "e":
		// Edit selected site
		if len(m.state.Sites) > 0 && m.state.SitesListIndex >= 0 && m.state.SitesListIndex < len(m.state.Sites) {
			site := m.state.Sites[m.state.SitesListIndex]
			m.state.AddNotification("Edit site: "+site.Name+" (not yet implemented)", "info")
		}
		return m, nil

	case "d":
		// Delete selected site
		if len(m.state.Sites) > 0 && m.state.SitesListIndex >= 0 && m.state.SitesListIndex < len(m.state.Sites) {
			site := m.state.Sites[m.state.SitesListIndex]
			return m.handleDeleteSite(site.ID)
		}
		return m, nil
	}

	return m, nil
}

// handleSiteCreateKeys handles keys on the site creation form
func (m Model) handleSiteCreateKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check if we're on a dropdown field (Domain=1, Node=2)
	isDropdownField := m.state.CurrentFieldIndex == 1 || m.state.CurrentFieldIndex == 2

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
			maxIndex := 0
			if m.state.CurrentFieldIndex == 1 {
				maxIndex = len(m.state.Domains) - 1
			} else if m.state.CurrentFieldIndex == 2 {
				maxIndex = len(m.state.Nodes) - 1
			}
			if m.state.DropdownIndex < maxIndex {
				m.state.DropdownIndex++
			}
			return m, nil

		case tea.KeyEnter, tea.KeyTab:
			// Confirm selection and close dropdown
			if m.state.CurrentFieldIndex == 1 && len(m.state.Domains) > 0 {
				m.state.FormFields[1] = m.state.Domains[m.state.DropdownIndex].Name
			} else if m.state.CurrentFieldIndex == 2 && len(m.state.Nodes) > 0 {
				m.state.FormFields[2] = m.state.Nodes[m.state.DropdownIndex].Name
			}
			m.state.DropdownOpen = false

			// If Tab, move to next field
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
		// Add space to current field (only domain name field 0)
		if m.state.CurrentFieldIndex == 0 {
			m.state.FormFields[0] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to domain name field only
		if m.state.CurrentFieldIndex == 0 {
			m.state.FormFields[0] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from domain name field only
		if m.state.CurrentFieldIndex == 0 && len(m.state.FormFields[0]) > 0 {
			m.state.FormFields[0] = m.state.FormFields[0][:len(m.state.FormFields[0])-1]
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
		// Add space to current field (only domain name field 0)
		if m.state.CurrentFieldIndex == 0 {
			m.state.FormFields[0] += " "
		}
		return m, nil

	case tea.KeyRunes:
		// Add character to domain name field only
		if m.state.CurrentFieldIndex == 0 {
			m.state.FormFields[0] += string(msg.Runes)
		}
		return m, nil

	case tea.KeyBackspace:
		// Remove last character from domain name field only
		if m.state.CurrentFieldIndex == 0 && len(m.state.FormFields[0]) > 0 {
			m.state.FormFields[0] = m.state.FormFields[0][:len(m.state.FormFields[0])-1]
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
		// TODO: Save config to file
		m.state.AddNotification("Save to file not yet implemented", "info")
		return m, nil
	}

	return m, nil
}

// handleSettingsKeys handles keys on the settings form
func (m Model) handleSettingsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeySpace:
		// Add space to current field
		if m.state.CurrentFieldIndex < len(m.state.FormFields) {
			m.state.FormFields[m.state.CurrentFieldIndex] += " "
		}
		return m, nil

	case tea.KeyRunes:
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
		// Move to next field
		m.state.CurrentFieldIndex = (m.state.CurrentFieldIndex + 1) % len(m.state.FormFields)
		return m, nil

	case tea.KeyShiftTab:
		// Move to previous field
		m.state.CurrentFieldIndex--
		if m.state.CurrentFieldIndex < 0 {
			m.state.CurrentFieldIndex = len(m.state.FormFields) - 1
		}
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
	// Validate required fields (first 5 fields, last 2 are optional)
	for i := 0; i < 5; i++ {
		if m.state.FormFields[i] == "" {
			m.state.AddNotification("Required fields (Name, Domain, Node, Image, Port) must be filled", "error")
			return m, nil
		}
	}

	// Parse port
	var port int
	_, err := fmt.Sscanf(m.state.FormFields[4], "%d", &port)
	if err != nil || port < 1 || port > 65535 {
		m.state.AddNotification("Invalid port number", "error")
		return m, nil
	}

	// Find domain by name
	var domainID uuid.UUID
	domainFound := false
	for _, domain := range m.state.Domains {
		if domain.Name == m.state.FormFields[1] {
			domainID = domain.ID
			domainFound = true
			break
		}
	}
	if !domainFound {
		m.state.AddNotification("Domain not found: "+m.state.FormFields[1], "error")
		return m, nil
	}

	// Find node by name
	var nodeID uuid.UUID
	nodeFound := false
	for _, node := range m.state.Nodes {
		if node.Name == m.state.FormFields[2] {
			nodeID = node.ID
			nodeFound = true
			break
		}
	}
	if !nodeFound {
		m.state.AddNotification("Node not found: "+m.state.FormFields[2], "error")
		return m, nil
	}

	// Create new site
	site := models.NewSite(m.state.FormFields[0], domainID, nodeID, m.state.FormFields[3], port)

	// Parse environment variables (field 5) if provided
	if m.state.FormFields[5] != "" {
		envVars := strings.Split(m.state.FormFields[5], "|")
		for _, envVar := range envVars {
			envVar = strings.TrimSpace(envVar)
			if envVar == "" {
				continue
			}
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key != "" {
					site.EnvironmentVars[key] = value
				}
			}
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

	m.state.Sites = append(m.state.Sites, *site)

	m.state.AddNotification("Site created: "+site.Name, "success")

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

	// Create DNS provider based on type
	var provider models.DnsProvider
	switch providerType {
	case "cloudflare":
		provider = models.DnsProvider{Type: models.DnsProviderCloudflare}
	case "route53":
		provider = models.DnsProvider{Type: models.DnsProviderRoute53}
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

	// Update DNS provider based on type
	var provider models.DnsProvider
	switch providerType {
	case "cloudflare":
		provider = models.DnsProvider{Type: models.DnsProviderCloudflare}
	case "route53":
		provider = models.DnsProvider{Type: models.DnsProviderRoute53}
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
	// Update state with new API keys
	m.state.CloudflareAPIKey = m.state.FormFields[0]
	m.state.CloudflareAPIToken = m.state.FormFields[1]
	m.state.Route53AccessKey = m.state.FormFields[2]
	m.state.Route53SecretKey = m.state.FormFields[3]

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
			CloudflareAPIKey:        m.state.CloudflareAPIKey,
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

// handleDeleteSite removes a site from the state
func (m Model) handleDeleteSite(siteID uuid.UUID) (tea.Model, tea.Cmd) {
	// Find and remove site
	for i, site := range m.state.Sites {
		if site.ID == siteID {
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

// handleDeleteNode removes a node from the state
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
