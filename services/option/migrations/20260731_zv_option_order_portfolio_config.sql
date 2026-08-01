-- Persist the exact portfolio-risk parameter version resolved by order admission.
-- Historical orders are intentionally not backfilled: 0/0 means not applicable or
-- created before this evidence became mandatory.

SET @schema_name = DATABASE();

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name AND table_name='t_option_order'
      AND column_name='portfolio_risk_config_id'
  ),
  'SELECT 1',
  'ALTER TABLE `t_option_order` ADD COLUMN `portfolio_risk_config_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''组合保证金准入采用的参数ID；非组合保证金卖单或迁移前历史单为0'' AFTER `margin_coin`'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=@schema_name AND table_name='t_option_order'
      AND column_name='portfolio_risk_config_version'
  ),
  'SELECT 1',
  'ALTER TABLE `t_option_order` ADD COLUMN `portfolio_risk_config_version` BIGINT NOT NULL DEFAULT 0 COMMENT ''组合保证金准入采用的参数版本；与参数ID成对保存'' AFTER `portfolio_risk_config_id`'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_schema=@schema_name AND table_name='t_option_order'
      AND constraint_name='chk_option_order_portfolio_config_pair'
  ),
  'SELECT 1',
  'ALTER TABLE `t_option_order` ADD CONSTRAINT `chk_option_order_portfolio_config_pair` CHECK ((`portfolio_risk_config_id`=0 AND `portfolio_risk_config_version`=0) OR (`portfolio_risk_config_id`>0 AND `portfolio_risk_config_version`>0))'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DROP TRIGGER IF EXISTS `trg_option_order_portfolio_config_insert`;
DROP TRIGGER IF EXISTS `trg_option_order_portfolio_config_update`;

DELIMITER $$

CREATE TRIGGER `trg_option_order_portfolio_config_insert`
BEFORE INSERT ON `t_option_order`
FOR EACH ROW
BEGIN
  DECLARE portfolio_mode BIGINT DEFAULT 0;
  DECLARE config_match_count BIGINT DEFAULT 0;

  IF NOT (
    (NEW.portfolio_risk_config_id=0 AND NEW.portfolio_risk_config_version=0)
    OR (NEW.portfolio_risk_config_id>0 AND NEW.portfolio_risk_config_version>0)
  ) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='portfolio risk config id/version must be 0/0 or both positive';
  END IF;

  SELECT `seller_margin_mode`
    INTO portfolio_mode
  FROM `t_option_contract`
  WHERE `tenant_id`=NEW.tenant_id AND `id`=NEW.contract_id
  LIMIT 1;

  IF NEW.side=2 AND portfolio_mode=3 THEN
    IF NEW.portfolio_risk_config_id<=0 OR NEW.portfolio_risk_config_version<=0 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='portfolio seller order requires admission risk config evidence';
    END IF;

    SELECT COUNT(1)
      INTO config_match_count
    FROM `t_option_portfolio_risk_config` config
    JOIN `t_option_contract` contract
      ON contract.tenant_id=config.tenant_id
     AND contract.id=NEW.contract_id
     AND contract.settle_coin=config.settle_coin
    WHERE config.id=NEW.portfolio_risk_config_id
      AND config.version=NEW.portfolio_risk_config_version
      AND config.tenant_id=NEW.tenant_id
      AND config.status IN (2,4)
      AND config.effective_from<=NEW.create_times
      AND (config.effective_until=0 OR config.effective_until>NEW.create_times);

    IF config_match_count<>1 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='portfolio seller order references invalid or inactive risk config';
    END IF;
  ELSEIF NEW.portfolio_risk_config_id<>0 OR NEW.portfolio_risk_config_version<>0 THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='non-portfolio-seller order must not reference portfolio risk config';
  END IF;
END$$

CREATE TRIGGER `trg_option_order_portfolio_config_update`
BEFORE UPDATE ON `t_option_order`
FOR EACH ROW
BEGIN
  IF NEW.tenant_id<>OLD.tenant_id
     OR NEW.contract_id<>OLD.contract_id
     OR NEW.side<>OLD.side
     OR NEW.create_times<>OLD.create_times THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='order portfolio admission identity is immutable';
  END IF;

  IF NEW.portfolio_risk_config_id<>OLD.portfolio_risk_config_id
     OR NEW.portfolio_risk_config_version<>OLD.portfolio_risk_config_version THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='order portfolio risk config evidence is immutable';
  END IF;
END$$

DELIMITER ;
