ALTER TABLE `t_option_asset_instruction`
  ADD COLUMN `margin_lot_id` BIGINT NOT NULL DEFAULT 0 COMMENT '卖方保证金批次ID' AFTER `position_id`,
  ADD COLUMN `liquidation_id` BIGINT NOT NULL DEFAULT 0 COMMENT '强平记录ID' AFTER `margin_lot_id`,
  ADD KEY `idx_instruction_margin_lot` (`tenant_id`, `margin_lot_id`),
  ADD KEY `idx_instruction_liquidation` (`tenant_id`, `liquidation_id`, `status`, `id`);
