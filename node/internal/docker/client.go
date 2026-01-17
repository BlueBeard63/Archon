package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/models"
)

type Client struct {
	cli         *client.Client
	networkName string
}

// NewClient creates a new Docker client
func NewClient(host, networkName string) (*Client, error) {
	var cli *client.Client
	var err error

	if host == "" {
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	} else {
		cli, err = client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Client{
		cli:         cli,
		networkName: networkName,
	}, nil
}

// EnsureNetwork creates the archon network if it doesn't exist
func (c *Client) EnsureNetwork(ctx context.Context) error {
	networks, err := c.cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list networks: %w", err)
	}

	// Check if network exists
	for _, n := range networks {
		if n.Name == c.networkName {
			return nil
		}
	}

	// Create network
	_, err = c.cli.NetworkCreate(ctx, c.networkName, types.NetworkCreate{
		Driver: "bridge",
	})
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	return nil
}

// DeploySite deploys a site as a Docker container
func (c *Client) DeploySite(ctx context.Context, req *models.DeployRequest, dataDir string) (*models.DeployResponse, error) {
	// Ensure network exists
	if err := c.EnsureNetwork(ctx); err != nil {
		return nil, err
	}

	encodedJSON, err := json.Marshal(req.Docker.Credentials)
	if err != nil {
		log.Fatal(err)
	}

	// Get authentication for pulling image
	authStr := base64.StdEncoding.EncodeToString(encodedJSON)

	// Pull image
	reader, err := c.cli.ImagePull(ctx, req.Docker.Image, image.PullOptions{
		RegistryAuth: authStr,
	})
	if err != nil {
		return &models.DeployResponse{
			SiteID:  req.ID,
			Status:  models.SiteStatusFailed,
			Message: fmt.Sprintf("Failed to pull image: %v", err),
		}, nil
	}
	defer reader.Close()

	// Consume pull output
	io.Copy(io.Discard, reader)

	// Prepare container name
	containerName := fmt.Sprintf("archon-%s", req.Name)

	// Prepare environment variables
	var envVars []string
	for k, v := range req.EnvironmentVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
	}

	// Prepare port bindings for all domain mappings
	primaryDomainMapping := req.DomainMappings[0]
	exposedPorts := nat.PortSet{}
	portBindings := nat.PortMap{}

	// Expose all ports from domain mappings
	for _, mapping := range req.DomainMappings {
		containerPort := nat.Port(fmt.Sprintf("%d/tcp", mapping.Port))
		exposedPorts[containerPort] = struct{}{}

		// Determine host port (use HostPort if specified, otherwise use Port)
		hostPort := mapping.Port
		if mapping.HostPort > 0 {
			hostPort = mapping.HostPort
		}

		portBindings[containerPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: fmt.Sprintf("%d", hostPort),
			},
		}
	}

	// Prepare labels
	labels := map[string]string{
		"archon.site.id":     req.ID.String(),
		"archon.site.name":   req.Name,
		"archon.site.domain": primaryDomainMapping.Domain,
	}

	// Add Traefik labels if provided
	for k, v := range req.TraefikLabels {
		labels[k] = v
	}

	// Create container config
	containerConfig := &container.Config{
		Image:        req.Docker.Image,
		Env:          envVars,
		ExposedPorts: exposedPorts,
		Labels:       labels,
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			c.networkName: {},
		},
	}

	// Handle config files
	if len(req.ConfigFiles) > 0 {
		siteDataDir := filepath.Join(dataDir, "sites", req.ID.String())
		if err := os.MkdirAll(siteDataDir, 0755); err != nil {
			return &models.DeployResponse{
				SiteID:  req.ID,
				Status:  models.SiteStatusFailed,
				Message: fmt.Sprintf("Failed to create site data directory: %v", err),
			}, nil
		}

		var binds []string
		for _, cf := range req.ConfigFiles {
			hostPath := filepath.Join(siteDataDir, cf.Name)
			if err := os.WriteFile(hostPath, []byte(cf.Content), 0644); err != nil {
				return &models.DeployResponse{
					SiteID:  req.ID,
					Status:  models.SiteStatusFailed,
					Message: fmt.Sprintf("Failed to write config file %s: %v", cf.Name, err),
				}, nil
			}
			binds = append(binds, fmt.Sprintf("%s:%s:ro", hostPath, cf.ContainerPath))
		}
		hostConfig.Binds = binds
	}

	// Stop and remove existing container if it exists
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return &models.DeployResponse{
			SiteID:  req.ID,
			Status:  models.SiteStatusFailed,
			Message: fmt.Sprintf("Failed to list containers: %v", err),
		}, nil
	}

	for _, cont := range containers {
		for _, name := range cont.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				// Stop container
				timeout := 10
				c.cli.ContainerStop(ctx, cont.ID, container.StopOptions{Timeout: &timeout})
				// Remove container
				c.cli.ContainerRemove(ctx, cont.ID, container.RemoveOptions{Force: true})
				break
			}
		}
	}

	// Create container
	resp, err := c.cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, containerName)
	if err != nil {
		return &models.DeployResponse{
			SiteID:  req.ID,
			Status:  models.SiteStatusFailed,
			Message: fmt.Sprintf("Failed to create container: %v", err),
		}, nil
	}

	// Start container
	if err := c.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return &models.DeployResponse{
			SiteID:  req.ID,
			Status:  models.SiteStatusFailed,
			Message: fmt.Sprintf("Failed to start container: %v", err),
		}, nil
	}

	return &models.DeployResponse{
		SiteID:      req.ID,
		Status:      models.SiteStatusRunning,
		ContainerID: resp.ID,
		Message:     "Site deployed successfully",
	}, nil
}

