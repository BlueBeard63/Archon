package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui/components"
)

// RenderDockerCredentialsList renders the Docker credentials list screen
func RenderDockerCredentialsList(s *state.AppState) string {
	return RenderDockerCredentialsListWithZones(s, nil)
}

// RenderDockerCredentialsListWithZones renders Docker credentials list with button zones
func RenderDockerCredentialsListWithZones(s *state.AppState, zm *zone.Manager) string {
	title := titleStyle.Render("🐳 Docker Registry Credentials")

	// Create button group
	buttonGroup := &components.ButtonGroup{
		Buttons: []components.Button{
			{ID: "create-docker-credential", Label: "➕ Add Credential", Primary: true},
		},
	}

	var buttons string
	if zm != nil {
		buttons = buttonGroup.RenderWithZones(zm)
	} else {
		buttons = buttonGroup.Render()
	}

	var content string
	if len(s.DockerCredentials) == 0 {
		content = helpStyle.Render("No Docker credentials configured.\n\nDocker credentials are used to pull private images from container registries.\nClick 'Add Credential' or press 'n' to add one.")
	} else {
		// Build table rows
		var rows []table.Row
		for _, cred := range s.DockerCredentials {
			rows = append(rows, table.Row{
				truncateDockerCred(cred.Name, 20),
				truncateDockerCred(cred.Registry, 25),
				truncateDockerCred(cred.Username, 20),
				"••••••••", // Masked token
			})
		}

		// Initialize/update table
		if s.DockerCredentialsTable == nil {
			columns := []table.Column{
				{Title: "Name", Width: 20},
				{Title: "Registry", Width: 25},
				{Title: "Username", Width: 20},
				{Title: "Token", Width: 10},
			}
			s.DockerCredentialsTable = components.NewTableComponent(columns, rows)
			s.DockerCredentialsTable.SetCursor(s.DockerCredentialsListIndex)
		} else {
			s.DockerCredentialsTable.SetRows(rows)
			s.DockerCredentialsTable.SetCursor(s.DockerCredentialsListIndex)
		}

		// Render table view
		tableView := s.DockerCredentialsTable.View()

		// Build action buttons column
		var actionsColumn strings.Builder
		actionsColumn.WriteString("\n\n") // Header padding

		for _, cred := range s.DockerCredentials {
			editBtn := components.Button{
				ID:      "edit-docker-credential-" + cred.ID.String(),
				Label:   "✏️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}
			deleteBtn := components.Button{
				ID:      "delete-docker-credential-" + cred.ID.String(),
				Label:   "🗑️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}

			var actionLine string
			if zm != nil {
				actionLine = editBtn.RenderWithZone(zm) + " " + deleteBtn.RenderWithZone(zm)
			} else {
				actionLine = editBtn.Render() + " " + deleteBtn.Render()
			}

			actionsColumn.WriteString(actionLine + "\n")
		}

		// Join table + actions horizontally
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			tableView,
			actionsColumn.String(),
		)
	}

	help := helpStyle.Render("\n\nPress j/k or arrows to navigate • e to edit • d to delete • n to create • Esc to go back")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		buttons,
		"",
		content,
		help,
	)
}

// truncateDockerCred truncates a string to maxLen characters
func truncateDockerCred(s string, maxLen int) string {
	if len(s) > maxLen {
		if maxLen <= 3 {
			return s[:maxLen]
		}
		return s[:maxLen-3] + "..."
	}
	if len(s) < maxLen {
		return s + strings.Repeat(" ", maxLen-len(s))
	}
	return s
}

// RenderDockerCredentialCreate renders the credential creation form
func RenderDockerCredentialCreate(s *state.AppState) string {
	return RenderDockerCredentialCreateWithZones(s, nil)
}

