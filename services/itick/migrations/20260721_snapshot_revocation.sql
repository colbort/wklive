CREATE TABLE IF NOT EXISTS `t_itick_snapshot_revocation` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `snapshot_id` VARCHAR(64) NOT NULL,
  `replacement_snapshot_id` VARCHAR(64) NOT NULL DEFAULT '',
  `reason` VARCHAR(512) NOT NULL,
  `create_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_snapshot_revocation` (`snapshot_id`),
  KEY `idx_snapshot_revocation_rebuild` (`id`),
  CONSTRAINT `chk_snapshot_revocation_reason` CHECK (CHAR_LENGTH(`reason`) > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
