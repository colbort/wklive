-- P2-004：公司行动、后继合约、逐持仓及保证金批次迁移审计。
-- 自动路径仅支持单一现金结算后继合约和可精确表示的换算；其他事件保持停牌并进入人工处理。

SET @option_halt_check_exists = (
  SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA=DATABASE()
    AND TABLE_NAME='t_option_trading_halt'
    AND CONSTRAINT_NAME='chk_option_trading_halt'
    AND CONSTRAINT_TYPE='CHECK'
);
SET @option_halt_drop_check_sql = IF(
  @option_halt_check_exists>0,
  'ALTER TABLE t_option_trading_halt DROP CHECK chk_option_trading_halt',
  'SELECT 1'
);
PREPARE option_halt_drop_check_stmt FROM @option_halt_drop_check_sql;
EXECUTE option_halt_drop_check_stmt;
DEALLOCATE PREPARE option_halt_drop_check_stmt;
ALTER TABLE t_option_trading_halt ADD CONSTRAINT chk_option_trading_halt CHECK (
  `tenant_id` > 0 AND `halt_no` <> '' AND `active_key` <> '' AND `contract_id` > 0
  AND `source` IN (1,2,3,4) AND `status` IN (1,2) AND `reason` <> ''
  AND `started_at` > 0 AND `cancel_total` >= 0 AND `cancel_success` >= 0 AND `cancel_failed` >= 0
  AND `cancel_success` + `cancel_failed` <= `cancel_total`
  AND (
    (`status` = 1 AND `active_key` = CONCAT('CONTRACT:',`contract_id`) AND `lifted_at` = 0 AND `lifted_by` = 0)
    OR
    (`status` = 2 AND `active_key` = CONCAT('HALT:',`halt_no`) AND `lifted_at` > 0 AND `lifted_by` >= 0 AND `lift_reason` <> '')
  )
);

