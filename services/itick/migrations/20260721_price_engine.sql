CREATE TABLE IF NOT EXISTS `t_itick_price_formula` (
  `id` BIGINT NOT NULL AUTO_INCREMENT, `formula_no` VARCHAR(64) NOT NULL,
  `authority` VARCHAR(32) NOT NULL DEFAULT 'price-engine', `snapshot_kind` VARCHAR(32) NOT NULL,
  `category_code` VARCHAR(64) NOT NULL DEFAULT '', `market` VARCHAR(32) NOT NULL DEFAULT '', `symbol` VARCHAR(64) NOT NULL,
  `algorithm` VARCHAR(32) NOT NULL, `formula_version` VARCHAR(64) NOT NULL, `components` JSON NOT NULL,
  `max_lookback_ms` BIGINT NOT NULL DEFAULT 30000, `max_deviation_bps` INT NOT NULL DEFAULT 0,
  `interval_ms` BIGINT NOT NULL DEFAULT 1000, `last_target_time` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 2 COMMENT '1 active,2 inactive,3 revoked', `version` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL, `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`), UNIQUE KEY `uk_price_formula_no` (`formula_no`),
  UNIQUE KEY `uk_price_formula_output` (`authority`,`snapshot_kind`,`category_code`,`market`,`symbol`,`formula_version`),
  KEY `idx_price_formula_due` (`status`,`last_target_time`),
  CONSTRAINT `chk_price_formula` CHECK (`snapshot_kind` IN ('MARK','INDEX','FUNDING','DELIVERY') AND `algorithm` IN ('WEIGHTED_MEAN','MEDIAN','PREMIUM_RATE') AND `max_lookback_ms` > 0 AND `interval_ms` > 0 AND `status` IN (1,2,3))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
ALTER TABLE `t_itick_authoritative_snapshot` DROP CHECK `chk_authoritative_snapshot`;
ALTER TABLE `t_itick_authoritative_snapshot` ADD CONSTRAINT `chk_authoritative_snapshot` CHECK ((`snapshot_kind`='FUNDING' OR `price` > 0) AND `source_timestamp` > 0 AND `snapshot_timestamp` > 0 AND `revision` > 0);
INSERT INTO `t_itick_authority_registry` (`authority`,`producer_type`,`allowed_kinds`,`status`,`version`,`create_times`,`update_times`)
VALUES ('price-engine','PRICE_ENGINE',JSON_ARRAY('MARK','INDEX','FUNDING','DELIVERY'),1,0,0,0)
ON DUPLICATE KEY UPDATE `producer_type`=VALUES(`producer_type`),`allowed_kinds`=VALUES(`allowed_kinds`);
