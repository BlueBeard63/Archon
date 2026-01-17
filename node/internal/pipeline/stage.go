package pipeline

import "context"

// Stage represents a single step in the deployment pipeline.
// Each stage can execute work and optionally rollback on failure.
type Stage interface {
	// Name returns a human-readable identifier for this stage
	Name() string

	// Execute performs the stage's work, modifying state as needed.
	// Returns an error if the stage fails.
	Execute(ctx context.Context, state *DeploymentState) error

	// Rollback undoes any changes made by Execute.
	// Called in reverse order when a later stage fails.
	// Should be idempotent and never return errors that stop rollback.
	Rollback(ctx context.Context, state *DeploymentState) error
}

// BaseStage provides a no-op Rollback for stages that don't need cleanup.
// Embed this in stages that only validate or check state without side effects.
type BaseStage struct {
	name string
}

// NewBaseStage creates a new BaseStage with the given name
func NewBaseStage(name string) BaseStage {
	return BaseStage{name: name}
}

// Name returns the stage name
func (s BaseStage) Name() string {
	return s.name
}

// Rollback is a no-op for stages without side effects
func (s BaseStage) Rollback(ctx context.Context, state *DeploymentState) error {
	return nil
}
