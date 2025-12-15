package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
)

var _ = table.Model{} // Suppress unused import

// RenderNodesList renders the nodes list screen with table
func RenderNodesList(s *state.AppState) string {
	// TODO: Implement nodes table using bubbles.table
	// Columns: Name | IP Address | Status | Docker | Traefik | Last Check
	//
	// Example structure:
	// columns := []table.Column{
	//     {Title: "Name", Width: 20},
	//     {Title: "IP Address", Width: 16},
	//     {Title: "Status", Width: 10},
	//     {Title: "Docker", Width: 12},
	//     {Title: "Traefik", Width: 12},
	//     {Title: "Last Check", Width: 20},
	// }
	//
	// rows := []table.Row{}
	// for _, node := range s.Nodes {
	//     dockerInfo := "N/A"
	//     if node.DockerInfo != nil {
	//         dockerInfo = fmt.Sprintf("v%s (%d)", node.DockerInfo.Version, node.DockerInfo.ContainersRunning)
	//     }
	//
	//     traefikInfo := "N/A"
	//     if node.TraefikInfo != nil {
	//         traefikInfo = fmt.Sprintf("v%s", node.TraefikInfo.Version)
	//     }
	//
	//     lastCheck := "Never"
	//     if node.LastHealthCheck != nil {
	//         lastCheck = node.LastHealthCheck.Format("2006-01-02 15:04")
	//     }
	//
	//     rows = append(rows, table.Row{
	//         node.Name,
	//         node.IPAddress.String(),
	//         string(node.Status),
	//         dockerInfo,
	//         traefikInfo,
	//         lastCheck,
	//     })
	// }
	//
	// Color-code status: online=green, offline=red, degraded=yellow

	title := titleStyle.Render("🖥️  Nodes")

	var content string
	if len(s.Nodes) == 0 {
		content = helpStyle.Render("No nodes yet. Press 'n' to create your first node.")
	} else {
		content = fmt.Sprintf("Total Nodes: %d\n\n", len(s.Nodes))
		for i, node := range s.Nodes {
			content += fmt.Sprintf("%d. %s (%s - %s)\n", i+1, node.Name, node.IPAddress.String(), node.APIEndpoint)
		}
	}

	help := helpStyle.Render("\nPress n to create • Esc to go back")

	return title + "\n\n" + content + "\n" + help
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

	help := helpStyle.Render("\nPress h to refresh health check • Esc to go back")

	return title + "\n\n" + content + "\n" + help
}
