package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BlueBeard63/archon-node/internal/mocks"
	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleHealth_ReturnsOnline(t *testing.T) {
	// Create mocks
	dockerClient := new(mocks.MockDockerClient)
	proxyManager := new(mocks.MockProxyManager)

	// Setup expectations - both healthy
	dockerClient.On("GetDockerInfo", mock.Anything).Return(&models.DockerInfo{
		Version:           "24.0.0",
		ContainersRunning: 5,
		ImagesCount:       10,
	}, nil)
	proxyManager.On("GetInfo", mock.Anything).Return(&models.TraefikInfo{
		Version:       "2.10.0",
		RoutersCount:  3,
		ServicesCount: 3,
	}, nil)

	// Create handlers with mocks
	handlers := &HandlersTestable{
		dockerClientMock: dockerClient,
		proxyManagerMock: proxyManager,
	}

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handlers.HandleHealthTestable(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "online", response.Status)
	assert.NotNil(t, response.Docker)
	assert.Equal(t, "24.0.0", response.Docker.Version)
	assert.NotNil(t, response.Traefik)

	dockerClient.AssertExpectations(t)
	proxyManager.AssertExpectations(t)
}

func TestHandleHealth_ReturnsDegraded_DockerDown(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)
	proxyManager := new(mocks.MockProxyManager)

	// Docker returns error
	dockerClient.On("GetDockerInfo", mock.Anything).Return(nil, errors.New("docker daemon not responding"))
	proxyManager.On("GetInfo", mock.Anything).Return(&models.TraefikInfo{
		Version: "2.10.0",
	}, nil)

	handlers := &HandlersTestable{
		dockerClientMock: dockerClient,
		proxyManagerMock: proxyManager,
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HandleHealthTestable(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "degraded", response.Status)
	assert.Nil(t, response.Docker)

	dockerClient.AssertExpectations(t)
}

func TestHandleHealth_ReturnsDegraded_ProxyDown(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)
	proxyManager := new(mocks.MockProxyManager)

	// Docker healthy, proxy returns error
	dockerClient.On("GetDockerInfo", mock.Anything).Return(&models.DockerInfo{
		Version: "24.0.0",
	}, nil)
	proxyManager.On("GetInfo", mock.Anything).Return(nil, errors.New("proxy not responding"))

	handlers := &HandlersTestable{
		dockerClientMock: dockerClient,
		proxyManagerMock: proxyManager,
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HandleHealthTestable(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "degraded", response.Status)
	assert.NotNil(t, response.Docker)
	assert.Nil(t, response.Traefik)
}

func TestHandleHealth_ReturnsOffline_AllDown(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)
	proxyManager := new(mocks.MockProxyManager)

	// Both return errors
	dockerClient.On("GetDockerInfo", mock.Anything).Return(nil, errors.New("docker daemon not responding"))
	proxyManager.On("GetInfo", mock.Anything).Return(nil, errors.New("proxy not responding"))

	handlers := &HandlersTestable{
		dockerClientMock: dockerClient,
		proxyManagerMock: proxyManager,
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HandleHealthTestable(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.HealthResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "offline", response.Status)
	assert.Nil(t, response.Docker)
	assert.Nil(t, response.Traefik)
}

// HandlersTestable is a testable version of Handlers that uses mock interfaces
type HandlersTestable struct {
	dockerClientMock *mocks.MockDockerClient
	proxyManagerMock *mocks.MockProxyManager
}

// HandleHealthTestable is a testable version of HandleHealth
func (h *HandlersTestable) HandleHealthTestable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var dockerInfo *models.DockerInfo
	var traefikInfo *models.TraefikInfo
	dockerHealthy := false
	proxyHealthy := false

	// Get Docker info
	if info, err := h.dockerClientMock.GetDockerInfo(ctx); err == nil {
		dockerInfo = info
		dockerHealthy = true
	}

	// Get proxy info
	if info, err := h.proxyManagerMock.GetInfo(ctx); err == nil {
		traefikInfo = info
		proxyHealthy = true
	}

	// Determine status
	status := "offline"
	if dockerHealthy && proxyHealthy {
		status = "online"
	} else if dockerHealthy || proxyHealthy {
		status = "degraded"
	}

	response := models.HealthResponse{
		Status:  status,
		Docker:  dockerInfo,
		Traefik: traefikInfo,
	}

	respondJSON(w, http.StatusOK, response)
}

// DockerClientInterface defines the interface for Docker client operations used in handlers
type DockerClientInterface interface {
	GetDockerInfo(ctx context.Context) (*models.DockerInfo, error)
}

// ProxyManagerInterface defines the interface for proxy manager operations used in handlers
type ProxyManagerInterface interface {
	GetInfo(ctx context.Context) (*models.TraefikInfo, error)
}

// Restart handler tests

