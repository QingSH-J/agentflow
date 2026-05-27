package store

import (
	"context"
	"encoding/json"
)

type CreateStepParams struct {
	ID          string
	RunID       string
	StepKey     string
	StepType    string
	Input       json.RawMessage
	MaxAttempts int
	DependsOn   json.RawMessage
}

func (s *Store) CreateStep(ctx context.Context, params CreateStepParams) (*Step, error) {
	const query = `
		INSERT INTO steps (
			id,
			run_id,
			step_key,
			step_type,
			status,
			input,
			max_attempts,
			depends_on
			)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING
			id,
			run_id,
			step_key,
			step_type,
			status,
			input,
			output,
			error,
			attempt_count,
			max_attempts,
			depends_on,
			created_at,
			started_at,
			finished_at
			`
	var step Step
	err := s.db.QueryRow(ctx, query,
		params.ID,
		params.RunID,
		params.StepKey,
		params.StepType,
		StepStatusPending,
		params.Input,
		params.MaxAttempts,
		params.DependsOn,
	).Scan(
		&step.ID,
		&step.RunID,
		&step.StepKey,
		&step.StepType,
		&step.Status,
		&step.Input,
		&step.Output,
		&step.Error,
		&step.AttemptCount,
		&step.MaxAttempts,
		&step.DependsOn,
		&step.CreatedAt,
		&step.StartedAt,
		&step.FinishedAt,
	)
	if err != nil {
		return nil, err
	}
	return &step, nil
}

// Get Pending Steps By RunID
func (s *Store) ListPendingStepsByRun(ctx context.Context, runID string) ([]Step, error) {
	const query = `
		SELECT
			id,
			run_id,
			step_key,
			step_type,
			status,
			input,
			output,
			error,
			attempt_count,
			max_attempts,
			depends_on,
			created_at,
			started_at,
			finished_at
		FROM steps
		WHERE run_id = $1 AND status = $2
		ORDER BY created_at ASC
	`
	rows, err := s.db.Query(ctx, query, runID, StepStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var steps []Step
	for rows.Next() {
		var step Step
		err := rows.Scan(
			&step.ID,
			&step.RunID,
			&step.StepKey,
			&step.StepType,
			&step.Status,
			&step.Input,
			&step.Output,
			&step.Error,
			&step.AttemptCount,
			&step.MaxAttempts,
			&step.DependsOn,
			&step.CreatedAt,
			&step.StartedAt,
			&step.FinishedAt,
		)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

//Get Step By ID
func (s *Store) GetStep(ctx context.Context, stepID string) (*Step, error) {
	const query = `
		SELECT
			id,
			run_id,
			step_key,
			step_type,
			status,
			input,
			output,
			error,
			attempt_count,
			max_attempts,
			depends_on,
			created_at,
			started_at,
			finished_at
		FROM steps
		WHERE id = $1
	`
	var step Step
	err := s.db.QueryRow(ctx, query, stepID).Scan(
		&step.ID,
		&step.RunID,
		&step.StepKey,
		&step.StepType,
		&step.Status,
		&step.Input,
		&step.Output,
		&step.Error,
		&step.AttemptCount,
		&step.MaxAttempts,
		&step.DependsOn,
		&step.CreatedAt,
		&step.StartedAt,
		&step.FinishedAt,
	)
	if err != nil {
		return nil, err
	}
	return &step, nil
}