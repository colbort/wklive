SET @option_fee_config_sql = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE() AND table_name = 't_option_contract'
        AND column_name = 'maker_fee_rate'
    ),
    'SELECT 1',
    'ALTER TABLE `t_option_contract`
       ADD COLUMN `maker_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT ''Maker成交手续费率'' AFTER `is_auto_exercise`,
       ADD COLUMN `taker_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT ''Taker成交手续费率'' AFTER `maker_fee_rate`,
       ADD COLUMN `exercise_fee_rate` DECIMAL(20,10) NOT NULL DEFAULT 0 COMMENT ''行权手续费率'' AFTER `taker_fee_rate`,
       ADD COLUMN `fee_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''平台手续费归集用户ID'' AFTER `exercise_fee_rate`,
       ADD COLUMN `fee_account_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''平台手续费归集Option账户ID'' AFTER `fee_user_id`'
  )
);
PREPARE option_fee_config_stmt FROM @option_fee_config_sql;
EXECUTE option_fee_config_stmt;
DEALLOCATE PREPARE option_fee_config_stmt;
