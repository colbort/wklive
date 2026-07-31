-- 组合保证金模型治理：不可覆盖的参数版本、四眼审批、有效期和风险快照可追溯字段。

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_risk_account'
      AND column_name='portfolio_risk_config_id'),
  'SELECT 1',
  'ALTER TABLE t_option_risk_account
    ADD COLUMN portfolio_risk_config_id BIGINT NOT NULL DEFAULT 0
      COMMENT ''本次计算使用的组合风险参数版本ID'' AFTER portfolio_risk_method'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_risk_account'
      AND column_name='portfolio_risk_config_version'),
  'SELECT 1',
  'ALTER TABLE t_option_risk_account
    ADD COLUMN portfolio_risk_config_version BIGINT NOT NULL DEFAULT 0
      COMMENT ''本次计算使用的组合风险参数版本号'' AFTER portfolio_risk_config_id'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_risk_account'
      AND column_name='portfolio_concentration_addon'),
  'SELECT 1',
  'ALTER TABLE t_option_risk_account
    ADD COLUMN portfolio_concentration_addon DECIMAL(32,16) NOT NULL DEFAULT 0
      COMMENT ''组合集中度附加保证金'' AFTER portfolio_short_floor'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_risk_account'
      AND column_name='portfolio_liquidity_addon'),
  'SELECT 1',
  'ALTER TABLE t_option_risk_account
    ADD COLUMN portfolio_liquidity_addon DECIMAL(32,16) NOT NULL DEFAULT 0
      COMMENT ''组合流动性附加保证金'' AFTER portfolio_concentration_addon'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_schema=DATABASE() AND table_name='t_option_risk_account'
      AND constraint_name='chk_option_risk_account' AND constraint_type='CHECK'),
  'ALTER TABLE t_option_risk_account DROP CHECK chk_option_risk_account',
  'SELECT 1'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

ALTER TABLE t_option_risk_account
  ADD CONSTRAINT `chk_option_risk_account` CHECK (
    `position_margin` >= 0 AND `maintenance_margin` >= 0
    AND `portfolio_risk_method` IN (0,1)
    AND `portfolio_risk_config_id` >= 0 AND `portfolio_risk_config_version` >= 0
    AND `portfolio_scenario_loss` >= 0 AND `portfolio_short_floor` >= 0
    AND `portfolio_concentration_addon` >= 0 AND `portfolio_liquidity_addon` >= 0
    AND `risk_rate` >= 0 AND `status` IN (1,2,3,4,5)
  );

CREATE TABLE IF NOT EXISTS `t_option_portfolio_risk_config` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `settle_coin` VARCHAR(16) NOT NULL DEFAULT '',
  `version` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `model_method` TINYINT NOT NULL DEFAULT 1,
  `initial_shock_rate` DECIMAL(20,10) NOT NULL DEFAULT 0,
  `maintenance_shock_rate` DECIMAL(20,10) NOT NULL DEFAULT 0,
  `scenario_shocks` VARCHAR(500) NOT NULL DEFAULT '',
  `concentration_threshold` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `concentration_addon_rate` DECIMAL(20,10) NOT NULL DEFAULT 0,
  `liquidity_addon_rate` DECIMAL(20,10) NOT NULL DEFAULT 0,
  `effective_from` BIGINT NOT NULL DEFAULT 0,
  `effective_until` BIGINT NOT NULL DEFAULT 0,
  `supersedes_id` BIGINT NOT NULL DEFAULT 0,
  `change_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '',
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `reviewed_by` BIGINT NOT NULL DEFAULT 0,
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_portfolio_risk_config_version` (`tenant_id`, `settle_coin`, `version`),
  KEY `idx_portfolio_risk_config_active`
    (`tenant_id`, `settle_coin`, `status`, `effective_from`, `effective_until`),
  CONSTRAINT `chk_option_portfolio_risk_config` CHECK (
    `tenant_id` > 0 AND `settle_coin` <> '' AND `version` > 0
    AND `status` IN (1,2,3,4) AND `model_method` = 1
    AND `initial_shock_rate` > 0 AND `initial_shock_rate` <= 10
    AND `maintenance_shock_rate` > 0
    AND `maintenance_shock_rate` <= `initial_shock_rate`
    AND `scenario_shocks` <> ''
    AND `concentration_threshold` >= 0
    AND `concentration_addon_rate` >= 0 AND `concentration_addon_rate` <= 1
    AND `liquidity_addon_rate` >= 0 AND `liquidity_addon_rate` <= 1
    AND `effective_from` > 0
    AND (`effective_until` = 0 OR `effective_until` > `effective_from`)
    AND `change_reason` <> '' AND `created_by` > 0
    AND (
      (`status` = 1 AND `reviewed_by` = 0 AND `reviewed_at` = 0)
      OR (`status` IN (2,3,4) AND `reviewed_by` > 0
        AND `reviewed_by` <> `created_by` AND `reviewed_at` > 0)
    )
    AND (`status` NOT IN (2,4) OR `evidence_ref` <> '')
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='组合保证金风险模型参数版本';

DROP TRIGGER IF EXISTS trg_option_portfolio_config_insert_guard;
DELIMITER $$
CREATE TRIGGER trg_option_portfolio_config_insert_guard
BEFORE INSERT ON t_option_portfolio_risk_config
FOR EACH ROW
BEGIN
  IF NEW.status <> 1 OR NEW.reviewed_by <> 0 OR NEW.reviewed_at <> 0
    OR NEW.effective_until <> 0 OR NEW.supersedes_id <> 0
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'portfolio risk config must start as a pending draft';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_portfolio_config_immutable;
DELIMITER $$
CREATE TRIGGER trg_option_portfolio_config_immutable
BEFORE UPDATE ON t_option_portfolio_risk_config
FOR EACH ROW
BEGIN
  IF NEW.tenant_id <> OLD.tenant_id
    OR NEW.settle_coin <> OLD.settle_coin
    OR NEW.version <> OLD.version
    OR NEW.model_method <> OLD.model_method
    OR NEW.initial_shock_rate <> OLD.initial_shock_rate
    OR NEW.maintenance_shock_rate <> OLD.maintenance_shock_rate
    OR NEW.scenario_shocks <> OLD.scenario_shocks
    OR NEW.concentration_threshold <> OLD.concentration_threshold
    OR NEW.concentration_addon_rate <> OLD.concentration_addon_rate
    OR NEW.liquidity_addon_rate <> OLD.liquidity_addon_rate
    OR NEW.effective_from <> OLD.effective_from
    OR NEW.change_reason <> OLD.change_reason
    OR NEW.evidence_ref <> OLD.evidence_ref
    OR NEW.created_by <> OLD.created_by
    OR NEW.create_times <> OLD.create_times
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'portfolio risk parameter content is immutable';
  END IF;

  IF (OLD.status = 1 AND NEW.status NOT IN (2,3))
    OR (OLD.status = 2 AND (NEW.status <> 4 OR NEW.effective_until <= NEW.effective_from))
    OR OLD.status IN (3,4)
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'invalid portfolio risk config status transition';
  END IF;

  IF OLD.status = 1 AND NEW.status IN (2,3)
    AND (NEW.reviewed_by <= 0 OR NEW.reviewed_by = OLD.created_by OR NEW.reviewed_at <= 0)
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'portfolio risk config requires an independent reviewer';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_portfolio_config_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_portfolio_config_no_delete
BEFORE DELETE ON t_option_portfolio_risk_config
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'portfolio risk config history cannot be deleted';
END$$
DELIMITER ;
