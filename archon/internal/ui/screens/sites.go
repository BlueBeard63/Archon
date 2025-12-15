package screens

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

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
		content = fmt.Sprintf("Total Sites: %d\n\n", len(s.Sites))

		// Manual table with row action buttons
		headerStyle := lipgloss.NewStyle().Bold(true).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("240"))
		rowStyle := lipgloss.NewStyle().PaddingRight(2)

		header := headerStyle.Render(fmt.Sprintf("%-25s %-18s %-25s %-6s %-10s %-45s", "Name", "Domain", "Node", "Port", "Status", "Actions"))
		content += header + "\n"

		for _, site := range s.Sites {
			// Get domain and node names (would need lookups in real implementation)
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

			// Create row action buttons (icon only)
			editBtn := components.Button{ID: "edit-site-" + site.ID.String(), Label: "✏️", Primary: false, Border: false, Icon: true}
			deleteBtn := components.Button{ID: "delete-site-" + site.ID.String(), Label: "🗑️", Primary: false, Border: false, Icon: true}

			var editBtnStr, deleteBtnStr string
			if zm != nil {
				editBtnStr = editBtn.RenderWithZone(zm)
				deleteBtnStr = deleteBtn.RenderWithZone(zm)
			} else {
				editBtnStr = editBtn.Render()
				deleteBtnStr = deleteBtn.Render()
			}

			actions := editBtnStr + " " + deleteBtnStr

			rowText := fmt.Sprintf("%-25s %-18s %-25s %-6d %-10s %-45s",
				truncate(site.Name, 25),
				truncate(domainName, 18),
				truncate(nodeName, 25),
				site.Port,
				site.Status,
				actions,
			)

			content += rowStyle.Render(rowText) + "\n"
		}
	}

	help := helpStyle.Render("\n\nPress n to create • Esc to go back")

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
