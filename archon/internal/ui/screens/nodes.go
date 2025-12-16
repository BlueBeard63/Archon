package screens

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
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
		// 1. Build table rows (data only, NO buttons)
		var rows []table.Row
		for _, node := range s.Nodes {
			rows = append(rows, table.Row{
				truncateNode(node.Name, 20),
				node.IPAddress.String(),
				truncateNode(node.APIEndpoint, 28),
				string(node.Status),
			})
		}

		// 2. Initialize/update table
		if s.NodesTable == nil {
			columns := []table.Column{
				{Title: "Name", Width: 20},
				{Title: "IP Address", Width: 15},
				{Title: "API Endpoint", Width: 28},
				{Title: "Status", Width: 10},
			}
			s.NodesTable = components.NewTableComponent(columns, rows)
			s.NodesTable.SetCursor(s.NodesListIndex)
		} else {
			s.NodesTable.SetRows(rows)
			s.NodesTable.SetCursor(s.NodesListIndex)
		}

		// 3. Render table view
		tableView := s.NodesTable.View()

		// 4. Build action buttons column (aligned with rows)
		var actionsColumn strings.Builder
		actionsColumn.WriteString("\n\n") // Header padding

		for _, node := range s.Nodes {
			viewBtn := components.Button{
				ID:      "view-node-" + node.ID.String(),
				Label:   "👁️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}
			editBtn := components.Button{
				ID:      "edit-node-" + node.ID.String(),
				Label:   "✏️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}
			deleteBtn := components.Button{
				ID:      "delete-node-" + node.ID.String(),
				Label:   "🗑️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}

			var actionLine string
			if zm != nil {
				actionLine = viewBtn.RenderWithZone(zm) + " " + editBtn.RenderWithZone(zm) + " " + deleteBtn.RenderWithZone(zm)
			} else {
				actionLine = viewBtn.Render() + " " + editBtn.Render() + " " + deleteBtn.Render()
			}

			actionsColumn.WriteString(actionLine + "\n")
		}

		// 5. Join table + actions horizontally
		mainContent := lipgloss.JoinHorizontal(
			lipgloss.Top,
			tableView,
			actionsColumn.String(),
		)

		// 6. Build sidebar for selected node
		var sidebar string
		if len(s.Nodes) > 0 && s.NodesListIndex >= 0 && s.NodesListIndex < len(s.Nodes) {
			node := &s.Nodes[s.NodesListIndex]
			sidebar = renderNodeSidebar(s, node)
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

	help := helpStyle.Render("\n\nPress j/k or arrows to navigate • e to edit • d to delete • enter to view • n to create • Esc to go back")

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

// generateAPIKey generates a random 32-character API key
func generateAPIKey() string {
	bytes := make([]byte, 24) // 24 bytes = 32 base64 characters
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to less secure but working key
		return fmt.Sprintf("%032x", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// RenderNodeCreate renders the node creation form
func RenderNodeCreate(s *state.AppState) string {
	// Initialize form if needed (2 editable fields + 1 generated field)
	if len(s.FormFields) != 3 {
		s.FormFields = []string{"", "", generateAPIKey()}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Node")

	labels := []string{"Name:", "API Endpoint:", "API Key (auto-generated):"}

	// Render each field
	var fields string
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value

		// Show cursor if focused (but not for API key field)
		if i == s.CurrentFieldIndex && i < 2 {
			displayValue = value + "_"
			label = "> " + label // Show arrow for focused field
		} else {
			label = "  " + label
		}

		// Show API key as read-only
		if i == 2 {
			displayValue = lipgloss.NewStyle().Faint(true).Render(value)
		}

		fields += label + " " + displayValue + "\n"
	}

	help := helpStyle.Render("\nTab to navigate, Enter to create, Esc to cancel")
	note := helpStyle.Render("Note: IP address will be determined from API endpoint")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// RenderNodeCreateWithZones renders the node creation form with clickable fields
func RenderNodeCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Fall back to regular rendering if no zone manager
	if zm == nil {
		return RenderNodeCreate(s)
	}

	// Initialize form if needed (2 editable fields + 1 generated field)
	if len(s.FormFields) != 3 {
		s.FormFields = []string{"", "", generateAPIKey()}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Node")

	labels := []string{"Name:", "API Endpoint:", "API Key (auto-generated):"}

	// Render each field with zones
	var fields string
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value

		// Show cursor if focused (but not for API key field)
		if i == s.CurrentFieldIndex && i < 2 {
			displayValue = value + "_"
			label = "> " + label // Show arrow for focused field
		} else {
			label = "  " + label
		}

		// Show API key as read-only
		if i == 2 {
			displayValue = lipgloss.NewStyle().Faint(true).Render(value)
		}

		// Wrap the entire field line in a clickable zone (only for editable fields)
		fieldLine := label + " " + displayValue + "\n"
		if i < 2 {
			fields += zm.Mark(fmt.Sprintf("field:%d", i), fieldLine)
		} else {
			fields += fieldLine
		}
	}

	help := helpStyle.Render("\nTab to navigate, Enter to create, Esc to cancel")
	note := helpStyle.Render("Note: IP address will be determined from API endpoint")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
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

// renderNodeSidebar renders a sidebar showing sites deployed on the selected node
func renderNodeSidebar(s *state.AppState, node *models.Node) string {
	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(35)

	title := lipgloss.NewStyle().Bold(true).Render("📦 Deployed Sites")

	// Find sites on this node
	var deployedSites []string
	for _, site := range s.Sites {
		if site.NodeID == node.ID {
			deployedSites = append(deployedSites,
				fmt.Sprintf("• %s (%s)", site.Name, site.Status))
		}
	}

	var content string
	if len(deployedSites) == 0 {
		content = lipgloss.NewStyle().Faint(true).
			Render("No sites deployed")
	} else {
		content = strings.Join(deployedSites, "\n")
	}

	return sidebarStyle.Render(title + "\n\n" + content)
}
