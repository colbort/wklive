-- Price Engine 异常剔除后最少有效输入数；安全幂等升级。
SET @schema_name = DATABASE();
SET @column_exists = (
  SELECT COUNT(1)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = @schema_name
    AND TABLE_NAME = 't_itick_price_formula'
    AND COLUMN_NAME = 'min_input_count'
);
SET @ddl = IF(
  @column_exists = 0,
  'ALTER TABLE `t_itick_price_formula` ADD COLUMN `min_input_count` INT NOT NULL DEFAULT 1 AFTER `max_deviation_bps`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
