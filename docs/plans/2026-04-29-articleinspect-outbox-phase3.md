# ArticleInspect Outbox Phase 3 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Upgrade the current articleinspect outbox skeleton into an operationally complete control plane with claim/lease, retry/backoff, dead-letter handling, cleanup, and manual recovery guidance.

**Architecture:** Keep the current module-local outbox approach. Extend the existing `xt_article_inspect_task_outbox` single-table model with claim/lease and terminal-state retention fields, centralize dispatch state transitions in `internal/modules/articleinspect/task_outbox.go`, and let `internal/scheduler` own relay and cleanup loops. Do not add new HTTP/API/UI surfaces in this phase.

**Tech Stack:** Go, GORM, Huma-adjacent module wiring, Asynq, cron, MySQL 5.7-compatible SQL, Markdown docs, SQL helper scripts.

---

### Task 1: Lock the phase-3 outbox state machine with failing tests

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Modify: `internal/scheduler/scheduler_test.go`

**Step 1: Write failing outbox lifecycle tests**

Add tests for:

- claim success writes `claimed` state and lease metadata
- expired `claimed` rows can be reclaimed
- retryable dispatch failure returns row to `pending` with `next_attempt_at`
- poison payload moves row to `dead_letter`
- rows over `max_attempts` move to `dead_letter`
- cleanup deletes expired `dispatched` / `dead_letter` rows

Target helper names to introduce in tests:

- `TestTaskOutboxRelayClaimsPendingMessage`
- `TestTaskOutboxRelayReclaimsExpiredLease`
- `TestTaskOutboxRelayRetryableFailureSchedulesNextAttempt`
- `TestTaskOutboxRelayMovesPoisonMessageToDeadLetter`
- `TestTaskOutboxCleanupDeletesExpiredTerminalRows`

**Step 2: Run focused tests to verify they fail**

Run:

```bash
go test ./internal/modules/articleinspect ./internal/scheduler -run 'TestTaskOutbox|TestArticleInspectTaskOutbox' -v
```

Expected:

- compile errors for missing fields / constants / config
- or failing assertions around missing `claimed`, `dead_letter`, `next_attempt_at`, cleanup behavior

**Step 3: Commit the red test changes**

```bash
git add internal/modules/articleinspect/articleinspect_test.go internal/scheduler/scheduler_test.go
git commit -m "test(articleinspect): lock outbox phase 3 state machine"
```

### Task 2: Extend the outbox schema, model, and constants

**Files:**
- Modify: `internal/modules/articleinspect/model.go`
- Modify: `internal/modules/articleinspect/constants.go`
- Create: `migrations/20260429_01_article_inspect_task_outbox_phase3.sql`
- Modify: `internal/app/bootstrap/schema.go`

**Step 1: Add schema fields and state constants**

Add constants for:

- `TaskOutboxStatusClaimed`
- `TaskOutboxStatusDeadLetter`
- retry error codes such as `TaskOutboxErrorDispatch`, `TaskOutboxErrorPayloadDecode`

Extend `InspectionTaskOutboxMessage` with:

- `ClaimedBy string`
- `ClaimedAt *time.Time`
- `ClaimUntil *time.Time`
- `NextAttemptAt *time.Time`
- `LastErrorCode string`
- `DeadLetteredAt *time.Time`
- `RetainedUntil *time.Time`

**Step 2: Add the migration**

Migration must:

- `ALTER TABLE xt_article_inspect_task_outbox` to add the new columns
- add helpful composite indexes for:
  - `status + next_attempt_at + id`
  - `status + claim_until + id`
  - `status + retained_until + id`

Keep SQL compatible with MySQL 5.7.

**Step 3: Re-run the focused tests**

Run:

```bash
go test ./internal/modules/articleinspect ./internal/scheduler -run 'TestTaskOutbox|TestArticleInspectTaskOutbox' -v
```

Expected:

- tests still fail, but now on behavior rather than missing fields

**Step 4: Commit**

```bash
git add internal/modules/articleinspect/model.go internal/modules/articleinspect/constants.go internal/app/bootstrap/schema.go migrations/20260429_01_article_inspect_task_outbox_phase3.sql
git commit -m "feat(articleinspect): extend outbox schema for lease and dead-letter"
```

### Task 3: Implement claim/lease, retry/backoff, and dead-letter transitions

**Files:**
- Modify: `internal/modules/articleinspect/task_outbox.go`
- Modify: `internal/modules/articleinspect/service_tasks.go`
- Modify: `internal/queue/asynq/articleinspect_dispatcher.go`
- Modify: `internal/modules/articleinspect/articleinspect_test.go`

**Step 1: Add explicit dispatch classification helpers**

In `internal/modules/articleinspect/task_outbox.go`, introduce focused helpers such as:

