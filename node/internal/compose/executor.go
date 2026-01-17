package compose

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/models"
)

// Executor handles Docker Compose deployments
type Executor struct {
	tempDir     string
	networkName string
}

// NewExecutor creates a new compose executor
func NewExecutor(tempDir, networkName string) *Executor {
	return &Executor{
		tempDir:     tempDir,
		networkName: networkName,
	}
}

// writeComposeFile writes compose content to a temp file and returns the path
func (e *Executor) writeComposeFile(siteID uuid.UUID, content string) (string, error) {
	// Create site-specific temp directory
	siteDir := filepath.Join(e.tempDir, "compose", siteID.String())
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create compose temp dir: %w", err)
	}

	composePath := filepath.Join(siteDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	return composePath, nil
}

// composeUp runs docker compose up -d for the given compose file
func (e *Executor) composeUp(ctx context.Context, composePath, projectName string) ([]byte, error) {
	args := []string{
		"compose",
		"-f", composePath,
		"-p", projectName,
		"up", "-d",
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("docker compose up failed: %w", err)
	}

	return output, nil
}

// composeDown runs docker compose down for the given compose file
func (e *Executor) composeDown(ctx context.Context, composePath, projectName string) ([]byte, error) {
	args := []string{
		"compose",
		"-f", composePath,
		"-p", projectName,
		"down",
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("docker compose down failed: %w", err)
	}

	return output, nil
}

// cleanup removes the temp compose directory for a site
func (e *Executor) cleanup(siteID uuid.UUID) error {
	siteDir := filepath.Join(e.tempDir, "compose", siteID.String())
	if err := os.RemoveAll(siteDir); err != nil {
		return fmt.Errorf("failed to cleanup compose temp dir: %w", err)
	}
	return nil
}

// DeploySite deploys a site using Docker Compose
func (e *Executor) DeploySite(ctx context.Context, req *models.DeployRequest) (*models.DeployResponse, error) {
	// Validate this is a compose deployment
	if !req.IsCompose() {
		return nil, fmt.Errorf("not a compose deployment")
	}

	projectName := fmt.Sprintf("archon-%s", req.Name)

	// Write compose file to temp location
	composePath, err := e.writeComposeFile(req.ID, req.ComposeContent)
	if err != nil {
		return &models.DeployResponse{
			SiteID:  req.ID,
			Status:  models.SiteStatusFailed,
			Message: fmt.Sprintf("Failed to write compose file: %v", err),
		}, nil
	}

	// If redeploying, bring down existing services first
	// Ignore errors since services may not exist
	_, _ = e.composeDown(ctx, composePath, projectName)

	// Run docker compose up
	output, err := e.composeUp(ctx, composePath, projectName)
	if err != nil {
		return &models.DeployResponse{
			SiteID:  req.ID,
			Status:  models.SiteStatusFailed,
			Message: fmt.Sprintf("Docker compose up failed: %v\nOutput: %s", err, string(output)),
		}, nil
	}

	// Cleanup temp files after successful deployment
	if cleanupErr := e.cleanup(req.ID); cleanupErr != nil {
		// Log but don't fail - deployment succeeded
		// log.Printf("Warning: failed to cleanup compose temp files: %v", cleanupErr)
	}

	return &models.DeployResponse{
		SiteID:  req.ID,
		Status:  models.SiteStatusRunning,
		Message: "Compose deployment successful",
	}, nil
}

// StopSite stops a compose deployment by project name
func (e *Executor) StopSite(ctx context.Context, siteID uuid.UUID, siteName string) error {
	projectName := fmt.Sprintf("archon-%s", siteName)

	// Use docker compose stop with just project name (no file needed)
	args := []string{
		"compose",
		"-p", projectName,
		"stop",
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose stop failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// DeleteSite removes a compose deployment by project name
func (e *Executor) DeleteSite(ctx context.Context, siteID uuid.UUID, siteName string) error {
	projectName := fmt.Sprintf("archon-%s", siteName)

	// Use docker compose down with just project name
	args := []string{
		"compose",
		"-p", projectName,
		"down",
		"--volumes",       // Remove named volumes
		"--remove-orphans",
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose down failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetStatus returns the status of a compose deployment
func (e *Executor) GetStatus(ctx context.Context, siteID uuid.UUID, siteName string) (*models.SiteStatusResponse, error) {
	projectName := fmt.Sprintf("archon-%s", siteName)

	// Use docker compose ps to check status
	args := []string{
		"compose",
		"-p", projectName,
		"ps",
		"--format", "json",
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If command fails, services probably don't exist
		return &models.SiteStatusResponse{
			SiteID:    siteID,
			Status:    models.SiteStatusInactive,
			IsRunning: false,
			Message:   "Compose services not found",
		}, nil
	}

	// If we got output, services exist
	// Check if any are running (simplified check)
	isRunning := len(output) > 0 && string(output) != "[]\n"
	status := models.SiteStatusStopped
	if isRunning {
		status = models.SiteStatusRunning
	}

	return &models.SiteStatusResponse{
		SiteID:    siteID,
		Status:    status,
		IsRunning: isRunning,
	}, nil
}
