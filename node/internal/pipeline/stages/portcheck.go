package stages

import (
	"context"

	"github.com/BlueBeard63/archon-node/internal/docker"
	"github.com/BlueBeard63/archon-node/internal/pipeline"
)

// PortCheckStage checks for port conflicts before deployment
type PortCheckStage struct {
	pipeline.BaseStage
	dockerClient *docker.Client
}

// NewPortCheckStage creates a new port check stage
func NewPortCheckStage(dockerClient *docker.Client) *PortCheckStage {
	return &PortCheckStage{
		BaseStage:    pipeline.NewBaseStage("port-check"),
		dockerClient: dockerClient,
	}
}

// Execute checks for port conflicts
func (s *PortCheckStage) Execute(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	// Extract host ports from domain mappings
	hostPorts := make([]int, 0, len(req.DomainMappings))
	for _, mapping := range req.DomainMappings {
		hostPort := mapping.Port
		if mapping.HostPort > 0 {
			hostPort = mapping.HostPort
		}
		hostPorts = append(hostPorts, hostPort)
	}

	// Check for conflicts (excludes this site's existing container)
	return s.dockerClient.CheckPortConflicts(ctx, hostPorts, req.ID)
}
