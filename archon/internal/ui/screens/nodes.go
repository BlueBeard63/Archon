package screens

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui/components"
)

// RenderNodesList renders the nodes list screen
func RenderNodesList(s *state.AppState) string {
	return RenderNodesListWithZones(s, nil)
}

// RenderNodesListWithZones renders nodes list with button zones
func RenderNodesListWithZones(s *state.AppState, zm *zone.Manager) string {
	title := titleStyle.Render("🖥️  Nodes")

	// Create button group
	buttonGroup := &components.ButtonGroup{
		Buttons: []components.Button{
			{ID: "create-node", Label: "➕ Create Node", Primary: true},
		},
	}

	var buttons string
	if zm != nil {
		buttons = buttonGroup.RenderWithZones(zm)
	} else {
		buttons = buttonGroup.Render()
	}

	var content string
	if len(s.Nodes) == 0 {
		content = helpStyle.Render("No nodes yet. Click 'Create Node' or press 'n'.")
	} else {
		content = fmt.Sprintf("Total Nodes: %d\n\n", len(s.Nodes))

		// Manual table with row action buttons
		headerStyle := lipgloss.NewStyle().Bold(true).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("240"))
		rowStyle := lipgloss.NewStyle().PaddingRight(2)

		header := headerStyle.Render(fmt.Sprintf("%-20s %-15s %-28s %-10s %-45s", "Name", "IP Address", "API Endpoint", "Status", "Actions"))
		content += header + "\n"

		for _, node := range s.Nodes {
			// Create row action buttons (icon only)
			viewBtn := components.Button{ID: "view-node-" + node.ID.String(), Label: "👁️", Primary: false, Border: false, Icon: true}
			editBtn := components.Button{ID: "edit-node-" + node.ID.String(), Label: "✏️", Primary: false, Border: false, Icon: true}
			deleteBtn := components.Button{ID: "delete-node-" + node.ID.String(), Label: "🗑️", Primary: false, Border: false, Icon: true}

			var viewBtnStr, editBtnStr, deleteBtnStr string
			if zm != nil {
				viewBtnStr = viewBtn.RenderWithZone(zm)
				editBtnStr = editBtn.RenderWithZone(zm)
				deleteBtnStr = deleteBtn.RenderWithZone(zm)
			} else {
				viewBtnStr = viewBtn.Render()
				editBtnStr = editBtn.Render()
				deleteBtnStr = deleteBtn.Render()
			}

			actions := viewBtnStr + " " + editBtnStr + " " + deleteBtnStr

			rowText := fmt.Sprintf("%-20s %-15s %-28s %-10s %-45s",
				truncateNode(node.Name, 20),
				node.IPAddress.String(),
				truncateNode(node.APIEndpoint, 28),
				node.Status,
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

// truncateNode truncates a string to maxLen characters
func truncateNode(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// RenderNodeCreate renders the node creation form
func RenderNodeCreate(s *state.AppState) string {
	// Initialize form if needed (4 fields)
	if len(s.FormFields) != 4 {
		s.FormFields = []string{"", "", "", ""}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Node")

	labels := []string{"Name:", "API Endpoint:", "API Key:", "IP Address:"}

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

	return title + "\n\n" + fields + help
}

// RenderNodeCreateWithZones renders the node creation form with clickable fields
func RenderNodeCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Fall back to regular rendering if no zone manager
	if zm == nil {
		return RenderNodeCreate(s)
	}

	// Initialize form if needed (4 fields)
	if len(s.FormFields) != 4 {
		s.FormFields = []string{"", "", "", ""}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Node")

	labels := []string{"Name:", "API Endpoint:", "API Key:", "IP Address:"}

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

	return title + "\n\n" + fields + help
}

// RenderNodeDetails renders detailed information about a node
func RenderNodeDetails(s *state.AppState, nodeID string) string {
	title := titleStyle.Render("Node Details")

	// Find the node
	var node *models.Node
	for i := range s.Nodes {
		if s.Nodes[i].ID.String() == nodeID {
			node = &s.Nodes[i]
			break
		}
	}

	if node == nil {
		return title + "\n\n" + "Node not found\n\n" + helpStyle.Render("Press Esc to go back")
	}

	// Build node info section
	content := fmt.Sprintf("Name: %s\n", node.Name)
	content += fmt.Sprintf("Endpoint: %s\n", node.APIEndpoint)
	content += fmt.Sprintf("IP Address: %s\n", node.IPAddress.String())
	content += fmt.Sprintf("Status: %s\n\n", node.Status)

	// Docker info section
	content += "Docker Information:\n"
	if node.DockerInfo != nil {
		content += fmt.Sprintf("  Version: %s\n", node.DockerInfo.Version)
		content += fmt.Sprintf("  Containers Running: %d\n", node.DockerInfo.ContainersRunning)
		content += fmt.Sprintf("  Images: %d\n\n", node.DockerInfo.ImagesCount)
	} else {
		content += "  No Docker info available\n\n"
	}

	// Traefik info section
	content += "Proxy Information:\n"
	if node.TraefikInfo != nil {
		content += fmt.Sprintf("  Version: %s\n", node.TraefikInfo.Version)
		content += fmt.Sprintf("  Routers: %d\n", node.TraefikInfo.RoutersCount)
		content += fmt.Sprintf("  Services: %d\n\n", node.TraefikInfo.ServicesCount)
	} else {
		content += "  No proxy info available\n\n"
	}

	// Find sites deployed on this node
	content += "Deployed Sites:\n"
	sitesFound := false
	for _, site := range s.Sites {
		if site.NodeID == node.ID {
			content += fmt.Sprintf("  - %s (%s)\n", site.Name, site.DockerImage)
			sitesFound = true
		}
	}
	if !sitesFound {
		content += "  No sites deployed\n"
	}

	// Last health check
	if node.LastHealthCheck != nil {
		content += fmt.Sprintf("\nLast Health Check: %s\n", node.LastHealthCheck.Format("2006-01-02 15:04:05"))
	}

	help := helpStyle.Render("\nPress c to view config • h to refresh health check • Esc to go back")

	return title + "\n\n" + content + "\n" + help
}

// RenderNodeConfig renders the TOML configuration for a node
func RenderNodeConfig(s *state.AppState) string {
	title := titleStyle.Render("📄 Node Configuration")

	// Find the node by selected ID
	var node *models.Node
	for i := range s.Nodes {
		if s.Nodes[i].ID == s.SelectedNodeID {
			node = &s.Nodes[i]
			break
		}
	}

	if node == nil {
		return title + "\n\n" + "Node not found\n\n" + helpStyle.Render("Press Esc to go back")
	}

	// Generate the TOML config
	configContent := node.GenerateNodeConfigTOML()

	// Show instructions at the top
	instructions := helpStyle.Render(
		"Copy this configuration to /etc/archon/node-config.toml on your server (" + node.IPAddress.String() + ")",
	)

	help := helpStyle.Render("\nPress Esc to go back • Press s to save to file")

	return title + "\n\n" + instructions + "\n\n" + configContent + "\n" + help
}
