package screens

import (
	"github.com/BlueBeard63/archon/internal/state"
	"github.com/BlueBeard63/archon/internal/ui"
)

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

	title := ui.TitleStyle.Render("Domains")
	help := ui.HelpStyle.Render("Press n to create, d to delete, s to sync DNS")

	return title + "\n\n" +
		"TODO: Implement domains table\n" +
		"Show warning icon for manual DNS domains\n\n" +
		help
}

// RenderDomainCreate renders the domain creation form
func RenderDomainCreate(s *state.AppState) string {
	// TODO: Implement single-field domain creation form
	// Field: Domain Name (e.g., example.com)
	//
	// Show warning about manual DNS:
	// "Note: Domain will be created with Manual DNS. Edit config.toml to add
	//  Cloudflare or Route53 API credentials."

	title := ui.TitleStyle.Render("Create New Domain")
	warning := ui.NotificationWarningStyle.Render(
		"⚠ Domain will be created with Manual DNS configuration",
	)
	help := ui.HelpStyle.Render("Enter domain name (e.g., example.com), Enter to submit, Esc to cancel")

	return title + "\n\n" +
		warning + "\n\n" +
		"Domain Name: " + "\n\n" + // TODO: Add actual input field
		help + "\n\n" +
		ui.HelpStyle.Render("To use Cloudflare or Route53, edit config.toml after creation")
}

// RenderDomainDnsRecords renders DNS records for a domain
func RenderDomainDnsRecords(s *state.AppState, domainID string) string {
	// TODO: Implement DNS records table
	// Columns: Type | Name | Value | TTL | Proxied
	//
	// Allow creating, editing, deleting records
	// If manual DNS, show instructions for manual configuration

	title := ui.TitleStyle.Render("DNS Records")
	help := ui.HelpStyle.Render("Press n to create record, e to edit, d to delete")

	return title + "\n\n" +
		"TODO: Implement DNS records table\n\n" +
		help
}