CREATE TABLE IF NOT EXISTS `t_option_corporate_action` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `event_no` VARCHAR(96) NOT NULL DEFAULT '',
  `external_event_ref` VARCHAR(128) NOT NULL DEFAULT '',
  `version` BIGINT NOT NULL DEFAULT 1,
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '',
  `action_type` TINYINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `announcement_time` BIGINT NOT NULL DEFAULT 0,
  `ex_time` BIGINT NOT NULL DEFAULT 0,
  `record_time` BIGINT NOT NULL DEFAULT 0,
  `effective_time` BIGINT NOT NULL DEFAULT 0,
  `pay_time` BIGINT NOT NULL DEFAULT 0,
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '',
  `evidence_hash` VARCHAR(128) NOT NULL DEFAULT '',
  `description` VARCHAR(1000) NOT NULL DEFAULT '',
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `reviewed_by` BIGINT NOT NULL DEFAULT 0,
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0,
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_event_no` (`tenant_id`,`event_no`),
  UNIQUE KEY `uk_corporate_action_external_version` (`tenant_id`,`external_event_ref`,`version`),
  KEY `idx_corporate_action_due` (`tenant_id`,`status`,`effective_time`,`id`),
  KEY `idx_corporate_action_underlying` (`tenant_id`,`underlying_symbol`,`id`),
  CONSTRAINT `chk_option_corporate_action` CHECK (
    `tenant_id` > 0 AND `event_no` <> '' AND `external_event_ref` <> '' AND `version` > 0
    AND `underlying_symbol` <> '' AND `action_type` IN (1,2,3,4,5,6,7,8,9,10)
    AND `status` IN (1,2,3,4,5,6,7)
    AND `announcement_time` > 0 AND `effective_time` > 0
    AND `evidence_ref` <> '' AND `evidence_hash` <> '' AND `description` <> ''
    AND `created_by` > 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='不可覆盖的期权公司行动事件版本';

CREATE TABLE IF NOT EXISTS `t_option_corporate_action_contract` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `action_id` BIGINT NOT NULL DEFAULT 0,
  `source_contract_id` BIGINT NOT NULL DEFAULT 0,
  `successor_contract_id` BIGINT NOT NULL DEFAULT 0,
  `execution_mode` TINYINT NOT NULL DEFAULT 0,
  `quantity_numerator` DECIMAL(32,0) NOT NULL DEFAULT 1,
  `quantity_denominator` DECIMAL(32,0) NOT NULL DEFAULT 1,
  `halt_id` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `position_total` BIGINT NOT NULL DEFAULT 0,
  `position_completed` BIGINT NOT NULL DEFAULT 0,
  `position_failed` BIGINT NOT NULL DEFAULT 0,
  `last_position_id` BIGINT NOT NULL DEFAULT 0,
  `retry_count` INT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_source` (`tenant_id`,`action_id`,`source_contract_id`),
  KEY `idx_corporate_action_contract_status` (`tenant_id`,`action_id`,`status`,`id`),
  CONSTRAINT `chk_option_corporate_action_contract` CHECK (
    `tenant_id` > 0 AND `action_id` > 0 AND `source_contract_id` > 0
    AND `execution_mode` IN (1,2)
    AND ((`execution_mode` = 1 AND `successor_contract_id` > 0) OR `execution_mode` = 2)
    AND `quantity_numerator` > 0 AND `quantity_denominator` > 0
    AND `status` IN (1,2,3,4,5,6)
    AND `position_total` >= 0 AND `position_completed` >= 0 AND `position_failed` >= 0
    AND `retry_count` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公司行动受影响合约及后继映射';

CREATE TABLE IF NOT EXISTS `t_option_corporate_action_position` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `action_id` BIGINT NOT NULL DEFAULT 0,
  `action_contract_id` BIGINT NOT NULL DEFAULT 0,
  `source_position_id` BIGINT NOT NULL DEFAULT 0,
  `successor_position_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `account_id` BIGINT NOT NULL DEFAULT 0,
  `side` TINYINT NOT NULL DEFAULT 0,
  `source_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `successor_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `source_available_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `successor_available_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `source_open_avg_price` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `successor_open_avg_price` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `source_effective_multiplier` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `successor_effective_multiplier` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `cost_basis_before` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `cost_basis_after` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `cash_difference` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 1,
  `retry_count` INT NOT NULL DEFAULT 0,
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `completed_at` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_position` (`tenant_id`,`action_contract_id`,`source_position_id`),
  KEY `idx_corporate_action_position_status` (`tenant_id`,`action_contract_id`,`status`,`source_position_id`),
  CONSTRAINT `chk_option_corporate_action_position` CHECK (
    `tenant_id` > 0 AND `action_id` > 0 AND `action_contract_id` > 0
    AND `source_position_id` > 0 AND `user_id` > 0 AND `side` IN (1,2)
    AND `source_quantity` > 0 AND `successor_quantity` > 0
    AND `source_available_quantity` >= 0 AND `successor_available_quantity` >= 0
    AND `source_effective_multiplier` > 0 AND `successor_effective_multiplier` > 0
    AND `cost_basis_before` >= 0 AND `cost_basis_after` >= 0
    AND `status` IN (1,2,3) AND `retry_count` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公司行动逐持仓不可覆盖迁移审计';

CREATE TABLE IF NOT EXISTS `t_option_corporate_action_margin_lot` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `action_position_id` BIGINT NOT NULL DEFAULT 0,
  `margin_lot_id` BIGINT NOT NULL DEFAULT 0,
  `source_contract_id` BIGINT NOT NULL DEFAULT 0,
  `successor_contract_id` BIGINT NOT NULL DEFAULT 0,
  `source_position_id` BIGINT NOT NULL DEFAULT 0,
  `successor_position_id` BIGINT NOT NULL DEFAULT 0,
  `source_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `successor_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `source_remaining_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `successor_remaining_quantity` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_corporate_action_margin_lot` (`tenant_id`,`action_position_id`,`margin_lot_id`),
  KEY `idx_corporate_action_margin_lot` (`tenant_id`,`margin_lot_id`,`id`),
  CONSTRAINT `chk_option_corporate_action_margin_lot` CHECK (
    `tenant_id` > 0 AND `action_position_id` > 0 AND `margin_lot_id` > 0
    AND `source_contract_id` > 0 AND `successor_contract_id` > 0
    AND `source_position_id` > 0 AND `successor_position_id` > 0
    AND `source_quantity` > 0 AND `successor_quantity` > 0
    AND `source_remaining_quantity` >= 0 AND `successor_remaining_quantity` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公司行动保证金批次换算审计';

SET @option_margin_origin_contract_exists = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='t_option_margin_lot' AND COLUMN_NAME='origin_contract_id'
);
SET @option_margin_origin_contract_sql = IF(
  @option_margin_origin_contract_exists=0,
  'ALTER TABLE t_option_margin_lot ADD COLUMN origin_contract_id BIGINT NOT NULL DEFAULT 0 AFTER position_id',
  'SELECT 1'
);
PREPARE option_margin_origin_contract_stmt FROM @option_margin_origin_contract_sql;
EXECUTE option_margin_origin_contract_stmt;
DEALLOCATE PREPARE option_margin_origin_contract_stmt;

SET @option_margin_origin_position_exists = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='t_option_margin_lot' AND COLUMN_NAME='origin_position_id'
);
SET @option_margin_origin_position_sql = IF(
  @option_margin_origin_position_exists=0,
  'ALTER TABLE t_option_margin_lot ADD COLUMN origin_position_id BIGINT NOT NULL DEFAULT 0 AFTER origin_contract_id',
  'SELECT 1'
);
PREPARE option_margin_origin_position_stmt FROM @option_margin_origin_position_sql;
EXECUTE option_margin_origin_position_stmt;
DEALLOCATE PREPARE option_margin_origin_position_stmt;

SET @option_margin_action_position_exists = (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='t_option_margin_lot' AND COLUMN_NAME='corporate_action_position_id'
);
SET @option_margin_action_position_sql = IF(
  @option_margin_action_position_exists=0,
  'ALTER TABLE t_option_margin_lot ADD COLUMN corporate_action_position_id BIGINT NOT NULL DEFAULT 0 AFTER origin_position_id',
  'SELECT 1'
);
PREPARE option_margin_action_position_stmt FROM @option_margin_action_position_sql;
EXECUTE option_margin_action_position_stmt;
DEALLOCATE PREPARE option_margin_action_position_stmt;

UPDATE t_option_margin_lot
SET origin_contract_id=contract_id
WHERE origin_contract_id=0;

UPDATE t_option_margin_lot
SET origin_position_id=position_id
WHERE origin_position_id=0 AND position_id>0;

DROP TRIGGER IF EXISTS trg_option_margin_lot_origin_insert;
DELIMITER $$
CREATE TRIGGER trg_option_margin_lot_origin_insert
BEFORE INSERT ON t_option_margin_lot
FOR EACH ROW
BEGIN
  IF NEW.origin_contract_id=0 THEN
    SET NEW.origin_contract_id=NEW.contract_id;
  END IF;
  IF NEW.origin_position_id=0 AND NEW.position_id>0 THEN
    SET NEW.origin_position_id=NEW.position_id;
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_margin_lot_origin_update;
DELIMITER $$
CREATE TRIGGER trg_option_margin_lot_origin_update
BEFORE UPDATE ON t_option_margin_lot
FOR EACH ROW
BEGIN
  IF NEW.origin_contract_id<>OLD.origin_contract_id
    OR (OLD.origin_position_id>0 AND NEW.origin_position_id<>OLD.origin_position_id)
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='margin lot origin identity is immutable';
  END IF;
  IF OLD.origin_position_id=0 AND NEW.position_id>0 THEN
    SET NEW.origin_position_id=NEW.position_id;
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_immutable;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_immutable
BEFORE UPDATE ON t_option_corporate_action
FOR EACH ROW
BEGIN
  IF NEW.tenant_id<>OLD.tenant_id OR NEW.event_no<>OLD.event_no
    OR NEW.external_event_ref<>OLD.external_event_ref OR NEW.version<>OLD.version
    OR NEW.underlying_symbol<>OLD.underlying_symbol OR NEW.action_type<>OLD.action_type
    OR NEW.announcement_time<>OLD.announcement_time OR NEW.ex_time<>OLD.ex_time
    OR NEW.record_time<>OLD.record_time OR NEW.effective_time<>OLD.effective_time
    OR NEW.pay_time<>OLD.pay_time OR NEW.evidence_ref<>OLD.evidence_ref
    OR NEW.evidence_hash<>OLD.evidence_hash OR NEW.description<>OLD.description
    OR NEW.created_by<>OLD.created_by OR NEW.create_times<>OLD.create_times
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='corporate action evidence and economics are immutable';
  END IF;
  IF OLD.status IN (3,5,6) AND NEW.status<>OLD.status THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='terminal corporate action is immutable';
  END IF;
  IF NOT (
    NEW.status=OLD.status
    OR (OLD.status=1 AND NEW.status IN (2,3,6))
    OR (OLD.status=2 AND NEW.status IN (4,6,7))
    OR (OLD.status=4 AND NEW.status IN (5,6,7))
    OR (OLD.status=7 AND NEW.status IN (4,6,7))
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='invalid corporate action status transition';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_contract_immutable;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_contract_immutable
BEFORE UPDATE ON t_option_corporate_action_contract
FOR EACH ROW
BEGIN
  IF NEW.tenant_id<>OLD.tenant_id OR NEW.action_id<>OLD.action_id
    OR NEW.source_contract_id<>OLD.source_contract_id
    OR NEW.successor_contract_id<>OLD.successor_contract_id
    OR NEW.execution_mode<>OLD.execution_mode
    OR NEW.quantity_numerator<>OLD.quantity_numerator
    OR NEW.quantity_denominator<>OLD.quantity_denominator
    OR NEW.create_times<>OLD.create_times
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='corporate action contract mapping is immutable';
  END IF;
  IF NOT (NEW.halt_id=OLD.halt_id OR (OLD.halt_id=0 AND NEW.halt_id>0)) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='corporate action halt identity is immutable';
  END IF;
  IF NEW.position_total<OLD.position_total
    OR NEW.position_completed<OLD.position_completed
    OR NEW.position_failed<OLD.position_failed
    OR NEW.last_position_id<OLD.last_position_id
    OR NEW.retry_count<OLD.retry_count
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='corporate action progress is monotonic';
  END IF;
  IF NOT (
    NEW.status=OLD.status
    OR (OLD.status=1 AND NEW.status IN (2,5))
    OR (OLD.status=2 AND NEW.status IN (3,5,6))
    OR (OLD.status=3 AND NEW.status IN (4,5,6))
    OR (OLD.status=6 AND NEW.status IN (3,5,6))
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT='invalid corporate action contract status transition';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_position_immutable;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_position_immutable
BEFORE UPDATE ON t_option_corporate_action_position
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT='corporate action position audit is append only';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_margin_lot_immutable;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_margin_lot_immutable
BEFORE UPDATE ON t_option_corporate_action_margin_lot
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT='corporate action margin lot audit is append only';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_no_delete
BEFORE DELETE ON t_option_corporate_action
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT='corporate action history cannot be deleted';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_contract_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_contract_no_delete
BEFORE DELETE ON t_option_corporate_action_contract
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT='corporate action contract history cannot be deleted';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_position_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_position_no_delete
BEFORE DELETE ON t_option_corporate_action_position
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT='corporate action position history cannot be deleted';
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_corporate_action_margin_lot_no_delete;
DELIMITER $$
CREATE TRIGGER trg_option_corporate_action_margin_lot_no_delete
BEFORE DELETE ON t_option_corporate_action_margin_lot
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT='corporate action margin lot history cannot be deleted';
END$$
DELIMITER ;
