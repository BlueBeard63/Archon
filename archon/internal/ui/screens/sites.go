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
			domainName := site.DomainID.String()[:8] + "..."
			nodeName := site.NodeID.String()[:8] + "..."

			// Find actual names if possible
			for _, d := range s.Domains {
				if d.ID == site.DomainID {
					domainName = d.Name
					break
				}
			}
			for _, n := range s.Nodes {
				if n.ID == site.NodeID {
					nodeName = n.Name
					break
				}
			}

			rows = append(rows, table.Row{
				truncate(site.Name, 25),
				truncate(domainName, 18),
				truncate(nodeName, 25),
				fmt.Sprintf("%d", site.Port),
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
				actionLine = editBtn.RenderWithZone(zm) + " " + deleteBtn.RenderWithZone(zm)
			} else {
				actionLine = editBtn.Render() + " " + deleteBtn.Render()
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

	help := helpStyle.Render("\n\nPress j/k or arrows to navigate • e to edit • d to delete • n to create • Esc to go back")

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
	// Initialize form if needed (5 basic fields for now)
	if len(s.FormFields) != 5 {
		s.FormFields = []string{"", "", "", "", "8080"}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Site")

	labels := []string{
		"Name:",
		"Domain (name):",
		"Node (name):",
		"Docker Image:",
		"Port:",
	}

	// Render each field
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

		fields += label + " " + displayValue + "\n"
	}

	help := helpStyle.Render("\nTab/Shift+Tab to navigate, Enter to create, Esc to cancel")
	note := helpStyle.Render("Note: Domain and Node must already exist")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// RenderSiteCreateWithZones renders the site creation form with clickable fields
func RenderSiteCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Fall back to regular rendering if no zone manager
	if zm == nil {
		return RenderSiteCreate(s)
	}

	// Initialize form if needed (5 basic fields for now)
	if len(s.FormFields) != 5 {
		s.FormFields = []string{"", "", "", "", "8080"}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Site")

	labels := []string{
		"Name:",
		"Domain (name):",
		"Node (name):",
		"Docker Image:",
		"Port:",
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
	}

	help := helpStyle.Render("\nTab/Shift+Tab to navigate, Enter to create, Esc to cancel")
	note := helpStyle.Render("Note: Domain and Node must already exist")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
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

	// Find domain
	var domainInfo string
	for _, d := range s.Domains {
		if d.ID == site.DomainID {
			domainInfo = fmt.Sprintf("🌍 Domain: %s\n   Provider: %s",
				d.Name, d.ProviderName())
			break
		}
	}
	if domainInfo == "" {
		domainInfo = "🌍 Domain: Not found"
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
