package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/docker"
	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/BlueBeard63/archon-node/internal/proxy"
	"github.com/BlueBeard63/archon-node/internal/ssl"
)

type Handlers struct {
	dockerClient *docker.Client
	proxyManager proxy.ProxyManager
	sslManager   *ssl.Manager
	dataDir      string
}

func NewHandlers(dockerClient *docker.Client, proxyManager proxy.ProxyManager, sslManager *ssl.Manager, dataDir string) *Handlers {
	return &Handlers{
		dockerClient: dockerClient,
		proxyManager: proxyManager,
		sslManager:   sslManager,
		dataDir:      dataDir,
	}
}

// HandleHealth returns the health status of the node
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get Docker info
	dockerInfo, err := h.dockerClient.GetDockerInfo(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get Docker info")
		return
	}

	// Get proxy info
	traefikInfo, _ := h.proxyManager.GetInfo(ctx)

	response := models.HealthResponse{
		Status:  "healthy",
		Docker:  dockerInfo,
		Traefik: traefikInfo,
	}

	respondJSON(w, http.StatusOK, response)
}

// HandleDeploySite deploys a new site
func (h *Handlers) HandleDeploySite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req models.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" || req.Domain == "" || req.DockerImage == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Ensure SSL certificates if needed
	var certPath, keyPath string
	var err error
	if req.SSLEnabled {
		certPath, keyPath, err = h.sslManager.EnsureCertificate(ctx, req.ID, req.Domain, req.SSLCert, req.SSLKey)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to setup SSL: "+err.Error())
			return
		}
	}

	// Deploy container
	deployResp, err := h.dockerClient.DeploySite(ctx, &req, h.dataDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to deploy site: "+err.Error())
		return
	}

	// Configure reverse proxy
	if err := h.proxyManager.Configure(ctx, &req, certPath, keyPath); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to configure proxy: "+err.Error())
		return
	}

	// Reload proxy
	if err := h.proxyManager.Reload(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reload proxy: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, deployResp)
}

// HandleGetSiteStatus returns the status of a site
func (h *Handlers) HandleGetSiteStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get site ID from URL
	siteIDStr := chi.URLParam(r, "siteID")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	// Get status
	status, err := h.dockerClient.GetSiteStatus(ctx, siteID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get site status: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, status)
}

// HandleStopSite stops a running site
func (h *Handlers) HandleStopSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get site ID from URL
	siteIDStr := chi.URLParam(r, "siteID")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	// Stop site
	if err := h.dockerClient.StopSite(ctx, siteID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to stop site: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Site stopped successfully"})
}

// HandleRestartSite restarts a site
func (h *Handlers) HandleRestartSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get site ID from URL
	siteIDStr := chi.URLParam(r, "siteID")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	// Restart site
	if err := h.dockerClient.RestartSite(ctx, siteID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to restart site: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Site restarted successfully"})
}

// HandleDeleteSite deletes a site
func (h *Handlers) HandleDeleteSite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get site ID and domain from URL
	siteIDStr := chi.URLParam(r, "siteID")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		respondError(w, http.StatusBadRequest, "Missing domain parameter")
		return
	}

	// Delete container
	if err := h.dockerClient.DeleteSite(ctx, siteID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete site: "+err.Error())
		return
	}

	// Remove proxy configuration
	if err := h.proxyManager.Remove(ctx, siteID, domain); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to remove proxy config: "+err.Error())
		return
	}

	// Reload proxy
	if err := h.proxyManager.Reload(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reload proxy: "+err.Error())
		return
	}

	// Remove SSL certificates
	if err := h.sslManager.RemoveCertificate(siteID); err != nil {
		// Log but don't fail
		// fmt.Printf("Warning: failed to remove SSL certificate: %v\n", err)
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Site deleted successfully"})
}

// HandleGetLogs retrieves container logs
func (h *Handlers) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get site ID from URL
	siteIDStr := chi.URLParam(r, "siteID")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid site ID")
		return
	}

	// Get logs
	logs, err := h.dockerClient.GetContainerLogs(ctx, siteID, 100)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get logs: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs": logs,
	})
}
