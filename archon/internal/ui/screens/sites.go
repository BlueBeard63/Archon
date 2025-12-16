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

			rows = append(rows, table.Row{
				truncate(site.Name, 25),
				truncate(domainDisplay, 18),
				truncate(nodeName, 25),
				portDisplay,
				string(site.Status),
			})
		}

		// 2. Initialize/update table
		if s.SitesTable == nil {
			columns := []table.Column{
				{Title: "Name", Width: 25},
				{Title: "Domain", Width: 18},
				{Title: "Node", Width: 25},
				{Title: "Port", Width: 6},
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
			deployBtn := components.Button{
				ID:      "deploy-site-" + site.ID.String(),
				Label:   "🚀",
				Primary: false,
				Border:  false,
				Icon:    true,
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

			var actionLine string
			if zm != nil {
				actionLine = deployBtn.RenderWithZone(zm) + " " + editBtn.RenderWithZone(zm) + " " + deleteBtn.RenderWithZone(zm)
			} else {
				actionLine = deployBtn.Render() + " " + editBtn.Render() + " " + deleteBtn.Render()
			}

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

	help := helpStyle.Render("\n\nPress j/k or arrows to navigate • Space/Enter to deploy • e to edit • d to delete • n to create • Esc to go back")

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

// truncate truncates a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// RenderSiteCreate renders the site creation form
func RenderSiteCreate(s *state.AppState) string {
	return RenderSiteCreateWithZones(s, nil)
}

// RenderSiteCreateWithZones renders the site creation form with clickable fields
func RenderSiteCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Initialize form if needed (8 fields: basic + SSL email + env vars + config file)
	if len(s.FormFields) != 8 {
		s.FormFields = []string{"", "", "", "", "8080", "", "", ""}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Site")

	labels := []string{
		"Name:",
		"Domain:",
		"Node:",
		"Docker Image:",
		"Port:",
		"SSL Email (for Let's Encrypt):",
		"Env Vars (KEY=VALUE, one per line):",
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

		// Show dropdown options for Domain (index 1) and Node (index 2) when focused
		if i == s.CurrentFieldIndex && i == 1 && s.DropdownOpen {
			// Domain dropdown
			dropdownOptions := renderDropdownOptions(s, s.Domains, s.DropdownIndex, func(d models.Domain) string {
				return d.Name
			})
			fields += dropdownOptions + "\n"
		} else if i == s.CurrentFieldIndex && i == 2 && s.DropdownOpen {
			// Node dropdown
			dropdownOptions := renderDropdownOptions(s, s.Nodes, s.DropdownIndex, func(n models.Node) string {
				return n.Name
			})
			fields += dropdownOptions + "\n"
		}
	}

	helpText := "\nTab/Shift+Tab to navigate, Enter to create, Esc to cancel"
	if s.CurrentFieldIndex == 1 || s.CurrentFieldIndex == 2 {
		// On dropdown fields
		if s.DropdownOpen {
			helpText = "\nUp/Down to select, Enter/Tab to confirm, Esc to cancel"
		} else {
			helpText = "\nPress Enter or Down to open dropdown, Tab to skip"
		}
	} else if s.CurrentFieldIndex == 5 {
		helpText = "\nEmail for Let's Encrypt SSL certificate notifications (e.g., admin@example.com)"
	} else if s.CurrentFieldIndex == 6 {
		helpText = "\nEnter env vars as KEY=VALUE (use | to separate multiple vars)"
	} else if s.CurrentFieldIndex == 7 {
		helpText = "\nEnter full path to config file (will be loaded when site is created)"
	}

	help := helpStyle.Render(helpText)
	note := helpStyle.Render("Note: Domain/Node use dropdowns • Env vars and config file are optional")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// renderDropdownOptions renders a dropdown list of options
func renderDropdownOptions[T any](s *state.AppState, items []T, selectedIndex int, getName func(T) string) string {
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

// boolToYesNo converts boolean to "Yes"/"No" string
func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
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
