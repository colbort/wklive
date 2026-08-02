-- OPT-P0-007: governed platform-backstop policies and transaction-atomic limits.
-- Existing backstop cover rows are preserved as policy_id=0 historical evidence;
-- the insert trigger requires every new row to carry an approved policy snapshot.

CREATE TABLE IF NOT EXISTS `t_asset_backstop_policy` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL,
  `coin` VARCHAR(32) NOT NULL,
  `request_no` VARCHAR(96) NOT NULL COMMENT '创建幂等键',
  `version` BIGINT NOT NULL,
  `mode` TINYINT NOT NULL COMMENT '1禁用 2预注资 3信用底线',
  `per_request_limit` DECIMAL(36,18) NOT NULL DEFAULT 0,
  `daily_limit` DECIMAL(36,18) NOT NULL DEFAULT 0,
  `balance_floor` DECIMAL(36,18) NOT NULL DEFAULT 0,
  `effective_from` BIGINT NOT NULL COMMENT '毫秒UTC',
  `effective_until` BIGINT NOT NULL COMMENT '毫秒UTC，有限有效期',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1草稿 2批准 3拒绝',
  `reason` VARCHAR(255) NOT NULL,
  `evidence_ref` VARCHAR(255) NOT NULL,
  `created_by` BIGINT NOT NULL,
  `reviewed_by` BIGINT NOT NULL DEFAULT 0,
  `review_reason` VARCHAR(255) NOT NULL DEFAULT '',
  `create_times` BIGINT NOT NULL,
  `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_backstop_policy_request` (`tenant_id`,`request_no`),
  UNIQUE KEY `uk_backstop_policy_version` (`tenant_id`,`coin`,`version`),
  KEY `idx_backstop_policy_effective` (`tenant_id`,`coin`,`status`,`effective_from`,`effective_until`,`version`),
  CONSTRAINT `chk_asset_backstop_policy` CHECK (
    `tenant_id` > 0 AND `coin` <> '' AND `request_no` <> '' AND `version` > 0
    AND `mode` IN (1,2,3) AND `effective_from` > 0 AND `effective_until` > `effective_from`
    AND `effective_until` - `effective_from` <= 31622400000
    AND `reason` <> '' AND `evidence_ref` <> '' AND `created_by` > 0
    AND (
      (`mode` = 1 AND `per_request_limit` = 0 AND `daily_limit` = 0 AND `balance_floor` = 0)
      OR (`mode` = 2 AND `per_request_limit` > 0 AND `per_request_limit` <= `daily_limit` AND `balance_floor` = 0)
      OR (`mode` = 3 AND `per_request_limit` > 0 AND `per_request_limit` <= `daily_limit` AND `balance_floor` < 0)
    )
    AND (
      (`status` = 1 AND `reviewed_by` = 0 AND `review_reason` = '')
      OR (`status` IN (2,3) AND `reviewed_by` > 0 AND `reviewed_by` <> `created_by` AND `review_reason` <> '')
    )
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option平台兜底版本化资金政策';

CREATE TABLE IF NOT EXISTS `t_asset_backstop_usage_daily` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT NOT NULL,
  `coin` VARCHAR(32) NOT NULL,
  `usage_day` CHAR(8) NOT NULL COMMENT 'Asset服务端UTC YYYYMMDD',
  `covered_amount` DECIMAL(36,18) NOT NULL,
  `last_policy_id` BIGINT NOT NULL,
  `create_times` BIGINT NOT NULL,
  `update_times` BIGINT NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_backstop_usage_day` (`tenant_id`,`coin`,`usage_day`),
  KEY `idx_backstop_usage_policy` (`tenant_id`,`last_policy_id`,`usage_day`),
  CONSTRAINT `chk_asset_backstop_usage_daily` CHECK (
    `tenant_id` > 0 AND `coin` <> '' AND `usage_day` REGEXP '^[0-9]{8}$'
    AND `covered_amount` > 0 AND `last_policy_id` > 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option平台兜底UTC日累计硬额度';

SET @asset_backstop_cover_policy_sql = (
  SELECT IF(
    EXISTS (
      SELECT 1 FROM information_schema.columns
      WHERE table_schema = DATABASE() AND table_name = 't_asset_backstop_cover'
        AND column_name = 'policy_id'
    ),
    'SELECT 1',
    'ALTER TABLE `t_asset_backstop_cover`
       ADD COLUMN `policy_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''0仅表示迁移前历史记录'' AFTER `liquidation_no`,
       ADD COLUMN `policy_version` BIGINT NOT NULL DEFAULT 0 AFTER `policy_id`,
       ADD COLUMN `policy_mode` TINYINT NOT NULL DEFAULT 0 COMMENT ''1禁用 2预注资 3信用底线'' AFTER `policy_version`,
       ADD COLUMN `daily_used_before` DECIMAL(36,18) NOT NULL DEFAULT 0 AFTER `covered_amount`,
       ADD COLUMN `daily_used_after` DECIMAL(36,18) NOT NULL DEFAULT 0 AFTER `daily_used_before`,
       ADD COLUMN `balance_floor` DECIMAL(36,18) NOT NULL DEFAULT 0 AFTER `daily_used_after`,
       ADD COLUMN `balance_before` DECIMAL(36,18) NOT NULL DEFAULT 0 AFTER `balance_floor`,
       ADD COLUMN `balance_after` DECIMAL(36,18) NOT NULL DEFAULT 0 AFTER `balance_before`,
       ADD KEY `idx_backstop_policy_time` (`tenant_id`,`policy_id`,`create_times`)'
  )
);
PREPARE asset_backstop_cover_policy_stmt FROM @asset_backstop_cover_policy_sql;
EXECUTE asset_backstop_cover_policy_stmt;
DEALLOCATE PREPARE asset_backstop_cover_policy_stmt;

