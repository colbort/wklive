SET @option_margin_instruction_sql = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE() AND table_name = 't_option_asset_instruction'
        AND column_name = 'margin_lot_id'
    ),
    'SELECT 1',
    'ALTER TABLE `t_option_asset_instruction`
       ADD COLUMN `margin_lot_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''卖方保证金批次ID'' AFTER `position_id`,
       ADD COLUMN `liquidation_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''强平记录ID'' AFTER `margin_lot_id`,
       ADD KEY `idx_instruction_margin_lot` (`tenant_id`, `margin_lot_id`),
       ADD KEY `idx_instruction_liquidation` (`tenant_id`, `liquidation_id`, `status`, `id`)'
  )
);
PREPARE option_margin_instruction_stmt FROM @option_margin_instruction_sql;
EXECUTE option_margin_instruction_stmt;
DEALLOCATE PREPARE option_margin_instruction_stmt;
