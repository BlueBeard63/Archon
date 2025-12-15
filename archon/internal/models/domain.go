package models

import (
	"time"

	"github.com/google/uuid"
)

type DnsProviderType string

const (
	DnsProviderCloudflare DnsProviderType = "cloudflare"
	DnsProviderRoute53    DnsProviderType = "route53"
	DnsProviderManual     DnsProviderType = "manual"
)

type DnsProvider struct {
	Type         DnsProviderType `json:"type" toml:"type"`
	APIToken     string          `json:"api_token,omitempty" toml:"api_token,omitempty"`             // Cloudflare
	ZoneID       string          `json:"zone_id,omitempty" toml:"zone_id,omitempty"`                 // Cloudflare
	AccessKey    string          `json:"access_key,omitempty" toml:"access_key,omitempty"`           // Route53
	SecretKey    string          `json:"secret_key,omitempty" toml:"secret_key,omitempty"`           // Route53
	HostedZoneID string          `json:"hosted_zone_id,omitempty" toml:"hosted_zone_id,omitempty"`   // Route53
}

type Domain struct {
	ID             uuid.UUID   `json:"id" toml:"id"`
	Name           string      `json:"name" toml:"name"`
	DnsProvider    DnsProvider `json:"dns_provider" toml:"dns_provider"`
	DnsRecords     []DnsRecord `json:"dns_records" toml:"dns_records"`
	TraefikEnabled bool        `json:"traefik_enabled" toml:"traefik_enabled"`
	CreatedAt      time.Time   `json:"created_at" toml:"created_at"`
}

// IsManualDNS returns true if this domain uses manual DNS configuration
func (d *Domain) IsManualDNS() bool {
	return d.DnsProvider.Type == DnsProviderManual
}

// ProviderName returns a human-readable name for the DNS provider
func (d *Domain) ProviderName() string {
	switch d.DnsProvider.Type {
	case DnsProviderCloudflare:
		return "Cloudflare"
	case DnsProviderRoute53:
		return "AWS Route53"
	case DnsProviderManual:
		return "Manual"
	default:
		return "Unknown"
	}
}

// NewDomain creates a new Domain with default values
func NewDomain(name string, provider DnsProvider) *Domain {
	return &Domain{
		ID:             uuid.New(),
		Name:           name,
		DnsProvider:    provider,
		DnsRecords:     []DnsRecord{},
		TraefikEnabled: true,
		CreatedAt:      time.Now(),
	}
}
