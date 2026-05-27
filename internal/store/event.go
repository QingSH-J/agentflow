package store

import (
	"context"
	"encoding/json"
)

const (
	EventRunCreated    = "run.created"
	EventRunStarted    = "run.started"
	EventRunCompleted  = "run.completed"
	EventRunFailed     = "run.failed"
	EventStepQueued    = "step.queued"
	EventStepStarted   = "step.started"
	EventStepCompleted = "step.completed"
	EventStepFailed    = "step.failed"
	EventLLMStarted    = "llm.started"
	EventLLMCompleted  = "llm.completed"
	EventLLMFailed     = "llm.failed"
)

type CreateEventParams struct {
	RunID   string
	StepID  *string
	Type    string
	Payload json.RawMessage
}

func (s *Store) CreateEvent(ctx context.Context, params CreateEventParams) (*Event, error) {
	const query = `
		INSERT INTO events (
			run_id,
			step_id,
			type,
			payload
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			run_id,
			step_id,
			type,
			payload,
			created_at
	`

	var event Event
	err := s.db.QueryRow(
		ctx,
		query,
		params.RunID,
		params.StepID,
		params.Type,
		params.Payload,
	).Scan(
		&event.ID,
		&event.RunID,
		&event.StepID,
		&event.Type,
		&event.Payload,
		&event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Store) ListEventsByRun(ctx context.Context, runID string, afterID int64) ([]Event, error) {
	const query = `
		SELECT
			id,
			run_id,
			step_id,
			type,
			payload,
			created_at
		FROM events
		WHERE run_id = $1
			AND id > $2
		ORDER BY id ASC
	`

	rows, err := s.db.Query(ctx, query, runID, afterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.ID,
			&event.RunID,
			&event.StepID,
			&event.Type,
			&event.Payload,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}