package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
)

// RenderQuickConfig renders the quick configure screen for sharing node configs via dpaste.org
func RenderQuickConfig(s *state.AppState) string {
	// Find the selected node
	var node *models.Node
	for i := range s.Nodes {
		if s.Nodes[i].ID == s.SelectedNodeID {
			node = &s.Nodes[i]
			break
		}
	}

	if node == nil {
		return titleStyle.Render("Quick Configure") + "\n\n" +
			"Node not found\n\n" +
			helpStyle.Render("Press Esc to go back")
	}

	title := titleStyle.Render(fmt.Sprintf("Quick Configure: %s", node.Name))

	// Box style for URL/command display
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(60)

	var content strings.Builder

	if s.QuickConfigURL == "" {
		// Config not yet uploaded
		content.WriteString("Config will be uploaded to dpaste.org for sharing.\n\n")
		content.WriteString("The new node can fetch this config using:\n")
		content.WriteString(boxStyle.Render("archon-node configure --from-url <URL>"))
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("Press Enter to upload config"))
	} else {
		// Config uploaded - show URL
		content.WriteString("Config uploaded to dpaste.org\n\n")
		content.WriteString("On your new server, run:\n")
		content.WriteString(boxStyle.Render(fmt.Sprintf("archon-node configure --from-url \\\n  %s", s.QuickConfigURL)))
		content.WriteString("\n\n")

		content.WriteString("Then start the node:\n")
		content.WriteString(boxStyle.Render("archon-node"))
		content.WriteString("\n\n")

		// Status
		statusIcon := "..."
		statusText := "Waiting for node health check..."
		statusColor := lipgloss.Color("243")

		if s.QuickConfigHealthConfirmed {
			statusIcon = "OK"
			statusText = "Node health check passed!"
			statusColor = lipgloss.Color("42")
		}

		statusStyle := lipgloss.NewStyle().Foreground(statusColor)
		content.WriteString(fmt.Sprintf("Status: %s %s\n", statusStyle.Render(statusIcon), statusText))

		// Expiration
		if s.QuickConfigExpiresAt != "" {
			content.WriteString(fmt.Sprintf("Paste expires: %s\n", s.QuickConfigExpiresAt))
		}
	}

	// Different help text depending on state
	var help string
	if s.QuickConfigURL == "" {
		help = helpStyle.Render("\n\n[Enter] Upload config  [Esc] Cancel")
	} else {
		help = helpStyle.Render("\n\n[r] Refresh status  [Enter] Done  [Esc] Cancel")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		content.String(),
		help,
	)
}
