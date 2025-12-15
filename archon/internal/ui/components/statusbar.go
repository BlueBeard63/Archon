package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/BlueBeard63/archon/internal/state"
)

// RenderStatusBar renders the bottom status bar with notifications
func RenderStatusBar(s *state.AppState, windowWidth int) string {
	// Define styles inline to avoid circular import
	statusBarStyle := lipgloss.NewStyle().Padding(0, 1)

	// Left side: current screen name
	screenName := getScreenName(s.CurrentScreen)
	left := statusBarStyle.Render(screenName)

	// Center: latest notification (if any)
	center := ""
	if len(s.Notifications) > 0 {
		latest := s.Notifications[len(s.Notifications)-1]
		center = renderNotification(latest.Message, latest.Level)
	}

	// Right side: Help hint
	right := "Press ? for help"

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	centerWidth := lipgloss.Width(center)
	rightWidth := lipgloss.Width(right)

	// Calculate padding to distribute content
	totalContent := leftWidth + centerWidth + rightWidth
	if totalContent >= windowWidth {
		// Not enough space, just show left and right
		padding := windowWidth - leftWidth - rightWidth - 2
		if padding < 0 {
			padding = 0
		}
		return left + strings.Repeat(" ", padding) + right
	}

	// Distribute with center aligned
	leftPadding := (windowWidth - totalContent) / 2
	rightPadding := windowWidth - leftWidth - leftPadding - centerWidth - rightWidth

	if leftPadding < 0 {
		leftPadding = 0
	}
	if rightPadding < 0 {
		rightPadding = 0
	}

	return left +
		strings.Repeat(" ", leftPadding) +
		center +
		strings.Repeat(" ", rightPadding) +
		right
}

// RenderNotificationList renders a list of recent notifications
func RenderNotificationList(s *state.AppState) string {
	titleStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)

	if len(s.Notifications) == 0 {
		return "No notifications"
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Recent Notifications") + "\n\n")

	// Show last 10 notifications
	start := 0
	if len(s.Notifications) > 10 {
		start = len(s.Notifications) - 10
	}

	for i := start; i < len(s.Notifications); i++ {
		notif := s.Notifications[i]
		line := renderNotification(notif.Message, notif.Level)
		b.WriteString(line + "\n")
	}

	return b.String()
}

// renderNotification renders a notification with the appropriate style
func renderNotification(message, level string) string {
	// Simple prefix based on level, no colors
	prefix := ""
	switch level {
	case "success":
		prefix = "[OK] "
	case "error":
		prefix = "[ERROR] "
	case "warning":
		prefix = "[WARN] "
	case "info":
		prefix = "[INFO] "
	}
	return prefix + message
}

// getScreenName returns a human-readable name for a screen
func getScreenName(screen state.Screen) string {
	switch screen {
	case state.ScreenDashboard:
		return "Dashboard"
	case state.ScreenSitesList:
		return "Sites"
	case state.ScreenSiteCreate:
		return "Create Site"
	case state.ScreenDomainsList:
		return "Domains"
	case state.ScreenDomainCreate:
		return "Create Domain"
	case state.ScreenNodesList:
		return "Nodes"
	case state.ScreenNodeCreate:
		return "Create Node"
	case state.ScreenHelp:
		return "Help"
	default:
		return string(screen)
	}
}
