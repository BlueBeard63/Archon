package pipeline

import (
	"time"

	"github.com/google/uuid"

	"github.com/BlueBeard63/archon-node/internal/models"
)

// DeploymentState holds all data needed during pipeline execution.
// Stages read input from Request and write results to intermediate fields.
type DeploymentState struct {
	// Input (set before pipeline execution)
	Request *models.DeployRequest
	DataDir string

	// Runtime tracking
	CurrentStage    string
	CompletedStages []string
	StartTime       time.Time

	// Intermediate results (set by stages)
	CertPath    string // SSL certificate path (set by SSLStage)
	KeyPath     string // SSL key path (set by SSLStage)
	ContainerID string // Docker container ID (set by DeploymentStage for container type)
	ProjectName string // Compose project name (set by DeploymentStage for compose type)

	// Output
	Response *models.DeployResponse
	Error    error

	// Progress callback (optional, for WebSocket progress updates)
	OnProgress func(stage, status, message string)
}

// NewDeploymentState creates a new state for a deployment request
func NewDeploymentState(req *models.DeployRequest, dataDir string) *DeploymentState {
	return &DeploymentState{
		Request:         req,
		DataDir:         dataDir,
		CompletedStages: make([]string, 0),
		StartTime:       time.Now(),
	}
}

// EmitProgress sends a progress update if a callback is registered
func (s *DeploymentState) EmitProgress(stage, status, message string) {
	if s.OnProgress != nil {
		s.OnProgress(stage, status, message)
	}
}

// SiteID returns the site ID from the request for convenience
func (s *DeploymentState) SiteID() uuid.UUID {
	return s.Request.ID
}

// IsCompose returns true if this is a compose deployment
func (s *DeploymentState) IsCompose() bool {
	return s.Request.IsCompose()
}
