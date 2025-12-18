package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BlueBeard63/archon/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// HTTPNodeClient implements NodeClient using standard net/http
type HTTPNodeClient struct {
	client *http.Client
}

// NewHTTPNodeClient creates a new HTTP-based node client
func NewHTTPNodeClient() *HTTPNodeClient {
	return &HTTPNodeClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: false,
			},
		},
	}
}

// DeploymentMessage represents a message sent during WebSocket deployment
type DeploymentMessage struct {
	Type    string `json:"type"`    // "progress", "success", "error"
	Message string `json:"message"` // Human-readable message
	Step    string `json:"step"`    // Current step (e.g., "ssl", "docker", "proxy")
	Error   string `json:"error,omitempty"`
}

// DeploymentProgressCallback is called for each progress update during deployment
type DeploymentProgressCallback func(msg DeploymentMessage)

// DeploySite sends a deployment request to a node
func (c *HTTPNodeClient) DeploySite(endpoint, apiKey string, site *models.Site, domainName string) error {
	// Build deploy request with domain mappings support
	domainMappings := convertToNodeDomainMappings(site, domainName)

	req := struct {
		ID              uuid.UUID           `json:"id"`
		Name            string              `json:"name"`
		Docker          Docker              `json:"docker"`
		EnvironmentVars map[string]string   `json:"environment_vars"`
		DomainMappings  []DomainMapping     `json:"domain_mappings"`
		SSLEnabled      bool                `json:"ssl_enabled"`
		SSLEmail        string              `json:"ssl_email,omitempty"`
		ConfigFiles     []models.ConfigFile `json:"config_files"`
		TraefikLabels   map[string]string   `json:"traefik_labels,omitempty"`
	}{
		ID:   site.ID,
		Name: site.Name,
		Docker: Docker{
			Image: site.DockerImage,
			Credentials: DockerCredentials{
				Username: site.DockerUsername,
				Password: site.DockerToken,
			},
		},
		EnvironmentVars: site.EnvironmentVars,
		DomainMappings:  domainMappings,
		SSLEnabled:      site.SSLEnabled,
		SSLEmail:        site.SSLEmail,
		ConfigFiles:     site.ConfigFiles,
		TraefikLabels:   site.GenerateTraefikLabels(domainName),
	}

	url := fmt.Sprintf("%s/api/v1/sites/deploy", endpoint)
	resp, err := c.doRequest("POST", url, apiKey, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deploy failed with status %d", resp.StatusCode)
	}

	return nil
}

// DeploySiteWebSocket sends a deployment request via WebSocket with progress updates
func (c *HTTPNodeClient) DeploySiteWebSocket(endpoint, apiKey string, site *models.Site, domainName string, progressCallback DeploymentProgressCallback) error {
	// Convert HTTP/HTTPS endpoint to WebSocket URL
	wsURL, err := convertToWebSocketURL(endpoint, "/api/v1/sites/deploy/ws")
	if err != nil {
		return fmt.Errorf("failed to convert to WebSocket URL: %w", err)
	}

	// Establish WebSocket connection with Authorization header
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	// Set Authorization header
	headers := http.Header{}
	if apiKey != "" {
		headers.Set("Authorization", "Bearer "+apiKey)
	}

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()

	// Build deploy request with domain mappings support
	domainMappings := convertToNodeDomainMappings(site, domainName)

	req := struct {
		ID              uuid.UUID           `json:"id"`
		Name            string              `json:"name"`
		Docker          Docker              `json:"docker"`
		EnvironmentVars map[string]string   `json:"environment_vars"`
		DomainMappings  []DomainMapping     `json:"domain_mappings"`
		SSLEnabled      bool                `json:"ssl_enabled"`
		SSLEmail        string              `json:"ssl_email,omitempty"`
		ConfigFiles     []models.ConfigFile `json:"config_files"`
		TraefikLabels   map[string]string   `json:"traefik_labels,omitempty"`
	}{
		ID:   site.ID,
		Name: site.Name,
		Docker: Docker{
			Image: site.DockerImage,
			Credentials: DockerCredentials{
				Username: site.DockerUsername,
				Password: site.DockerToken,
			},
		},
		EnvironmentVars: site.EnvironmentVars,
		DomainMappings:  domainMappings,
		SSLEnabled:      site.SSLEnabled,
		SSLEmail:        site.SSLEmail,
		ConfigFiles:     site.ConfigFiles,
		TraefikLabels:   site.GenerateTraefikLabels(domainName),
	}

	// Send deployment request as first message
	if err := conn.WriteJSON(req); err != nil {
		return fmt.Errorf("failed to send deployment request: %w", err)
	}

	// Listen for progress messages
	for {
		var msg DeploymentMessage
		if err := conn.ReadJSON(&msg); err != nil {
			// Connection closed or error reading
			return fmt.Errorf("WebSocket read error: %w", err)
		}

		// Call progress callback if provided
		if progressCallback != nil {
			progressCallback(msg)
		}

		// Handle message types
		switch msg.Type {
		case "success":
			// Deployment completed successfully
			return nil
		case "error":
			// Deployment failed
			if msg.Error != "" {
				return fmt.Errorf("deployment failed: %s", msg.Error)
			}
			return fmt.Errorf("deployment failed: %s", msg.Message)
		case "progress":
			// Continue listening for more messages
			continue
		default:
			// Unknown message type, log and continue
			continue
		}
	}
}

// convertToWebSocketURL converts an HTTP/HTTPS endpoint to WebSocket URL
func convertToWebSocketURL(endpoint, path string) (string, error) {
	// Parse the endpoint URL
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Convert scheme to WebSocket
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	default:
		// Assume http if no scheme
		u.Scheme = "ws"
	}

	// Set the path
	u.Path = strings.TrimSuffix(u.Path, "/") + path

	return u.String(), nil
}

