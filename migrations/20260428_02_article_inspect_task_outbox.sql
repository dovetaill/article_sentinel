CREATE TABLE IF NOT EXISTS `xt_article_inspect_task_outbox` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `orgid` BIGINT UNSIGNED NOT NULL,
  `task_id` BIGINT UNSIGNED NOT NULL,
  `message_type` VARCHAR(64) NOT NULL,
  `status` VARCHAR(32) NOT NULL,
  `payload` LONGTEXT NULL,
  `attempt_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `last_error` TEXT NULL,
  `last_attempt_at` DATETIME NULL,
  `dispatched_at` DATETIME NULL,
  `create_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_status_id` (`status`, `id`),
  KEY `idx_org_task` (`orgid`, `task_id`),
  KEY `idx_message_type_status` (`message_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
