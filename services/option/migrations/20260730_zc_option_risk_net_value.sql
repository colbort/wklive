-- 保存风险权益公式中的净期权市值，便于审计和对账。
SET @schema_name = DATABASE();
SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name
      AND table_name='t_option_risk_account'
      AND column_name='net_option_value'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_risk_account
     ADD COLUMN net_option_value DECIMAL(32,16) NOT NULL DEFAULT 0
       COMMENT ''多头市值减空头市值'' AFTER equity'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
