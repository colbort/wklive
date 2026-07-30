-- Admin WebSocket operational-alert acknowledgement and escalation receipts.
-- The stable incident identity is tenant + event type + alert key. Repeated
-- firing events update the same row; resolved events close it without deleting
-- acknowledgement history.
CREATE TABLE IF NOT EXISTS `sys_admin_notification_incident` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `event_type` VARCHAR(64) NOT NULL,
  `alert_key` VARCHAR(191) NOT NULL,
  `last_event_id` VARCHAR(255) NOT NULL,
  `severity` VARCHAR(32) NOT NULL,
  `title` VARCHAR(255) NOT NULL,
  `message` VARCHAR(2000) NOT NULL,
  `source` VARCHAR(64) NOT NULL,
  `payload_json` LONGTEXT NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1=open 2=acknowledged 3=resolved',
  `first_seen_at` BIGINT NOT NULL DEFAULT 0,
  `last_seen_at` BIGINT NOT NULL DEFAULT 0,
  `acknowledged_at` BIGINT NOT NULL DEFAULT 0,
  `acknowledged_by` BIGINT NOT NULL DEFAULT 0,
  `acknowledged_username` VARCHAR(64) NOT NULL DEFAULT '',
  `acknowledge_reason` VARCHAR(255) NOT NULL DEFAULT '',
  `escalation_level` INT NOT NULL DEFAULT 0,
  `last_escalated_at` BIGINT NOT NULL DEFAULT 0,
  `next_escalate_at` BIGINT NOT NULL DEFAULT 0,
  `resolved_at` BIGINT NOT NULL DEFAULT 0,
  `version` BIGINT NOT NULL DEFAULT 1,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_admin_notification_incident` (`tenant_id`, `event_type`, `alert_key`),
  KEY `idx_admin_notification_due` (`status`, `next_escalate_at`, `id`),
  KEY `idx_admin_notification_last_event` (`last_event_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Admin operational alert acknowledgement and escalation incident';