// GetSiteStatus returns the status of a deployed site
func (c *Client) GetSiteStatus(ctx context.Context, siteID uuid.UUID) (*models.SiteStatusResponse, error) {
	// Find container by label
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	siteIDStr := siteID.String()
	for _, cont := range containers {
		if id, ok := cont.Labels["archon.site.id"]; ok && id == siteIDStr {
			isRunning := cont.State == "running"
			status := models.SiteStatusStopped
			if isRunning {
				status = models.SiteStatusRunning
			}

			return &models.SiteStatusResponse{
				SiteID:      siteID,
				Status:      status,
				ContainerID: cont.ID,
				IsRunning:   isRunning,
			}, nil
		}
	}

	return &models.SiteStatusResponse{
		SiteID:    siteID,
		Status:    models.SiteStatusInactive,
		IsRunning: false,
		Message:   "Container not found",
	}, nil
}

// StopSite stops a running site
func (c *Client) StopSite(ctx context.Context, siteID uuid.UUID) error {
	status, err := c.GetSiteStatus(ctx, siteID)
	if err != nil {
		return err
	}

	if status.ContainerID == "" {
		return fmt.Errorf("container not found")
	}

	timeout := 10
	if err := c.cli.ContainerStop(ctx, status.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

// RestartSite restarts a site
func (c *Client) RestartSite(ctx context.Context, siteID uuid.UUID) error {
	status, err := c.GetSiteStatus(ctx, siteID)
	if err != nil {
		return err
	}

	if status.ContainerID == "" {
		return fmt.Errorf("container not found")
	}

	timeout := 10
	if err := c.cli.ContainerRestart(ctx, status.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}

	return nil
}

// DeleteSite stops and removes a site
func (c *Client) DeleteSite(ctx context.Context, siteID uuid.UUID) error {
	status, err := c.GetSiteStatus(ctx, siteID)
	if err != nil {
		return err
	}

	if status.ContainerID == "" {
		return fmt.Errorf("container not found")
	}

	// Stop container
	timeout := 10
	c.cli.ContainerStop(ctx, status.ContainerID, container.StopOptions{Timeout: &timeout})

	// Remove container
	if err := c.cli.ContainerRemove(ctx, status.ContainerID, container.RemoveOptions{
		Force: true,
	}); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}

// GetContainerLogs retrieves logs from a container
func (c *Client) GetContainerLogs(ctx context.Context, siteID uuid.UUID, lines int) ([]string, error) {
	status, err := c.GetSiteStatus(ctx, siteID)
	if err != nil {
		return nil, err
	}

	if status.ContainerID == "" {
		return nil, fmt.Errorf("container not found")
	}

	tail := fmt.Sprintf("%d", lines)
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}

	reader, err := c.cli.ContainerLogs(ctx, status.ContainerID, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs: %w", err)
	}

	// Split into lines
	logLines := strings.Split(string(logs), "\n")
	return logLines, nil
}

// GetDockerInfo returns information about Docker
func (c *Client) GetDockerInfo(ctx context.Context) (*models.DockerInfo, error) {
	info, err := c.cli.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker info: %w", err)
	}

	version, err := c.cli.ServerVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker version: %w", err)
	}

	return &models.DockerInfo{
		Version:           version.Version,
		ContainersRunning: info.ContainersRunning,
		ImagesCount:       info.Images,
	}, nil
}

// CheckPortConflicts validates that host ports are available
// Returns an error if any of the requested host ports are already in use by other containers
func (c *Client) CheckPortConflicts(ctx context.Context, hostPorts []int, excludeSiteID uuid.UUID) error {
	// List all containers (including stopped ones)
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "managed-by=archon"),
		),
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	usedPorts := make(map[int]string) // port -> container name

	for _, cont := range containers {
		// Skip the container being updated
		if siteID, ok := cont.Labels["archon.site.id"]; ok && siteID == excludeSiteID.String() {
			continue
		}

		// Extract ports from container
		for _, portBinding := range cont.Ports {
			if portBinding.PublicPort > 0 {
				containerName := strings.TrimPrefix(cont.Names[0], "/")
				usedPorts[int(portBinding.PublicPort)] = containerName
			}
		}
	}

	// Check for conflicts
	var conflicts []string
	for _, port := range hostPorts {
		if containerName, exists := usedPorts[port]; exists {
			conflicts = append(conflicts, fmt.Sprintf("port %d (used by %s)", port, containerName))
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("port conflicts detected: %s", strings.Join(conflicts, ", "))
	}

	return nil
}

// Close closes the Docker client
func (c *Client) Close() error {
	return c.cli.Close()
}
