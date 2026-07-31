-- 行权治理：独立截止时间、自动行权阈值、版本化 DNE/相反指令，
-- 并补充持仓已实现收益分项。可重复执行。

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract'
      AND column_name='exercise_cutoff_time'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN exercise_cutoff_time BIGINT NOT NULL DEFAULT 0 COMMENT ''主动行权及到期指令截止时间'' AFTER deliver_time'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_contract'
      AND column_name='auto_exercise_threshold'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_contract
     ADD COLUMN auto_exercise_threshold DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''每单位内在价值自动行权阈值'' AFTER exercise_cutoff_time'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE t_option_contract
SET exercise_cutoff_time = expire_time
WHERE exercise_cutoff_time = 0
  AND expire_time > 0;

CREATE TABLE IF NOT EXISTS `t_option_exercise_instruction` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `user_id` BIGINT NOT NULL DEFAULT 0 COMMENT '用户ID',
  `account_id` BIGINT NOT NULL DEFAULT 0 COMMENT '交易账户ID',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '合约ID',
  `position_id` BIGINT NOT NULL DEFAULT 0 COMMENT '多头持仓ID',
  `client_instruction_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端幂等号',
  `instruction_type` TINYINT NOT NULL DEFAULT 1 COMMENT '指令：1自动 2放弃 3相反行权',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '持仓指令版本',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1生效 2已被新版本替代',
  `supersedes_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被本版本替代的指令ID',
  `cutoff_time` BIGINT NOT NULL DEFAULT 0 COMMENT '提交时适用的截止时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_exercise_instruction_client` (`tenant_id`, `user_id`, `client_instruction_id`),
  UNIQUE KEY `uk_exercise_instruction_version` (`tenant_id`, `position_id`, `version`),
  KEY `idx_exercise_instruction_active` (`tenant_id`, `contract_id`, `position_id`, `status`, `version`),
  CONSTRAINT `chk_option_exercise_instruction` CHECK (
    `instruction_type` IN (1,2,3)
    AND `version` > 0
    AND `status` IN (1,2)
    AND `client_instruction_id` <> ''
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权到期行权指令版本';

-- 已提交指令只允许状态迁移和更新时间变化；更换指令必须新增版本，
-- 避免 DNE/相反行权在截止后被原地改写。
DROP TRIGGER IF EXISTS trg_option_exercise_instruction_immutable;
DELIMITER $$
CREATE TRIGGER trg_option_exercise_instruction_immutable
BEFORE UPDATE ON t_option_exercise_instruction
FOR EACH ROW
BEGIN
  IF NOT (
    OLD.tenant_id <=> NEW.tenant_id
    AND OLD.user_id <=> NEW.user_id
    AND OLD.account_id <=> NEW.account_id
    AND OLD.contract_id <=> NEW.contract_id
    AND OLD.position_id <=> NEW.position_id
    AND OLD.client_instruction_id <=> NEW.client_instruction_id
    AND OLD.instruction_type <=> NEW.instruction_type
    AND OLD.version <=> NEW.version
    AND OLD.supersedes_id <=> NEW.supersedes_id
    AND OLD.cutoff_time <=> NEW.cutoff_time
    AND OLD.create_times <=> NEW.create_times
  ) THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'exercise instruction economic fields are immutable; create a new version';
  END IF;
END$$
DELIMITER ;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_position'
      AND column_name='trade_realized_pnl'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_position
     ADD COLUMN trade_realized_pnl DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''平仓产生的权利金交易毛盈亏'' AFTER realized_pnl'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_position'
      AND column_name='settlement_realized_pnl'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_position
     ADD COLUMN settlement_realized_pnl DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''行权/到期产生的结算毛盈亏'' AFTER trade_realized_pnl'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_position'
      AND column_name='fee_paid'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_position
     ADD COLUMN fee_paid DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''归属于持仓的累计交易/行权/强平费用'' AFTER settlement_realized_pnl'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_position'
      AND column_name='total_return'
  ),
  'SELECT 1',
  'ALTER TABLE t_option_position
     ADD COLUMN total_return DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT ''累计已实现总收益=交易+结算-费用'' AFTER fee_paid'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 历史 realized_pnl 无法可靠拆分来源，保守归入交易毛盈亏并保持总数不变。
UPDATE t_option_position
SET trade_realized_pnl = realized_pnl,
    settlement_realized_pnl = 0,
    fee_paid = 0,
    total_return = realized_pnl
WHERE trade_realized_pnl = 0
  AND settlement_realized_pnl = 0
  AND fee_paid = 0
  AND total_return = 0
  AND realized_pnl <> 0;
