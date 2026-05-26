# CODEX.md

This file guides Codex when working on this repository.

The project is a Go-based lightweight AI workflow runtime. The v1 goal is to let users define AI workflows in YAML, submit a run through an HTTP API, execute workflow steps through a Redis-backed worker queue, persist run/step state in PostgreSQL, and stream execution events back to clients.

The project should stay small, explicit, and verifiable. Do not turn v1 into a full distributed system, Kubernetes replacement, Temporal clone, or SaaS platform.

## Core Product Idea

Build a lightweight backend runtime for reliable AI workflows.

Users provide a YAML workflow definition. The system parses it into a DAG, creates a run instance, persists step state, schedules executable steps, sends step tasks to workers, executes LLM/tool steps, records events, retries failed steps when allowed, and exposes run status and progress events.

The essence of the project is not YAML parsing. YAML is only the declaration format. The real value is the runtime that turns a workflow definition into a recoverable, observable, executable state machine.

## V1 Scope

V1 should support:

- YAML-defined workflows
- Step dependencies through `depends_on`
- Basic DAG validation
- Run creation
- Step creation
- PostgreSQL-backed run and step state
- Redis/Asynq-backed task queue
- Worker process for step execution
- LLM step execution through OpenAI-compatible APIs
- Basic retry support
- Basic timeout support if simple to implement
- Event recording
- Run status query
- SSE endpoint for run events
- Docker Compose for local PostgreSQL and Redis
- BYOK model API keys through environment variables

V1 should not support:

- Multi-tenant accounts
- Billing
- Hosted SaaS behavior
- Kubernetes scheduling
- Raft, consensus, or distributed locks
- Complex visual workflow editor
- Complex permission system
- Full sandbox runtime
- Multi-region deployment
- Advanced memory system
- Complex prompt templating engine
- Complex plugin marketplace
- Premature abstractions for providers, tools, or schedulers

## Recommended Tech Stack

Use Go.

Use these components unless there is a clear reason not to:

- HTTP server: `gin` or standard `net/http`
- Queue: `github.com/hibiken/asynq`
- Redis client: Asynq's Redis integration, plus `go-redis` only if needed
- PostgreSQL driver: `github.com/jackc/pgx/v5/pgxpool`
- YAML parsing: `gopkg.in/yaml.v3`
- Config: environment variables, optionally `github.com/caarlos0/env/v11`
- IDs: `github.com/google/uuid`
- Logging: standard log or `zerolog`

Prefer boring, stable dependencies. Do not add frameworks unless they directly simplify the current task.

## Initial Project Structure

Keep this structure unless the codebase clearly outgrows it:

```text
cmd/
  server/
    main.go
  worker/
    main.go

internal/
  api/          # HTTP handlers and request/response types
  config/       # environment config loading
  store/        # PostgreSQL access
  queue/        # Asynq enqueue/client code
  worker/       # worker handlers and step execution dispatch
  workflow/     # YAML parser, validation, DAG helpers
  scheduler/    # state transition and ready-step scheduling
  llm/          # OpenAI-compatible client
  events/       # event recording and SSE helpers
  tools/        # built-in tool steps, initially minimal

examples/
  doc-summary/
    workflow.yaml

migrations/
  001_init.sql

docker-compose.yml
.env.example
README.md
CODEX.md
```

Do not split into microservices. `server` and `worker` are separate processes but share internal packages.

## Local Infrastructure

Use Docker Compose for local development:

- PostgreSQL 16
- Redis 7

Expected default connection values:

```env
DATABASE_URL=postgres://agentflow:agentflow@localhost:5433/agentflow?sslmode=disable
REDIS_ADDR=localhost:6379
OPENAI_API_KEY=
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
```

Prefer port `5433:5432` for PostgreSQL to avoid conflict with a local PostgreSQL instance running on `5432`.

## Core Data Model

Start with these tables:

- `runs`
- `steps`
- `events`
- `artifacts`

Minimum `runs` fields:

```text
id
workflow_name
status
input
output
error
created_at
updated_at
```

