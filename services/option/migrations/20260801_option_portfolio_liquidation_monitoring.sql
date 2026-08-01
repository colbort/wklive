-- Cover OPT-A034 portfolio-liquidation monitoring by scope, wallet and sequence.
-- Kept separate from immutable historical monitoring migrations.

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_liquidation'
      AND index_name='idx_option_liquidation_portfolio_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_liquidation
     ADD INDEX idx_option_liquidation_portfolio_monitor
       (liquidation_scope,tenant_id,user_id,id,status,contract_id,update_times,create_times)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
