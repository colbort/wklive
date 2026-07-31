-- 做市商保护：报价组、滚动成交/不利损失阈值、自动撤单和冷静期。

SET @ddl = IF(
  EXISTS(SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='t_option_order' AND column_name='mmp_group'),
  'SELECT 1',
  'ALTER TABLE t_option_order ADD COLUMN mmp_group VARCHAR(32) NOT NULL DEFAULT ''''
    COMMENT ''MMP报价组，mmp=1时必填'' AFTER mmp'
);
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- 历史 mmp=1 仅是未实现时期保存的标志，不具备保护语义，不能冒充已配置 MMP。
UPDATE t_option_order SET mmp = 2, mmp_group = ''
WHERE mmp = 1 AND mmp_group = '';

CREATE TABLE IF NOT EXISTS `t_option_mmp_config` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL DEFAULT 0,
  `contract_id` BIGINT NOT NULL DEFAULT 0,
  `group_code` VARCHAR(32) NOT NULL DEFAULT '',
  `enabled` TINYINT NOT NULL DEFAULT 2,
  `qty_threshold` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `trade_count_threshold` BIGINT NOT NULL DEFAULT 0,
  `loss_threshold` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `window_seconds` BIGINT NOT NULL DEFAULT 0,
  `cooldown_seconds` BIGINT NOT NULL DEFAULT 0,
  `status` TINYINT NOT NULL DEFAULT 3,
  `window_start` BIGINT NOT NULL DEFAULT 0,
  `accumulated_qty` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `trade_count` BIGINT NOT NULL DEFAULT 0,
  `accumulated_loss` DECIMAL(32,16) NOT NULL DEFAULT 0,
  `triggered_at` BIGINT NOT NULL DEFAULT 0,
  `cooldown_until` BIGINT NOT NULL DEFAULT 0,
  `trigger_reason` VARCHAR(64) NOT NULL DEFAULT '',
  `last_error_msg` VARCHAR(500) NOT NULL DEFAULT '',
  `created_by` BIGINT NOT NULL DEFAULT 0,
  `updated_by` BIGINT NOT NULL DEFAULT 0,
  `create_times` BIGINT NOT NULL DEFAULT 0,
  `update_times` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_mmp_config` (`tenant_id`, `user_id`, `contract_id`, `group_code`),
  KEY `idx_option_mmp_status` (`tenant_id`, `status`, `id`),
  KEY `idx_option_mmp_contract` (`tenant_id`, `contract_id`, `id`),
  CONSTRAINT `chk_option_mmp_config` CHECK (
    `tenant_id` > 0 AND `user_id` > 0 AND `contract_id` > 0 AND `group_code` <> ''
    AND `enabled` IN (1,2) AND `status` IN (1,2,3)
    AND `qty_threshold` >= 0 AND `trade_count_threshold` >= 0 AND `loss_threshold` >= 0
    AND `window_seconds` > 0 AND `cooldown_seconds` >= 0
    AND (`enabled` = 2 OR (`qty_threshold` > 0 OR `trade_count_threshold` > 0 OR `loss_threshold` > 0))
    AND `accumulated_qty` >= 0 AND `trade_count` >= 0 AND `accumulated_loss` >= 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权做市商保护配置与实时窗口';

DROP TRIGGER IF EXISTS trg_option_order_mmp_guard;
DELIMITER $$
CREATE TRIGGER trg_option_order_mmp_guard
BEFORE INSERT ON t_option_order
FOR EACH ROW
BEGIN
  IF (NEW.mmp = 1 AND NEW.mmp_group = '')
    OR (NEW.mmp <> 1 AND NEW.mmp_group <> '')
  THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'MMP order and group must be configured together';
  END IF;
END$$
DELIMITER ;

DROP TRIGGER IF EXISTS trg_option_order_mmp_guard_update;
DELIMITER $$
CREATE TRIGGER trg_option_order_mmp_guard_update
BEFORE UPDATE ON t_option_order
FOR EACH ROW
BEGIN
  IF NEW.mmp <> OLD.mmp OR NEW.mmp_group <> OLD.mmp_group THEN
    SIGNAL SQLSTATE '45000'
      SET MESSAGE_TEXT = 'MMP identity is immutable after order creation';
  END IF;
END$$
DELIMITER ;
