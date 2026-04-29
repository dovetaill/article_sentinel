# ArticleInspect Outbox Upgrade Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a pragmatic articleinspect outbox skeleton, improve router queue-init observability, and update maintainer docs to match the refactored architecture.

**Architecture:** Keep the existing Asynq worker flow intact. Move task creation to a transaction that writes the task graph plus an outbox record, then add optimistic relay and scheduler-based retry around that persisted outbox state. Keep the implementation scoped to the `articleinspect` module rather than extracting a generic outbox framework.

**Tech Stack:** Go, GORM, Huma, Asynq, cron, Markdown docs.

---

### Task 1: Lock the new outbox contract with failing tests

**Files:**
- Modify: `internal/modules/articleinspect/articleinspect_test.go`
- Modify: `internal/api/register/router_test.go`
- Modify: `internal/scheduler/scheduler_test.go`

**Step 1: Write failing tests**

Add coverage for:

- task creation persists an outbox message with `pending` status
- optimistic relay success marks the message dispatched
- optimistic relay failure still returns created task and leaves pending outbox
- router logs dispatcher init failure
- scheduler registers/runs an outbox relay job when configured

**Step 2: Run tests to verify they fail**

Run:

```bash
go test ./internal/modules/articleinspect ./internal/api/register ./internal/scheduler -run 'Test(TaskCreation|TaskCreate|Router|RegisterJobs)' -v
```

**Step 3: Implement the minimal code to make them pass**

Do not widen scope beyond outbox skeleton, observability, and docs hooks.

### Task 2: Add the outbox model and transactional writer

**Files:**
- Modify: `internal/modules/articleinspect/model.go`
- Modify: `internal/app/bootstrap/schema.go`
- Create: `migrations/20260428_02_article_inspect_task_outbox.sql`
- Modify: `internal/modules/articleinspect/service_tasks.go`

**Step 1: Add the failing migration/model assumptions if needed in tests**

**Step 2: Implement the outbox record and transactional task writer**

Write task + task keywords + outbox row in one transaction.

**Step 3: Re-run targeted tests**

### Task 3: Add optimistic relay and scheduler retry skeleton

**Files:**
- Create: `internal/modules/articleinspect/task_outbox.go`
- Modify: `internal/modules/articleinspect/task_routes.go`
- Modify: `internal/api/register/router.go`
- Modify: `internal/scheduler/jobs.go`
- Modify: `internal/scheduler/scheduler.go`
- Modify: `pkg/config/config.go`
- Modify: `configs/config.example.yaml`

**Step 1: Keep handler thin**

Route should call service / relay orchestration rather than inline queue code.

**Step 2: Add optimistic relay and batch relay**

- single-message dispatch after create
- batch pending relay for scheduler retry

**Step 3: Add router observability**

Log queue client initialization failures explicitly.

**Step 4: Re-run focused tests**

### Task 4: Update README and maintainer docs

**Files:**
- Modify: `README.md`
- Modify: `docs/maintainer-development-flow.md`
- Modify: `docs/README.md`

**Step 1: Update architecture wording**

Reflect:
- route files + module wiring
- outbox ownership
- scheduler retry responsibility
- queue-init observability caveat

**Step 2: Re-read docs for stale `handler.go` references**

Remove or rewrite outdated guidance.

### Task 5: Full verification

**Files:**
- Verify only

**Step 1: Run targeted tests**

```bash
go test ./internal/modules/articleinspect ./internal/api/register ./internal/scheduler -v
```

**Step 2: Run full suite**

```bash
go test ./...
```

**Step 3: Inspect diff**

```bash
git status --short
git diff --stat
```
