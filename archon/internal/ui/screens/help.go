package screens

import (
	"github.com/charmbracelet/lipgloss"
)

// No color constants needed - using default white on black

// RenderHelp renders the help screen with all key bindings in 2 columns
func RenderHelp() string {
	title := titleStyle.Render("Help - Keyboard Shortcuts")

	// Left Column Sections
	globalSection := titleStyle.Render("Global Keys") + "\n" +
		formatKeyBinding("?", "Show this help screen") + "\n" +
		formatKeyBinding("Esc", "Go back / Cancel") + "\n" +
		formatKeyBinding("Ctrl+C, q", "Quit application") + "\n" +
		formatKeyBinding("Ctrl+S", "Save configuration")

	navigationSection := titleStyle.Render("Navigation") + "\n" +
		formatKeyBinding("Click Tabs", "Navigate with mouse") + "\n" +
		formatKeyBinding("1, s", "Sites list") + "\n" +
		formatKeyBinding("2, d", "Domains list") + "\n" +
		formatKeyBinding("3, n", "Nodes list") + "\n" +
		formatKeyBinding("0", "Dashboard")

	listsSection := titleStyle.Render("Lists (Sites/Domains/Nodes)") + "\n" +
		formatKeyBinding("j, Down", "Select next item") + "\n" +
		formatKeyBinding("k, Up", "Select previous item") + "\n" +
		formatKeyBinding("n, c", "Create new item") + "\n" +
		formatKeyBinding("d", "Delete selected item") + "\n" +
		formatKeyBinding("Enter", "View/Deploy selected item") + "\n" +
		formatKeyBinding("Click", "Select item (mouse)")

	formsSection := titleStyle.Render("Forms (Create/Edit)") + "\n" +
		formatKeyBinding("Tab", "Next field") + "\n" +
		formatKeyBinding("Shift+Tab", "Previous field") + "\n" +
		formatKeyBinding("Enter", "Submit form") + "\n" +
		formatKeyBinding("Esc", "Cancel") + "\n" +
		formatKeyBinding("Click", "Focus field (mouse)")

	// Right Column Sections
	sitesSection := titleStyle.Render("Sites Specific") + "\n" +
		formatKeyBinding("Enter", "Deploy site to node") + "\n" +
		formatKeyBinding("s", "Stop site") + "\n" +
		formatKeyBinding("r", "Restart site") + "\n" +
		formatKeyBinding("l", "View logs")

	domainsSection := titleStyle.Render("Domains Specific") + "\n" +
		formatKeyBinding("s", "Sync DNS records") + "\n" +
		formatKeyBinding("e", "Edit DNS records") + "\n" +
		formatKeyBinding("Enter", "View DNS records")

	nodesSection := titleStyle.Render("Nodes Specific") + "\n" +
		formatKeyBinding("v", "View node config") + "\n" +
		formatKeyBinding("h", "Health check") + "\n" +
		formatKeyBinding("Enter", "View node details") + "\n" +
		formatKeyBinding("m", "View metrics")

	mouseSection := titleStyle.Render("Mouse Support") + "\n" +
		"• Click on tabs to navigate screens\n" +
		"• Click on table rows to select items\n" +
		"• Click on form fields to focus them\n" +
		"• Click on buttons to activate them\n" +
		"• Scroll to navigate long lists"

	// Build left column
	leftColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		globalSection,
		"",
		navigationSection,
		"",
		listsSection,
		"",
		formsSection,
	)

	// Build right column
	rightColumn := lipgloss.JoinVertical(
		lipgloss.Left,
		sitesSection,
		"",
		domainsSection,
		"",
		nodesSection,
		"",
		mouseSection,
	)

	// Style columns with padding
	columnStyle := lipgloss.NewStyle().
		Width(40).
		PaddingRight(2)

	leftColumnStyled := columnStyle.Render(leftColumn)
	rightColumnStyled := columnStyle.Render(rightColumn)

	// Join columns horizontally
	columns := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftColumnStyled,
		rightColumnStyled,
	)

	// Combine title, columns, and footer
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		columns,
		"",
		helpStyle.Render("Press Esc or ? to close this help screen"),
	)

	return content
}

// formatKeyBinding formats a key binding line
func formatKeyBinding(key, description string) string {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Width(15)

	return keyStyle.Render(key) + description
}
