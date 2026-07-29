-- P2-01 全仓账户级风险投影。Asset 仍是资金账本唯一事实源；
-- 本表只固化 Trade 按保证金币种聚合后的可重建风险快照。
ALTER TABLE `t_contract_margin_snapshot`
  ADD COLUMN `maintenance_margin` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT '同保证金币种全仓持仓维持保证金合计' AFTER `order_margin`,
  ADD COLUMN `account_equity` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT '钱包余额、全仓仓位保证金与未实现盈亏合计' AFTER `maintenance_margin`,
  ADD COLUMN `available_margin` DECIMAL(36,18) NOT NULL DEFAULT 0
    COMMENT 'Asset可用余额叠加全仓未实现盈亏后的风险可用额' AFTER `account_equity`,
  ADD COLUMN `risk_rate` DECIMAL(20,10) NOT NULL DEFAULT 0
    COMMENT '账户级维持保证金/账户权益；非正权益使用上限值' AFTER `available_margin`,
  ADD COLUMN `position_count` INT NOT NULL DEFAULT 0
    COMMENT '参与当前快照的开放全仓仓位数' AFTER `risk_rate`,
  ADD COLUMN `asset_version` BIGINT NOT NULL DEFAULT 0
    COMMENT '生成快照时读取的Asset账户版本' AFTER `position_count`,
  ADD CONSTRAINT `chk_cross_margin_snapshot_risk`
    CHECK (`maintenance_margin` >= 0 AND `risk_rate` >= 0 AND
           `position_count` >= 0 AND `asset_version` >= 0);