DROP TRIGGER IF EXISTS `trg_asset_backstop_policy_insert`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_policy_update`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_policy_delete`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_usage_insert`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_usage_update`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_usage_delete`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_cover_insert_policy`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_cover_update`;
DROP TRIGGER IF EXISTS `trg_asset_backstop_cover_delete`;

DELIMITER $$
CREATE TRIGGER `trg_asset_backstop_policy_insert`
BEFORE INSERT ON `t_asset_backstop_policy`
FOR EACH ROW
BEGIN
  IF NEW.status <> 1 OR NEW.reviewed_by <> 0 OR NEW.review_reason <> ''
     OR NEW.effective_from <= CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'backstop policy must be created as an unreviewed draft';
  END IF;
END$$

CREATE TRIGGER `trg_asset_backstop_policy_update`
BEFORE UPDATE ON `t_asset_backstop_policy`
FOR EACH ROW
BEGIN
  IF OLD.status <> 1 OR NEW.status NOT IN (2,3)
     OR NEW.reviewed_by <= 0 OR NEW.reviewed_by = OLD.created_by OR NEW.review_reason = ''
     OR NOT (OLD.tenant_id <=> NEW.tenant_id)
     OR NOT (OLD.coin <=> NEW.coin)
     OR NOT (OLD.request_no <=> NEW.request_no)
     OR NOT (OLD.version <=> NEW.version)
     OR NOT (OLD.mode <=> NEW.mode)
     OR NOT (OLD.per_request_limit <=> NEW.per_request_limit)
     OR NOT (OLD.daily_limit <=> NEW.daily_limit)
     OR NOT (OLD.balance_floor <=> NEW.balance_floor)
     OR NOT (OLD.effective_from <=> NEW.effective_from)
     OR NOT (OLD.effective_until <=> NEW.effective_until)
     OR NOT (OLD.reason <=> NEW.reason)
     OR NOT (OLD.evidence_ref <=> NEW.evidence_ref)
     OR NOT (OLD.created_by <=> NEW.created_by)
     OR NOT (OLD.create_times <=> NEW.create_times)
     OR NEW.update_times < OLD.update_times
     OR (NEW.status = 2 AND NEW.effective_until <= CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED)) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'invalid or non-four-eyes backstop policy transition';
  END IF;
END$$

CREATE TRIGGER `trg_asset_backstop_policy_delete`
BEFORE DELETE ON `t_asset_backstop_policy`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'backstop policy evidence cannot be deleted';
END$$

CREATE TRIGGER `trg_asset_backstop_usage_insert`
BEFORE INSERT ON `t_asset_backstop_usage_daily`
FOR EACH ROW
BEGIN
  DECLARE matching_policy BIGINT DEFAULT 0;
  SELECT COUNT(*) INTO matching_policy
    FROM `t_asset_backstop_policy` p
   WHERE p.id = NEW.last_policy_id
     AND p.tenant_id = NEW.tenant_id
     AND BINARY p.coin = BINARY NEW.coin
     AND p.status = 2 AND p.mode IN (2,3)
     AND p.effective_from <= NEW.update_times AND p.effective_until > NEW.update_times
     AND NEW.covered_amount <= p.daily_limit;
  IF NEW.covered_amount <= 0 OR NEW.last_policy_id <= 0 OR matching_policy <> 1 THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'invalid backstop daily usage';
  END IF;
END$$

CREATE TRIGGER `trg_asset_backstop_usage_update`
BEFORE UPDATE ON `t_asset_backstop_usage_daily`
FOR EACH ROW
BEGIN
  DECLARE matching_policy BIGINT DEFAULT 0;
  SELECT COUNT(*) INTO matching_policy
    FROM `t_asset_backstop_policy` p
   WHERE p.id = NEW.last_policy_id
     AND p.tenant_id = NEW.tenant_id
     AND BINARY p.coin = BINARY NEW.coin
     AND p.status = 2 AND p.mode IN (2,3)
     AND p.effective_from <= NEW.update_times AND p.effective_until > NEW.update_times
     AND NEW.covered_amount <= p.daily_limit;
  IF NOT (OLD.tenant_id <=> NEW.tenant_id)
     OR NOT (OLD.coin <=> NEW.coin)
     OR NOT (OLD.usage_day <=> NEW.usage_day)
     OR NOT (OLD.create_times <=> NEW.create_times)
     OR NEW.covered_amount <= OLD.covered_amount OR NEW.last_policy_id <= 0
     OR matching_policy <> 1 THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'invalid backstop daily usage update';
  END IF;
END$$

CREATE TRIGGER `trg_asset_backstop_usage_delete`
BEFORE DELETE ON `t_asset_backstop_usage_daily`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'backstop daily usage cannot be deleted';
END$$

CREATE TRIGGER `trg_asset_backstop_cover_insert_policy`
BEFORE INSERT ON `t_asset_backstop_cover`
FOR EACH ROW
BEGIN
  DECLARE matching_policy BIGINT DEFAULT 0;
  DECLARE matching_account BIGINT DEFAULT 0;
  DECLARE matching_usage BIGINT DEFAULT 0;
  DECLARE matching_flow BIGINT DEFAULT 0;
  DECLARE utc_usage_day CHAR(8);

  SET utc_usage_day = DATE_FORMAT(
    DATE_ADD('1970-01-01 00:00:00', INTERVAL (NEW.create_times DIV 1000) SECOND),
    '%Y%m%d'
  );
  SELECT COUNT(*) INTO matching_policy
    FROM `t_asset_backstop_policy` p
   WHERE p.id = NEW.policy_id
     AND p.tenant_id = NEW.tenant_id
     AND BINARY p.coin = BINARY NEW.coin
     AND p.version = NEW.policy_version AND p.mode = NEW.policy_mode
     AND p.status = 2 AND p.mode IN (2,3)
     AND p.effective_from <= NEW.create_times AND p.effective_until > NEW.create_times
     AND NEW.covered_amount <= p.per_request_limit
     AND NEW.daily_used_after <= p.daily_limit
     AND NEW.balance_floor = p.balance_floor;
  SELECT COUNT(*) INTO matching_account
    FROM `t_asset_platform_account` a
   WHERE a.id = NEW.platform_account_id
     AND a.tenant_id = NEW.tenant_id
     AND BINARY a.coin = BINARY NEW.coin
     AND a.account_type = 'OPTION_BACKSTOP' AND a.status = 1
     AND a.available_amount = NEW.balance_after;
  SELECT COUNT(*) INTO matching_usage
    FROM `t_asset_backstop_usage_daily` u
   WHERE u.tenant_id = NEW.tenant_id
     AND BINARY u.coin = BINARY NEW.coin
     AND u.usage_day = utc_usage_day
     AND u.covered_amount = NEW.daily_used_after
     AND u.last_policy_id = NEW.policy_id;
  SELECT COUNT(*) INTO matching_flow
    FROM `t_asset_platform_flow` f
   WHERE f.tenant_id = NEW.tenant_id
     AND f.platform_account_id = NEW.platform_account_id
     AND BINARY f.coin = BINARY NEW.coin
     AND f.account_type = 'OPTION_BACKSTOP'
     AND f.op_type = 2
     AND f.biz_type = 'platform_backstop'
     AND f.scene_type = 'platform_backstop_cover'
     AND f.biz_id = NEW.liquidation_id
     AND f.biz_no = NEW.liquidation_no
     AND f.amount = NEW.covered_amount
     AND f.before_available = NEW.balance_before
     AND f.after_available = NEW.balance_after;
  IF NEW.policy_id <= 0 OR NEW.policy_version <= 0 OR NEW.policy_mode NOT IN (2,3)
     OR NEW.daily_used_before < 0
     OR NEW.daily_used_after <> NEW.daily_used_before + NEW.covered_amount
     OR NEW.balance_after <> NEW.balance_before - NEW.covered_amount
     OR NEW.balance_after < NEW.balance_floor
     OR matching_policy <> 1 OR matching_account <> 1
     OR matching_usage <> 1 OR matching_flow <> 1 THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'invalid governed backstop cover evidence';
  END IF;
END$$

CREATE TRIGGER `trg_asset_backstop_cover_update`
BEFORE UPDATE ON `t_asset_backstop_cover`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'backstop cover evidence cannot be updated';
END$$

CREATE TRIGGER `trg_asset_backstop_cover_delete`
BEFORE DELETE ON `t_asset_backstop_cover`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'backstop cover evidence cannot be deleted';
END$$
DELIMITER ;
