package stages

import (
	"context"
	"fmt"

	"github.com/BlueBeard63/archon-node/internal/pipeline"
)

// ValidationStage validates the deployment request
type ValidationStage struct {
	pipeline.BaseStage
}

// NewValidationStage creates a new validation stage
func NewValidationStage() *ValidationStage {
	return &ValidationStage{
		BaseStage: pipeline.NewBaseStage("validation"),
	}
}

// Execute validates the deployment request
func (s *ValidationStage) Execute(ctx context.Context, state *pipeline.DeploymentState) error {
	req := state.Request

	// Basic validation
	if req.Name == "" {
		return fmt.Errorf("site name is required")
	}

	if len(req.DomainMappings) == 0 {
		return fmt.Errorf("at least one domain mapping is required")
	}

	// Type-specific validation
	if req.IsCompose() {
		if req.ComposeContent == "" {
			return fmt.Errorf("compose content is required for compose deployments")
		}
	} else {
		if req.Docker.Image == "" {
			return fmt.Errorf("docker image is required for container deployments")
		}
	}

	return nil
}
