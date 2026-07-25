ALTER TABLE `t_trade_order`
  MODIFY COLUMN `source` TINYINT NOT NULL DEFAULT 1 COMMENT '订单来源：1App 2Web 3API 4System 5Liquidity',
  ADD INDEX `idx_source_status_updated` (`source`, `status`, `update_times`);

-- 仅用于给升级前已经存在的内部做市报价补来源；升级后的订单由 RPC 强制写入来源 5。
UPDATE `t_trade_order`
SET `source` = 5
WHERE `source` = 4
  AND `client_order_id` LIKE 'LQC%';

CREATE TABLE `t_trade_order_archive` LIKE `t_trade_order`;
ALTER TABLE `t_trade_order_archive`
  ADD COLUMN `archived_at` BIGINT NOT NULL DEFAULT 0 COMMENT '归档时间' AFTER `update_times`,
  ADD COLUMN `archive_reason` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '归档原因' AFTER `archived_at`,
  ADD INDEX `idx_archived_at` (`archived_at`);

CREATE TABLE `t_trade_order_spot_archive` LIKE `t_trade_order_spot`;
ALTER TABLE `t_trade_order_spot_archive`
  ADD COLUMN `archived_at` BIGINT NOT NULL DEFAULT 0 COMMENT '归档时间' AFTER `update_times`,
  ADD COLUMN `archive_reason` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '归档原因' AFTER `archived_at`,
  ADD INDEX `idx_archived_at` (`archived_at`);

CREATE TABLE `t_trade_order_contract_archive` LIKE `t_trade_order_contract`;
ALTER TABLE `t_trade_order_contract_archive`
  ADD COLUMN `archived_at` BIGINT NOT NULL DEFAULT 0 COMMENT '归档时间' AFTER `update_times`,
  ADD COLUMN `archive_reason` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '归档原因' AFTER `archived_at`,
  ADD INDEX `idx_archived_at` (`archived_at`);

CREATE TABLE `t_trade_cancel_log_archive` LIKE `t_trade_cancel_log`;
ALTER TABLE `t_trade_cancel_log_archive`
  ADD COLUMN `archived_at` BIGINT NOT NULL DEFAULT 0 COMMENT '归档时间' AFTER `create_times`,
  ADD COLUMN `archive_reason` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '归档原因' AFTER `archived_at`,
  ADD INDEX `idx_archived_at` (`archived_at`);
