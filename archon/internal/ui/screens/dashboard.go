package screens

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/BlueBeard63/archon/internal/state"
)

// Inline styles to avoid circular import
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle()

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(1, 2)
)

// RenderDashboard renders the main dashboard with 3-column layout
func RenderDashboard(s *state.AppState) string {
	title := titleStyle.Render("📊 Dashboard")

	// Render summaries
	leftColumn := renderSitesSummary(s)
	middleColumn := renderNodesSummary(s)
	rightColumn := renderDomainsSummary(s)

	// Join columns horizontally with spacing
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumn,
		"  ",
		middleColumn,
		"  ",
		rightColumn,
	)

	help := helpStyle.Render("\nPress 1 or s for Sites • 2 or d for Domains • 3 or n for Nodes • ? for Help • q to Quit")

	return title + "\n\n" + columns + "\n" + help
}

// renderBox renders content in a box with title
func renderBox(title, content string) string {
	titleText := titleStyle.Render(title)
	return boxStyle.Render(titleText + "\n" + content)
}

// renderSitesSummary renders the sites summary box
func renderSitesSummary(s *state.AppState) string {
	total := len(s.Sites)

	content := fmt.Sprintf(
		"Total Sites: %d\n\n"+
			"Press 's' or '1' to manage sites",
		total,
	)

	return renderBox("🌐 Sites", content)
}

// renderNodesSummary renders the nodes summary box
func renderNodesSummary(s *state.AppState) string {
	total := len(s.Nodes)

	content := fmt.Sprintf(
		"Total Nodes: %d\n\n"+
			"Press 'n' or '3' to manage nodes",
		total,
	)

	return renderBox("🖥️  Nodes", content)
}

// renderDomainsSummary renders the domains summary box
func renderDomainsSummary(s *state.AppState) string {
	total := len(s.Domains)

	content := fmt.Sprintf(
		"Total Domains: %d\n\n"+
			"Press 'd' or '2' to manage domains",
		total,
	)

	return renderBox("🌍 Domains", content)
}
