CREATE TABLE IF NOT EXISTS `t_asset_platform_account` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL,
  `account_type` VARCHAR(32) NOT NULL COMMENT 'INSURANCE_FUND/FUNDING_DIFFERENCE/FEE_REVENUE',
  `coin` VARCHAR(32) NOT NULL,
  `available_amount` DECIMAL(36,18) NOT NULL DEFAULT 0,
  `frozen_amount` DECIMAL(36,18) NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 2禁用',
  `version` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL,
  `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_type_coin` (`tenant_id`,`account_type`,`coin`),
  KEY `idx_platform_account_status` (`tenant_id`,`status`,`account_type`),
  CONSTRAINT `chk_asset_platform_account` CHECK (`tenant_id` > 0 AND `account_type` <> '' AND `coin` <> '' AND `available_amount` >= 0 AND `frozen_amount` >= 0 AND `status` IN (1,2) AND `version` >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Asset平台自有资金账户';

CREATE TABLE IF NOT EXISTS `t_asset_platform_flow` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL,
  `platform_account_id` BIGINT NOT NULL,
  `account_type` VARCHAR(32) NOT NULL,
  `coin` VARCHAR(32) NOT NULL,
  `op_type` TINYINT NOT NULL COMMENT '1增加 2扣减',
  `amount` DECIMAL(36,18) NOT NULL,
  `before_available` DECIMAL(36,18) NOT NULL,
  `after_available` DECIMAL(36,18) NOT NULL,
  `biz_type` VARCHAR(32) NOT NULL,
  `scene_type` VARCHAR(64) NOT NULL,
  `biz_id` BIGINT NOT NULL DEFAULT 0,
  `biz_no` VARCHAR(96) NOT NULL,
  `remark` VARCHAR(255) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_platform_flow_biz` (`tenant_id`,`platform_account_id`,`scene_type`,`biz_no`),
  KEY `idx_platform_flow_account_time` (`platform_account_id`,`create_times`),
  CONSTRAINT `chk_asset_platform_flow` CHECK (`op_type` IN (1,2) AND `amount` > 0 AND `before_available` >= 0 AND `after_available` >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Asset平台自有资金流水';

ALTER TABLE `t_asset_insurance_cover`
  ADD COLUMN `platform_account_id` BIGINT NOT NULL DEFAULT 0 AFTER `tenant_id`,
  DROP INDEX `idx_fund_asset_time`,
  ADD KEY `idx_fund_asset_time` (`tenant_id`,`platform_account_id`,`coin`,`create_times`),
  DROP COLUMN `fund_user_id`,
  DROP COLUMN `wallet_type`;

-- 旧保险基金来自普通用户资产，无法在 SQL 中安全推断迁移关系。
-- 升级后必须先创建并充值平台保险基金账户，再启用自动保险基金赔付。
