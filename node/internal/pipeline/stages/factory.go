package stages

import (
	"github.com/BlueBeard63/archon-node/internal/compose"
	"github.com/BlueBeard63/archon-node/internal/docker"
	"github.com/BlueBeard63/archon-node/internal/pipeline"
	"github.com/BlueBeard63/archon-node/internal/proxy"
	"github.com/BlueBeard63/archon-node/internal/ssl"
)

// Dependencies holds all dependencies needed to create pipeline stages
type Dependencies struct {
	DockerClient    *docker.Client
	ComposeExecutor *compose.Executor
	ProxyManager    proxy.ProxyManager
	SSLManager      *ssl.Manager
}

// NewDeploymentPipeline creates the standard deployment pipeline
func NewDeploymentPipeline(deps *Dependencies) *pipeline.Pipeline {
	return pipeline.NewPipeline(
		NewValidationStage(),
		NewPortCheckStage(deps.DockerClient),
		NewSSLStage(deps.ProxyManager, deps.SSLManager),
		NewDeploymentStage(deps.DockerClient, deps.ComposeExecutor),
		NewProxyStage(deps.ProxyManager),
	)
}