- `claimDispatchableMessage(...)`
- `completeDispatch(...)`
- `scheduleRetry(...)`
- `moveToDeadLetter(...)`
- `backoffForAttempt(...)`
- `isNonRetryableOutboxError(...)`

Suggested backoff policy:

- attempts `1..3` => `15s`
- attempts `4..10` => `1m`
- attempts `11+` => `5m`

**Step 2: Make optimistic relay share the same state machine**

`TryDispatchMessage` and scheduler-driven dispatch must both go through the same claim/dispatch path.

Requirements:

- optimistic relay should claim before dispatching
- dispatch success must clear claim fields and set `retained_until`
- retryable failures must clear claim fields and set `next_attempt_at`
- poison payload / unsupported message types must go directly to `dead_letter`
- duplicate `TaskID` enqueue conflicts from Asynq must still be treated as success

**Step 3: Run targeted tests**

Run:

```bash
go test ./internal/modules/articleinspect -run 'TestTaskOutbox|TestTaskCreate' -v
```

Expected:

- all new outbox state machine tests pass

**Step 4: Commit**

```bash
git add internal/modules/articleinspect/task_outbox.go internal/modules/articleinspect/service_tasks.go internal/queue/asynq/articleinspect_dispatcher.go internal/modules/articleinspect/articleinspect_test.go
git commit -m "feat(articleinspect): add outbox lease, retry, and dead-letter flow"
```

### Task 4: Add scheduler relay and cleanup control-plane jobs

**Files:**
- Modify: `internal/scheduler/jobs.go`
- Modify: `internal/scheduler/scheduler.go`
- Modify: `internal/scheduler/scheduler_test.go`
- Modify: `cmd/scheduler/main.go`
- Modify: `pkg/config/config.go`
- Modify: `configs/config.example.yaml`

**Step 1: Expand outbox config**

Add config fields under `queue.outbox`:

- `lease_duration_seconds`
- `max_attempts`
- `cleanup_spec`
- `dispatched_retention_hours`
- `dead_letter_retention_hours`

**Step 2: Add cleanup job seam**

Add an interface in `internal/scheduler/jobs.go` similar to:

```go
type ArticleInspectTaskOutboxCleaner interface {
    CleanupArticleInspectTaskOutbox(ctx context.Context, limit int) (int, error)
}
```

Add `NewArticleInspectTaskOutboxCleanupJob(...)`.

**Step 3: Register the new job**

In `internal/scheduler/scheduler.go`, register:

- relay job from `relay_spec`
- cleanup job from `cleanup_spec`

Only when outbox is enabled.

**Step 4: Run focused scheduler tests**

Run:

```bash
go test ./internal/scheduler -v
```

Expected:

- scheduler tests pass with relay and cleanup registration

**Step 5: Commit**

```bash
git add internal/scheduler/jobs.go internal/scheduler/scheduler.go internal/scheduler/scheduler_test.go cmd/scheduler/main.go pkg/config/config.go configs/config.example.yaml
git commit -m "feat(scheduler): add outbox relay and cleanup control-plane jobs"
```

### Task 5: Add manual recovery script and operations docs

**Files:**
- Create: `scripts/articleinspect_outbox_requeue.sql`
- Modify: `README.md`
- Modify: `docs/README.md`
- Modify: `docs/maintainer-development-flow.md`

**Step 1: Add a conservative SQL recovery template**

Create `scripts/articleinspect_outbox_requeue.sql` with commented templates for:

- requeue by outbox `id`
- requeue by `task_id`
- clear claim fields
- set `status = 'pending'`
- reset `next_attempt_at = NOW()`

Do not silently zero out `attempt_count`; preserve history by default.

**Step 2: Update docs**

Document:

- how to diagnose pending backlog
- how to detect expired claims
- how to inspect dead-letter rows
- how to use the requeue SQL safely
- what cleanup removes and what it must not remove

**Step 3: Run docs sanity and full tests**

Run:

```bash
go test ./...
```

Expected:

- full suite passes

**Step 4: Commit**

```bash
git add scripts/articleinspect_outbox_requeue.sql README.md docs/README.md docs/maintainer-development-flow.md
git commit -m "docs(articleinspect): document outbox operations and recovery flow"
```

### Task 6: Final verification and diff review

**Files:**
- Verify only

**Step 1: Run the full suite again**

```bash
go test ./...
```

Expected:

- all packages pass

**Step 2: Inspect worktree**

```bash
git status --short
git diff --stat HEAD~4..HEAD
```

Expected:

- empty worktree
- four coherent commits matching the intended phase-3 slice

**Step 3: Prepare review summary**

Summarize:

- state machine changes
- new scheduler jobs
- recovery script path
- config additions

---

Plan complete and saved to `docs/plans/2026-04-29-articleinspect-outbox-phase3.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
