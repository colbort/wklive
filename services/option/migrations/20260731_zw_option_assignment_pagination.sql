-- Support deterministic keyset pagination for American exercise FIFO assignment.

SET @schema_name = DATABASE();
SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=@schema_name AND table_name='t_option_position'
      AND index_name='idx_option_position_assignment_fifo'
  ),
  'SELECT 1',
  'ALTER TABLE `t_option_position` ADD INDEX `idx_option_position_assignment_fifo` (`tenant_id`,`contract_id`,`side`,`status`,`create_times`,`id`)'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
