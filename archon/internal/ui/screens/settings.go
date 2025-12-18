package screens

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/state"
)

// RenderSettings renders the settings screen
func RenderSettings(s *state.AppState) string {
	return RenderSettingsWithZones(s, nil)
}

// RenderSettingsWithZones renders the settings screen with clickable fields
func RenderSettingsWithZones(s *state.AppState, zm *zone.Manager) string {
	// Initialize form if needed (3 fields for API keys - Zone ID is now per-domain)
	if len(s.FormFields) != 3 {
		s.FormFields = []string{
			s.CloudflareAPIToken,
			s.Route53AccessKey,
			s.Route53SecretKey,
		}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("⚙️  Settings")

	labels := []string{
		"Cloudflare API Token:",
		"Route53 Access Key:",
		"Route53 Secret Key:",
	}

	helpTexts := []string{
		"Cloudflare API Token (with DNS edit permissions) - optional global default",
		"AWS access key for Route53 - optional global default",
		"AWS secret key for Route53 - optional global default",
	}

	// Render each field
	var fields string
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value

		// Mask the value if it's not empty and not focused
		if value != "" && i != s.CurrentFieldIndex {
			displayValue = "••••••••••••••••"
		}

		// Show cursor if focused
		if i == s.CurrentFieldIndex {
			// Show cursor at position
			cursor := s.CursorPosition
			if cursor < 0 {
				cursor = 0
			}
			if cursor > len(value) {
				cursor = len(value)
			}
			displayValue = value[:cursor] + "_" + value[cursor:]
			label = "> " + label // Show arrow for focused field
		} else {
			label = "  " + label
		}

		fieldLine := label + " " + displayValue + "\n"
		helpLine := "  " + lipgloss.NewStyle().Faint(true).Render(helpTexts[i]) + "\n\n"

		if zm != nil {
			fields += zm.Mark(fmt.Sprintf("field:%d", i), fieldLine) + helpLine
		} else {
			fields += fieldLine + helpLine
		}
	}

	help := helpStyle.Render("\nTab/Shift+Tab to navigate, Enter to save, Esc to cancel")
	note := helpStyle.Render("Note: Global credentials are stored in config.toml and used as defaults.\nZone IDs are configured per-domain when creating or editing domains.")

	return title + "\n\n" + fields + help + "\n" + note
}
