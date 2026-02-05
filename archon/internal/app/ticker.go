package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"

	"github.com/BlueBeard63/archon/internal/api"
	"github.com/BlueBeard63/archon/internal/models"
)

// DefaultHealthCheckInterval is the default interval between health checks (5 minutes)
const DefaultHealthCheckInterval = 5 * time.Minute

// GetHealthCheckInterval returns the health check interval from config or default
func GetHealthCheckInterval(intervalSecs int) time.Duration {
	if intervalSecs <= 0 {
		return DefaultHealthCheckInterval
	}
	return time.Duration(intervalSecs) * time.Second
}

// StartHealthCheckTicker returns a command that sends TickMsg at the specified interval
func StartHealthCheckTicker(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// PerformNodeHealthChecks returns commands to check health of all nodes
func PerformNodeHealthChecks(nodes []models.Node, nodeClient api.NodeClient) []tea.Cmd {
	cmds := make([]tea.Cmd, 0, len(nodes))
	for _, node := range nodes {
		nodeID := node.ID
		endpoint := node.APIEndpoint
		apiKey := node.APIKey

		cmd := func() tea.Msg {
			result, err := nodeClient.HealthCheck(endpoint, apiKey)
			return NodeHealthCheckResultMsg{
				NodeID: nodeID,
				Result: result,
				Error:  err,
			}
		}
		cmds = append(cmds, cmd)
	}
	return cmds
}

// PerformSiteStatusChecks returns commands to check status of all sites
func PerformSiteStatusChecks(sites []models.Site, nodes []models.Node, nodeClient api.NodeClient) []tea.Cmd {
	cmds := make([]tea.Cmd, 0)

	// Create a map of nodes by ID for quick lookup
	nodeMap := make(map[uuid.UUID]*models.Node)
	for i := range nodes {
		nodeMap[nodes[i].ID] = &nodes[i]
	}

	for _, site := range sites {
		// Skip sites that are currently deploying (in progress)
		if site.Status == models.SiteStatusDeploying {
			continue
		}

		// Get the node for this site
		node, ok := nodeMap[site.NodeID]
		if !ok {
			continue
		}

		siteID := site.ID
		siteName := site.Name
		siteType := site.GetSiteType()
		endpoint := node.APIEndpoint
		apiKey := node.APIKey

		cmd := func() tea.Msg {
			status, err := nodeClient.GetSiteStatus(endpoint, apiKey, siteID, siteName, siteType)
			var statusStr string
			if status != nil {
				// SiteStatus is a string type, so dereference directly
				statusStr = string(*status)
			}
			return SiteStatusCheckResultMsg{
				SiteID: siteID,
				Status: statusStr,
				Error:  err,
			}
		}
		cmds = append(cmds, cmd)
	}
	return cmds
}
