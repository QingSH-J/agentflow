package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	RunStatusQueued    = "queued"
	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"

	StepStatusPending   = "pending"
	StepStatusQueued    = "queued"
	StepStatusRunning   = "running"
	StepStatusCompleted = "completed"
	StepStatusFailed    = "failed"
	StepStatusSkipped   = "skipped"
	StepStatusCancelled = "cancelled"
)

type Store struct {
	db *pgxpool.Pool
}

func Connect(ctx context.Context, databaseURL string) (*Store, error) {
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func New(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func (s *Store) Close() {
	s.db.Close()
}

type Run struct {
	ID           string          `json:"id"`
	WorkflowName string          `json:"workflow_name"`
	Status       string          `json:"status"`
	Input        json.RawMessage `json:"input"`
	Output       json.RawMessage `json:"output,omitempty"`
	Error        *string         `json:"error,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type Step struct {
	ID           string          `json:"id"`
	RunID        string          `json:"run_id"`
	StepKey      string          `json:"step_key"`
	StepType     string          `json:"step_type"`
	Status       string          `json:"status"`
	Input        json.RawMessage `json:"input"`
	Output       json.RawMessage `json:"output,omitempty"`
	Error        *string         `json:"error,omitempty"`
	AttemptCount int             `json:"attempt_count"`
	MaxAttempts  int             `json:"max_attempts"`
	DependsOn    json.RawMessage `json:"depends_on"`
	CreatedAt    time.Time       `json:"created_at"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
}

type Event struct {
	ID        int64           `json:"id"`
	RunID     string          `json:"run_id"`
	StepID    *string         `json:"step_id,omitempty"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}