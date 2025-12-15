package models

type DnsRecordType string

const (
	DnsRecordTypeA     DnsRecordType = "A"
	DnsRecordTypeAAAA  DnsRecordType = "AAAA"
	DnsRecordTypeCNAME DnsRecordType = "CNAME"
	DnsRecordTypeMX    DnsRecordType = "MX"
	DnsRecordTypeTXT   DnsRecordType = "TXT"
	DnsRecordTypeSRV   DnsRecordType = "SRV"
)

type DnsRecord struct {
	ID         *string       `json:"id,omitempty" toml:"id,omitempty"`
	RecordType DnsRecordType `json:"record_type" toml:"record_type"`
	Name       string        `json:"name" toml:"name"`
	Value      string        `json:"value" toml:"value"`
	TTL        int           `json:"ttl" toml:"ttl"`
	Proxied    bool          `json:"proxied" toml:"proxied"` // Cloudflare-specific
}

// NewDnsRecord creates a new DNS record with default values
func NewDnsRecord(recordType DnsRecordType, name, value string, ttl int) *DnsRecord {
	return &DnsRecord{
		RecordType: recordType,
		Name:       name,
		Value:      value,
		TTL:        ttl,
		Proxied:    false,
	}
}
