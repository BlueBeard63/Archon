package pastebin

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DPasteClient is a client for dpaste.org API
type DPasteClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewDPasteClient creates a new dpaste.org client
func NewDPasteClient() *DPasteClient {
	return &DPasteClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://dpaste.org",
	}
}

// Upload uploads content to dpaste.org and returns the raw URL for fetching.
// expiresSeconds specifies how long the paste should live (max 31536000 = 1 year).
// Returns the raw URL like https://dpaste.org/XXXX/raw
func (c *DPasteClient) Upload(content string, expiresSeconds int) (string, error) {
	// Build form data
	// dpaste.org API expects: content, lexer, expires
	data := url.Values{}
	data.Set("content", content)
	data.Set("lexer", "text") // Plain text
	data.Set("expires", fmt.Sprintf("%d", expiresSeconds))

	// Create request
	req, err := http.NewRequest(
		http.MethodPost,
		c.baseURL+"/api/",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload to dpaste: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("dpaste returned status %d: %s", resp.StatusCode, string(body))
	}

	// Response is the URL like "https://dpaste.org/XXXX\n"
	pasteURL := strings.TrimSpace(string(body))
	if pasteURL == "" {
		return "", fmt.Errorf("dpaste returned empty URL")
	}

	// Validate it looks like a dpaste URL
	if !strings.HasPrefix(pasteURL, "https://dpaste.org/") && !strings.HasPrefix(pasteURL, "http://dpaste.org/") {
		return "", fmt.Errorf("unexpected response from dpaste: %s", pasteURL)
	}

	// Return the raw URL for direct TOML fetching
	rawURL := pasteURL + "/raw"

	return rawURL, nil
}
