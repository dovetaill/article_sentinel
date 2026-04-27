# Dev Stop And Orgid Consistency Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `scripts/dev.sh stop` reliably kill stale `go run` temp binaries and verify the admin HTTP API contract consistently uses `orgid`.

**Architecture:** Extend stale process detection in `scripts/dev.sh` to recognize `/tmp/go-build.../exe/{server,worker,scheduler}` processes launched from the repo root, then lock the behavior with shell regression tests. Keep frontend API types/contracts on `orgid`, and verify there is no remaining `org_id` contract in `web/admin`.

**Tech Stack:** Bash, ripgrep, Go test runner, Vitest

---

### Task 1: Add stale temp-binary regression test

**Files:**
- Modify: `scripts/dev_test.sh`

**Step 1: Write the failing test**
- Add a shell test that starts a fake stale process with argv[0] set to `/tmp/go-build123/b001/exe/server` from the repo root, runs `scripts/dev.sh stop`, and expects the process to be terminated.

**Step 2: Run test to verify it fails**
- Run: `bash scripts/dev_test.sh`
- Expected: FAIL in the new stale temp-binary test because `dev.sh` does not yet match that process shape.

### Task 2: Implement stale process matching

**Files:**
- Modify: `scripts/dev.sh`

**Step 1: Write minimal implementation**
- Extend `stale_process_matches()` so repo-root processes whose argv[0] looks like `/tmp/go-build.../exe/server|worker|scheduler` are considered stale dev processes.

**Step 2: Run test to verify it passes**
- Run: `bash scripts/dev_test.sh`
- Expected: PASS

### Task 3: Verify `orgid` contract and full regression suite

**Files:**
- Inspect: `web/admin/src/services/*.ts`

**Step 1: Verify frontend contract**
- Run: `rg -n '\borg_id\b' web/admin/src`
- Expected: no matches

**Step 2: Run full verification**
- Run: `go test ./...`
- Expected: PASS
- Run: `cd web/admin && npm test -- --runInBand`
- Expected: PASS

### Task 4: Commit

**Step 1: Commit the validated change**
- Run: `git add docs/plans/2026-04-27-dev-stop-orgid-implementation.md scripts/dev.sh scripts/dev_test.sh`
- Run: `git commit -m "fix(dev): stop stale go-run temp processes"`
