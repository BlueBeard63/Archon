package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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

	// Log the received request for debugging
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("========================================")
	log.Printf("Received Deploy Request:")
	log.Printf("----------------------------------------")
	log.Printf("%s", string(reqJSON))
	log.Printf("========================================")

	// Validate request
	if req.Name == "" || req.Domain == "" || req.Docker.Image == "" {
		respondError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Ensure SSL certificates if needed
	var certPath, keyPath string
	var err error
	if req.SSLEnabled {
		log.Printf("Setting up SSL certificate for domain: %s (email: %s)", req.Domain, req.SSLEmail)
		log.Printf("[DEBUG] ProxyManager type: %T", h.proxyManager)
		// For Let's Encrypt, configure proxy first to serve validation challenges
		log.Printf("Configuring reverse proxy for Let's Encrypt validation")
		if err := h.proxyManager.ConfigureForValidation(ctx, &req); err != nil {
			log.Printf("[ERROR] Failed to configure proxy for validation: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to configure proxy for validation: "+err.Error())
			return
		}

		// Reload proxy with validation configuration
		log.Printf("Reloading reverse proxy with validation configuration")
		if err := h.proxyManager.Reload(ctx); err != nil {
			log.Printf("[ERROR] Failed to reload proxy: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to reload proxy: "+err.Error())
			return
		}

		// Wait for DNS propagation before attempting SSL certificate
		log.Printf("Waiting for DNS propagation for: %s", req.Domain)
		if err := waitForDNSPropagation(req.Domain, 60*time.Second); err != nil {
			log.Printf("[ERROR] DNS propagation timeout for %s: %v", req.Domain, err)
			respondError(w, http.StatusInternalServerError, "DNS propagation timeout for "+req.Domain+": "+err.Error())
			return
		}
		log.Printf("DNS propagation verified for: %s", req.Domain)

		certPath, keyPath, err = h.sslManager.EnsureCertificate(ctx, req.ID, req.Domain, req.SSLCert, req.SSLKey, req.SSLEmail)
		if err != nil {
			log.Printf("[ERROR] Failed to setup SSL: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to setup SSL: "+err.Error())
			return
		}
		log.Printf("SSL certificate obtained successfully: cert=%s, key=%s", certPath, keyPath)
	}

	// Deploy container
	log.Printf("Deploying Docker container: image=%s, port=%d", req.Docker.Image, req.Port)
	deployResp, err := h.dockerClient.DeploySite(ctx, &req, h.dataDir)
	if err != nil {
		log.Printf("[ERROR] Failed to deploy site: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to deploy site: "+err.Error())
		return
	}
	log.Printf("Docker container deployed successfully: containerID=%s", deployResp.ContainerID)

	// Configure reverse proxy
	log.Printf("Configuring reverse proxy for domain: %s", req.Domain)
	if err := h.proxyManager.Configure(ctx, &req, certPath, keyPath); err != nil {
		log.Printf("[ERROR] Failed to configure proxy: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to configure proxy: "+err.Error())
		return
	}

	// Reload proxy
	log.Printf("Reloading reverse proxy")
	if err := h.proxyManager.Reload(ctx); err != nil {
		log.Printf("[ERROR] Failed to reload proxy: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to reload proxy: "+err.Error())
		return
	}
	log.Printf("Site deployment completed successfully")

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
