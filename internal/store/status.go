package store

import (
	"context"
	"encoding/json"
)

func (s *Store) MarkRunRunning(ctx context.Context, runID string) error {
	const query = `
		UPDATE runs
		SET status = $2,
			updated_at = now()
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, runID, RunStatusRunning)
	return err
}

func (s *Store) MarkRunCompleted(ctx context.Context, runID string, output json.RawMessage) error {
	const query = `
		UPDATE runs
		SET status = $2,
			output = $3,
			updated_at = now()
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, runID, RunStatusCompleted, output)
	return err
}

func (s *Store) MarkRunFailed(ctx context.Context, runID string, message string) error {
	const query = `
		UPDATE runs
		SET status = $2,
			error = $3,
			updated_at = now()
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, runID, RunStatusFailed, message)
	return err
}

func (s *Store) MarkStepQueued(ctx context.Context, stepID string) error {
	const query = `
		UPDATE steps
		SET status = $2
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, stepID, StepStatusQueued)
	return err
}

func (s *Store) MarkStepRunning(ctx context.Context, stepID string) error {
	const query = `
		UPDATE steps
		SET status = $2,
			started_at = now()
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, stepID, StepStatusRunning)
	return err
}

func (s *Store) MarkStepCompleted(ctx context.Context, stepID string, output json.RawMessage) error {
	const query = `
		UPDATE steps
		SET status = $2,
			output = $3,
			finished_at = now()
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, stepID, StepStatusCompleted, output)
	return err
}

func (s *Store) MarkStepFailed(ctx context.Context, stepID string, message string) error {
	const query = `
		UPDATE steps
		SET status = $2,
			error = $3,
			finished_at = now()
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, stepID, StepStatusFailed, message)
	return err
}