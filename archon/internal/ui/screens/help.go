package screens

import (
	"github.com/charmbracelet/lipgloss"
)

// No color constants needed - using default white on black

// RenderHelp renders the help screen with all key bindings
func RenderHelp() string {
	title := titleStyle.Render("Help - Keyboard Shortcuts")

	globalSection := titleStyle.Render("Global Keys") + "\n" +
		formatKeyBinding("?", "Show this help screen") + "\n" +
		formatKeyBinding("Esc", "Go back / Cancel") + "\n" +
		formatKeyBinding("Ctrl+C, q", "Quit application") + "\n" +
		formatKeyBinding("Ctrl+S", "Save configuration")

	navigationSection := titleStyle.Render("Navigation") + "\n" +
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
		formatKeyBinding("h", "Health check") + "\n" +
		formatKeyBinding("Enter", "View node details") + "\n" +
		formatKeyBinding("m", "View metrics")

	formsSection := titleStyle.Render("Forms (Create/Edit)") + "\n" +
		formatKeyBinding("Tab", "Next field") + "\n" +
		formatKeyBinding("Shift+Tab", "Previous field") + "\n" +
		formatKeyBinding("Enter", "Submit form") + "\n" +
		formatKeyBinding("Esc", "Cancel") + "\n" +
		formatKeyBinding("Click", "Focus field (mouse)")

	mouseSection := titleStyle.Render("Mouse Support") + "\n" +
		"• Click on table rows to select items\n" +
		"• Click on form fields to focus them\n" +
		"• Click on buttons to activate them\n" +
		"• Scroll to navigate long lists"

	// Join all sections with vertical spacing
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		globalSection,
		"",
		navigationSection,
		"",
		listsSection,
		"",
		sitesSection,
		"",
		domainsSection,
		"",
		nodesSection,
		"",
		formsSection,
		"",
		mouseSection,
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
