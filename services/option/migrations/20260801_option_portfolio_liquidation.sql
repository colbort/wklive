-- Account-level portfolio-margin liquidation audit snapshot.
-- Existing rows are isolated-position liquidations by construction. Every DDL
-- operation is baseline-safe so release automation can run this migration more
-- than once against either the old schema or the canonical schema.
SET @schema_name = DATABASE();

SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='liquidation_scope'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `liquidation_scope` TINYINT NOT NULL DEFAULT 1 COMMENT ''强平范围：1逐仓仓位 2组合钱包'' AFTER `deficit_resolution`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='portfolio_risk_config_id'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `portfolio_risk_config_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''组合强平风险参数快照ID'' AFTER `liquidation_scope`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='portfolio_risk_config_version'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `portfolio_risk_config_version` BIGINT NOT NULL DEFAULT 0 COMMENT ''组合强平风险参数版本'' AFTER `portfolio_risk_config_id`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='portfolio_maintenance_before'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `portfolio_maintenance_before` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''移除仓位前组合维持保证金'' AFTER `portfolio_risk_config_version`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='portfolio_maintenance_after'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `portfolio_maintenance_after` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''移除仓位后组合维持保证金'' AFTER `portfolio_maintenance_before`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='portfolio_initial_after'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `portfolio_initial_after` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''移除仓位后组合初始保证金'' AFTER `portfolio_maintenance_after`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='portfolio_collateral_before'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `portfolio_collateral_before` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''组合强平前可用抵押'' AFTER `portfolio_initial_after`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND column_name='portfolio_collateral_after'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD COLUMN `portfolio_collateral_after` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''本次消费后保留的组合抵押'' AFTER `portfolio_collateral_before`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
SET @ddl = IF(EXISTS(SELECT 1 FROM information_schema.statistics WHERE table_schema=@schema_name AND table_name='t_option_liquidation' AND index_name='idx_liquidation_portfolio_wallet'),
  'SELECT 1', 'ALTER TABLE `t_option_liquidation` ADD KEY `idx_liquidation_portfolio_wallet` (`tenant_id`,`user_id`,`liquidation_scope`,`status`,`id`)');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

ALTER TABLE `t_option_liquidation`
  MODIFY COLUMN `status` TINYINT NOT NULL DEFAULT 1
    COMMENT '状态：1待处理 2执行中 3完成 4失败 5破产 6人工 7快照失效取消';

-- CANCELED=7 is a terminal state for an execution-time stale portfolio risk
-- snapshot. The pre-existing check allowed only 1..6, which would correctly
-- detect recovery but reject the safe terminal transition.
SET @check_name = (
  SELECT tc.constraint_name
  FROM information_schema.table_constraints tc
  WHERE tc.constraint_schema=@schema_name
    AND tc.table_name='t_option_liquidation'
    AND tc.constraint_type='CHECK'
    AND tc.constraint_name='chk_option_liquidation'
  LIMIT 1
);
SET @ddl = IF(@check_name IS NULL,
  'SELECT 1',
  'ALTER TABLE `t_option_liquidation` DROP CHECK `chk_option_liquidation`');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

ALTER TABLE `t_option_liquidation`
  ADD CONSTRAINT `chk_option_liquidation` CHECK (
    `quantity` > 0 AND `deficit_amount` >= 0 AND `liquidation_fee` >= 0
    AND `status` IN (1,2,3,4,5,6,7)
    AND `retry_count` >= 0 AND `insurance_attempt` >= 0
    AND `backstop_amount` >= 0 AND `deficit_resolution` IN (1,2,3,4,5)
  );

DROP TRIGGER IF EXISTS `trg_option_liquidation_evidence_insert`;
DROP TRIGGER IF EXISTS `trg_option_liquidation_evidence_update`;

DELIMITER $$

CREATE TRIGGER `trg_option_liquidation_evidence_insert`
BEFORE INSERT ON `t_option_liquidation`
FOR EACH ROW
BEGIN
  DECLARE config_match_count BIGINT DEFAULT 0;

  IF NEW.liquidation_scope=1 THEN
    IF NEW.portfolio_risk_config_id<>0 OR NEW.portfolio_risk_config_version<>0
       OR NEW.portfolio_maintenance_before<>0 OR NEW.portfolio_maintenance_after<>0
       OR NEW.portfolio_initial_after<>0 OR NEW.portfolio_collateral_before<>0
       OR NEW.portfolio_collateral_after<>0 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='isolated liquidation cannot contain portfolio evidence';
    END IF;
  ELSEIF NEW.liquidation_scope=2 THEN
    IF NEW.account_id<>0 OR NEW.portfolio_risk_config_id<=0 OR NEW.portfolio_risk_config_version<=0
       OR NEW.portfolio_maintenance_before<=NEW.portfolio_maintenance_after
       OR NEW.portfolio_maintenance_after<0 OR NEW.portfolio_initial_after<0
       OR NEW.portfolio_collateral_before<NEW.portfolio_collateral_after
       OR NEW.portfolio_collateral_after<NEW.portfolio_initial_after THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='invalid portfolio liquidation risk and collateral evidence';
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
      AND config.status=2
      AND config.effective_from<=NEW.create_times
      AND (config.effective_until=0 OR config.effective_until>NEW.create_times);
    IF config_match_count<>1 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='portfolio liquidation references invalid active risk config';
    END IF;
  ELSE
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='invalid option liquidation scope';
  END IF;
END$$

CREATE TRIGGER `trg_option_liquidation_evidence_update`
BEFORE UPDATE ON `t_option_liquidation`
FOR EACH ROW
BEGIN
  IF NEW.tenant_id<>OLD.tenant_id OR NEW.liquidation_no<>OLD.liquidation_no
     OR NEW.user_id<>OLD.user_id OR NEW.account_id<>OLD.account_id
     OR NEW.contract_id<>OLD.contract_id OR NEW.position_id<>OLD.position_id
     OR NEW.quantity<>OLD.quantity OR NEW.mark_price<>OLD.mark_price
     OR NEW.liquidation_scope<>OLD.liquidation_scope
     OR NEW.portfolio_risk_config_id<>OLD.portfolio_risk_config_id
     OR NEW.portfolio_risk_config_version<>OLD.portfolio_risk_config_version
     OR NEW.portfolio_maintenance_before<>OLD.portfolio_maintenance_before
     OR NEW.portfolio_maintenance_after<>OLD.portfolio_maintenance_after
     OR NEW.portfolio_initial_after<>OLD.portfolio_initial_after
     OR NEW.portfolio_collateral_before<>OLD.portfolio_collateral_before
     OR NEW.portfolio_collateral_after<>OLD.portfolio_collateral_after THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='liquidation identity and portfolio evidence are immutable';
  END IF;
END$$

DELIMITER ;
