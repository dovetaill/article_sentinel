-- ArticleInspect outbox manual recovery templates.
-- Use with care in a transaction, and always inspect rows before updating them.
-- Default policy: preserve attempt_count history; do not silently zero it out.

-- 1) Inspect pending backlog.
SELECT id, orgid, task_id, status, attempt_count, next_attempt_at, last_error_code, update_at
FROM xt_article_inspect_task_outbox
WHERE status = 'pending'
ORDER BY next_attempt_at IS NULL DESC, next_attempt_at ASC, id ASC
LIMIT 200;

-- 2) Inspect expired claims that are eligible for reclaim.
SELECT id, orgid, task_id, status, claimed_by, claimed_at, claim_until, attempt_count, last_error_code
FROM xt_article_inspect_task_outbox
WHERE status = 'claimed'
  AND claim_until IS NOT NULL
  AND claim_until < UTC_TIMESTAMP()
ORDER BY claim_until ASC, id ASC
LIMIT 200;

-- 3) Inspect dead-letter rows before manual recovery.
SELECT id, orgid, task_id, status, attempt_count, last_error_code, last_error, dead_lettered_at, retained_until
FROM xt_article_inspect_task_outbox
WHERE status = 'dead_letter'
ORDER BY dead_lettered_at DESC, id DESC
LIMIT 200;

-- 4) Requeue one row by outbox id.
-- Replace the sample id before running.
START TRANSACTION;
SELECT id, orgid, task_id, status, attempt_count, last_error_code, last_error
FROM xt_article_inspect_task_outbox
WHERE id = 12345
FOR UPDATE;

UPDATE xt_article_inspect_task_outbox
SET status = 'pending',
    claimed_by = '',
    claimed_at = NULL,
    claim_until = NULL,
    next_attempt_at = UTC_TIMESTAMP(),
    dead_lettered_at = NULL,
    retained_until = NULL,
    dispatched_at = NULL,
    update_at = UTC_TIMESTAMP()
WHERE id = 12345;
COMMIT;

-- 5) Requeue rows by task_id after inspection.
-- Replace the sample task id before running.
START TRANSACTION;
SELECT id, orgid, task_id, status, attempt_count, last_error_code, last_error
FROM xt_article_inspect_task_outbox
WHERE task_id = 67890
ORDER BY id ASC
FOR UPDATE;

UPDATE xt_article_inspect_task_outbox
SET status = 'pending',
    claimed_by = '',
    claimed_at = NULL,
    claim_until = NULL,
    next_attempt_at = UTC_TIMESTAMP(),
    dead_lettered_at = NULL,
    retained_until = NULL,
    dispatched_at = NULL,
    update_at = UTC_TIMESTAMP()
WHERE task_id = 67890
  AND status IN ('pending', 'claimed', 'dead_letter');
COMMIT;

-- 6) If you only need to clear a stuck lease, prefer this narrower update.
-- Replace the sample id before running.
START TRANSACTION;
SELECT id, orgid, task_id, status, claimed_by, claim_until
FROM xt_article_inspect_task_outbox
WHERE id = 12345
FOR UPDATE;

UPDATE xt_article_inspect_task_outbox
SET status = 'pending',
    claimed_by = '',
    claimed_at = NULL,
    claim_until = NULL,
    next_attempt_at = UTC_TIMESTAMP(),
    update_at = UTC_TIMESTAMP()
WHERE id = 12345
  AND status = 'claimed';
COMMIT;
