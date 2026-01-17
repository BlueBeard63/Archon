package pipeline

import (
	"context"
	"fmt"
	"log"
)

// Pipeline executes a sequence of stages with automatic rollback on failure.
type Pipeline struct {
	stages []Stage
}

// NewPipeline creates a new pipeline with the given stages.
// Stages are executed in order; rollback happens in reverse order.
func NewPipeline(stages ...Stage) *Pipeline {
	return &Pipeline{
		stages: stages,
	}
}

// Execute runs all stages in sequence.
// If any stage fails, previously executed stages are rolled back in reverse order.
// Returns the first error encountered, or nil if all stages succeed.
func (p *Pipeline) Execute(ctx context.Context, state *DeploymentState) (err error) {
	var executed []Stage

	// Defer rollback in case of error (including panic)
	defer func() {
		if err != nil {
			p.rollback(ctx, state, executed)
		}
	}()

	for _, stage := range p.stages {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
		}

		state.CurrentStage = stage.Name()
		state.EmitProgress(stage.Name(), "started", fmt.Sprintf("Starting %s", stage.Name()))

		if err = stage.Execute(ctx, state); err != nil {
			state.Error = err
			state.EmitProgress(stage.Name(), "failed", fmt.Sprintf("Failed: %v", err))
			err = fmt.Errorf("stage %s failed: %w", stage.Name(), err)
			return
		}

		executed = append(executed, stage)
		state.CompletedStages = append(state.CompletedStages, stage.Name())
		state.EmitProgress(stage.Name(), "completed", fmt.Sprintf("Completed %s", stage.Name()))
	}

	return nil
}

// rollback calls Rollback on executed stages in reverse order.
// Logs errors but continues rollback for remaining stages.
func (p *Pipeline) rollback(ctx context.Context, state *DeploymentState, executed []Stage) {
	for i := len(executed) - 1; i >= 0; i-- {
		stage := executed[i]
		state.EmitProgress(stage.Name(), "rollback", fmt.Sprintf("Rolling back %s", stage.Name()))

		if rollbackErr := stage.Rollback(ctx, state); rollbackErr != nil {
			// Log but continue - don't let rollback errors stop other rollbacks
			log.Printf("[ROLLBACK ERROR] Stage %s: %v", stage.Name(), rollbackErr)
		}
	}
}

// Stages returns the list of stages in this pipeline (for testing/inspection)
func (p *Pipeline) Stages() []Stage {
	return p.stages
}
