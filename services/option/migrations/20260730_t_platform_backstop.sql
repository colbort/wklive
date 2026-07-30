-- 平台全额兜底策略及强平缺口审计字段。可重复执行。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @schema_name
      AND table_name = 't_option_contract'
      AND column_name = 'liquidation_deficit_policy'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN liquidation_deficit_policy TINYINT NOT NULL DEFAULT 1
       COMMENT ''保险不足策略：1人工 2平台兜底''
       AFTER insurance_account_id'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @schema_name
      AND table_name = 't_option_liquidation'
      AND column_name = 'backstop_amount'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_liquidation
     ADD COLUMN backstop_amount DECIMAL(32,16) NOT NULL DEFAULT 0
       COMMENT ''平台兜底负债金额''
       AFTER insurance_attempt'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @schema_name
      AND table_name = 't_option_liquidation'
      AND column_name = 'deficit_resolution'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_liquidation
     ADD COLUMN deficit_resolution TINYINT NOT NULL DEFAULT 1
       COMMENT ''缺口处置：1无 2保险 3平台兜底 4保险加兜底 5人工''
       AFTER backstop_amount'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @check_name = (
  SELECT tc.constraint_name
  FROM information_schema.table_constraints tc
  WHERE tc.constraint_schema = @schema_name
    AND tc.table_name = 't_option_liquidation'
    AND tc.constraint_type = 'CHECK'
    AND tc.constraint_name = 'chk_option_liquidation'
  LIMIT 1
);
SET @ddl = IF(
  @check_name IS NULL,
  'SELECT 1',
  'ALTER TABLE t_option_liquidation DROP CHECK chk_option_liquidation'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE t_option_liquidation
  ADD CONSTRAINT chk_option_liquidation CHECK (
    quantity > 0 AND deficit_amount >= 0 AND liquidation_fee >= 0
    AND status IN (1,2,3,4,5,6) AND retry_count >= 0 AND insurance_attempt >= 0
    AND backstop_amount >= 0 AND deficit_resolution IN (1,2,3,4,5)
  );
