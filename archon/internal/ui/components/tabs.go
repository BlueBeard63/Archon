package components

import (
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/state"
)

var (
	// Active tab border style
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	// Inactive tab border style
	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}

	// Active tab style
	activeTabStyle = lipgloss.NewStyle().
			Border(activeTabBorder, true).
			BorderForeground(lipgloss.Color("36")).
			Padding(0, 1).
			Bold(true)

	// Inactive tab style
	inactiveTabStyle = lipgloss.NewStyle().
				Border(tabBorder, true).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1)

	// Tab gap style (fills space between tabs and edge)
	tabGapStyle = lipgloss.NewStyle().
			Border(lipgloss.Border{Bottom: "─"}, false, false, true, false).
			BorderForeground(lipgloss.Color("240"))
)

// Tab represents a single tab item
type Tab struct {
	ID     string
	Label  string
	Screen state.Screen
}

// TabBar represents a set of navigation tabs
type TabBar struct {
	Tabs   []Tab
	Active state.Screen
	Width  int
}

// NewTabBar creates a new tab bar with default tabs
func NewTabBar() *TabBar {
	return &TabBar{
		Tabs: []Tab{
			{ID: "dashboard", Label: "📊 Dashboard", Screen: state.ScreenDashboard},
			{ID: "sites", Label: "🌐 Sites", Screen: state.ScreenSitesList},
			{ID: "domains", Label: "🌍 Domains", Screen: state.ScreenDomainsList},
			{ID: "nodes", Label: "🖥️  Nodes", Screen: state.ScreenNodesList},
			{ID: "help", Label: "❓ Help", Screen: state.ScreenHelp},
		},
		Active: state.ScreenDashboard,
		Width:  80,
	}
}

// Render renders the tab bar (without zones)
func (t *TabBar) Render(currentScreen state.Screen) string {
	t.Active = currentScreen

	var renderedTabs []string
	for _, tab := range t.Tabs {
		var style lipgloss.Style
		if tab.Screen == t.Active {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(tab.Label))
	}

	// Join tabs horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Calculate gap to fill remaining width
	gap := t.Width - lipgloss.Width(row) - 2
	if gap > 0 {
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, tabGapStyle.Render(lipgloss.PlaceHorizontal(gap, lipgloss.Left, "")))
	}

	return row
}

// RenderWithZones renders the tab bar with clickable zones
func (t *TabBar) RenderWithZones(currentScreen state.Screen, zm *zone.Manager) string {
	if zm == nil {
		return t.Render(currentScreen)
	}

	t.Active = currentScreen

	var renderedTabs []string
	for _, tab := range t.Tabs {
		var style lipgloss.Style
		if tab.Screen == t.Active {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		// Wrap each tab in a clickable zone
		renderedTab := style.Render(tab.Label)
		renderedTabs = append(renderedTabs, zm.Mark("tab:"+tab.ID, renderedTab))
	}

	// Join tabs horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// Calculate gap to fill remaining width
	gap := t.Width - lipgloss.Width(row) - 2
	if gap > 0 {
		row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, tabGapStyle.Render(lipgloss.PlaceHorizontal(gap, lipgloss.Left, "")))
	}

	return row
}

// GetScreenByID returns the screen associated with a tab ID
func (t *TabBar) GetScreenByID(id string) *state.Screen {
	for _, tab := range t.Tabs {
		if tab.ID == id {
			return &tab.Screen
		}
	}
	return nil
}
