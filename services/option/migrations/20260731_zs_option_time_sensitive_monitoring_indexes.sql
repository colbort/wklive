-- Incremental indexes for OPT-A009/A020/A026 time-sensitive monitoring.
-- Do not append these to 20260731_zr: recorded migration checksums are immutable.

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_exercise'
      AND index_name='idx_option_exercise_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_exercise
     ADD INDEX idx_option_exercise_monitor
       (status,tenant_id,contract_id,create_times,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_asset_instruction'
      AND index_name='idx_option_asset_instruction_control_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_asset_instruction
     ADD INDEX idx_option_asset_instruction_control_monitor
       (action,status,tenant_id,user_id,create_times,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE()
      AND table_name='t_option_physical_delivery_unit'
      AND index_name='idx_option_physical_delivery_monitor'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_physical_delivery_unit
     ADD INDEX idx_option_physical_delivery_monitor
       (status,cure_deadline,tenant_id,id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
