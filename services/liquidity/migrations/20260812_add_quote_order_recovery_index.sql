SET @quote_order_recovery_index_exists = (
  SELECT COUNT(1)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 't_liquidity_quote_order'
    AND index_name = 'idx_status_update_id'
);

SET @quote_order_recovery_index_ddl = IF(
  @quote_order_recovery_index_exists = 0,
  'ALTER TABLE `t_liquidity_quote_order` ADD INDEX `idx_status_update_id` (`status`, `update_times`, `id`)',
  'SELECT 1'
);

PREPARE quote_order_recovery_index_stmt FROM @quote_order_recovery_index_ddl;
EXECUTE quote_order_recovery_index_stmt;
DEALLOCATE PREPARE quote_order_recovery_index_stmt;
