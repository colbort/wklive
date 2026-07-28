-- 为资金费历史恢复提供明确业务时点和仓位版本。
-- 既有记录只能以写入时间回填，不能据此自动补算迁移前的资金费周期。
ALTER TABLE `t_contract_position_history`
  ADD COLUMN `margin_asset` VARCHAR(32) NOT NULL DEFAULT ''
    COMMENT '仓位结算币种快照' AFTER `position_side`,
  ADD COLUMN `business_time` BIGINT NOT NULL DEFAULT 0
    COMMENT '仓位变更所属业务时点；用于历史结算重建' AFTER `action_key`,
  ADD COLUMN `before_version` BIGINT NOT NULL DEFAULT 0
    COMMENT '变更前仓位版本' AFTER `business_time`,
  ADD COLUMN `after_version` BIGINT NOT NULL DEFAULT 0
    COMMENT '变更后仓位版本' AFTER `before_version`;

UPDATE `t_contract_position_history`
SET `business_time` = `create_times`
WHERE `business_time` = 0;

UPDATE `t_contract_position_history` AS h
JOIN `t_contract_position` AS p
  ON p.tenant_id = h.tenant_id
 AND p.id = h.position_id
SET h.margin_asset = p.margin_asset
WHERE h.margin_asset = '';

ALTER TABLE `t_contract_position_history`
  ADD INDEX `idx_tenant_symbol_business_position`
    (`tenant_id`, `symbol_id`, `business_time`, `position_id`, `id`);
