package screens

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
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
				truncateNode(node.IPAddress.String(), 20),
				truncateNode(node.APIEndpoint, 28),
				string(node.Status),
			})
		}

		// 2. Initialize/update table
		if s.NodesTable == nil {
			columns := []table.Column{
				{Title: "Name", Width: 20},
				{Title: "IP Address", Width: 20},
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
	// Initialize form if needed (3 editable fields + 1 generated field: Name, Endpoint, Proxy, APIKey)
	if len(s.FormFields) != 4 {
		s.FormFields = []string{"", "", "nginx", generateAPIKey()}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Node")

	labels := []string{"Name:", "API Endpoint:", "Reverse Proxy:", "API Key (auto-generated):"}

	// Render each field
	var fields string
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value

		// Show cursor if focused (but not for proxy or API key field)
		if i == s.CurrentFieldIndex && i < 2 {
			displayValue = value + "_"
			label = "> " + label // Show arrow for focused field
		} else if i == s.CurrentFieldIndex && i == 2 {
			// Proxy field is focused but no cursor (uses dropdown)
			label = "> " + label
		} else {
			label = "  " + label
		}

		// Show API key as read-only
		if i == 3 {
			displayValue = lipgloss.NewStyle().Faint(true).Render(value)
		}

		fields += label + " " + displayValue + "\n"

		// Show dropdown options for Proxy field (index 2) when focused
		if i == s.CurrentFieldIndex && i == 2 && s.DropdownOpen {
			proxies := []string{"nginx", "apache", "traefik"}
			dropdownOptions := renderProxyDropdown(proxies, s.DropdownIndex)
			fields += dropdownOptions + "\n"
		}
	}

	helpText := "\nTab to navigate, Enter to create, Esc to cancel"
	if s.CurrentFieldIndex == 2 {
		// On proxy field
		if s.DropdownOpen {
			helpText = "\nUp/Down to select, Enter/Tab to confirm, Esc to cancel"
		} else {
			helpText = "\nPress Enter or Down to open proxy dropdown"
		}
	}

	help := helpStyle.Render(helpText)
	note := helpStyle.Render("Note: IP address will be determined from API endpoint")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// renderProxyDropdown renders a dropdown list of proxy options
func renderProxyDropdown(proxies []string, selectedIndex int) string {
	var options strings.Builder
	options.WriteString("     ┌─────────────────────────────────┐\n")

	proxyLabels := map[string]string{
		"nginx":   "Nginx",
		"apache":  "Apache2",
		"traefik": "Traefik",
	}

	for i, proxy := range proxies {
		label := proxyLabels[proxy]
		if label == "" {
			label = proxy
		}

		if i == selectedIndex {
			options.WriteString(fmt.Sprintf("     │ ▶ %-29s │\n", label))
		} else {
			options.WriteString(fmt.Sprintf("     │   %-29s │\n", label))
		}
	}

	options.WriteString("     └─────────────────────────────────┘")
	return options.String()
}

// RenderNodeCreateWithZones renders the node creation form with clickable fields
func RenderNodeCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Fall back to regular rendering if no zone manager
	if zm == nil {
		return RenderNodeCreate(s)
	}

	// Initialize form if needed (3 editable fields + 1 generated field: Name, Endpoint, Proxy, APIKey)
	if len(s.FormFields) != 4 {
		s.FormFields = []string{"", "", "nginx", generateAPIKey()}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Node")

	labels := []string{"Name:", "API Endpoint:", "Reverse Proxy:", "API Key (auto-generated):"}

	// Render each field with zones
	var fields string
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value

		// Show cursor if focused (but not for proxy or API key field)
		if i == s.CurrentFieldIndex && i < 2 {
			displayValue = value + "_"
			label = "> " + label // Show arrow for focused field
		} else if i == s.CurrentFieldIndex && i == 2 {
			// Proxy field is focused but no cursor (uses dropdown)
			label = "> " + label
		} else {
			label = "  " + label
		}

		// Show API key as read-only
		if i == 3 {
			displayValue = lipgloss.NewStyle().Faint(true).Render(value)
		}

		// Wrap the entire field line in a clickable zone (only for editable fields)
		fieldLine := label + " " + displayValue + "\n"
		if i < 3 {
			fields += zm.Mark(fmt.Sprintf("field:%d", i), fieldLine)
		} else {
			fields += fieldLine
		}

		// Show dropdown options for Proxy field (index 2) when focused
		if i == s.CurrentFieldIndex && i == 2 && s.DropdownOpen {
			proxies := []string{"nginx", "apache", "traefik"}
			dropdownOptions := renderProxyDropdown(proxies, s.DropdownIndex)
			fields += dropdownOptions + "\n"
		}
	}

	helpText := "\nTab to navigate, Enter to create, Esc to cancel"
	if s.CurrentFieldIndex == 2 {
		// On proxy field
		if s.DropdownOpen {
			helpText = "\nUp/Down to select, Enter/Tab to confirm, Esc to cancel"
		} else {
			helpText = "\nPress Enter or Down to open proxy dropdown"
		}
	}

	help := helpStyle.Render(helpText)
	note := helpStyle.Render("Note: IP address will be determined from API endpoint")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// RenderNodeEdit renders the node edit form
func RenderNodeEdit(s *state.AppState) string {
	return RenderNodeEditWithZones(s, nil)
}

// RenderNodeEditWithZones renders the node edit form with clickable fields
func RenderNodeEditWithZones(s *state.AppState, zm *zone.Manager) string {
	// Find the node
	var node *models.Node
	for i := range s.Nodes {
		if s.Nodes[i].ID == s.SelectedNodeID {
			node = &s.Nodes[i]
			break
		}
	}

	if node == nil {
		return titleStyle.Render("Edit Node") + "\n\n" + "Node not found\n\n" + helpStyle.Render("Press Esc to go back")
	}

	// Initialize form if needed (3 editable fields: Name, Endpoint, Proxy)
	if len(s.FormFields) != 3 {
		s.FormFields = []string{node.Name, node.APIEndpoint, string(node.ProxyType)}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Edit Node: " + node.Name)

	labels := []string{"Name:", "API Endpoint:", "Reverse Proxy:"}

	// Render each field
	var fields string
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value

		// Show cursor if focused (but not for proxy field)
		if i == s.CurrentFieldIndex && i < 2 {
			displayValue = value + "_"
			label = "> " + label
		} else if i == s.CurrentFieldIndex && i == 2 {
			// Proxy field is focused but no cursor (uses dropdown)
			label = "> " + label
		} else {
			label = "  " + label
		}

		// Wrap the field line in a clickable zone
		fieldLine := label + " " + displayValue + "\n"
		if zm != nil {
			fields += zm.Mark(fmt.Sprintf("field:%d", i), fieldLine)
		} else {
			fields += fieldLine
		}

		// Show dropdown options for Proxy field (index 2) when focused
		if i == s.CurrentFieldIndex && i == 2 && s.DropdownOpen {
			proxies := []string{"nginx", "apache", "traefik"}
			dropdownOptions := renderProxyDropdown(proxies, s.DropdownIndex)
			fields += dropdownOptions + "\n"
		}
	}

	helpText := "\nTab/Shift+Tab to navigate, Enter to save, Esc to cancel"
	if s.CurrentFieldIndex == 2 {
		// On proxy field
		if s.DropdownOpen {
			helpText = "\nUp/Down to select, Enter/Tab to confirm, Esc to cancel"
		} else {
			helpText = "\nPress Enter or Down to open proxy dropdown"
		}
	}

	help := helpStyle.Render(helpText)
	note := helpStyle.Render("Note: Changing proxy will require reconfiguring the node")

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

// RenderNodeConfigSave renders the file save dialog for node config
func RenderNodeConfigSave(s *state.AppState) string {
	return RenderNodeConfigSaveWithZones(s, nil)
}

// RenderNodeConfigSaveWithZones renders the file save dialog with clickable fields
func RenderNodeConfigSaveWithZones(s *state.AppState, zm *zone.Manager) string {
	// Find the node
	var node *models.Node
	for i := range s.Nodes {
		if s.Nodes[i].ID == s.SelectedNodeID {
			node = &s.Nodes[i]
			break
		}
	}

	if node == nil {
		return titleStyle.Render("Save Node Config") + "\n\n" + "Node not found\n\n" + helpStyle.Render("Press Esc to go back")
	}

	// Initialize form if needed (1 field: file path)
	if len(s.FormFields) != 1 {
		// Get home directory and set default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "~"
		}
		filename := fmt.Sprintf("node-config-%s.toml", node.Name)
		defaultPath := filepath.Join(homeDir, filename)
		s.FormFields = []string{defaultPath}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("💾 Save Node Configuration")

	// Render the file path field
	label := "  File Path:"
	if s.CurrentFieldIndex == 0 {
		label = "> File Path:"
	}

	value := s.FormFields[0]
	displayValue := value
	if s.CurrentFieldIndex == 0 {
		// Show cursor at position
		cursor := s.CursorPosition
		if cursor < 0 {
			cursor = 0
		}
		if cursor > len(value) {
			cursor = len(value)
		}
		displayValue = value[:cursor] + "_" + value[cursor:]
	}

	// Wrap in zone for click support
	fieldLine := label + " " + displayValue + "\n"
	var fields string
	if zm != nil {
		fields = zm.Mark("field:0", fieldLine)
	} else {
		fields = fieldLine
	}

	help := helpStyle.Render("\nEnter to save • Esc to cancel")
	note := helpStyle.Render("Note: Use absolute path or ~ for home directory")

	return title + "\n\n" + fields + "\n" + help + "\n" + note
}

// RenderNodeConfig renders the TOML configuration for a node with scrollable viewport
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

	// Build the full content
	fullContent := instructions + "\n\n" + configContent

	// Initialize viewport if needed
	if s.NodeConfigViewport.Width == 0 {
		// Set viewport size based on window dimensions
		// Leave room for title (3 lines), help (2 lines), and some padding
		viewportHeight := s.WindowHeight - 7
		if viewportHeight < 10 {
			viewportHeight = 10
		}

		s.NodeConfigViewport = viewport.New(s.WindowWidth-4, viewportHeight)
		s.NodeConfigViewport.SetContent(fullContent)
	} else {
		// Update content if viewport already exists
		s.NodeConfigViewport.SetContent(fullContent)
	}

	help := helpStyle.Render("\n↑/↓ to scroll • PgUp/PgDn for page • Home/End to jump • Esc to go back • s to save")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		s.NodeConfigViewport.View(),
		help,
	)
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
