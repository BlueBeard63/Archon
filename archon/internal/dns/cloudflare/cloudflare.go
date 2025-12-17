package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BlueBeard63/archon/internal/models"
)

const cloudflareAPIBase = "https://api.cloudflare.com/client/v4"

// Provider implements dns.Provider for Cloudflare
type Provider struct {
	apiToken string
	zoneID   string
	client   *http.Client
}

// NewCloudflareProvider creates a new Cloudflare DNS provider
func NewCloudflareProvider(apiToken, zoneID string) *Provider {
	return &Provider{
		apiToken: apiToken,
		zoneID:   zoneID,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ListRecords retrieves all DNS records for a zone
func (p *Provider) ListRecords(domain string) ([]models.DnsRecord, error) {
	// TODO: Implement Cloudflare DNS record listing
	// Steps:
	// 1. Build request: GET {cloudflareAPIBase}/zones/{zoneID}/dns_records
	// 2. Set Authorization: Bearer {apiToken} header
	// 3. Parse response JSON
	// 4. Convert Cloudflare response format to []models.DnsRecord
	// 5. Handle pagination if necessary (Cloudflare uses page/per_page)

	// Example structure:
	// url := fmt.Sprintf("%s/zones/%s/dns_records", cloudflareAPIBase, p.zoneID)
	// req, err := http.NewRequest("GET", url, nil)
	// req.Header.Set("Authorization", "Bearer "+p.apiToken)
	// req.Header.Set("Content-Type", "application/json")
	// ...
	// Parse Cloudflare response format:
	// {
	//   "result": [
	//     {
	//       "id": "...",
	//       "type": "A",
	//       "name": "example.com",
	//       "content": "192.0.2.1",
	//       "ttl": 300,
	//       "proxied": false
	//     }
	//   ],
	//   "success": true
	// }

	return nil, nil
}

// CreateRecord creates a new DNS record in Cloudflare
func (p *Provider) CreateRecord(domain string, record *models.DnsRecord) (*models.DnsRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records", cloudflareAPIBase, p.zoneID)

	// Convert to Cloudflare format
	cfRecord := toCloudflareRecord(record)

	// Marshal request body
	body, err := json.Marshal(cfRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Body = http.NoBody
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	req.ContentLength = int64(len(body))

	// Make request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var cfResp cloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare API error: %s", cfResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("cloudflare API request failed")
	}

	// Parse the created record from response
	var createdRecord cloudflareRecord
	if err := json.Unmarshal(cfResp.Result, &createdRecord); err != nil {
		return nil, fmt.Errorf("failed to parse created record: %w", err)
	}

	// Convert back to models.DnsRecord
	resultRecord := fromCloudflareRecord(createdRecord)
	return &resultRecord, nil
}

// UpdateRecord updates an existing DNS record
func (p *Provider) UpdateRecord(domain string, record *models.DnsRecord) (*models.DnsRecord, error) {
	// TODO: Implement DNS record update
	// PUT {cloudflareAPIBase}/zones/{zoneID}/dns_records/{recordID}
	// record.ID must be set to the Cloudflare record ID

	if record.ID == nil || *record.ID == "" {
		return nil, fmt.Errorf("record ID is required for updates")
	}

	return nil, nil
}

// DeleteRecord removes a DNS record from Cloudflare
func (p *Provider) DeleteRecord(domain string, recordID string) error {
	// TODO: Implement DNS record deletion
	// DELETE {cloudflareAPIBase}/zones/{zoneID}/dns_records/{recordID}

	if recordID == "" {
		return fmt.Errorf("record ID is required for deletion")
	}

	return nil
}

// cloudflareResponse represents the standard Cloudflare API response format
type cloudflareResponse struct {
	Success bool              `json:"success"`
	Errors  []cloudflareError `json:"errors"`
	Result  json.RawMessage   `json:"result"`
}

type cloudflareError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// cloudflareRecord represents a DNS record in Cloudflare's format
type cloudflareRecord struct {
	ID      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

// toCloudflareRecord converts models.DnsRecord to Cloudflare format
func toCloudflareRecord(record *models.DnsRecord) cloudflareRecord {
	cfRec := cloudflareRecord{
		Type:    string(record.RecordType),
		Name:    record.Name,
		Content: record.Value,
		TTL:     record.TTL,
		Proxied: record.Proxied,
	}

	// Include ID if it exists (for updates)
	if record.ID != nil {
		cfRec.ID = *record.ID
	}

	// Set default TTL if not specified
	if cfRec.TTL == 0 {
		cfRec.TTL = 300 // 5 minutes default
	}

	return cfRec
}

// fromCloudflareRecord converts Cloudflare format to models.DnsRecord
func fromCloudflareRecord(cf cloudflareRecord) models.DnsRecord {
	id := cf.ID
	return models.DnsRecord{
		ID:         &id,
		RecordType: models.DnsRecordType(cf.Type),
		Name:       cf.Name,
		Value:      cf.Content,
		TTL:        cf.TTL,
		Proxied:    cf.Proxied,
	}
}
