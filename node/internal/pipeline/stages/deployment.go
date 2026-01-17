package stages

import (
	"context"
	"fmt"
	"log"

	"github.com/BlueBeard63/archon-node/internal/compose"
	"github.com/BlueBeard63/archon-node/internal/docker"
	"github.com/BlueBeard63/archon-node/internal/models"
	"github.com/BlueBeard63/archon-node/internal/pipeline"
)

// DeploymentStage deploys containers or compose stacks
type DeploymentStage struct {
	name            string
	dockerClient    *docker.Client
	composeExecutor *compose.Executor
}

// NewDeploymentStage creates a new deployment stage
func NewDeploymentStage(dockerClient *docker.Client, composeExecutor *compose.Executor) *DeploymentStage {
	return &DeploymentStage{
		name:            "deployment",
		dockerClient:    dockerClient,
		composeExecutor: composeExecutor,
	}
}

// Name returns the stage name
func (s *DeploymentStage) Name() string {
	return s.name
}

// Execute deploys the site (container or compose based on type)
func (s *DeploymentStage) Execute(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	if req.IsCompose() {
		return s.deployCompose(ctx, state)
	}
	return s.deployContainer(ctx, state)
}

// deployContainer deploys a single Docker container
func (s *DeploymentStage) deployContainer(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	log.Printf("[DEPLOY] Deploying container: image=%s", req.Docker.Image)

	resp, err := s.dockerClient.DeploySite(ctx, req, state.DataDir)
	if err != nil {
		return fmt.Errorf("failed to deploy container: %w", err)
	}

	if resp.Status == models.SiteStatusFailed {
		return fmt.Errorf("deployment failed: %s", resp.Message)
	}

	// Store container ID for potential rollback
	state.ContainerID = resp.ContainerID
	state.Response = resp

	log.Printf("[DEPLOY] Container deployed: %s", resp.ContainerID)
	return nil
}

// deployCompose deploys a Docker Compose stack
func (s *DeploymentStage) deployCompose(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	log.Printf("[DEPLOY] Deploying compose stack: %s", req.Name)

	resp, err := s.composeExecutor.DeploySite(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to deploy compose stack: %w", err)
	}

	if resp.Status == models.SiteStatusFailed {
		return fmt.Errorf("compose deployment failed: %s", resp.Message)
	}

	// Store project name for potential rollback
	state.ProjectName = fmt.Sprintf("archon-%s", req.Name)
	state.Response = resp

	log.Printf("[DEPLOY] Compose stack deployed: %s", state.ProjectName)
	return nil
}

// Rollback removes the deployed container or compose stack
func (s *DeploymentStage) Rollback(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	if req.IsCompose() && state.ProjectName != "" {
		log.Printf("[ROLLBACK] Removing compose stack: %s", state.ProjectName)
		if err := s.composeExecutor.DeleteSite(ctx, req.ID, req.Name); err != nil {
			log.Printf("[ROLLBACK] Warning: failed to remove compose stack: %v", err)
		}
		return nil
	}

	if state.ContainerID != "" {
		log.Printf("[ROLLBACK] Removing container: %s", state.ContainerID)
		if err := s.dockerClient.DeleteSite(ctx, req.ID); err != nil {
			log.Printf("[ROLLBACK] Warning: failed to remove container: %v", err)
		}
	}

	return nil
}
