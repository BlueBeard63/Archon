package screens

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/BlueBeard63/archon/internal/state"
)

var (
	notificationWarningStyle = lipgloss.NewStyle().Bold(true)
)

var _ = table.Model{} // Suppress unused import

// RenderDomainsList renders the domains list screen with table
func RenderDomainsList(s *state.AppState) string {
	// TODO: Implement domains table using bubbles.table
	// Columns: Name | Provider | Records | Traefik | Created
	//
	// Example structure similar to sites table:
	// columns := []table.Column{
	//     {Title: "Name", Width: 30},
	//     {Title: "Provider", Width: 15},
	//     {Title: "Records", Width: 10},
	//     {Title: "Traefik", Width: 8},
	//     {Title: "Created", Width: 12},
	// }
	//
	// rows := []table.Row{}
	// for _, domain := range s.Domains {
	//     rows = append(rows, table.Row{
	//         domain.Name,
	//         domain.ProviderName(),
	//         fmt.Sprintf("%d", len(domain.DnsRecords)),
	//         boolToYesNo(domain.TraefikEnabled),
	//         domain.CreatedAt.Format("2006-01-02"),
	//     })
	// }
	//
	// If domain is manual DNS, show warning indicator

	title := titleStyle.Render("🌍 Domains")

	var content string
	if len(s.Domains) == 0 {
		content = helpStyle.Render("No domains yet. Press 'n' to create your first domain.")
	} else {
		content = fmt.Sprintf("Total Domains: %d\n\n", len(s.Domains))
		for i, domain := range s.Domains {
			content += fmt.Sprintf("%d. %s (%s)\n", i+1, domain.Name, domain.ProviderName())
		}
	}

	help := helpStyle.Render("\nPress n to create • Esc to go back")

	return title + "\n\n" + content + "\n" + help
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
		displayValue = domainName + "_"  // Show cursor
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
