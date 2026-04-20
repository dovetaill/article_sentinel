-- Article Sentinel phase-1 demo seed
-- Import with:
-- mysql -h127.0.0.1 -P3307 -uroot -p"${MYSQL_ROOT_PASSWORD:-root}" article_sentinel < scripts/article_inspection_seed.sql

SET NAMES utf8mb4;

DELETE FROM `xt_article_inspect_field_change_logs` WHERE `orgid` = 100 AND `article_id` IN (9001001, 9001002);
DELETE FROM `xt_article_inspect_operation_logs` WHERE `orgid` = 100 AND `article_id` IN (9001001, 9001002);
DELETE FROM `xt_article_inspect_actions` WHERE `orgid` = 100 AND `action_no` IN ('offline-20260420-01', 'rectify-20260420-01');
DELETE FROM `xt_article_inspect_result_hits` WHERE `orgid` = 100 AND `result_id` IN (30001, 30002);
DELETE FROM `xt_article_inspect_results` WHERE `orgid` = 100 AND `id` IN (30001, 30002);
DELETE FROM `xt_article_inspect_task_keywords` WHERE `orgid` = 100 AND `task_id` = 20001;
DELETE FROM `xt_article_inspect_tasks` WHERE `orgid` = 100 AND `id` = 20001;
DELETE FROM `xt_article_inspect_keyword_scopes` WHERE `orgid` = 100 AND `keyword_id` IN (10001, 10002);
DELETE FROM `xt_article_inspect_keywords` WHERE `orgid` = 100 AND `id` IN (10001, 10002);

INSERT INTO `xt_article_inspect_keywords`
  (`id`, `orgid`, `name`, `category`, `match_type`, `risk_level`, `suggest_action`, `enabled`, `remark`, `creator_id`, `creator_name`, `updater_id`, `updater_name`)
VALUES
  (10001, 100, 'spam', 'policy', 'contains', 'high', 'offline', 1, 'seed rule for acceptance', 7, 'auditor', 7, 'auditor'),
  (10002, 100, 'scam', 'policy', 'contains', 'medium', 'process', 1, 'secondary demo rule', 7, 'auditor', 7, 'auditor');

INSERT INTO `xt_article_inspect_keyword_scopes` (`orgid`, `keyword_id`, `scope`)
VALUES
  (100, 10001, 'title'),
  (100, 10001, 'body'),
  (100, 10002, 'title');

INSERT INTO `xt_article_inspect_tasks`
  (`id`, `orgid`, `task_no`, `status`, `article_state_filter`, `include_body`, `request_snapshot`, `rule_snapshot`, `total_scanned`, `hit_articles`, `hit_count`, `skip_count`, `fail_count`, `batch_count`, `creator_id`, `creator_name`, `started_at`, `finished_at`, `duration_ms`)
VALUES
  (20001, 100, 'inspect-20260420-20001', 'success', '9', 1, '{"orgid":100,"article_state":9}', '{"keyword_ids":[10001,10002]}', 12, 2, 3, 0, 0, 1, 7, 'auditor', '2026-04-20 10:00:00', '2026-04-20 10:02:30', 150000);

INSERT INTO `xt_article_inspect_task_keywords` (`orgid`, `task_id`, `keyword_id`)
VALUES
  (100, 20001, 10001),
  (100, 20001, 10002);

INSERT INTO `xt_article_inspect_results`
  (`id`, `orgid`, `task_id`, `article_id`, `article_title`, `article_state`, `publish_at_time`, `risk_level`, `suggest_action`, `hit_fields_count`, `hit_keywords_count`, `hit_count`, `disposition_status`, `latest_action_id`, `latest_operator_id`, `latest_operator_name`, `latest_action_at`)
VALUES
  (30001, 100, 20001, 9001001, 'Spam headline needs audit', 9, '2026-04-20 09:00:00', 'high', 'offline', 2, 1, 2, 'pending', 40001, 7, 'auditor', '2026-04-20 10:05:00'),
  (30002, 100, 20001, 9001002, 'Scam promo draft', 1, '2026-04-20 09:10:00', 'medium', 'process', 1, 1, 1, 'processed', 40002, 7, 'auditor', '2026-04-20 10:08:00');

