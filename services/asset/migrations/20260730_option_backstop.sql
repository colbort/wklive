ALTER TABLE `t_asset_platform_account`
  DROP CHECK `chk_asset_platform_account`,
  ADD CONSTRAINT `chk_asset_platform_account` CHECK (
    `tenant_id` > 0 AND `account_type` <> '' AND `coin` <> ''
    AND (`account_type` = 'OPTION_BACKSTOP' OR `available_amount` >= 0)
    AND `frozen_amount` >= 0 AND `status` IN (1,2) AND `version` >= 0
  );

ALTER TABLE `t_asset_platform_flow`
  DROP CHECK `chk_asset_platform_flow`,
  ADD CONSTRAINT `chk_asset_platform_flow` CHECK (
    `op_type` IN (1,2) AND `amount` > 0
    AND (`account_type` = 'OPTION_BACKSTOP' OR (`before_available` >= 0 AND `after_available` >= 0))
  );

CREATE TABLE IF NOT EXISTS `t_asset_backstop_cover` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL,
  `platform_account_id` BIGINT NOT NULL,
  `coin` VARCHAR(32) NOT NULL,
  `liquidation_id` BIGINT NOT NULL,
  `liquidation_no` VARCHAR(96) NOT NULL,
  `covered_amount` DECIMAL(36,18) NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1已赔付',
  `create_times` BIGINT NOT NULL,
  `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_backstop_liquidation_no` (`tenant_id`,`liquidation_no`),
  KEY `idx_backstop_account_time` (`tenant_id`,`platform_account_id`,`coin`,`create_times`),
  CONSTRAINT `chk_asset_backstop_cover` CHECK (`covered_amount` > 0 AND `status` = 1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='平台兜底穿仓赔付及幂等结果';
