-- Extend the durable cross-margin account liquidation Saga with negative
-- equity insurance-fund and ADL checkpoints.

ALTER TABLE `t_contract_account_liquidation`
  ADD COLUMN `deficit_amount` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT '用户余额扣尽后的账户穿仓缺口' AFTER `user_debit`,
  ADD COLUMN `adl_relief_amount` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT 'ADL累计缓释金额' AFTER `insurance_fund_amount`,
  ADD COLUMN `adl_qty` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT 'ADL累计接管数量' AFTER `adl_relief_amount`;

ALTER TABLE `t_contract_account_liquidation_item`
  ADD COLUMN `deficit_amount` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT '分摊到仓位的ADL目标缺口' AFTER `liquidation_fee`,
  ADD COLUMN `bankruptcy_price` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT '按分摊缺口冻结的合成破产价' AFTER `deficit_amount`,
  ADD COLUMN `adl_relief_amount` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT '仓位ADL累计缓释金额' AFTER `bankruptcy_price`,
  ADD COLUMN `adl_qty` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT '仓位ADL累计接管数量' AFTER `adl_relief_amount`;

ALTER TABLE `t_contract_account_liquidation`
  DROP CHECK `chk_cross_account_liquidation`,
  ADD CONSTRAINT `chk_cross_account_liquidation` CHECK (
    `margin_snapshot_id` > 0 AND `margin_snapshot_version` > 0 AND
    `asset_version` >= 0 AND `wallet_balance` >= 0 AND
    `position_margin` >= 0 AND `maintenance_margin` >= 0 AND
    `risk_rate` >= 0 AND `liquidation_fee` >= 0 AND
    `user_credit` >= 0 AND `user_debit` >= 0 AND
    `deficit_amount` >= 0 AND `insurance_fund_amount` >= 0 AND
    `adl_relief_amount` >= 0 AND `adl_qty` >= 0 AND
    `position_count` >= 0 AND
    `status` IN (1,2,3,4,5,6,7) AND `version` >= 0
  );

ALTER TABLE `t_contract_account_liquidation_item`
  DROP CHECK `chk_cross_account_liquidation_item`,
  ADD CONSTRAINT `chk_cross_account_liquidation_item` CHECK (
    `position_version` > 0 AND `position_side` IN (1,2,3) AND
    `trigger_qty` > 0 AND `trigger_mark_price` > 0 AND
    `position_margin` >= 0 AND `maintenance_margin` >= 0 AND
    `liquidation_fee` >= 0 AND `deficit_amount` >= 0 AND
    `bankruptcy_price` >= 0 AND `adl_relief_amount` >= 0 AND
    `adl_qty` >= 0 AND `status` IN (1,2)
  );
