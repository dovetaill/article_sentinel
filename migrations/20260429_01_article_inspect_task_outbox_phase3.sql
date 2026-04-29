SET @has_outbox_claimed_by_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND column_name = 'claimed_by'
);

SET @add_outbox_claimed_by_column := IF(
  @has_outbox_claimed_by_column > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD COLUMN `claimed_by` VARCHAR(64) NOT NULL DEFAULT '''''' AFTER `attempt_count`'
);

PREPARE article_inspect_add_outbox_claimed_by_column FROM @add_outbox_claimed_by_column;
EXECUTE article_inspect_add_outbox_claimed_by_column;
DEALLOCATE PREPARE article_inspect_add_outbox_claimed_by_column;

SET @has_outbox_claimed_at_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND column_name = 'claimed_at'
);

SET @add_outbox_claimed_at_column := IF(
  @has_outbox_claimed_at_column > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD COLUMN `claimed_at` DATETIME NULL AFTER `claimed_by`'
);

PREPARE article_inspect_add_outbox_claimed_at_column FROM @add_outbox_claimed_at_column;
EXECUTE article_inspect_add_outbox_claimed_at_column;
DEALLOCATE PREPARE article_inspect_add_outbox_claimed_at_column;

SET @has_outbox_claim_until_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND column_name = 'claim_until'
);

SET @add_outbox_claim_until_column := IF(
  @has_outbox_claim_until_column > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD COLUMN `claim_until` DATETIME NULL AFTER `claimed_at`'
);

PREPARE article_inspect_add_outbox_claim_until_column FROM @add_outbox_claim_until_column;
EXECUTE article_inspect_add_outbox_claim_until_column;
DEALLOCATE PREPARE article_inspect_add_outbox_claim_until_column;

SET @has_outbox_next_attempt_at_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND column_name = 'next_attempt_at'
);

SET @add_outbox_next_attempt_at_column := IF(
  @has_outbox_next_attempt_at_column > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD COLUMN `next_attempt_at` DATETIME NULL AFTER `claim_until`'
);

PREPARE article_inspect_add_outbox_next_attempt_at_column FROM @add_outbox_next_attempt_at_column;
EXECUTE article_inspect_add_outbox_next_attempt_at_column;
DEALLOCATE PREPARE article_inspect_add_outbox_next_attempt_at_column;

SET @has_outbox_last_error_code_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND column_name = 'last_error_code'
);

SET @add_outbox_last_error_code_column := IF(
  @has_outbox_last_error_code_column > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD COLUMN `last_error_code` VARCHAR(64) NOT NULL DEFAULT '''''' AFTER `last_error`'
);

PREPARE article_inspect_add_outbox_last_error_code_column FROM @add_outbox_last_error_code_column;
EXECUTE article_inspect_add_outbox_last_error_code_column;
DEALLOCATE PREPARE article_inspect_add_outbox_last_error_code_column;

SET @has_outbox_dead_lettered_at_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND column_name = 'dead_lettered_at'
);

SET @add_outbox_dead_lettered_at_column := IF(
  @has_outbox_dead_lettered_at_column > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD COLUMN `dead_lettered_at` DATETIME NULL AFTER `last_attempt_at`'
);

PREPARE article_inspect_add_outbox_dead_lettered_at_column FROM @add_outbox_dead_lettered_at_column;
EXECUTE article_inspect_add_outbox_dead_lettered_at_column;
DEALLOCATE PREPARE article_inspect_add_outbox_dead_lettered_at_column;

SET @has_outbox_retained_until_column := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND column_name = 'retained_until'
);

SET @add_outbox_retained_until_column := IF(
  @has_outbox_retained_until_column > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD COLUMN `retained_until` DATETIME NULL AFTER `dispatched_at`'
);

PREPARE article_inspect_add_outbox_retained_until_column FROM @add_outbox_retained_until_column;
EXECUTE article_inspect_add_outbox_retained_until_column;
DEALLOCATE PREPARE article_inspect_add_outbox_retained_until_column;

SET @has_outbox_status_next_attempt_index := (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND index_name = 'idx_status_next_attempt_id'
);

SET @add_outbox_status_next_attempt_index := IF(
  @has_outbox_status_next_attempt_index > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD INDEX `idx_status_next_attempt_id` (`status`, `next_attempt_at`, `id`)'
);

PREPARE article_inspect_add_outbox_status_next_attempt_index FROM @add_outbox_status_next_attempt_index;
EXECUTE article_inspect_add_outbox_status_next_attempt_index;
DEALLOCATE PREPARE article_inspect_add_outbox_status_next_attempt_index;

SET @has_outbox_status_claim_until_index := (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND index_name = 'idx_status_claim_until_id'
);

SET @add_outbox_status_claim_until_index := IF(
  @has_outbox_status_claim_until_index > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD INDEX `idx_status_claim_until_id` (`status`, `claim_until`, `id`)'
);

PREPARE article_inspect_add_outbox_status_claim_until_index FROM @add_outbox_status_claim_until_index;
EXECUTE article_inspect_add_outbox_status_claim_until_index;
DEALLOCATE PREPARE article_inspect_add_outbox_status_claim_until_index;

SET @has_outbox_status_retained_until_index := (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'xt_article_inspect_task_outbox'
    AND index_name = 'idx_status_retained_until_id'
);

SET @add_outbox_status_retained_until_index := IF(
  @has_outbox_status_retained_until_index > 0,
  'SELECT 1',
  'ALTER TABLE `xt_article_inspect_task_outbox` ADD INDEX `idx_status_retained_until_id` (`status`, `retained_until`, `id`)'
);

PREPARE article_inspect_add_outbox_status_retained_until_index FROM @add_outbox_status_retained_until_index;
EXECUTE article_inspect_add_outbox_status_retained_until_index;
DEALLOCATE PREPARE article_inspect_add_outbox_status_retained_until_index;