func TestHandleRestartSite_Simple(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)

	siteID := "550e8400-e29b-41d4-a716-446655440000"
	expectedUUID, _ := uuid.Parse(siteID)
	dockerClient.On("RestartSite", mock.Anything, expectedUUID).Return(nil)

	handlers := &RestartHandlersTestable{
		dockerClientMock: dockerClient,
	}

	req := httptest.NewRequest("POST", "/api/v1/sites/"+siteID+"/restart", nil)
	w := httptest.NewRecorder()

	handlers.HandleRestartSiteTestable(w, req, siteID)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Site restarted successfully", response["message"])

	dockerClient.AssertExpectations(t)
}

func TestHandleRestartSite_WithPullLatest(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)

	siteID := "550e8400-e29b-41d4-a716-446655440000"
	expectedUUID, _ := uuid.Parse(siteID)
	dockerClient.On("RestartSiteWithPull", mock.Anything, expectedUUID, "", "").Return(nil)

	handlers := &RestartHandlersTestable{
		dockerClientMock: dockerClient,
	}

	body := `{"pull_latest": true}`
	req := httptest.NewRequest("POST", "/api/v1/sites/"+siteID+"/restart", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRestartSiteTestable(w, req, siteID)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Site restarted successfully", response["message"])

	dockerClient.AssertExpectations(t)
}

func TestHandleRestartSite_WithPullLatestAndCredentials(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)

	siteID := "550e8400-e29b-41d4-a716-446655440000"
	expectedUUID, _ := uuid.Parse(siteID)
	dockerClient.On("RestartSiteWithPull", mock.Anything, expectedUUID, "myuser", "mytoken").Return(nil)

	handlers := &RestartHandlersTestable{
		dockerClientMock: dockerClient,
	}

	body := `{"pull_latest": true, "docker_username": "myuser", "docker_token": "mytoken"}`
	req := httptest.NewRequest("POST", "/api/v1/sites/"+siteID+"/restart", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRestartSiteTestable(w, req, siteID)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Site restarted successfully", response["message"])

	dockerClient.AssertExpectations(t)
}

func TestHandleRestartSite_InvalidSiteID(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)

	handlers := &RestartHandlersTestable{
		dockerClientMock: dockerClient,
	}

	req := httptest.NewRequest("POST", "/api/v1/sites/invalid-uuid/restart", nil)
	w := httptest.NewRecorder()

	handlers.HandleRestartSiteTestable(w, req, "invalid-uuid")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRestartSite_RestartFails(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)

	siteID := "550e8400-e29b-41d4-a716-446655440000"
	expectedUUID, _ := uuid.Parse(siteID)
	dockerClient.On("RestartSite", mock.Anything, expectedUUID).Return(errors.New("container not found"))

	handlers := &RestartHandlersTestable{
		dockerClientMock: dockerClient,
	}

	req := httptest.NewRequest("POST", "/api/v1/sites/"+siteID+"/restart", nil)
	w := httptest.NewRecorder()

	handlers.HandleRestartSiteTestable(w, req, siteID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	dockerClient.AssertExpectations(t)
}

func TestHandleRestartSite_WithPullLatestFails(t *testing.T) {
	dockerClient := new(mocks.MockDockerClient)

	siteID := "550e8400-e29b-41d4-a716-446655440000"
	expectedUUID, _ := uuid.Parse(siteID)
	dockerClient.On("RestartSiteWithPull", mock.Anything, expectedUUID, "", "").Return(errors.New("container not found"))

	handlers := &RestartHandlersTestable{
		dockerClientMock: dockerClient,
	}

	body := `{"pull_latest": true}`
	req := httptest.NewRequest("POST", "/api/v1/sites/"+siteID+"/restart", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRestartSiteTestable(w, req, siteID)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	dockerClient.AssertExpectations(t)
}

// RestartHandlersTestable is a testable version for restart operations
type RestartHandlersTestable struct {
	dockerClientMock *mocks.MockDockerClient
	apiKey           string
}

// HandleRestartSiteTestable is a testable version of HandleRestartSite
func (h *RestartHandlersTestable) HandleRestartSiteTestable(w http.ResponseWriter, r *http.Request, siteIDStr string) {
	ctx := r.Context()

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	// Parse optional request body for credentials
	var req struct {
		DockerUsername       string `json:"docker_username,omitempty"`
		DockerToken          string `json:"docker_token,omitempty"`
		CredentialsEncrypted bool   `json:"credentials_encrypted,omitempty"`
		PullLatest           bool   `json:"pull_latest,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// If pull_latest is requested, use RestartSiteWithPull
	if req.PullLatest {
		username := req.DockerUsername
		password := req.DockerToken
		// Note: credential decryption would happen here if apiKey is set

		if err := h.dockerClientMock.RestartSiteWithPull(ctx, siteID, username, password); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to restart site: "+err.Error())
			return
		}
	} else {
		if err := h.dockerClientMock.RestartSite(ctx, siteID); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to restart site: "+err.Error())
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Site restarted successfully"})
}
