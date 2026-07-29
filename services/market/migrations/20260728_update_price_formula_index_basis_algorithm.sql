-- 允许生产 MARK 使用“指数价 + 有界基差”算法（algorithm=4）。
SET @schema_name = DATABASE();
SET @check_exists = (
  SELECT COUNT(1)
  FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = @schema_name
    AND TABLE_NAME = 't_market_price_formula'
    AND CONSTRAINT_NAME = 'chk_price_formula'
    AND CONSTRAINT_TYPE = 'CHECK'
);
SET @ddl = IF(
  @check_exists > 0,
  'ALTER TABLE `t_market_price_formula` DROP CHECK `chk_price_formula`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE `t_market_price_formula`
  ADD CONSTRAINT `chk_price_formula`
  CHECK (
    `snapshot_kind` IN ('MARK','INDEX','FUNDING','DELIVERY')
    AND `algorithm` IN (1,2,3,4)
    AND `max_lookback_ms` > 0
    AND `min_input_count` > 0
    AND `interval_ms` > 0
    AND `status` IN (1,2,3)
  );
