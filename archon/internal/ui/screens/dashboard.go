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

	help := helpStyle.Render("\nPress 1 or s for Sites • 2 or d for Domains • 3 or n for Nodes • 4 or c for Settings • ? for Help • q to Quit")

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

	// Show first few site names for debugging
	siteNames := ""
	if total > 0 {
		siteNames = "\n\nSites:\n"
		for i, site := range s.Sites {
			if i < 3 {
				siteNames += fmt.Sprintf("• %s\n", site.Name)
			}
		}
		if total > 3 {
			siteNames += fmt.Sprintf("... and %d more", total-3)
		}
	}

	content := fmt.Sprintf(
		"Total Sites: %d%s\n\n"+
			"Press 's' or '1' to manage sites",
		total,
		siteNames,
	)

	return renderBox("🌐 Sites", content)
}

// renderNodesSummary renders the nodes summary box
func renderNodesSummary(s *state.AppState) string {
	total := len(s.Nodes)

	// Show first few node names for debugging
	nodeNames := ""
	if total > 0 {
		nodeNames = "\n\nNodes:\n"
		for i, node := range s.Nodes {
			if i < 3 {
				nodeNames += fmt.Sprintf("• %s\n", node.Name)
			}
		}
		if total > 3 {
			nodeNames += fmt.Sprintf("... and %d more", total-3)
		}
	}

	content := fmt.Sprintf(
		"Total Nodes: %d%s\n\n"+
			"Press 'n' or '3' to manage nodes",
		total,
		nodeNames,
	)

	return renderBox("🖥️  Nodes", content)
}

// renderDomainsSummary renders the domains summary box
func renderDomainsSummary(s *state.AppState) string {
	total := len(s.Domains)

	// Show first few domain names for debugging
	domainNames := ""
	if total > 0 {
		domainNames = "\n\nDomains:\n"
		for i, domain := range s.Domains {
			if i < 3 {
				domainNames += fmt.Sprintf("• %s\n", domain.Name)
			}
		}
		if total > 3 {
			domainNames += fmt.Sprintf("... and %d more", total-3)
		}
	}

	content := fmt.Sprintf(
		"Total Domains: %d%s\n\n"+
			"Press 'd' or '2' to manage domains",
		total,
		domainNames,
	)

	return renderBox("🌍 Domains", content)
}
