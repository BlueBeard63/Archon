package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui/components"
)

// RenderSitesList renders the sites list screen with buttons
func RenderSitesList(s *state.AppState) string {
	return RenderSitesListWithZones(s, nil)
}

// RenderSitesListWithZones renders sites list with optional button zones
func RenderSitesListWithZones(s *state.AppState, zm *zone.Manager) string {
	title := titleStyle.Render("🌐 Sites")

	// Create button group
	buttonGroup := &components.ButtonGroup{
		Buttons: []components.Button{
			{ID: "create-site", Label: "➕ Create Site", Primary: true},
		},
	}

	var buttons string
	if zm != nil {
		buttons = buttonGroup.RenderWithZones(zm)
	} else {
		buttons = buttonGroup.Render()
	}

	var content string
	if len(s.Sites) == 0 {
		content = helpStyle.Render("No sites yet. Click 'Create Site' or press 'n'.")
	} else {
		// 1. Build table rows (data only, NO buttons)
		var rows []table.Row
		for _, site := range s.Sites {
			// Get domain and node names
			nodeName := site.NodeID.String()[:8] + "..."

			// Find node name
			for _, n := range s.Nodes {
				if n.ID == site.NodeID {
					nodeName = n.Name
					break
				}
			}

			// Get domain mappings (supports multiple domains)
			mappings := site.GetDomainMappings()
			var domainDisplay, portDisplay string

			if len(mappings) == 0 {
				domainDisplay = "(none)"
				portDisplay = "-"
			} else if len(mappings) == 1 {
				// Single domain - show name and port
				domainName := mappings[0].DomainID.String()[:8] + "..."
				for _, d := range s.Domains {
					if d.ID == mappings[0].DomainID {
						domainName = d.Name
						break
					}
				}
				domainDisplay = domainName
				portDisplay = fmt.Sprintf("%d", mappings[0].Port)
			} else {
				// Multiple domains - show count
				domainDisplay = fmt.Sprintf("Multiple (%d)", len(mappings))
				portDisplay = fmt.Sprintf("%d+", mappings[0].Port)
			}

			// Ensure status has a valid value
			statusDisplay := string(site.Status)
			if statusDisplay == "" {
				statusDisplay = "inactive"
			}

			rows = append(rows, table.Row{
				truncate(site.Name, 20),
				truncate(domainDisplay, 35),
				truncate(nodeName, 20),
				truncate(portDisplay, 20),
				truncate(statusDisplay, 10),
			})
		}

		// 2. Initialize/update table
		if s.SitesTable == nil {
			columns := []table.Column{
				{Title: "Name", Width: 20},
				{Title: "Domain", Width: 35},
				{Title: "Node", Width: 20},
				{Title: "Port", Width: 20},
				{Title: "Status", Width: 10},
			}
			s.SitesTable = components.NewTableComponent(columns, rows)
			s.SitesTable.SetCursor(s.SitesListIndex)
		} else {
			s.SitesTable.SetRows(rows)
			s.SitesTable.SetCursor(s.SitesListIndex)
		}

		// 3. Render table view
		tableView := s.SitesTable.View()

		// 4. Build action buttons column (aligned with rows)
		var actionsColumn strings.Builder
		actionsColumn.WriteString("\n\n") // Header padding

		for _, site := range s.Sites {
			var buttons []string

			// Only show deploy and DNS buttons for inactive/failed sites
			if site.Status == models.SiteStatusInactive || site.Status == models.SiteStatusFailed || site.Status == "" {
				// DNS setup button
				dnsBtn := components.Button{
					ID:      "setup-dns-" + site.ID.String(),
					Label:   "🌐",
					Primary: false,
					Border:  false,
					Icon:    true,
				}
				if zm != nil {
					buttons = append(buttons, dnsBtn.RenderWithZone(zm))
				} else {
					buttons = append(buttons, dnsBtn.Render())
				}

				// Deploy button
				deployBtn := components.Button{
					ID:      "deploy-site-" + site.ID.String(),
					Label:   "🚀",
					Primary: false,
					Border:  false,
					Icon:    true,
				}
				if zm != nil {
					buttons = append(buttons, deployBtn.RenderWithZone(zm))
				} else {
					buttons = append(buttons, deployBtn.Render())
				}
			}

			// Show stop button for running sites, play button for stopped/failed sites
			var controlBtn components.Button
			switch site.Status {
			case models.SiteStatusRunning, models.SiteStatusDeploying:
				controlBtn = components.Button{
					ID:      "stop-site-" + site.ID.String(),
					Label:   "⏹️",
					Primary: false,
					Border:  false,
					Icon:    true,
				}
			case models.SiteStatusStopped, models.SiteStatusFailed:
				controlBtn = components.Button{
					ID:      "restart-site-" + site.ID.String(),
					Label:   "▶️",
					Primary: false,
					Border:  false,
					Icon:    true,
				}
			}

			// Add control button if status is not inactive
			if (site.Status != models.SiteStatusInactive && site.Status != models.SiteStatusFailed) && site.Status != "" {
				if zm != nil {
					buttons = append(buttons, controlBtn.RenderWithZone(zm))
				} else {
					buttons = append(buttons, controlBtn.Render())
				}
			}

			editBtn := components.Button{
				ID:      "edit-site-" + site.ID.String(),
				Label:   "✏️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}
			deleteBtn := components.Button{
				ID:      "delete-site-" + site.ID.String(),
				Label:   "🗑️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}

			if zm != nil {
				buttons = append(buttons, editBtn.RenderWithZone(zm))
				buttons = append(buttons, deleteBtn.RenderWithZone(zm))
			} else {
				buttons = append(buttons, editBtn.Render())
				buttons = append(buttons, deleteBtn.Render())
			}

			actionLine := strings.Join(buttons, " ")
			actionsColumn.WriteString(actionLine + "\n")
		}

		// 5. Join table + actions horizontally
		mainContent := lipgloss.JoinHorizontal(
			lipgloss.Top,
			tableView,
			actionsColumn.String(),
		)

		// 6. Build sidebar for selected site
		var sidebar string
		if len(s.Sites) > 0 && s.SitesListIndex >= 0 && s.SitesListIndex < len(s.Sites) {
			site := &s.Sites[s.SitesListIndex]
			sidebar = renderSiteSidebar(s, site)
		}

		// 7. Join main content + sidebar
		if sidebar != "" {
			content = lipgloss.JoinHorizontal(
				lipgloss.Top,
				mainContent,
				"  ", // Spacing
				sidebar,
			)
		} else {
			content = mainContent
		}
	}

	help := helpStyle.Render("\n\nPress j/k or arrows to navigate • Space/Enter to deploy • s to start/stop • e to edit • d to delete • n to create • Esc to go back")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		buttons,
		"",
		content,
		help,
	)
}

// truncate truncates a string to maxLen characters and pads with spaces if shorter
func truncate(s string, maxLen int) string {
	// If string is longer than maxLen, truncate it
	if len(s) > maxLen {
		if maxLen <= 3 {
			return s[:maxLen]
		}
		return s[:maxLen-3] + "..."
	}

	// If string is shorter than maxLen, pad with spaces
	if len(s) < maxLen {
		return s + strings.Repeat(" ", maxLen-len(s))
	}

	// String is exactly maxLen
	return s
}

// RenderSiteCreate renders the site creation form
func RenderSiteCreate(s *state.AppState) string {
	return RenderSiteCreateWithZones(s, nil)
}

// RenderSiteCreateWithZones renders the site creation form with clickable fields
func RenderSiteCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Always ensure form is properly initialized (7 fields: name, node, docker image/compose path, docker username, docker token, ssl email, config file)
	if len(s.FormFields) != 7 {
		s.FormFields = []string{"", "", "", "", "", "", ""}
	}

	// Only reset field index if it's out of bounds (-1 is valid for site type selector)
	if s.CurrentFieldIndex < -1 || s.CurrentFieldIndex > 200 {
		s.CurrentFieldIndex = -1 // Start at site type selector
	}

	// Initialize site type selection if empty
	if s.SiteTypeSelection == "" {
		s.SiteTypeSelection = "container"
	}

	// Initialize compose input method if empty
	if s.ComposeInputMethod == "" {
		s.ComposeInputMethod = "file"
	}

	// Initialize ENV vars with one empty pair if needed
	if len(s.EnvVarPairs) == 0 {
		s.EnvVarPairs = []state.EnvVarPair{{Key: "", Value: ""}}
	}

	// Initialize domain mappings with one empty pair if needed
	if len(s.DomainMappingPairs) == 0 {
		s.DomainMappingPairs = []state.DomainMappingPair{{Subdomain: "", DomainName: "", DomainID: "", Port: "8080"}}
	}

	title := titleStyle.Render("Create New Site")

	isCompose := s.SiteTypeSelection == "compose"

	// Build fields string
	var fields string

	// Site Type selector (special field index -1, rendered first)
	siteTypeLabel := "  Deployment Type:"
	if s.CurrentFieldIndex == -1 {
		siteTypeLabel = "> Deployment Type:"
	}
	siteTypeValue := "Container"
	if isCompose {
		siteTypeValue = "Compose"
	}
	if s.CurrentFieldIndex == -1 {
		siteTypeValue += "_"
	}
	siteTypeLine := siteTypeLabel + " " + siteTypeValue + "\n"
	fields += zm.Mark("field:-1", siteTypeLine)

	// Show site type dropdown when focused
	if s.CurrentFieldIndex == -1 && s.DropdownOpen {
		options := []string{"Container", "Compose"}
		var dropdownStr string
		for i, opt := range options {
			prefix := "    "
			if i == s.DropdownIndex {
				prefix = "  > "
			}
			dropdownStr += prefix + opt + "\n"
		}
		fields += dropdownStr
	}

	// Define labels based on site type
	var labels []string
	if isCompose {
		labels = []string{
			"Name:",
			"Node:",
			"Compose File Path:",
			"", // Hidden (docker username)
			"", // Hidden (docker token)
			"SSL Email (for Let's Encrypt):",
			"", // Hidden (config file - not applicable for compose)
		}
	} else {
		labels = []string{
			"Name:",
			"Node:",
			"Docker Image:",
			"Docker Username:",
			"Docker Token:",
			"SSL Email (for Let's Encrypt):",
			"Config File Path (optional):",
		}
	}

	// Render each field with zones
	for i, label := range labels {
		// Skip hidden fields (empty labels)
		if label == "" {
			continue
		}

		value := s.FormFields[i]
		displayValue := value

		// Show cursor if focused
		if i == s.CurrentFieldIndex {
			displayValue = value + "_"
			label = "> " + label // Show arrow for focused field
		} else {
			label = "  " + label
		}

		// Wrap the entire field line in a clickable zone
		fieldLine := label + " " + displayValue + "\n"
		fields += zm.Mark(fmt.Sprintf("field:%d", i), fieldLine)

		// Show dropdown options for Node (index 1) when focused
		if i == s.CurrentFieldIndex && i == 1 && s.DropdownOpen {
			// Node dropdown
			dropdownOptions := renderDropdownOptions(s, s.Nodes, s.DropdownIndex, func(n models.Node) string {
				return n.Name
			})
			fields += dropdownOptions + "\n"
		}
	}

	// Render domain mappings section
	domainMappingsSection := renderDomainMappingsSection(s, zm)
	fields += "\n" + domainMappingsSection

	// Render ENV vars section (only for container deployments, compose has its own env handling)
	if !isCompose {
		envSection := renderEnvVarsSection(s, zm)
		fields += "\n" + envSection
	}

	helpText := "\nTab/Shift+Tab to navigate, Enter to create, Esc to cancel"
	switch s.CurrentFieldIndex {
	case -1:
		// Site type selector
		if s.DropdownOpen {
			helpText = "\nUp/Down to select, Enter/Tab to confirm"
		} else {
			helpText = "\nPress Enter or Down to open dropdown, Tab to skip"
		}
	case 1:
		// On Node dropdown field
		if s.DropdownOpen {
			helpText = "\nUp/Down to select, Enter/Tab to confirm, Esc to cancel"
		} else {
			helpText = "\nPress Enter or Down to open dropdown, Tab to skip"
		}
	case 2:
		if isCompose {
			helpText = "\nEnter path to docker-compose.yml file"
		} else {
			helpText = "\nDocker image to deploy (e.g., nginx:latest, myrepo/myimage:v1)"
		}
	case 3:
		helpText = "\nLeave blank to skip Docker Auth (if image is public)"
	case 4:
		helpText = "\nLeave blank to skip Docker Auth (if image is public)"
	case 5:
		helpText = "\nEmail for Let's Encrypt SSL certificate notifications (e.g., admin@example.com)"
	case 6:
		helpText = "\nEnter full path to config file (will be loaded when site is created)"
	case 100:
		// Special index for ENV vars
		helpText = "\nType key/value, Tab to switch between key/value, +/- buttons to add/remove pairs"
	case 200:
		// Special index for domain mappings
		helpText = "\nSelect subdomain/domain/port, Tab to switch fields, +/- buttons to add/remove mappings"
	}

	help := helpStyle.Render(helpText)

	var note string
	if isCompose {
		note = helpStyle.Render("Note: Compose deployment • Domain mappings apply to first service exposed port")
	} else {
		note = helpStyle.Render("Note: Node uses dropdown • Use + to add domain mappings/ENV vars, - to remove")
	}

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// RenderSiteEdit renders the site editing form
func RenderSiteEdit(s *state.AppState) string {
	return RenderSiteEditWithZones(s, nil)
}

// RenderSiteEditWithZones renders the site editing form with clickable fields
func RenderSiteEditWithZones(s *state.AppState, zm *zone.Manager) string {
	// Find the site being edited
	site := s.GetSiteByID(s.SelectedSiteID)
	if site == nil {
		return titleStyle.Render("Error: Site not found")
	}

	// Always initialize form fields with current site data (7 fields: name, node, docker image, docker username, docker token, ssl email, config file)
	// This ensures the form is always showing the correct data, even if it was previously used
	s.FormFields = make([]string, 7)
	s.FormFields[0] = site.Name
	s.FormFields[2] = site.DockerImage
	s.FormFields[3] = site.DockerUsername
	s.FormFields[4] = site.DockerToken
	s.FormFields[5] = site.SSLEmail

	// Find node name
	for _, n := range s.Nodes {
		if n.ID == site.NodeID {
			s.FormFields[1] = n.Name
			break
		}
	}

	// Config file path (leave blank or show first config file name)
	if len(site.ConfigFiles) > 0 {
		s.FormFields[6] = site.ConfigFiles[0].Name
	} else {
		s.FormFields[6] = ""
	}

	// Only reset field index if it's out of bounds
	if s.CurrentFieldIndex < 0 || s.CurrentFieldIndex > 200 {
		s.CurrentFieldIndex = 0
	}

	// Always initialize domain mapping pairs from current site data
	s.DomainMappingPairs = []state.DomainMappingPair{}
	mappings := site.GetDomainMappings()
	if len(mappings) > 0 {
		// Convert existing mappings to UI pairs
		for _, mapping := range mappings {
			// Find domain name
			domainName := ""
			for _, d := range s.Domains {
				if d.ID == mapping.DomainID {
					domainName = d.Name
					break
				}
			}
			s.DomainMappingPairs = append(s.DomainMappingPairs, state.DomainMappingPair{
				Subdomain:  mapping.Subdomain,
				DomainName: domainName,
				DomainID:   mapping.DomainID.String(),
				Port:       models.FormatPortMapping(mapping.Port, mapping.HostPort),
			})
		}
	} else {
		// Start with one empty pair
		s.DomainMappingPairs = []state.DomainMappingPair{{Subdomain: "", DomainName: "", DomainID: "", Port: "8080"}}
	}

	// Always initialize ENV vars from current site data
	s.EnvVarPairs = []state.EnvVarPair{}
	if len(site.EnvironmentVars) > 0 {
		// Convert map to pairs
		for key, value := range site.EnvironmentVars {
			s.EnvVarPairs = append(s.EnvVarPairs, state.EnvVarPair{
				Key:   key,
				Value: value,
			})
		}
	} else {
		// Start with one empty pair
		s.EnvVarPairs = []state.EnvVarPair{{Key: "", Value: ""}}
	}

	title := titleStyle.Render("Edit Site: " + site.Name)

	labels := []string{
		"Name:",
		"Node:",
		"Docker Image:",
		"Docker Username:",
		"Docker Token:",
		"SSL Email (for Let's Encrypt):",
		"Config File Path (optional):",
	}

	// Render each field with zones
	var fields string
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value

		// Show cursor if focused
		if i == s.CurrentFieldIndex {
			displayValue = value + "_"
			label = "> " + label // Show arrow for focused field
		} else {
			label = "  " + label
		}

		// Wrap the entire field line in a clickable zone
		fieldLine := label + " " + displayValue + "\n"
		fields += zm.Mark(fmt.Sprintf("field:%d", i), fieldLine)

		// Show dropdown options for Node (index 1) when focused
		if i == s.CurrentFieldIndex && i == 1 && s.DropdownOpen {
			// Node dropdown
			dropdownOptions := renderDropdownOptions(s, s.Nodes, s.DropdownIndex, func(n models.Node) string {
				return n.Name
			})
			fields += dropdownOptions + "\n"
		}
	}

	// Render domain mappings section
	domainMappingsSection := renderDomainMappingsSection(s, zm)
	fields += "\n" + domainMappingsSection

	// Render ENV vars section
	envSection := renderEnvVarsSection(s, zm)
	fields += "\n" + envSection

	helpText := "\nTab/Shift+Tab to navigate, Enter to create, Esc to cancel"
	switch s.CurrentFieldIndex {
	case 1:
		// On Node dropdown field
		if s.DropdownOpen {
			helpText = "\nUp/Down to select, Enter/Tab to confirm, Esc to cancel"
		} else {
			helpText = "\nPress Enter or Down to open dropdown, Tab to skip"
		}
	case 3:
		helpText = "\nLeave black to skip Docker Auth (if image is public)"
	case 4:
		helpText = "\nLeave black to skip Docker Auth (if image is public)"
	case 5:
		helpText = "\nEmail for Let's Encrypt SSL certificate notifications (e.g., admin@example.com)"
	case 6:
		helpText = "\nEnter full path to config file (will be loaded when site is created)"
	case 100:
		// Special index for ENV vars
		helpText = "\nType key/value, Tab to switch between key/value, +/- buttons to add/remove pairs"
	case 200:
		// Special index for domain mappings
		helpText = "\nSelect subdomain/domain/port, Tab to switch fields, +/- buttons to add/remove mappings"
	}

	help := helpStyle.Render(helpText)
	note := helpStyle.Render("Note: Node uses dropdown • Use + to add domain mappings/ENV vars, - to remove")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// renderEnvVarsSection renders the environment variables section with +/- buttons
func renderEnvVarsSection(s *state.AppState, zm *zone.Manager) string {
	var section strings.Builder

	section.WriteString("Environment Variables:\n")

	for i, pair := range s.EnvVarPairs {
		// Determine if this pair is focused
		isFocused := s.CurrentFieldIndex == 100 && s.EnvVarFocusedPair == i

		// Render key field
		keyValue := pair.Key
		if isFocused && s.EnvVarFocusedField == 0 {
			cursor := s.CursorPosition
			if cursor < 0 {
				cursor = 0
			}
			if cursor > len(keyValue) {
				cursor = len(keyValue)
			}
			keyValue = keyValue[:cursor] + "_" + keyValue[cursor:]
		}

		// Render value field
		valueDisplay := pair.Value
		if isFocused && s.EnvVarFocusedField == 1 {
			cursor := s.CursorPosition
			if cursor < 0 {
				cursor = 0
			}
			if cursor > len(valueDisplay) {
				cursor = len(valueDisplay)
			}
			valueDisplay = valueDisplay[:cursor] + "_" + valueDisplay[cursor:]
		}

		// Build the row
		prefix := "  "
		if isFocused {
			prefix = "> "
		}

		line := fmt.Sprintf("%s[%d] Key: %-20s Value: %-30s", prefix, i+1, keyValue, valueDisplay)

		// Add +/- buttons
		addBtn := "+ "
		removeBtn := "- "
		if len(s.EnvVarPairs) == 1 && i == 0 {
			removeBtn = "  " // Can't remove the last one
		}

		if zm != nil {
			addZoneID := fmt.Sprintf("env-add:%d", i)
			removeZoneID := fmt.Sprintf("env-remove:%d", i)
			keyZoneID := fmt.Sprintf("env-key:%d", i)
			valueZoneID := fmt.Sprintf("env-value:%d", i)

			section.WriteString(zm.Mark(keyZoneID, prefix+"Key: "))
			section.WriteString(zm.Mark(valueZoneID, fmt.Sprintf("%-20s Value: %-30s ", keyValue, valueDisplay)))
			section.WriteString(zm.Mark(addZoneID, addBtn))
			if len(s.EnvVarPairs) > 1 || i > 0 {
				section.WriteString(zm.Mark(removeZoneID, removeBtn))
			}
			section.WriteString("\n")
		} else {
			section.WriteString(line + " " + addBtn + removeBtn + "\n")
		}
	}

	return section.String()
}

// renderDomainMappingsSection renders the domain mappings section with +/- buttons
func renderDomainMappingsSection(s *state.AppState, zm *zone.Manager) string {
	var section strings.Builder

	section.WriteString("Domain Mappings:\n")

	for i, pair := range s.DomainMappingPairs {
		// Determine if this pair is focused
		isFocused := s.CurrentFieldIndex == 200 && s.DomainMappingFocusedPair == i

		// Render subdomain field
		subdomainValue := pair.Subdomain
		if isFocused && s.DomainMappingFocusedField == 0 {
			cursor := s.CursorPosition
			if cursor < 0 {
				cursor = 0
			}
			if cursor > len(subdomainValue) {
				cursor = len(subdomainValue)
			}
			subdomainValue = subdomainValue[:cursor] + "_" + subdomainValue[cursor:]
		}

		// Render domain field
		domainDisplay := pair.DomainName
		if domainDisplay == "" {
			domainDisplay = "(select)"
		}
		if isFocused && s.DomainMappingFocusedField == 1 {
			domainDisplay = domainDisplay + " ▼"
		}

		// Render port field
		portValue := pair.Port
		if isFocused && s.DomainMappingFocusedField == 2 {
			cursor := s.CursorPosition
			if cursor < 0 {
				cursor = 0
			}
			if cursor > len(portValue) {
				cursor = len(portValue)
			}
			portValue = portValue[:cursor] + "_" + portValue[cursor:]
		}

		// Build the row
		prefix := "  "
		if isFocused {
			prefix = "> "
		}

		line := fmt.Sprintf("%s[%d] Subdomain: %-15s Domain: %-25s Port (container:host): %-6s",
			prefix, i+1, subdomainValue, domainDisplay, portValue)

		// Add +/- buttons
		addBtn := "+ "
		removeBtn := "- "
		if len(s.DomainMappingPairs) == 1 && i == 0 {
			removeBtn = "  " // Can't remove the last one
		}

		if zm != nil {
			addZoneID := fmt.Sprintf("domain-add:%d", i)
			removeZoneID := fmt.Sprintf("domain-remove:%d", i)
			subdomainZoneID := fmt.Sprintf("domain-subdomain:%d", i)
			domainZoneID := fmt.Sprintf("domain-domain:%d", i)
			portZoneID := fmt.Sprintf("domain-port:%d", i)

			section.WriteString(zm.Mark(subdomainZoneID, prefix+fmt.Sprintf("[%d] Subdomain: %-15s ", i+1, subdomainValue)))
			section.WriteString(zm.Mark(domainZoneID, fmt.Sprintf("Domain: %-25s ", domainDisplay)))
			section.WriteString(zm.Mark(portZoneID, fmt.Sprintf("Port (container:host): %-6s ", portValue)))
			section.WriteString(zm.Mark(addZoneID, addBtn))
			if len(s.DomainMappingPairs) > 1 || i > 0 {
				section.WriteString(zm.Mark(removeZoneID, removeBtn))
			}
			section.WriteString("\n")
		} else {
			section.WriteString(line + " " + addBtn + removeBtn + "\n")
		}

		// Show domain dropdown if focused on domain field
		if isFocused && s.DomainMappingFocusedField == 1 && s.DropdownOpen {
			dropdownOptions := renderDropdownOptions(s, s.Domains, s.DropdownIndex, func(d models.Domain) string {
				return d.Name
			})
			section.WriteString(dropdownOptions + "\n")
		}
	}

	return section.String()
}

// renderDropdownOptions renders a dropdown list of options
func renderDropdownOptions[T any](_ *state.AppState, items []T, selectedIndex int, getName func(T) string) string {
	if len(items) == 0 {
		return "     (No options available)"
	}

	var options strings.Builder
	options.WriteString("     ┌─────────────────────────────────┐\n")

	maxVisible := 5 // Show up to 5 options at a time
	start := 0
	end := len(items)

	// Calculate scroll window if too many items
	if len(items) > maxVisible {
		// Keep selected item in middle of window
		start = selectedIndex - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(items) {
			end = len(items)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		name := getName(items[i])
		if i == selectedIndex {
			options.WriteString(fmt.Sprintf("     │ ▶ %-29s │\n", truncate(name, 29)))
		} else {
			options.WriteString(fmt.Sprintf("     │   %-29s │\n", truncate(name, 29)))
		}
	}

	options.WriteString("     └─────────────────────────────────┘")

	// Show scroll indicator if needed
	if len(items) > maxVisible {
		options.WriteString(fmt.Sprintf(" (%d/%d)", selectedIndex+1, len(items)))
	}

	return options.String()
}

// renderSiteSidebar renders a sidebar showing relationships for the selected site
func renderSiteSidebar(s *state.AppState, site *models.Site) string {
	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(35)

	title := lipgloss.NewStyle().Bold(true).Render("🔗 Relationships")

	// Get all domain mappings
	mappings := site.GetDomainMappings()
	var domainInfo string

	if len(mappings) == 0 {
		domainInfo = "🌍 Domains: None"
	} else if len(mappings) == 1 {
		// Single domain - show detailed info
		mapping := mappings[0]
		domainName := "Unknown"
		provider := "Unknown"
		for _, d := range s.Domains {
			if d.ID == mapping.DomainID {
				domainName = d.Name
				provider = d.ProviderName()
				break
			}
		}
		domainInfo = fmt.Sprintf("🌍 Domain: %s\n   Provider: %s\n   Port: %d",
			domainName, provider, mapping.Port)
	} else {
		// Multiple domains - list them all
		domainInfo = fmt.Sprintf("🌍 Domains (%d):\n", len(mappings))
		for i, mapping := range mappings {
			domainName := mapping.DomainID.String()[:8] + "..."
			for _, d := range s.Domains {
				if d.ID == mapping.DomainID {
					domainName = d.Name
					break
				}
			}
			domainInfo += fmt.Sprintf("   %d. %s:%d\n", i+1, domainName, mapping.Port)
		}
		domainInfo = strings.TrimSuffix(domainInfo, "\n")
	}

	// Find node
	var nodeInfo string
	for _, n := range s.Nodes {
		if n.ID == site.NodeID {
			nodeInfo = fmt.Sprintf("🖥️  Node: %s\n   IP: %s",
				n.Name, n.IPAddress.String())
			break
		}
	}
	if nodeInfo == "" {
		nodeInfo = "🖥️  Node: Not found"
	}

	content := domainInfo + "\n\n" + nodeInfo
	return sidebarStyle.Render(title + "\n\n" + content)
}