// DeleteSite removes a deployed site from a node
func (c *HTTPNodeClient) DeleteSite(endpoint, apiKey string, siteID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/sites/%s", endpoint, siteID.String())
	resp, err := c.doRequest("DELETE", url, apiKey, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetSiteStatus retrieves the current status of a deployed site
func (c *HTTPNodeClient) GetSiteStatus(endpoint, apiKey string, siteID uuid.UUID) (*models.SiteStatus, error) {
	url := fmt.Sprintf("%s/api/v1/sites/%s/status", endpoint, siteID.String())
	resp, err := c.doRequest("GET", url, apiKey, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get status failed with status %d", resp.StatusCode)
	}

	var status models.SiteStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode status response: %w", err)
	}

	return &status, nil
}

// StopSite stops a running site container
func (c *HTTPNodeClient) StopSite(endpoint, apiKey string, siteID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/sites/%s/stop", endpoint, siteID.String())
	resp, err := c.doRequest("POST", url, apiKey, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("stop failed with status %d", resp.StatusCode)
	}

	return nil
}

// RestartSite restarts a site container
func (c *HTTPNodeClient) RestartSite(endpoint, apiKey string, siteID uuid.UUID) error {
	url := fmt.Sprintf("%s/api/v1/sites/%s/restart", endpoint, siteID.String())
	resp, err := c.doRequest("POST", url, apiKey, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("restart failed with status %d", resp.StatusCode)
	}

	return nil
}

// HealthCheck performs a health check on a node
func (c *HTTPNodeClient) HealthCheck(endpoint, apiKey string) (*HealthResponse, error) {
	url := fmt.Sprintf("%s/health", endpoint)

	// Health endpoint is public, but we still include the API key if provided
	resp, err := c.doRequest("GET", url, apiKey, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode health response: %w", err)
	}

	return &health, nil
}

// GetDockerInfo retrieves Docker information from a node (via health check)
func (c *HTTPNodeClient) GetDockerInfo(endpoint, apiKey string) (*models.DockerInfo, error) {
	health, err := c.HealthCheck(endpoint, apiKey)
	if err != nil {
		return nil, err
	}

	if health.Docker == nil {
		return nil, fmt.Errorf("no Docker info in health response")
	}

	return health.Docker, nil
}

// GetTraefikInfo retrieves Traefik information from a node (via health check)
func (c *HTTPNodeClient) GetTraefikInfo(endpoint, apiKey string) (*models.TraefikInfo, error) {
	health, err := c.HealthCheck(endpoint, apiKey)
	if err != nil {
		return nil, err
	}

	if health.Traefik == nil {
		return nil, fmt.Errorf("no Traefik info in health response")
	}

	return health.Traefik, nil
}

// GetContainerLogs retrieves recent logs from a site's container
func (c *HTTPNodeClient) GetContainerLogs(endpoint, apiKey string, siteID uuid.UUID, lines int) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/sites/%s/logs", endpoint, siteID.String())
	resp, err := c.doRequest("GET", url, apiKey, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get logs failed with status %d", resp.StatusCode)
	}

	var logsResp struct {
		Logs []string `json:"logs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&logsResp); err != nil {
		return nil, fmt.Errorf("failed to decode logs response: %w", err)
	}

	return logsResp.Logs, nil
}

// GetContainerMetrics retrieves resource usage metrics for a site
func (c *HTTPNodeClient) GetContainerMetrics(endpoint, apiKey string, siteID uuid.UUID) (*ContainerMetrics, error) {
	// Note: This endpoint is not yet implemented in archon-node
	// For now, return a placeholder error
	return nil, fmt.Errorf("metrics endpoint not yet implemented")
}

// doRequest is a helper function to execute HTTP requests with auth
func (c *HTTPNodeClient) doRequest(method, url, apiKey string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	// Marshal body to JSON if provided
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	// Create request
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for error status codes
	if resp.StatusCode >= 400 {
		// Try to read error message from response
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(bodyBytes, &errResp); err == nil && errResp.Message != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Message)
		}

		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// DomainMapping represents a domain-to-port mapping for node API requests
type DomainMapping struct {
	Domain string `json:"domain"` // Full domain (e.g., "api.example.com")
	Port   int    `json:"port"`   // Container port for this domain
}

// convertToNodeDomainMappings converts site domain mappings to node API format
// Resolves domain UUIDs to actual domain names from the site's DomainMappings
func convertToNodeDomainMappings(site *models.Site, domainName string) []DomainMapping {
	// Extract the base domain by examining the mappings and domainName
	var baseDomain string

	// Look for a mapping with a subdomain that appears in domainName
	for _, mapping := range site.DomainMappings {
		if mapping.Subdomain != "" && strings.HasPrefix(domainName, mapping.Subdomain+".") {
			// Found a subdomain that matches - extract base by removing it
			baseDomain = strings.TrimPrefix(domainName, mapping.Subdomain+".")
			break
		}
	}

	// If no subdomain was found in domainName, domainName itself is likely the base
	if baseDomain == "" {
		baseDomain = domainName
	}

	// Create mappings for all domain mappings in the site
	mappings := make([]DomainMapping, 0, len(site.DomainMappings))
	for _, mapping := range site.DomainMappings {
		// Construct the full domain including subdomain if present
		fullDomain := baseDomain
		if mapping.Subdomain != "" {
			fullDomain = mapping.Subdomain + "." + baseDomain
		}
		mappings = append(mappings, DomainMapping{
			Domain: fullDomain,
			Port:   mapping.Port,
		})
	}
	return mappings
}