INSERT INTO `xt_article_inspect_result_hits`
  (`id`, `orgid`, `task_id`, `result_id`, `article_id`, `keyword_id`, `keyword_text`, `category`, `field_name`, `match_type`, `risk_level`, `suggest_action`, `matched_text`, `snippet`, `position_start`, `position_end`, `rule_snapshot`)
VALUES
  (50001, 100, 20001, 30001, 9001001, 10001, 'spam', 'policy', 'title', 'contains', 'high', 'offline', 'spam', 'spam headline needs audit', 1, 4, '{"scope":"title"}'),
  (50002, 100, 20001, 30001, 9001001, 10001, 'spam', 'policy', 'body', 'contains', 'high', 'offline', 'spam', 'body still contains spam keyword', 21, 24, '{"scope":"body"}'),
  (50003, 100, 20001, 30002, 9001002, 10002, 'scam', 'policy', 'title', 'contains', 'medium', 'process', 'scam', 'scam promo draft', 1, 4, '{"scope":"title"}');

INSERT INTO `xt_article_inspect_actions`
  (`id`, `orgid`, `action_no`, `action_type`, `task_id`, `batch_scope`, `target_count`, `success_count`, `fail_count`, `skip_count`, `status`, `reason`, `request_snapshot`, `operator_id`, `operator_name`, `request_id`, `source_ip`, `started_at`, `finished_at`)
VALUES
  (40001, 100, 'offline-20260420-01', 'offline', 20001, 'result_ids', 1, 1, 0, 0, 'success', 'manual batch offline', '{"result_ids":[30001]}', 7, 'auditor', 'req-seed-1', '127.0.0.1', '2026-04-20 10:05:00', '2026-04-20 10:05:05'),
  (40002, 100, 'rectify-20260420-01', 'rectify', 20001, 'single', 1, 1, 0, 0, 'success', 'cleanup and resubmit', '{"article_id":9001002,"target_article_state":1}', 7, 'auditor', 'req-seed-2', '127.0.0.1', '2026-04-20 10:08:00', '2026-04-20 10:08:10');

INSERT INTO `xt_article_inspect_operation_logs`
  (`id`, `orgid`, `action_id`, `task_id`, `result_id`, `article_id`, `operation_type`, `before_state`, `after_state`, `status`, `reason`, `summary`, `request_snapshot`, `operator_id`, `operator_name`, `request_id`, `source_ip`)
VALUES
  (60001, 100, 40001, 20001, 30001, 9001001, 'offline', 'online', 'offline', 'success', 'manual batch offline', 'Article sent offline after scan hit', '{"result_ids":[30001]}', 7, 'auditor', 'req-seed-1', '127.0.0.1'),
  (60002, 100, 40002, 20001, 30002, 9001002, 'rectify', 'audit', 'audit', 'success', 'cleanup and resubmit', 'Rectified article routed back to audit', '{"target_article_state":1}', 7, 'auditor', 'req-seed-2', '127.0.0.1');

INSERT INTO `xt_article_inspect_field_change_logs`
  (`id`, `orgid`, `action_id`, `task_id`, `result_id`, `article_id`, `field_name`, `before_value`, `after_value`, `diff_summary`, `operator_id`, `operator_name`, `request_id`, `source_ip`)
VALUES
  (70001, 100, 40002, 20001, 30002, 9001002, 'title', 'Scam promo draft', 'Scam promo updated', 'replace risky promo wording', 7, 'auditor', 'req-seed-2', '127.0.0.1'),
  (70002, 100, 40002, 20001, 30002, 9001002, 'body', '<p>scam offer body</p>', '<p>cleaned body waiting for audit</p>', 'remove risky body segment', 7, 'auditor', 'req-seed-2', '127.0.0.1');
