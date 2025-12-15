package screens

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui/components"
)

var (
	notificationWarningStyle = lipgloss.NewStyle().Bold(true)
)

// RenderDomainsList renders the domains list screen with table
func RenderDomainsList(s *state.AppState) string {
	return RenderDomainsListWithZones(s, nil)
}

// RenderDomainsListWithZones renders domains list with table and button zones
func RenderDomainsListWithZones(s *state.AppState, zm *zone.Manager) string {
	title := titleStyle.Render("🌍 Domains")

	// Create button group
	buttonGroup := &components.ButtonGroup{
		Buttons: []components.Button{
			{ID: "create-domain", Label: "➕ Create Domain", Primary: true},
		},
	}

	var buttons string
	if zm != nil {
		buttons = buttonGroup.RenderWithZones(zm)
	} else {
		buttons = buttonGroup.Render()
	}

	var content string
	if len(s.Domains) == 0 {
		content = helpStyle.Render("No domains yet. Click 'Create Domain' or press 'n'.")
	} else {
		content = fmt.Sprintf("Total Domains: %d\n\n", len(s.Domains))

		// Manual table with row action buttons
		headerStyle := lipgloss.NewStyle().Bold(true).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("240"))
		rowStyle := lipgloss.NewStyle().PaddingRight(2)

		header := headerStyle.Render(fmt.Sprintf("%-30s %-15s %-8s %-8s %-45s", "Name", "Provider", "Records", "Traefik", "Actions"))
		content += header + "\n"

		for _, domain := range s.Domains {
			providerName := domain.ProviderName()
			if domain.IsManualDNS() {
				providerName = "⚠ " + providerName
			}

			traefikStatus := "No"
			if domain.TraefikEnabled {
				traefikStatus = "Yes"
			}

			// Create row action buttons (icon only)
			editBtn := components.Button{ID: "edit-domain-" + domain.ID.String(), Label: "✏️", Primary: false, Border: false, Icon: true}
			deleteBtn := components.Button{ID: "delete-domain-" + domain.ID.String(), Label: "🗑️", Primary: false, Border: false, Icon: true}

			var editBtnStr, deleteBtnStr string
			if zm != nil {
				editBtnStr = editBtn.RenderWithZone(zm)
				deleteBtnStr = deleteBtn.RenderWithZone(zm)
			} else {
				editBtnStr = editBtn.Render()
				deleteBtnStr = deleteBtn.Render()
			}

			actions := editBtnStr + " " + deleteBtnStr

			rowText := fmt.Sprintf("%-30s %-15s %-8s %-8s %-45s",
				truncate(domain.Name, 30),
				truncate(providerName, 15),
				fmt.Sprintf("%d", len(domain.DnsRecords)),
				traefikStatus,
				actions,
			)

			content += rowStyle.Render(rowText) + "\n"
		}
	}

	help := helpStyle.Render("\n\nPress n to create • Esc to go back")

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

// RenderDomainCreate renders the domain creation form
func RenderDomainCreate(s *state.AppState) string {
	// Initialize form if needed
	if len(s.FormFields) == 0 {
		s.FormFields = []string{""}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Domain")
	warning := notificationWarningStyle.Render(
		"⚠ Domain will be created with Manual DNS configuration",
	)

	// Render input field
	domainName := s.FormFields[0]
	inputLabel := "Domain Name: "

	// Show cursor if focused
	displayValue := domainName
	if s.CurrentFieldIndex == 0 {
		displayValue = domainName + "_"
	}

	help := helpStyle.Render("Type domain name, press Enter to create, Esc to cancel")

	return title + "\n\n" +
		warning + "\n\n" +
		inputLabel + displayValue + "\n\n" +
		help + "\n\n" +
		helpStyle.Render("To use Cloudflare or Route53, edit config.toml after creation")
}

// RenderDomainCreateWithZones renders the domain creation form with clickable field
func RenderDomainCreateWithZones(s *state.AppState, zm *zone.Manager) string {
	// Fall back to regular rendering if no zone manager
	if zm == nil {
		return RenderDomainCreate(s)
	}

	// Initialize form if needed
	if len(s.FormFields) == 0 {
		s.FormFields = []string{""}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Create New Domain")
	warning := notificationWarningStyle.Render(
		"⚠ Domain will be created with Manual DNS configuration",
	)

	// Render input field with zone
	domainName := s.FormFields[0]
	displayValue := domainName
	if s.CurrentFieldIndex == 0 {
		displayValue = domainName + "_" // Show cursor
	}

	fieldLine := "Domain Name: " + displayValue + "\n"
	inputField := zm.Mark("field:0", fieldLine)

	help := helpStyle.Render("\nType domain name, press Enter to create, Esc to cancel")

	return title + "\n\n" +
		warning + "\n\n" +
		inputField + "\n" +
		help + "\n\n" +
		helpStyle.Render("To use Cloudflare or Route53, edit config.toml after creation")
}

// RenderDomainDnsRecords renders DNS records for a domain
func RenderDomainDnsRecords(s *state.AppState, domainID string) string {
	title := titleStyle.Render("DNS Records")

	// Find the domain
	var domain *models.Domain
	for i := range s.Domains {
		if s.Domains[i].ID.String() == domainID {
			domain = &s.Domains[i]
			break
		}
	}

	if domain == nil {
		return title + "\n\n" + "Domain not found\n\n" + helpStyle.Render("Press Esc to go back")
	}

	content := fmt.Sprintf("Domain: %s\n", domain.Name)
	content += fmt.Sprintf("Provider: %s\n\n", domain.ProviderName())

	// Show manual DNS warning if applicable
	if domain.IsManualDNS() {
		warning := notificationWarningStyle.Render("⚠ Manual DNS - Configure records manually at your DNS provider")
		content += warning + "\n\n"
	}

	// Display DNS records
	if len(domain.DnsRecords) == 0 {
		content += "No DNS records configured\n"
	} else {
		content += "DNS Records:\n\n"
		content += fmt.Sprintf("%-8s %-25s %-30s %-8s %-8s\n", "Type", "Name", "Value", "TTL", "Proxied")
		content += fmt.Sprintf("%s\n", "--------------------------------------------------------------------------------")

		for _, record := range domain.DnsRecords {
			proxied := "No"
			if record.Proxied {
				proxied = "Yes"
			}

			name := record.Name
			if len(name) > 25 {
				name = name[:22] + "..."
			}

			value := record.Value
			if len(value) > 30 {
				value = value[:27] + "..."
			}

			content += fmt.Sprintf("%-8s %-25s %-30s %-8d %-8s\n",
				record.RecordType,
				name,
				value,
				record.TTL,
				proxied,
			)
		}
	}

	help := helpStyle.Render("\nPress n to create record • Esc to go back")
	if domain.IsManualDNS() {
		help = helpStyle.Render("\nPress n to add record (manual config required) • Esc to go back")
	}

	return title + "\n\n" + content + "\n" + help
}