// RenderDockerCredentialCreateWithZones renders the credential creation form with clickable fields
func RenderDockerCredentialCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Initialize form if needed (4 fields: Name, Registry, Username, Token)
	if len(s.FormFields) != 4 {
		s.FormFields = []string{"", "docker.io", "", ""}
		s.CurrentFieldIndex = 0
		s.CursorPosition = 0
	}

	title := titleStyle.Render("Add Docker Registry Credential")

	labels := []string{"Name:", "Registry:", "Username:", "Token/Password:"}
	helpTexts := []string{
		"A friendly name for this credential (e.g., 'DockerHub', 'GitHub CR')",
		"Registry URL (e.g., docker.io, ghcr.io, registry.gitlab.com)",
		"Your registry username",
		"Access token or password (will be stored securely)",
	}

	// Render each field
	var fields strings.Builder
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value
		isFocused := i == s.CurrentFieldIndex

		// Mask token field when not focused
		if i == 3 && value != "" && !isFocused {
			displayValue = "••••••••••••••••"
		}

		// Show cursor if focused
		if isFocused {
			cursor := s.CursorPosition
			if cursor < 0 {
				cursor = 0
			}
			if cursor > len(value) {
				cursor = len(value)
			}
			displayValue = value[:cursor] + "_" + value[cursor:]
		}

		// Render label with focus styling
		styledLabel := renderFieldLabel(label, isFocused)

		fieldLine := styledLabel + " " + displayValue + "\n"
		helpLine := "  " + lipgloss.NewStyle().Faint(true).Render(helpTexts[i]) + "\n\n"

		if zm != nil {
			fields.WriteString(zm.Mark(fmt.Sprintf("field:%d", i), fieldLine) + helpLine)
		} else {
			fields.WriteString(fieldLine + helpLine)
		}
	}

	help := helpStyle.Render("\nTab/Shift+Tab to navigate, Enter to save, Esc to cancel")

	return title + "\n\n" + fields.String() + help
}

// RenderDockerCredentialEdit renders the credential edit form
func RenderDockerCredentialEdit(s *state.AppState) string {
	return RenderDockerCredentialEditWithZones(s, nil)
}

// RenderDockerCredentialEditWithZones renders the credential edit form with clickable fields
func RenderDockerCredentialEditWithZones(s *state.AppState, zm *zone.Manager) string {
	// Find the credential
	var cred *state.DockerCredential
	for i := range s.DockerCredentials {
		if s.DockerCredentials[i].ID == s.SelectedDockerCredentialID {
			cred = &s.DockerCredentials[i]
			break
		}
	}

	if cred == nil {
		return titleStyle.Render("Edit Docker Credential") + "\n\n" +
			"Credential not found\n\n" +
			helpStyle.Render("Press Esc to go back")
	}

	// Initialize form if needed (4 fields: Name, Registry, Username, Token)
	if len(s.FormFields) != 4 {
		s.FormFields = []string{cred.Name, cred.Registry, cred.Username, cred.Token}
		s.CurrentFieldIndex = 0
		s.CursorPosition = len(s.FormFields[0])
	}

	title := titleStyle.Render("Edit Docker Credential: " + cred.Name)

	labels := []string{"Name:", "Registry:", "Username:", "Token/Password:"}
	helpTexts := []string{
		"A friendly name for this credential",
		"Registry URL (e.g., docker.io, ghcr.io)",
		"Your registry username",
		"Access token or password",
	}

	// Render each field
	var fields strings.Builder
	for i, label := range labels {
		value := s.FormFields[i]
		displayValue := value
		isFocused := i == s.CurrentFieldIndex

		// Mask token field when not focused
		if i == 3 && value != "" && !isFocused {
			displayValue = "••••••••••••••••"
		}

		// Show cursor if focused
		if isFocused {
			cursor := s.CursorPosition
			if cursor < 0 {
				cursor = 0
			}
			if cursor > len(value) {
				cursor = len(value)
			}
			displayValue = value[:cursor] + "_" + value[cursor:]
		}

		// Render label with focus styling
		styledLabel := renderFieldLabel(label, isFocused)

		fieldLine := styledLabel + " " + displayValue + "\n"
		helpLine := "  " + lipgloss.NewStyle().Faint(true).Render(helpTexts[i]) + "\n\n"

		if zm != nil {
			fields.WriteString(zm.Mark(fmt.Sprintf("field:%d", i), fieldLine) + helpLine)
		} else {
			fields.WriteString(fieldLine + helpLine)
		}
	}

	help := helpStyle.Render("\nTab/Shift+Tab to navigate, Enter to save, Esc to cancel")

	return title + "\n\n" + fields.String() + help
}
