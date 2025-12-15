package ui

import (
	zone "github.com/lrstanley/bubblezone"

	"github.com/charmbracelet/lipgloss"
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui/components"
	"github.com/BlueBeard63/archon/internal/ui/screens"
)

// Render is the main rendering function that routes to appropriate screen (without zones)
func Render(s *state.AppState) string {
	// Render header
	header := RenderHeader()

	// Render navigation menu
	menu := RenderMenu(s.CurrentScreen, nil)

	// Render main content based on current screen
	content := RenderScreen(s, nil)

	// Render status bar
	statusBar := components.RenderStatusBar(s, s.WindowWidth)

	// Join all sections vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		menu,
		content,
		statusBar,
	)
}

// RenderWithZones is the main rendering function with bubblezone support
func RenderWithZones(s *state.AppState, zm *zone.Manager) string {
	// Render header
	header := RenderHeader()

	// Render navigation menu with zones
	menu := RenderMenu(s.CurrentScreen, zm)

	// Render main content based on current screen with zones
	content := RenderScreen(s, zm)

	// Render status bar
	statusBar := components.RenderStatusBar(s, s.WindowWidth)

	// Join all sections vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		menu,
		content,
		statusBar,
	)
}

// RenderHeader renders the top header bar
func RenderHeader() string {
	style := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	return style.Render("⚡ ARCHON TUI - Docker Site Manager")
}

// RenderMenu renders the navigation menu bar
func RenderMenu(currentScreen state.Screen, zm *zone.Manager) string {
	var items []string

	if zm != nil {
		// Wrap each menu item in a clickable zone
		items = []string{
			zm.Mark("menu:dashboard", formatMenuItem("0:Dashboard", currentScreen == state.ScreenDashboard)),
			zm.Mark("menu:sites", formatMenuItem("1:Sites", currentScreen == state.ScreenSitesList)),
			zm.Mark("menu:domains", formatMenuItem("2:Domains", currentScreen == state.ScreenDomainsList)),
			zm.Mark("menu:nodes", formatMenuItem("3:Nodes", currentScreen == state.ScreenNodesList)),
			zm.Mark("menu:help", formatMenuItem("?:Help", currentScreen == state.ScreenHelp)),
		}
	} else {
		// No zones - just render normally
		items = []string{
			formatMenuItem("0:Dashboard", currentScreen == state.ScreenDashboard),
			formatMenuItem("1:Sites", currentScreen == state.ScreenSitesList),
			formatMenuItem("2:Domains", currentScreen == state.ScreenDomainsList),
			formatMenuItem("3:Nodes", currentScreen == state.ScreenNodesList),
			formatMenuItem("?:Help", currentScreen == state.ScreenHelp),
		}
	}

	style := lipgloss.NewStyle().Padding(0, 1)
	return style.Render(lipgloss.JoinHorizontal(lipgloss.Left, items...))
}

// formatMenuItem formats a menu item with active/inactive styling
func formatMenuItem(text string, active bool) string {
	if active {
		return lipgloss.NewStyle().Bold(true).Render(" " + text + " ")
	}
	return " " + text + " "
}

// RenderScreen routes to the appropriate screen renderer
func RenderScreen(s *state.AppState, zm *zone.Manager) string {
	switch s.CurrentScreen {
	case state.ScreenDashboard:
		return screens.RenderDashboard(s)
	case state.ScreenSitesList:
		return screens.RenderSitesList(s)
	case state.ScreenSiteCreate:
		return screens.RenderSiteCreateWithZones(s, zm)
	case state.ScreenDomainsList:
		return screens.RenderDomainsList(s)
	case state.ScreenDomainCreate:
		return screens.RenderDomainCreateWithZones(s, zm)
	case state.ScreenNodesList:
		return screens.RenderNodesList(s)
	case state.ScreenNodeCreate:
		return screens.RenderNodeCreateWithZones(s, zm)
	case state.ScreenHelp:
		return screens.RenderHelp()
	default:
		return TitleStyle.Render("Unknown Screen")
	}
}
