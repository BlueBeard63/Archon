package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/state"
)

// Reuse styles from dashboard
var _ = table.Model{} // Suppress unused import

// RenderSitesList renders the sites list screen with table
func RenderSitesList(s *state.AppState) string {
	// TODO: Implement sites table using bubbles.table
	// Columns: Name | Domain | Node | Status | Port | SSL
	//
	// Example structure:
	// columns := []table.Column{
	//     {Title: "Name", Width: 20},
	//     {Title: "Domain", Width: 25},
	//     {Title: "Node", Width: 15},
	//     {Title: "Status", Width: 10},
	//     {Title: "Port", Width: 6},
	//     {Title: "SSL", Width: 5},
	// }
	//
	// rows := []table.Row{}
	// for _, site := range s.Sites {
	//     domain := s.GetDomainByID(site.DomainID)
	//     node := s.GetNodeByID(site.NodeID)
	//     rows = append(rows, table.Row{
	//         site.Name,
	//         domain.Name,
	//         node.Name,
	//         string(site.Status),
	//         fmt.Sprintf("%d", site.Port),
	//         boolToYesNo(site.SSLEnabled),
	//     })
	// }
	//
	// t := table.New(
	//     table.WithColumns(columns),
	//     table.WithRows(rows),
	//     table.WithFocused(true),
	// )
	//
	// // Apply styles
	// s := table.DefaultStyles()
	// s.Header = ui.TableHeaderStyle
	// s.Selected = ui.TableRowSelectedStyle
	// t.SetStyles(s)
	//
	// return ui.TitleStyle.Render("Sites") + "\n" + t.View()

	title := titleStyle.Render("🌐 Sites")

	var content string
	if len(s.Sites) == 0 {
		content = helpStyle.Render("No sites yet. Press 'n' to create your first site.")
	} else {
		content = fmt.Sprintf("Total Sites: %d\n\n", len(s.Sites))
		for i, site := range s.Sites {
			content += fmt.Sprintf("%d. %s (Port: %d, Image: %s)\n", i+1, site.Name, site.Port, site.DockerImage)
		}
	}

	help := helpStyle.Render("\nPress n to create • Esc to go back")

	return title + "\n\n" + content + "\n" + help
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
