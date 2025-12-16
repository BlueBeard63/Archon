package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
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
		// 1. Build table rows (data only, NO buttons)
		var rows []table.Row
		for _, domain := range s.Domains {
			providerName := domain.ProviderName()
			if domain.IsManualDNS() {
				providerName = "⚠ " + providerName
			}

			traefikStatus := "No"
			if domain.TraefikEnabled {
				traefikStatus = "Yes"
			}

			rows = append(rows, table.Row{
				truncate(domain.Name, 30),
				truncate(providerName, 15),
				fmt.Sprintf("%d", len(domain.DnsRecords)),
				traefikStatus,
			})
		}

		// 2. Initialize/update table
		if s.DomainsTable == nil {
			columns := []table.Column{
				{Title: "Name", Width: 30},
				{Title: "Provider", Width: 15},
				{Title: "Records", Width: 8},
				{Title: "Traefik", Width: 8},
			}
			s.DomainsTable = components.NewTableComponent(columns, rows)
			s.DomainsTable.SetCursor(s.DomainsListIndex)
		} else {
			s.DomainsTable.SetRows(rows)
			s.DomainsTable.SetCursor(s.DomainsListIndex)
		}

		// 3. Render table view
		tableView := s.DomainsTable.View()

		// 4. Build action buttons column (aligned with rows)
		var actionsColumn strings.Builder
		actionsColumn.WriteString("\n\n") // Header padding

		for _, domain := range s.Domains {
			editBtn := components.Button{
				ID:      "edit-domain-" + domain.ID.String(),
				Label:   "✏️",
				Primary: false,
				Border:  false,
				Icon:    true,
			}
			deleteBtn := components.Button{
				ID:      "delete-domain-" + domain.ID.String(),
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

		// 5. Join table + actions horizontally
		mainContent := lipgloss.JoinHorizontal(
			lipgloss.Top,
			tableView,
			actionsColumn.String(),
		)

		// 6. Build sidebar for selected domain
		var sidebar string
		if len(s.Domains) > 0 && s.DomainsListIndex >= 0 && s.DomainsListIndex < len(s.Domains) {
			domain := &s.Domains[s.DomainsListIndex]
			sidebar = renderDomainSidebar(s, domain)
		}

		// 7. Join main content + sidebar
		if sidebar != "" {
			content = lipgloss.JoinHorizontal(
				lipgloss.Top,
				mainContent,
				"  ", // Spacing
				sidebar,
			)
		} else {
			content = mainContent
		}
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

// RenderDomainEdit renders the domain edit form
func RenderDomainEdit(s *state.AppState) string {
	return RenderDomainEditWithZones(s, nil)
}

// RenderDomainEditWithZones renders the domain edit form with clickable field
func RenderDomainEditWithZones(s *state.AppState, zm *zone.Manager) string {
	// Find the domain
	var domain *models.Domain
	for i := range s.Domains {
		if s.Domains[i].ID == s.SelectedDomainID {
			domain = &s.Domains[i]
			break
		}
	}

	if domain == nil {
		return titleStyle.Render("Edit Domain") + "\n\n" + "Domain not found\n\n" + helpStyle.Render("Press Esc to go back")
	}

	// Initialize form if needed with current domain name
	if len(s.FormFields) == 0 {
		s.FormFields = []string{domain.Name}
		s.CurrentFieldIndex = 0
	}

	title := titleStyle.Render("Edit Domain: " + domain.Name)

	// Render input field
	domainName := s.FormFields[0]
	displayValue := domainName
	if s.CurrentFieldIndex == 0 {
		displayValue = domainName + "_"
	}

	fieldLine := "Domain Name: " + displayValue + "\n"
	var inputField string
	if zm != nil {
		inputField = zm.Mark("field:0", fieldLine)
	} else {
		inputField = fieldLine
	}

	help := helpStyle.Render("\nType domain name, press Enter to save, Esc to cancel")

	return title + "\n\n" +
		inputField + "\n" +
		help
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

// renderDomainSidebar renders a sidebar showing sites related to the selected domain
func renderDomainSidebar(s *state.AppState, domain *models.Domain) string {
	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(35)

	title := lipgloss.NewStyle().Bold(true).Render("📋 Related Sites")

	// Find sites using this domain
	var relatedSites []string
	for _, site := range s.Sites {
		if site.DomainID == domain.ID {
			relatedSites = append(relatedSites,
				fmt.Sprintf("• %s (Port %d)", site.Name, site.Port))
		}
	}

	var content string
	if len(relatedSites) == 0 {
		content = lipgloss.NewStyle().Faint(true).
			Render("No sites using this domain")
	} else {
		content = strings.Join(relatedSites, "\n")
	}

	return sidebarStyle.Render(title + "\n\n" + content)
}