Minimum `steps` fields:

```text
id
run_id
step_key
step_type
status
input
output
error
attempt_count
max_attempts
depends_on
created_at
started_at
finished_at
```

Minimum `events` fields:

```text
id
run_id
step_id
type
payload
created_at
```

Use JSONB for flexible `input`, `output`, `payload`, and `depends_on`.

## Workflow YAML Format

V1 YAML should remain simple:

```yaml
name: doc_summary

steps:
  - id: summarize
    type: llm
    prompt: |
      Summarize the following document:

      {{ input.document }}

  - id: extract_key_points
    type: llm
    depends_on: [summarize]
    prompt: |
      Extract 5 key points from this summary:

      {{ steps.summarize.output }}
```

Supported fields:

```text
id
type
depends_on
prompt
retry
timeout
```

Do not design a large DSL in v1. Do not build a full template engine unless needed. Basic replacement for `{{ input.xxx }}` and `{{ steps.step_id.output }}` is enough for the initial implementation.

## Workflow Validation

The `workflow` package should validate:

- Workflow name is present
- At least one step exists
- Every step has an id
- Every step has a type
- Step ids are unique
- Every dependency references an existing step
- The dependency graph is acyclic

If validation fails, return clear errors.

## Runtime State Model

Use explicit statuses.

Run statuses:

```text
queued
running
completed
failed
cancelled
```

Step statuses:

```text
pending
queued
running
completed
failed
skipped
cancelled
```

Avoid adding more statuses until necessary.

## Scheduling Logic

The scheduler should be simple.

When a run is created:

1. Persist the run.
2. Persist all steps as `pending`.
3. Find steps with no dependencies.
4. Mark them as `queued`.
5. Enqueue them as Asynq tasks.

After a step completes:

1. Persist the step output.
2. Mark the step as `completed`.
3. Record an event.
4. Find pending dependent steps whose dependencies are all completed.
5. Mark those steps as `queued`.
6. Enqueue them.
7. If all steps are completed, mark the run as `completed`.

After a step fails:

1. Increment attempt count.
2. If attempts remain, re-enqueue according to retry policy.
3. If attempts are exhausted, mark the step as `failed`.
4. Mark the run as `failed` for v1.
5. Record an event.

Do not build complex branching, compensation, or human-in-the-loop in v1.

## Queue Design

Use one Asynq task type initially:

```text
step:execute
```

Task payload should contain only stable identifiers:

```json
{
  "run_id": "run_xxx",
  "step_id": "step_xxx"
}
```

Workers should fetch authoritative state from PostgreSQL. Do not put full workflow definitions or large prompts into the queue payload unless there is a measured need.

## API Design

Start with these endpoints:

```http
POST /runs
GET /runs/:id
GET /runs/:id/events
POST /runs/:id/cancel
GET /health
```

Initial `POST /runs` request:

```json
{
  "workflow_path": "examples/doc-summary/workflow.yaml",
  "input": {
    "document": "..."
  }
}
```

This is acceptable for local v1. Later, this can evolve into workflow registration or inline workflow definitions.

Initial response:

```json
{
  "run_id": "run_xxx",
  "status": "queued"
}
```

`GET /runs/:id` should return run status and step statuses.

`GET /runs/:id/events` should use SSE. V1 can poll the database for new events every second. Do not introduce Redis Pub/Sub until needed.

## LLM Provider Policy

Use BYOK by default.

The user provides their own API key through environment variables. The project should not pay for model usage or implement billing in v1.

Support OpenAI-compatible APIs first:

```env
OPENAI_API_KEY=
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o-mini
```

Do not implement multi-provider routing in the first pass. Keep the `llm` package small, but avoid hardcoding in a way that prevents later OpenAI-compatible base URL support.

## Events

Record meaningful events:

```text
run.created
run.started
run.completed
run.failed
step.queued
step.started
step.completed
step.failed
llm.started
llm.completed
llm.failed
```

For v1, do not stream every token unless the code is already simple. Step-level progress events are enough. Token streaming can be added later.

