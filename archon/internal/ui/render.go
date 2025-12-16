package ui

import (
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui/components"
	"github.com/BlueBeard63/archon/internal/ui/screens"
	"github.com/charmbracelet/lipgloss"
)

var tabBar = components.NewTabBar()

// Render is the main rendering function that routes to appropriate screen (without zones)
func Render(s *state.AppState) string {
	// Update tab bar width based on window width
	tabBar.Width = s.WindowWidth
	if tabBar.Width == 0 {
		tabBar.Width = 80
	}

	// Render header
	header := RenderHeader()

	// Render tab navigation bar
	tabs := tabBar.Render(s.CurrentScreen)

	// Render main content based on current screen
	content := RenderScreen(s, nil)

	// Join all sections vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
	)
}

// RenderWithZones is the main rendering function with bubblezone support
func RenderWithZones(s *state.AppState, zm *zone.Manager) string {
	// Update tab bar width based on window width
	tabBar.Width = s.WindowWidth
	if tabBar.Width == 0 {
		tabBar.Width = 80
	}

	// Render header
	header := RenderHeader()

	// Render tab navigation bar with zones
	tabs := tabBar.RenderWithZones(s.CurrentScreen, zm)

	// Render main content based on current screen with zones
	content := RenderScreen(s, zm)

	// Join all sections vertically
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
	)
}

// RenderHeader renders the top header bar
func RenderHeader() string {
	style := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	return style.Render("⚡ ARCHON TUI - Docker Site Manager")
}

// RenderScreen routes to the appropriate screen renderer
func RenderScreen(s *state.AppState, zm *zone.Manager) string {
	switch s.CurrentScreen {
	case state.ScreenDashboard:
		return screens.RenderDashboard(s)
	case state.ScreenSitesList:
		return screens.RenderSitesListWithZones(s, zm)
	case state.ScreenSiteCreate:
		return screens.RenderSiteCreateWithZones(s, zm)
	case state.ScreenDomainsList:
		return screens.RenderDomainsListWithZones(s, zm)
	case state.ScreenDomainCreate:
		return screens.RenderDomainCreateWithZones(s, zm)
	case state.ScreenDomainEdit:
		return screens.RenderDomainEditWithZones(s, zm)
	case state.ScreenNodesList:
		return screens.RenderNodesListWithZones(s, zm)
	case state.ScreenNodeCreate:
		return screens.RenderNodeCreateWithZones(s, zm)
	case state.ScreenNodeConfig:
		return screens.RenderNodeConfig(s)
	case state.ScreenSettings:
		return screens.RenderSettingsWithZones(s, zm)
	case state.ScreenHelp:
		return screens.RenderHelp()
	default:
		return TitleStyle.Render("Unknown Screen")
	}
}
