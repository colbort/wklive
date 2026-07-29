-- dbinit:baseline-safe
-- 已有数据库首次接管时，将价格公式表收敛到当前基础 Schema。
-- 本迁移可重复执行；旧的非幂等迁移仍只作为基线登记。

SET @dbinit_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_market_price_formula'
    AND column_name = 'min_input_count'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_market_price_formula` ADD COLUMN `min_input_count` INT NOT NULL DEFAULT 1 AFTER `max_deviation_bps`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*)
  FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_market_price_formula'
    AND constraint_name = 'chk_price_formula'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_market_price_formula` DROP CHECK `chk_price_formula`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

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
