package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BlueBeard63/archon/internal/config"
	"github.com/BlueBeard63/archon/internal/crypto"
	"github.com/BlueBeard63/archon/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeploySite_EncryptsCredentials(t *testing.T) {
	apiKey := "test-api-key-12345"
	var capturedRequest struct {
		Docker struct {
			Credentials struct {
				Username  string `json:"username"`
				Password  string `json:"password"`
				Encrypted bool   `json:"encrypted"`
			} `json:"credentials"`
		} `json:"docker"`
	}

	// Create test server that captures the request body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedRequest)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"site_id": uuid.New().String(),
			"status":  "running",
		})
	}))
	defer server.Close()

	client := NewHTTPNodeClient()
	site := &models.Site{
		ID:             uuid.New(),
		Name:           "test-site",
		DockerImage:    "nginx:latest",
		DockerUsername: "testuser",
		DockerToken:    "secret-token-123",
		DomainMappings: []models.DomainMapping{
			{DomainID: uuid.New(), Port: 80},
		},
	}

	err := client.DeploySiteWithEncryption(server.URL, apiKey, site, "example.com", nil)
	require.NoError(t, err)

	// Verify credentials are encrypted
	assert.True(t, capturedRequest.Docker.Credentials.Encrypted)
	assert.NotEqual(t, "testuser", capturedRequest.Docker.Credentials.Username)
	assert.NotEqual(t, "secret-token-123", capturedRequest.Docker.Credentials.Password)

	// Verify we can decrypt back to original values
	decryptedUser, err := crypto.Decrypt(capturedRequest.Docker.Credentials.Username, apiKey)
	require.NoError(t, err)
	assert.Equal(t, "testuser", decryptedUser)

	decryptedPass, err := crypto.Decrypt(capturedRequest.Docker.Credentials.Password, apiKey)
	require.NoError(t, err)
	assert.Equal(t, "secret-token-123", decryptedPass)
}

func TestDeploySite_ResolvesCredentialID(t *testing.T) {
	apiKey := "test-api-key-12345"
	credID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	var capturedRequest struct {
		Docker struct {
			Credentials struct {
				Username  string `json:"username"`
				Password  string `json:"password"`
				Encrypted bool   `json:"encrypted"`
			} `json:"credentials"`
		} `json:"docker"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedRequest)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"site_id": uuid.New().String(),
			"status":  "running",
		})
	}))
	defer server.Close()

	// Settings with a stored credential
	settings := &config.Settings{
		DockerCredentials: []config.DockerCredential{
			{
				ID:       credID,
				Name:     "Test Registry",
				Registry: "ghcr.io",
				Username: "registry-user",
				Token:    "registry-token",
			},
		},
	}

	client := NewHTTPNodeClient()
	site := &models.Site{
		ID:                 uuid.New(),
		Name:               "test-site",
		DockerImage:        "ghcr.io/test/app:latest",
		DockerCredentialID: &credID,
		DomainMappings: []models.DomainMapping{
			{DomainID: uuid.New(), Port: 8080},
		},
	}

	err := client.DeploySiteWithEncryption(server.URL, apiKey, site, "example.com", settings)
	require.NoError(t, err)

	// Verify credentials were resolved from settings and encrypted
	assert.True(t, capturedRequest.Docker.Credentials.Encrypted)

	decryptedUser, err := crypto.Decrypt(capturedRequest.Docker.Credentials.Username, apiKey)
	require.NoError(t, err)
	assert.Equal(t, "registry-user", decryptedUser)

	decryptedPass, err := crypto.Decrypt(capturedRequest.Docker.Credentials.Password, apiKey)
	require.NoError(t, err)
	assert.Equal(t, "registry-token", decryptedPass)
}

func TestDeploySite_EmptyCredentials_NoEncryption(t *testing.T) {
	var capturedRequest struct {
		Docker struct {
			Credentials struct {
				Username  string `json:"username"`
				Password  string `json:"password"`
				Encrypted bool   `json:"encrypted"`
			} `json:"credentials"`
		} `json:"docker"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &capturedRequest)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"site_id": uuid.New().String(),
			"status":  "running",
		})
	}))
	defer server.Close()

	client := NewHTTPNodeClient()
	site := &models.Site{
		ID:          uuid.New(),
		Name:        "test-site",
		DockerImage: "nginx:latest", // Public image, no credentials
		DomainMappings: []models.DomainMapping{
			{DomainID: uuid.New(), Port: 80},
		},
	}

	err := client.DeploySiteWithEncryption(server.URL, "api-key", site, "example.com", nil)
	require.NoError(t, err)

	// Empty credentials should not be marked as encrypted
	assert.False(t, capturedRequest.Docker.Credentials.Encrypted)
	assert.Empty(t, capturedRequest.Docker.Credentials.Username)
	assert.Empty(t, capturedRequest.Docker.Credentials.Password)
}

func TestConvertToWebSocketURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		path     string
		want     string
		wantErr  bool
	}{
		{
			name:     "http to ws",
			endpoint: "http://localhost:8080",
			path:     "/api/v1/sites/deploy/ws",
			want:     "ws://localhost:8080/api/v1/sites/deploy/ws",
			wantErr:  false,
		},
		{
			name:     "https to wss",
			endpoint: "https://node.example.com",
			path:     "/api/v1/sites/deploy/ws",
			want:     "wss://node.example.com/api/v1/sites/deploy/ws",
			wantErr:  false,
		},
		{
			name:     "with trailing slash",
			endpoint: "http://localhost:8080/",
			path:     "/api/v1/test",
			want:     "ws://localhost:8080/api/v1/test",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToWebSocketURL(tt.endpoint, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
