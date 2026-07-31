-- 用户主动行权必须携带客户端幂等号；同一租户、用户和幂等号只生成一笔行权。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name
      AND table_name='t_option_exercise'
      AND column_name='client_exercise_id'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_exercise
     ADD COLUMN client_exercise_id VARCHAR(64) NOT NULL DEFAULT ''''
       COMMENT ''客户端行权幂等号；用户主动行权禁止为空'' AFTER exercise_no'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 历史自动行权使用确定性键；历史用户行权使用原行权单号，确保加唯一索引前均非空。
UPDATE t_option_exercise
SET client_exercise_id = CASE
  WHEN exercise_type = 2 THEN CONCAT('AUTO-', position_id, '-', exercise_time)
  ELSE exercise_no
END
WHERE client_exercise_id = '';

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=@schema_name
      AND table_name='t_option_exercise'
      AND index_name='uk_tenant_user_client_exercise'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_exercise
     ADD UNIQUE KEY uk_tenant_user_client_exercise
       (tenant_id, user_id, client_exercise_id)'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_schema=@schema_name
      AND table_name='t_option_exercise'
      AND constraint_name='chk_option_client_exercise_key'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_exercise
     ADD CONSTRAINT chk_option_client_exercise_key
       CHECK (exercise_type <> 1 OR client_exercise_id <> '''')'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
