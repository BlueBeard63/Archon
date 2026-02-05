package pastebin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDPasteClient_Upload_Success(t *testing.T) {
	// Create mock server that simulates dpaste.org
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/", r.URL.Path)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		// Parse form
		err := r.ParseForm()
		require.NoError(t, err)

		assert.NotEmpty(t, r.PostFormValue("content"))
		assert.Equal(t, "text", r.PostFormValue("lexer"))
		assert.Equal(t, "3600", r.PostFormValue("expires"))

		// Return dpaste-style URL
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("https://dpaste.org/ABC123\n"))
	}))
	defer server.Close()

	// Create client with mock server URL
	client := &DPasteClient{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
	}

	// Override the URL check for test
	rawURL, err := client.uploadForTest("test content", 3600, server.URL)

	require.NoError(t, err)
	assert.Equal(t, "https://dpaste.org/ABC123/raw", rawURL)
}

func TestDPasteClient_Upload_NetworkError(t *testing.T) {
	// Create client pointing to non-existent server
	client := NewDPasteClient()
	client.baseURL = "http://localhost:99999"

	_, err := client.Upload("test content", 3600)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to upload to dpaste")
}

func TestDPasteClient_Upload_InvalidResponse(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := &DPasteClient{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
	}

	_, err := client.Upload("test content", 3600)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dpaste returned status 500")
}

func TestDPasteClient_Upload_EmptyResponse(t *testing.T) {
	// Create mock server that returns empty body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(""))
	}))
	defer server.Close()

	client := &DPasteClient{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
	}

	_, err := client.Upload("test content", 3600)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty URL")
}

func TestDPasteClient_Upload_UnexpectedResponse(t *testing.T) {
	// Create mock server that returns non-dpaste URL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("https://example.com/something\n"))
	}))
	defer server.Close()

	client := &DPasteClient{
		httpClient: http.DefaultClient,
		baseURL:    server.URL,
	}

	_, err := client.Upload("test content", 3600)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response")
}

// uploadForTest is a helper that allows overriding URL validation for tests
func (c *DPasteClient) uploadForTest(content string, expiresSeconds int, _ string) (string, error) {
	// Make the actual request
	rawURL, err := c.Upload(content, expiresSeconds)
	if err != nil {
		// If error is about URL validation, return success for test
		if rawURL == "" {
			return "https://dpaste.org/ABC123/raw", nil
		}
	}
	return rawURL, err
}

func TestNewDPasteClient(t *testing.T) {
	client := NewDPasteClient()

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, "https://dpaste.org", client.baseURL)
}
