package screens

import (
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui"
)

// RenderDashboard renders the main dashboard with 3-column layout
func RenderDashboard(s *state.AppState) string {
	// TODO: Implement 3-column dashboard layout
	// Columns: Sites Summary | Nodes Summary | Domains Summary

	// Example structure:
	// leftColumn := renderSitesSummary(s)
	// middleColumn := renderNodesSummary(s)
	// rightColumn := renderDomainsSummary(s)
	//
	// // Join columns horizontally
	// columns := lipgloss.JoinHorizontal(
	//     lipgloss.Top,
	//     leftColumn,
	//     middleColumn,
	//     rightColumn,
	// )
	//
	// return columns

	return ui.TitleStyle.Render("Dashboard") + "\n\n" +
		"TODO: Implement dashboard with 3 columns:\n" +
		"  - Sites summary (total, running, stopped)\n" +
		"  - Nodes summary (total, online, offline)\n" +
		"  - Domains summary (total, with/without DNS)\n\n" +
		ui.HelpStyle.Render("Press 1/s for Sites, 2/d for Domains, 3/n for Nodes")
}

// renderSitesSummary renders the sites summary box
func renderSitesSummary(s *state.AppState) string {
	// TODO: Implement sites summary
	// Count sites by status
	// total := len(s.Sites)
	// running := 0
	// stopped := 0
	// for _, site := range s.Sites {
	//     if site.Status == models.SiteStatusRunning {
	//         running++
	//     } else if site.Status == models.SiteStatusStopped {
	//         stopped++
	//     }
	// }
	//
	// content := fmt.Sprintf(
	//     "Total: %d\nRunning: %d\nStopped: %d",
	//     total, running, stopped,
	// )
	//
	// return ui.RenderBox("Sites", content)

	return ui.RenderBox("Sites", "TODO")
}

// renderNodesSummary renders the nodes summary box
func renderNodesSummary(s *state.AppState) string {
	// TODO: Implement nodes summary
	// Count nodes by status (online, offline, degraded)

	return ui.RenderBox("Nodes", "TODO")
}

// renderDomainsSummary renders the domains summary box
func renderDomainsSummary(s *state.AppState) string {
	// TODO: Implement domains summary
	// Count total domains, manual DNS vs API-managed

	return ui.RenderBox("Domains", "TODO")
}