## Testing Strategy

Prefer small tests around core logic.

Important test areas:

- YAML parsing
- Workflow validation
- Duplicate step ids
- Unknown dependencies
- Cycle detection
- Scheduler ready-step selection
- Step status transitions
- Prompt rendering for basic variables

Do not require full Docker integration tests for every small change. Use unit tests first. Add integration tests only where valuable.

## Implementation Order

Follow this order:

1. Project scaffold
   - Verify: `go test ./...` passes.

2. Config loading and Docker Compose
   - Verify: PostgreSQL and Redis start locally.

3. Database migration
   - Verify: migration applies through `psql`.

4. Workflow parser and validation
   - Verify: unit tests for valid workflow, duplicate ids, missing dependency, cycle.

5. Store layer for runs, steps, events
   - Verify: basic create/read/update tests or manual local checks.

6. Queue client and worker skeleton
   - Verify: enqueue a fake `step:execute` task and worker logs it.

7. `POST /runs`
   - Verify: creates run, creates steps, queues root steps.

8. Worker executes fake step
   - Verify: pending -> queued -> running -> completed.

9. Scheduler advances dependent steps
   - Verify: a multi-step DAG completes in order.

10. LLM step execution
    - Verify: one simple LLM step stores output.

11. `GET /runs/:id`
    - Verify: returns run and step status.

12. `GET /runs/:id/events`
    - Verify: client receives SSE events.

Do not skip directly to LLM execution before the run/step state machine works with fake steps.

## Coding Guidelines for Codex

These guidelines are mandatory for this repository.

### Think Before Coding

Before implementing:

- State assumptions explicitly.
- If multiple interpretations exist, present them instead of silently choosing.
- If a simpler approach exists, say so.
- If something is unclear and affects correctness, ask before editing.

For trivial tasks, do not over-explain.

### Simplicity First

Write the minimum code that solves the current task.

Avoid:

- Speculative abstractions
- Generic interfaces before there are multiple implementations
- Configuration that was not requested
- Feature flags that are not needed
- Large helper packages for one call site
- Complex error handling for impossible scenarios

If a solution becomes much larger than expected, stop and simplify.

### Surgical Changes

Touch only what is required.

When editing existing code:

- Do not refactor unrelated code.
- Do not improve adjacent formatting unless necessary.
- Match existing project style.
- Remove imports, variables, and functions made unused by your own changes.
- Do not delete unrelated dead code unless explicitly asked.

Every changed line should map directly to the current task.

### Goal-Driven Execution

For multi-step tasks, write a brief plan with verification:

```text
1. Implement X -> verify with Y
2. Implement Z -> verify with go test ./...
```

When fixing bugs:

1. Reproduce with a test if practical.
2. Make the smallest fix.
3. Run the relevant tests.
4. Report what was changed and how it was verified.

### Verification

Prefer running:

```bash
go test ./...
go vet ./...
```

Run narrower commands when faster and sufficient.

If a command cannot be run because dependencies or services are missing, state that explicitly and provide the exact command the user should run locally.

## Boundaries

Do not:

- Rebuild Temporal.
- Add a custom distributed consensus system.
- Build a UI before the backend runtime works.
- Add user accounts or billing.
- Implement every LLM provider at once.
- Add Kubernetes deployment manifests in v1.
- Add a complex plugin architecture.
- Create a generic workflow language beyond the current YAML needs.

Do:

- Keep the runtime understandable.
- Keep state transitions explicit.
- Keep queue payloads small.
- Persist state before relying on it.
- Make failures visible through events.
- Prefer tests around DAG and scheduler behavior.

## Current Mental Model

The system should be understood as:

```text
Workflow YAML
  -> parsed Definition
  -> validated DAG
  -> persisted Run + Steps
  -> scheduler finds ready steps
  -> queue dispatches step tasks
  -> worker executes step
  -> store records output and events
  -> scheduler advances next steps
  -> API/SSE exposes progress
```

This is the v1 system. Keep implementation aligned with this model.
