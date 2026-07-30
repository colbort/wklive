-- dbinit:baseline-safe
-- 已过去且没有任何仓位事实的资金费时点允许生成显式空批次。
-- 空批次不伪造价格，仅用于审计该时点无资金影响并推进定时任务游标。

SET @dbinit_exists := (
  SELECT COUNT(*)
  FROM information_schema.table_constraints
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_funding_batch'
    AND constraint_name = 'chk_funding_batch'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_funding_batch` DROP CHECK `chk_funding_batch`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

ALTER TABLE `t_contract_funding_batch`
  ADD CONSTRAINT `chk_funding_batch`
  CHECK (
    (
      `mark_price` > 0
      OR (
        `total_positions` = 0 AND `status` = 3 AND `mark_price` = 0 AND
        `index_price` = 0 AND `funding_rate` = 0 AND
        `price_source` = 'NO_POSITION_HISTORY' AND `formula_version` = 'no-position-v1'
      )
    ) AND
    `index_price` >= 0 AND
    `settlement_time` > 0 AND
    `status` BETWEEN 1 AND 5 AND
    `total_positions` >= 0 AND
    `settled_positions` >= 0 AND
    `settled_positions` <= `total_positions` AND
    `version` >= 0
  );
