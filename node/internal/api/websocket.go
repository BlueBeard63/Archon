package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"

	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/BlueBeard63/archon-node/internal/ssl"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now (TODO: restrict in production)
	},
}

// DeploymentMessage represents a message sent during deployment
type DeploymentMessage struct {
	Type    string `json:"type"`    // "progress", "success", "error"
	Message string `json:"message"` // Human-readable message
	Step    string `json:"step"`    // Current step (e.g., "ssl", "docker", "proxy")
	Error   string `json:"error,omitempty"`
}

// waitForDNSPropagation waits for a DNS A record to resolve correctly
func waitForDNSPropagation(domain string, maxWaitTime time.Duration) error {
	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return &net.DNSError{
				Err:  "DNS propagation timeout",
				Name: domain,
			}
		case <-ticker.C:
			// Try to resolve the domain
			addrs, err := net.LookupHost(domain)
			if err == nil && len(addrs) > 0 {
				log.Printf("[DNS] Domain %s resolved to: %v", domain, addrs)
				return nil
			}
			// Log the attempt
			if err != nil {
				log.Printf("[DNS] Waiting for DNS propagation for %s: %v", domain, err)
			}
		}
	}
}

// HandleDeploySiteWebSocket handles WebSocket-based site deployment with progress updates
func (h *Handlers) HandleDeploySiteWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	ctx := r.Context()

	// Read deployment request from first WebSocket message
	var req models.DeployRequest
	if err := conn.ReadJSON(&req); err != nil {
		sendError(conn, "Failed to parse deployment request: "+err.Error())
		return
	}

	// Log the received request
	reqJSON, _ := json.MarshalIndent(req, "", "  ")
	log.Printf("========================================")
	log.Printf("WebSocket Deploy Request:")
	log.Printf("----------------------------------------")
	log.Printf("%s", string(reqJSON))
	log.Printf("========================================")

	// Validate request
	if req.Name == "" || len(req.DomainMappings) == 0 || req.Docker.Image == "" {
		sendError(conn, "Missing required fields")
		return
	}

	// Send progress: Starting deployment
	sendProgress(conn, "Starting deployment", "init")

	// Ensure SSL certificates if needed
	var certPath, keyPath string
	if req.SSLEnabled {
		// Get all domains to be configured
		domainMappings := getDomainMappingsForHandler(&req)
		domains := make([]string, 0, len(domainMappings))
		for _, mapping := range domainMappings {
			domains = append(domains, mapping.Domain)
		}

		sendProgress(conn, "Setting up SSL certificate for "+fmt.Sprintf("%d domains", len(domains)), "ssl")

		// Check if a valid SAN certificate already exists that covers all domains
		existingCertPath, existingKeyPath, coveredDomains, err := ssl.FindCertificateForDomains(domains)
		allCovered := false
		if err == nil && len(coveredDomains) > 0 {
			// Check if all requested domains are covered
			allCovered = true
			for _, d := range domains {
				found := false
				for _, covered := range coveredDomains {
					if covered == d {
						found = true
						break
					}
				}
				if !found {
					allCovered = false
					break
				}
			}
		}

		if allCovered {
			sendProgress(conn, "Existing SAN certificate covers all "+fmt.Sprintf("%d domains", len(domains)), "ssl")
			certPath = existingCertPath
			keyPath = existingKeyPath
		} else {
			// For Let's Encrypt, configure proxy first to serve validation challenges
			sendProgress(conn, "Configuring reverse proxy for Let's Encrypt validation", "ssl")
			if err := h.proxyManager.ConfigureForValidation(ctx, &req); err != nil {
				sendError(conn, "Failed to configure proxy for validation: "+err.Error())
				return
			}

			// Reload proxy with validation configuration
			sendProgress(conn, "Reloading reverse proxy with validation configuration", "ssl")
			if err := h.proxyManager.Reload(ctx); err != nil {
				sendError(conn, "Failed to reload proxy: "+err.Error())
				return
			}

			// Wait for DNS propagation for all domains
			for _, domain := range domains {
				sendProgress(conn, "Waiting for DNS propagation for: "+domain, "ssl")
				if err := waitForDNSPropagation(domain, 60*time.Second); err != nil {
					sendError(conn, "DNS propagation timeout for "+domain+": "+err.Error())
					return
				}
				sendProgress(conn, "DNS propagation verified for: "+domain, "ssl")
			}

			// Request SAN certificate for all domains
			sendProgress(conn, "Generating SSL certificate for "+fmt.Sprintf("%d domains", len(domains)), "ssl")
			var err error
			certPath, keyPath, err = h.sslManager.EnsureCertificateMulti(ctx, req.ID, domains, req.SSLCert, req.SSLKey, req.SSLEmail)
			if err != nil {
				sendError(conn, "Failed to setup SSL: "+err.Error())
				return
			}
			sendProgress(conn, "SSL certificate obtained successfully for "+fmt.Sprintf("%d domains", len(domains)), "ssl")
		}
	}

	// Deploy container
	sendProgress(conn, "Deploying Docker container: "+req.Docker.Image, "docker")
	deployResp, err := h.dockerClient.DeploySite(ctx, &req, h.dataDir)
	if err != nil {
		sendError(conn, "Failed to deploy site: "+err.Error())
		return
	}

	// Check if deployment failed (DeploySite returns error=nil even on failure)
	if deployResp.Status == models.SiteStatusFailed {
		sendError(conn, "Failed to deploy site: "+deployResp.Message)
		return
	}

	// Safely display container ID
	containerIDDisplay := deployResp.ContainerID
	if len(containerIDDisplay) > 12 {
		containerIDDisplay = containerIDDisplay[:12]
	}
	if containerIDDisplay == "" {
		containerIDDisplay = "(unknown)"
	}
	sendProgress(conn, "Docker container deployed successfully: "+containerIDDisplay, "docker")

	// Configure reverse proxy
	sendProgress(conn, "Configuring reverse proxy for "+fmt.Sprintf("%d domains", len(req.DomainMappings)), "proxy")

	// Check if proxy config already exists for all domains
	allProxyConfigsExist := true
	for _, mapping := range req.DomainMappings {
		domain := mapping.Domain
		apacheConfigPath := filepath.Join("/etc/apache2/sites-enabled", domain+".conf")
		nginxConfigPath := filepath.Join("/etc/nginx/sites-enabled", domain)
		if !fileExists(apacheConfigPath) && !fileExists(nginxConfigPath) {
			allProxyConfigsExist = false
			break
		}
	}

	if allProxyConfigsExist {
		sendProgress(conn, "Proxy configuration already exists for all "+fmt.Sprintf("%d domains", len(req.DomainMappings)), "proxy")
	} else {
		if err := h.proxyManager.Configure(ctx, &req, certPath, keyPath); err != nil {
			sendError(conn, "Failed to configure proxy: "+err.Error())
			return
		}

		// Reload proxy
		sendProgress(conn, "Reloading reverse proxy", "proxy")
		if err := h.proxyManager.Reload(ctx); err != nil {
			sendError(conn, "Failed to reload proxy: "+err.Error())
			return
		}
	}

	// Send success message with response
	sendSuccess(conn, "Site deployment completed successfully", deployResp)
}

// sendProgress sends a progress message to the WebSocket client
func sendProgress(conn *websocket.Conn, message, step string) {
	msg := DeploymentMessage{
		Type:    "progress",
		Message: message,
		Step:    step,
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send progress: %v", err)
	}
	log.Printf("[PROGRESS] %s", message)
}

// sendError sends an error message to the WebSocket client
func sendError(conn *websocket.Conn, message string) {
	msg := DeploymentMessage{
		Type:    "error",
		Message: "Deployment failed",
		Error:   message,
	}
	conn.WriteJSON(msg)
	log.Printf("[ERROR] %s", message)
}

// sendSuccess sends a success message with deployment response
func sendSuccess(conn *websocket.Conn, message string, response *models.DeployResponse) {
	// First send success message
	msg := DeploymentMessage{
		Type:    "success",
		Message: message,
		Step:    "complete",
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Failed to send success: %v", err)
		return
	}

	// Then send the deployment response
	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Failed to send response: %v", err)
	}

	// Give client time to receive messages before closing
	time.Sleep(100 * time.Millisecond)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
