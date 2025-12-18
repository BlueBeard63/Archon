package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	url := fmt.Sprintf("%s/zones/%s/dns_records", cloudflareAPIBase, p.zoneID)

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var cfResp struct {
		Success bool               `json:"success"`
		Errors  []cloudflareError  `json:"errors"`
		Result  []cloudflareRecord `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare API error: %s", cfResp.Errors[0].Message)
		}
		return nil, fmt.Errorf("cloudflare API request failed")
	}

	// Convert to models.DnsRecord
	records := make([]models.DnsRecord, 0, len(cfResp.Result))
	for _, cfRecord := range cfResp.Result {
		records = append(records, fromCloudflareRecord(cfRecord))
	}

	return records, nil
}

// CreateRecord creates a new DNS record in Cloudflare
func (p *Provider) CreateRecord(domain string, record *models.DnsRecord, tags []string) (*models.DnsRecord, error) {
	url := fmt.Sprintf("%s/zones/%s/dns_records", cloudflareAPIBase, p.zoneID)

	// Convert to Cloudflare format
	cfRecord := toCloudflareRecord(record, tags)
	// Marshal request body
	body, err := json.Marshal(cfRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with body
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")

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
func (p *Provider) UpdateRecord(domain string, record *models.DnsRecord, tags []string) (*models.DnsRecord, error) {
	if record.ID == nil || *record.ID == "" {
		return nil, fmt.Errorf("record ID is required for updates")
	}

	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPIBase, p.zoneID, *record.ID)

	// Convert to Cloudflare format
	cfRecord := toCloudflareRecord(record, tags)

	// Marshal request body
	body, err := json.Marshal(cfRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")

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

	// Parse the updated record from response
	var updatedRecord cloudflareRecord
	if err := json.Unmarshal(cfResp.Result, &updatedRecord); err != nil {
		return nil, fmt.Errorf("failed to parse updated record: %w", err)
	}

	// Convert back to models.DnsRecord
	resultRecord := fromCloudflareRecord(updatedRecord)
	return &resultRecord, nil
}

// DeleteRecord removes a DNS record from Cloudflare
func (p *Provider) DeleteRecord(domain string, recordID string) error {
	if recordID == "" {
		return fmt.Errorf("record ID is required for deletion")
	}

	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPIBase, p.zoneID, recordID)

	// Create HTTP request
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiToken)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var cfResp cloudflareResponse
	if err := json.NewDecoder(resp.Body).Decode(&cfResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !cfResp.Success {
		if len(cfResp.Errors) > 0 {
			return fmt.Errorf("cloudflare API error: %s", cfResp.Errors[0].Message)
		}
		return fmt.Errorf("cloudflare API request failed")
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
	Comment string `json:"comment,omitempty"`
}

// toCloudflareRecord converts models.DnsRecord to Cloudflare format
func toCloudflareRecord(record *models.DnsRecord, tags []string) cloudflareRecord {
	cfRec := cloudflareRecord{
		Type:    string(record.RecordType),
		Name:    record.Name,
		Content: record.Value,
		TTL:     record.TTL,
		Proxied: record.Proxied,
	}

	// Use first tag as comment if tags are provided
	if len(tags) > 0 {
		cfRec.Comment = tags[0]
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
