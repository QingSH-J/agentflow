package store

import (
	"context"
	"encoding/json"
)

type CreateRunParams struct {
	ID           string
	WorkflowName string
	Input        json.RawMessage
}

func (s *Store) CreateRun(ctx context.Context, params CreateRunParams) (*Run, error) {
	const query = `
		INSERT INTO runs (
			id,
			workflow_name,
			status,
			input
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			workflow_name,
			status,
			input,
			output,
			error,
			created_at,
			updated_at
			`
	var run Run
	err := s.db.QueryRow(ctx, query, params.ID, params.WorkflowName, RunStatusQueued, params.Input).Scan(
		&run.ID,
		&run.WorkflowName,
		&run.Status,
		&run.Input,
		&run.Output,
		&run.Error,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *Store) GetRun(ctx context.Context, id string) (*Run, error) {
	const query = `
		SELECT
			id,
			workflow_name,
			status,
			input,
			output,
			error,
			created_at,
			updated_at
		FROM runs
		WHERE id = $1
		`
	var run Run
	err := s.db.QueryRow(ctx, query, id).Scan(
		&run.ID,
		&run.WorkflowName,
		&run.Status,
		&run.Input,
		&run.Output,
		&run.Error,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (s *Store) ListStepByRun(ctx context.Context, runID string) ([]Step, error) {
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
		WHERE run_id = $1
		ORDER BY created_at ASC
	`
	rows, err := s.db.Query(ctx, query, runID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return steps, nil
}