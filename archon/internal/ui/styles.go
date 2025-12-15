package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color scheme for the TUI
var (
	// Primary colors
	ColorPrimary   = lipgloss.Color("#7C3AED") // Purple
	ColorSecondary = lipgloss.Color("#3B82F6") // Blue
	ColorSuccess   = lipgloss.Color("#10B981") // Green
	ColorWarning   = lipgloss.Color("#F59E0B") // Orange
	ColorError     = lipgloss.Color("#EF4444") // Red
	ColorInfo      = lipgloss.Color("#06B6D4") // Cyan

	// Text colors
	ColorText       = lipgloss.Color("#E5E7EB") // Light gray
	ColorTextMuted  = lipgloss.Color("#9CA3AF") // Medium gray
	ColorTextDim    = lipgloss.Color("#6B7280") // Dark gray
	ColorBackground = lipgloss.Color("#1F2937") // Dark background
	ColorBorder     = lipgloss.Color("#374151") // Border gray
)

// Base styles
var (
	// HeaderStyle is the style for the header bar
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorPrimary).
			Bold(true).
			Padding(0, 1)

	// MenuStyle is the style for the navigation menu
	MenuStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Background(ColorBackground).
			Padding(0, 1)

	// MenuItemStyle is the style for menu items
	MenuItemStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	// MenuItemActiveStyle is the style for the active menu item
	MenuItemActiveStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	// StatusBarStyle is the style for the bottom status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorBackground).
			Padding(0, 1)

	// TableHeaderStyle is the style for table headers
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				Bold(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorBorder)

	// TableRowStyle is the style for table rows
	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// TableRowSelectedStyle is the style for the selected table row
	TableRowSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				Background(lipgloss.Color("#2D3748"))

	// FormLabelStyle is the style for form field labels
	FormLabelStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Bold(true)

	// FormInputStyle is the style for form input fields
	FormInputStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	// FormInputFocusedStyle is the style for focused form input
	FormInputFocusedStyle = lipgloss.NewStyle().
				Foreground(ColorText).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(ColorPrimary).
				Padding(0, 1)

	// ButtonStyle is the style for buttons
	ButtonStyle = lipgloss.NewStyle().
			Foreground(ColorText).
			Background(ColorPrimary).
			Padding(0, 2).
			MarginRight(1).
			Bold(true)

	// ButtonDisabledStyle is the style for disabled buttons
	ButtonDisabledStyle = lipgloss.NewStyle().
				Foreground(ColorTextDim).
				Background(ColorBorder).
				Padding(0, 2).
				MarginRight(1)

	// NotificationSuccessStyle is the style for success notifications
	NotificationSuccessStyle = lipgloss.NewStyle().
					Foreground(ColorSuccess).
					Bold(true)

	// NotificationErrorStyle is the style for error notifications
	NotificationErrorStyle = lipgloss.NewStyle().
				Foreground(ColorError).
				Bold(true)

	// NotificationWarningStyle is the style for warning notifications
	NotificationWarningStyle = lipgloss.NewStyle().
					Foreground(ColorWarning).
					Bold(true)

	// NotificationInfoStyle is the style for info notifications
	NotificationInfoStyle = lipgloss.NewStyle().
				Foreground(ColorInfo)

	// BoxStyle is a generic box style for containers
	BoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1, 2)

	// TitleStyle is the style for section titles
	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			MarginBottom(1)

	// HelpStyle is the style for help text
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim).
			Italic(true)
)

// Helper functions for common operations

// RenderNotification renders a notification with the appropriate style
func RenderNotification(message, level string) string {
	var style lipgloss.Style
	switch level {
	case "success":
		style = NotificationSuccessStyle
	case "error":
		style = NotificationErrorStyle
	case "warning":
		style = NotificationWarningStyle
	case "info":
		style = NotificationInfoStyle
	default:
		style = NotificationInfoStyle
	}
	return style.Render(message)
}

// RenderButton renders a button with optional disabled state
func RenderButton(label string, enabled bool) string {
	if enabled {
		return ButtonStyle.Render(label)
	}
	return ButtonDisabledStyle.Render(label)
}

// RenderBox renders content in a box with optional title
func RenderBox(title, content string) string {
	if title != "" {
		titleText := TitleStyle.Render(title)
		return BoxStyle.Render(titleText + "\n" + content)
	}
	return BoxStyle.Render(content)
}

// MaxWidth returns the maximum width for content
func MaxWidth(windowWidth int) int {
	// Leave some margin on sides
	const margin = 4
	maxW := windowWidth - margin
	if maxW < 40 {
		return 40
	}
	return maxW
}

// MaxHeight returns the maximum height for content
func MaxHeight(windowHeight int) int {
	// Account for header (1) + menu (1) + status bar (1)
	const chrome = 3
	maxH := windowHeight - chrome
	if maxH < 10 {
		return 10
	}
	return maxH
}
