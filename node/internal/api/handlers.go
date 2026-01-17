package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/compose"
	"github.com/BlueBeard63/archon-node/internal/docker"
	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/BlueBeard63/archon-node/internal/pipeline"
	"github.com/BlueBeard63/archon-node/internal/pipeline/stages"
	"github.com/BlueBeard63/archon-node/internal/proxy"
	"github.com/BlueBeard63/archon-node/internal/ssl"
)

type Handlers struct {
	dockerClient    *docker.Client
	composeExecutor *compose.Executor
	proxyManager    proxy.ProxyManager
	sslManager      *ssl.Manager
	dataDir         string
}

func NewHandlers(
	dockerClient *docker.Client,
	composeExecutor *compose.Executor,
	proxyManager proxy.ProxyManager,
	sslManager *ssl.Manager,
	dataDir string,
) *Handlers {
	return &Handlers{
		dockerClient:    dockerClient,
		composeExecutor: composeExecutor,
		proxyManager:    proxyManager,
		sslManager:      sslManager,
		dataDir:         dataDir,
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

// HandleDeploySite deploys a new site using the deployment pipeline
func (h *Handlers) HandleDeploySite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req models.DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Log the received request for debugging
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("========================================")
	log.Printf("Received Deploy Request:")
	log.Printf("----------------------------------------")
	log.Printf("%s", string(reqJSON))
	log.Printf("========================================")

	// Create deployment pipeline
	deps := &stages.Dependencies{
		DockerClient:    h.dockerClient,
		ComposeExecutor: h.composeExecutor,
		ProxyManager:    h.proxyManager,
		SSLManager:      h.sslManager,
	}
	deployPipeline := stages.NewDeploymentPipeline(deps)

	// Create deployment state
	state := pipeline.NewDeploymentState(&req, h.dataDir)

	// Execute pipeline
	if err := deployPipeline.Execute(ctx, state); err != nil {
		log.Printf("[ERROR] Deployment failed: %v", err)

		// Determine appropriate status code
		statusCode := http.StatusInternalServerError
		if state.CurrentStage == "validation" {
			statusCode = http.StatusBadRequest
		} else if state.CurrentStage == "port-check" {
			statusCode = http.StatusConflict
		}

		respondError(w, statusCode, err.Error())
		return
	}

	log.Printf("Site deployment completed successfully")
	respondJSON(w, http.StatusOK, state.Response)
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

	// Get optional site name for compose (from query param)
	siteName := r.URL.Query().Get("name")
	siteType := r.URL.Query().Get("type")

	var status *models.SiteStatusResponse

	if siteType == "compose" && siteName != "" {
		// Use compose executor for compose sites
		status, err = h.composeExecutor.GetStatus(ctx, siteID, siteName)
	} else {
		// Default to docker client for container sites
		status, err = h.dockerClient.GetSiteStatus(ctx, siteID)
	}

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

	// Get optional parameters for compose
	siteName := r.URL.Query().Get("name")
	siteType := r.URL.Query().Get("type")

	if siteType == "compose" && siteName != "" {
		// Stop compose services
		if err := h.composeExecutor.StopSite(ctx, siteID, siteName); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to stop compose site: "+err.Error())
			return
		}
	} else {
		// Stop container
		if err := h.dockerClient.StopSite(ctx, siteID); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to stop site: "+err.Error())
			return
		}
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

	// Get optional parameters for compose
	siteName := r.URL.Query().Get("name")
	siteType := r.URL.Query().Get("type")

	// Delete container or compose stack
	if siteType == "compose" && siteName != "" {
		if err := h.composeExecutor.DeleteSite(ctx, siteID, siteName); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to delete compose site: "+err.Error())
			return
		}
	} else {
		if err := h.dockerClient.DeleteSite(ctx, siteID); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to delete site: "+err.Error())
			return
		}
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
		log.Printf("Warning: failed to remove SSL certificate: %v", err)
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

// getDomainMappingsForHandler extracts domain-port mappings from a DeployRequest
func getDomainMappingsForHandler(site *models.DeployRequest) []models.DomainMapping {
	return site.DomainMappings
}
