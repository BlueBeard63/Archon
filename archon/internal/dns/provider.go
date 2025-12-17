package dns

import (
	"fmt"

	"github.com/BlueBeard63/archon/internal/dns/cloudflare"
	"github.com/BlueBeard63/archon/internal/models"
)

// Provider defines the interface for DNS record management
// Implementations: Cloudflare, Route53 (future), Manual (returns nil)
type Provider interface {
	// ListRecords retrieves all DNS records for a domain
	ListRecords(domain string) ([]models.DnsRecord, error)

	// CreateRecord creates a new DNS record
	CreateRecord(domain string, record *models.DnsRecord) (*models.DnsRecord, error)

	// UpdateRecord updates an existing DNS record
	UpdateRecord(domain string, record *models.DnsRecord) (*models.DnsRecord, error)

	// DeleteRecord removes a DNS record by its ID
	DeleteRecord(domain string, recordID string) error
}

// CreateProvider is a factory function that creates the appropriate DNS provider
// based on the configuration. Returns nil for manual DNS.
func CreateProvider(provider *models.DnsProvider) (Provider, error) {
	switch provider.Type {
	case models.DnsProviderCloudflare:
		if provider.APIToken == "" || provider.ZoneID == "" {
			return nil, fmt.Errorf("Cloudflare provider requires APIToken and ZoneID")
		}
		return cloudflare.NewCloudflareProvider(provider.APIToken, provider.ZoneID), nil

	case models.DnsProviderRoute53:
		// TODO: Implement Route53 provider in future
		// return route53.NewRoute53Provider(provider.AccessKey, provider.SecretKey, provider.HostedZoneID), nil
		return nil, fmt.Errorf("Route53 provider not yet implemented")

	case models.DnsProviderManual:
		// Manual DNS means user manages records themselves
		// Return nil provider, TUI should show warnings
		return nil, nil

	default:
		return nil, fmt.Errorf("unknown DNS provider type: %s", provider.Type)
	}
}

// ValidateProvider checks if a DNS provider configuration is valid
func ValidateProvider(provider *models.DnsProvider) error {
	// TODO: Implement validation logic
	// Check that required fields are present based on provider type:
	// - Cloudflare: APIToken and ZoneID must be non-empty
	// - Route53: AccessKey, SecretKey, and HostedZoneID must be non-empty
	// - Manual: no validation needed

	switch provider.Type {
	case models.DnsProviderCloudflare:
		// if provider.APIToken == "" || provider.ZoneID == "" {
		//     return fmt.Errorf("Cloudflare provider requires APIToken and ZoneID")
		// }
	case models.DnsProviderRoute53:
		// if provider.AccessKey == "" || provider.SecretKey == "" || provider.HostedZoneID == "" {
		//     return fmt.Errorf("Route53 provider requires AccessKey, SecretKey, and HostedZoneID")
		// }
	}

	return nil
}
