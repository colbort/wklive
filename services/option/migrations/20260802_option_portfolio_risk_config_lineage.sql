-- 组合风险参数治理补强：持久化复制/回滚来源，并拒绝新草案及审批的追溯生效。

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_portfolio_risk_config'
      AND column_name='source_config_id'),
  'SELECT 1',
  'ALTER TABLE t_option_portfolio_risk_config
    ADD COLUMN source_config_id BIGINT NOT NULL DEFAULT 0
      COMMENT ''复制参数的历史版本ID；0表示非复制创建'' AFTER supersedes_id'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_schema=DATABASE() AND table_name='t_option_portfolio_risk_config'
      AND constraint_name='chk_option_portfolio_risk_config' AND constraint_type='CHECK'),
  'ALTER TABLE t_option_portfolio_risk_config DROP CHECK chk_option_portfolio_risk_config',
  'SELECT 1'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

ALTER TABLE t_option_portfolio_risk_config
  ADD CONSTRAINT `chk_option_portfolio_risk_config` CHECK (
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
    AND `source_config_id` >= 0
    AND `change_reason` <> '' AND `created_by` > 0
    AND (
      (`status` = 1 AND `reviewed_by` = 0 AND `reviewed_at` = 0)
      OR (`status` IN (2,3,4) AND `reviewed_by` > 0
        AND `reviewed_by` <> `created_by` AND `reviewed_at` > 0)
    )
    AND (`status` NOT IN (2,4) OR `evidence_ref` <> '')
  );

DROP TRIGGER IF EXISTS trg_option_portfolio_config_insert_guard;
DELIMITER $$
CREATE TRIGGER trg_option_portfolio_config_insert_guard
BEFORE INSERT ON t_option_portfolio_risk_config
FOR EACH ROW
BEGIN
  DECLARE source_matches BIGINT DEFAULT 0;
  IF NEW.status <> 1 OR NEW.reviewed_by <> 0 OR NEW.reviewed_at <> 0
    OR NEW.effective_until <> 0 OR NEW.supersedes_id <> 0
    OR NEW.effective_from <= NEW.create_times
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'portfolio risk config must start as a future pending draft';
  END IF;

  IF NEW.source_config_id > 0 THEN
    SELECT COUNT(*) INTO source_matches
    FROM t_option_portfolio_risk_config source
    WHERE source.id = NEW.source_config_id
      AND source.tenant_id = NEW.tenant_id
      AND source.settle_coin = NEW.settle_coin
      AND source.version < NEW.version
      AND source.status IN (2,4)
      AND source.model_method = NEW.model_method
      AND source.initial_shock_rate = NEW.initial_shock_rate
      AND source.maintenance_shock_rate = NEW.maintenance_shock_rate
      AND source.scenario_shocks = NEW.scenario_shocks
      AND source.concentration_threshold = NEW.concentration_threshold
      AND source.concentration_addon_rate = NEW.concentration_addon_rate
      AND source.liquidity_addon_rate = NEW.liquidity_addon_rate;
    IF source_matches <> 1 THEN
      SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'portfolio risk source config lineage is invalid';
    END IF;
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
    OR NEW.source_config_id <> OLD.source_config_id
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

  IF OLD.status = 1 AND NEW.status = 2 AND NEW.effective_from <= NEW.reviewed_at THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'approved portfolio risk config must remain future effective';
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
