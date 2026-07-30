-- 组合保证金风险快照。迁移可在已升级环境重复执行。
SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = @schema_name
      AND table_name = 't_option_risk_account'
      AND column_name = 'portfolio_risk_method'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_risk_account
     ADD COLUMN portfolio_risk_method TINYINT NOT NULL DEFAULT 0 COMMENT ''组合风险算法：0无 1到期损益情景V1'' AFTER status,
     ADD COLUMN portfolio_scenario_loss DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''组合价格情景最大损失'' AFTER portfolio_risk_method,
     ADD COLUMN portfolio_short_floor DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''组合裸空头最低保证金'' AFTER portfolio_scenario_loss'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @check_name = (
  SELECT tc.constraint_name
  FROM information_schema.table_constraints tc
  WHERE tc.constraint_schema = @schema_name
    AND tc.table_name = 't_option_risk_account'
    AND tc.constraint_type = 'CHECK'
    AND tc.constraint_name = 'chk_option_risk_account'
  LIMIT 1
);
SET @ddl = IF(
  @check_name IS NULL,
  'SELECT 1',
  'ALTER TABLE t_option_risk_account DROP CHECK chk_option_risk_account'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE t_option_risk_account
  ADD CONSTRAINT chk_option_risk_account CHECK (
    position_margin >= 0 AND maintenance_margin >= 0
    AND portfolio_risk_method IN (0,1)
    AND portfolio_scenario_loss >= 0 AND portfolio_short_floor >= 0
    AND risk_rate >= 0 AND status IN (1,2,3,4,5)
  );
