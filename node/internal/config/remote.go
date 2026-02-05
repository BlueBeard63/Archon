package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// RemoteFetchResponse is the response from the config fetch endpoint (JSON format)
type RemoteFetchResponse struct {
	Config string `json:"config"`
	NodeID string `json:"node_id"`
}

// FetchConfigFromURL fetches a config from a remote URL and saves it to the specified path.
// It supports both JSON response (from archon nodes) and plain text TOML (from dpaste.org).
// It validates the TOML before saving and returns the parsed config.
func FetchConfigFromURL(url, savePath string) (*Config, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Make the request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch config: server returned %d: %s", resp.StatusCode, string(body))
	}

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Determine if this is JSON or plain text TOML
	var configContent string
	contentType := resp.Header.Get("Content-Type")

	// Try to detect format:
	// 1. Check Content-Type header
	// 2. Check if body looks like JSON (starts with {)
	// 3. Otherwise treat as plain text TOML
	trimmedBody := strings.TrimSpace(string(body))
	isJSON := strings.Contains(contentType, "application/json") ||
		(len(trimmedBody) > 0 && trimmedBody[0] == '{')

	if isJSON {
		// Parse JSON response (archon node format)
		var fetchResp RemoteFetchResponse
		if err := json.Unmarshal(body, &fetchResp); err != nil {
			return nil, fmt.Errorf("failed to decode JSON response: %w", err)
		}
		configContent = fetchResp.Config
	} else {
		// Plain text TOML (dpaste.org format)
		configContent = string(body)
	}

	// Validate the TOML by parsing it
	var cfg Config
	if err := toml.Unmarshal([]byte(configContent), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(savePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
			// Ignore error if we're saving to current directory
		}
	}

	// Save the config with restricted permissions (0600)
	if err := os.WriteFile(savePath, []byte(configContent), 0600); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return &cfg, nil
}
