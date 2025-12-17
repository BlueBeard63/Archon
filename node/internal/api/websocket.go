package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/BlueBeard63/archon-node/internal/models"
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
	if req.Name == "" || req.Domain == "" || req.DockerImage == "" {
		sendError(conn, "Missing required fields")
		return
	}

	// Send progress: Starting deployment
	sendProgress(conn, "Starting deployment", "init")

	// Ensure SSL certificates if needed
	var certPath, keyPath string
	if req.SSLEnabled {
		sendProgress(conn, "Setting up SSL certificate for domain: "+req.Domain, "ssl")

		// Wait for DNS propagation before attempting SSL certificate
		sendProgress(conn, "Waiting for DNS propagation for: "+req.Domain, "ssl")
		if err := waitForDNSPropagation(req.Domain, 60*time.Second); err != nil {
			sendError(conn, "DNS propagation timeout for "+req.Domain+": "+err.Error())
			return
		}
		sendProgress(conn, "DNS propagation verified for: "+req.Domain, "ssl")

		certPath, keyPath, err = h.sslManager.EnsureCertificate(ctx, req.ID, req.Domain, req.SSLCert, req.SSLKey, req.SSLEmail)
		if err != nil {
			sendError(conn, "Failed to setup SSL: "+err.Error())
			return
		}
		sendProgress(conn, "SSL certificate obtained successfully", "ssl")
	}

	// Deploy container
	sendProgress(conn, "Deploying Docker container: "+req.DockerImage, "docker")
	deployResp, err := h.dockerClient.DeploySite(ctx, &req, h.dataDir)
	if err != nil {
		sendError(conn, "Failed to deploy site: "+err.Error())
		return
	}
	sendProgress(conn, "Docker container deployed successfully: "+deployResp.ContainerID[:12], "docker")

	// Configure reverse proxy
	sendProgress(conn, "Configuring reverse proxy for domain: "+req.Domain, "proxy")
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
