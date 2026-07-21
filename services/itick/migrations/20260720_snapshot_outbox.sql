CREATE TABLE IF NOT EXISTS `t_itick_snapshot_outbox` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `snapshot_id` VARCHAR(64) NOT NULL,
  `payload` JSON NOT NULL, `status` TINYINT NOT NULL DEFAULT 1,
  `retry_count` INT NOT NULL DEFAULT 0, `next_retry_at` BIGINT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '', `create_times` BIGINT NOT NULL, `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_snapshot_outbox` (`snapshot_id`),
  KEY `idx_snapshot_outbox_retry` (`status`,`next_retry_at`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO t_itick_snapshot_outbox(snapshot_id,payload,status,retry_count,next_retry_at,last_error_msg,create_times,update_times)
SELECT snapshot_id,JSON_OBJECT('snapshot',CAST(raw_payload AS JSON)),1,0,0,'migration repair',create_times,create_times
FROM t_itick_authoritative_snapshot;
