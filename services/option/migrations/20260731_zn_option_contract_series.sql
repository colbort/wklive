-- OPT-P2-005：到期/行权价系列、双人审批、原子草稿生成与不可变谱系。
-- 可重复执行；不修改已有业务数据。

CREATE TABLE IF NOT EXISTS `t_option_contract_series` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `request_key` VARCHAR(96) NOT NULL DEFAULT '' COMMENT '客户端幂等键',
  `series_code` VARCHAR(24) NOT NULL DEFAULT '' COMMENT '系列代码',
  `version` BIGINT NOT NULL DEFAULT 1 COMMENT '追加版本号',
  `supersedes_id` BIGINT NOT NULL DEFAULT 0 COMMENT '被替代系列版本ID',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1待复核 2已生成 3已拒绝',
  `template_contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '保留字段；V1使用内嵌参数快照',
  `template_snapshot` JSON NOT NULL COMMENT '创建时的完整模板经济参数快照',
  `underlying_symbol` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '模板标的',
  `reference_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '审批参考价快照',
  `reference_source` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '权威参考价来源',
  `reference_time` BIGINT NOT NULL DEFAULT 0 COMMENT '参考价时间',
  `evidence_ref` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '不可变审批证据引用',
  `change_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '创建/修订原因',
  `payload_hash` CHAR(64) NOT NULL DEFAULT '' COMMENT '规范化输入SHA-256',
  `expected_contract_count` BIGINT NOT NULL DEFAULT 0 COMMENT '预计生成合约数',
  `generated_contract_count` BIGINT NOT NULL DEFAULT 0 COMMENT '实际生成合约数',
  `created_by` BIGINT NOT NULL DEFAULT 0 COMMENT '创建管理员',
  `reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '复核管理员',
  `review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '复核意见',
  `reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '复核时间',
  `generated_at` BIGINT NOT NULL DEFAULT 0 COMMENT '原子生成完成时间',
  `launch_status` TINYINT NOT NULL DEFAULT 0 COMMENT '上市复核：0不适用 1待复核 2批准 3拒绝',
  `launch_reviewed_by` BIGINT NOT NULL DEFAULT 0 COMMENT '上市复核管理员',
  `launch_review_reason` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '上市复核意见',
  `launch_reviewed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '上市复核时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_request` (`tenant_id`, `request_key`),
  UNIQUE KEY `uk_option_contract_series_version` (`tenant_id`, `series_code`, `version`),
  KEY `idx_option_contract_series_status` (`tenant_id`, `status`, `id`),
  CONSTRAINT `chk_option_contract_series` CHECK (
    `tenant_id` > 0 AND `request_key` <> '' AND `series_code` <> ''
    AND `version` > 0 AND `status` IN (1,2,3)
    AND `template_contract_id` = 0 AND JSON_VALID(`template_snapshot`)
    AND `underlying_symbol` <> '' AND `reference_price` > 0
    AND `reference_source` <> '' AND `reference_time` > 0
    AND `evidence_ref` <> '' AND `change_reason` <> ''
    AND CHAR_LENGTH(`payload_hash`) = 64
    AND `expected_contract_count` BETWEEN 2 AND 500
    AND `generated_contract_count` BETWEEN 0 AND `expected_contract_count`
    AND `created_by` > 0
    AND (
      (`status` = 1 AND `reviewed_by` = 0 AND `reviewed_at` = 0 AND `generated_contract_count` = 0 AND `generated_at` = 0
          AND `launch_status` = 0 AND `launch_reviewed_by` = 0 AND `launch_reviewed_at` = 0)
      OR (`status` = 2 AND `reviewed_by` > 0 AND `reviewed_by` <> `created_by`
          AND `review_reason` <> '' AND `reviewed_at` > 0
          AND `generated_contract_count` = `expected_contract_count` AND `generated_at` > 0
          AND (
            (`launch_status` = 1 AND `launch_reviewed_by` = 0 AND `launch_reviewed_at` = 0)
            OR (`launch_status` IN (2,3) AND `launch_reviewed_by` > 0
                AND `launch_reviewed_by` <> `created_by`
                AND `launch_review_reason` <> '' AND `launch_reviewed_at` > 0)
          ))
      OR (`status` = 3 AND `reviewed_by` > 0 AND `reviewed_by` <> `created_by`
          AND `review_reason` <> '' AND `reviewed_at` > 0
          AND `generated_contract_count` = 0 AND `generated_at` = 0
          AND `launch_status` = 0 AND `launch_reviewed_by` = 0 AND `launch_reviewed_at` = 0)
    )
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='不可变合约系列版本及审批';

CREATE TABLE IF NOT EXISTS `t_option_contract_series_expiry` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `series_id` BIGINT NOT NULL DEFAULT 0 COMMENT '系列版本ID',
  `sequence_no` BIGINT NOT NULL DEFAULT 0 COMMENT '稳定到期序号',
  `cycle_code` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '运营确认的周期标签',
  `list_time` BIGINT NOT NULL DEFAULT 0 COMMENT '上市时间',
  `exercise_cutoff_time` BIGINT NOT NULL DEFAULT 0 COMMENT '行权指令截止时间',
  `expire_time` BIGINT NOT NULL DEFAULT 0 COMMENT '到期/最后交易时间',
  `deliver_time` BIGINT NOT NULL DEFAULT 0 COMMENT '交割时间',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_expiry_seq` (`tenant_id`, `series_id`, `sequence_no`),
  UNIQUE KEY `uk_option_contract_series_expiry_time` (`tenant_id`, `series_id`, `expire_time`),
  KEY `idx_option_contract_series_expiry` (`tenant_id`, `series_id`, `id`),
  CONSTRAINT `chk_option_contract_series_expiry` CHECK (
    `tenant_id` > 0 AND `series_id` > 0 AND `sequence_no` > 0 AND `cycle_code` <> ''
    AND `list_time` > 0 AND `exercise_cutoff_time` > `list_time`
    AND `expire_time` >= `exercise_cutoff_time` AND `deliver_time` >= `expire_time`
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系列显式到期规格';

CREATE TABLE IF NOT EXISTS `t_option_contract_series_strike_band` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `series_id` BIGINT NOT NULL DEFAULT 0 COMMENT '系列版本ID',
  `sequence_no` BIGINT NOT NULL DEFAULT 0 COMMENT '稳定梯度序号',
  `lower_strike` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '闭区间下界',
  `upper_strike` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '闭区间上界',
  `strike_step` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '精确步长',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_band_seq` (`tenant_id`, `series_id`, `sequence_no`),
  KEY `idx_option_contract_series_band` (`tenant_id`, `series_id`, `id`),
  CONSTRAINT `chk_option_contract_series_band` CHECK (
    `tenant_id` > 0 AND `series_id` > 0 AND `sequence_no` > 0
    AND `lower_strike` > 0 AND `upper_strike` >= `lower_strike` AND `strike_step` > 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系列绝对行权价梯度';

CREATE TABLE IF NOT EXISTS `t_option_contract_series_detail` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `series_id` BIGINT NOT NULL DEFAULT 0 COMMENT '系列版本ID',
  `expiry_id` BIGINT NOT NULL DEFAULT 0 COMMENT '到期规格ID',
  `option_type` TINYINT NOT NULL DEFAULT 0 COMMENT '1 Call 2 Put',
  `strike_price` DECIMAL(32,16) NOT NULL DEFAULT 0 COMMENT '行权价',
  `contract_code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '确定性合约代码',
  `contract_id` BIGINT NOT NULL DEFAULT 0 COMMENT '生成的PENDING合约ID',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_contract_series_detail_leg` (`tenant_id`, `series_id`, `expiry_id`, `option_type`, `strike_price`),
  UNIQUE KEY `uk_option_contract_series_detail_code` (`tenant_id`, `contract_code`),
  UNIQUE KEY `uk_option_contract_series_detail_contract` (`tenant_id`, `contract_id`),
  KEY `idx_option_contract_series_detail` (`tenant_id`, `series_id`, `id`),
  CONSTRAINT `chk_option_contract_series_detail` CHECK (
    `tenant_id` > 0 AND `series_id` > 0 AND `expiry_id` > 0
    AND `option_type` IN (1,2) AND `strike_price` > 0
    AND `contract_code` <> '' AND `contract_id` > 0
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系列生成合约不可变谱系';

DROP TRIGGER IF EXISTS `trg_option_contract_series_guard_update`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_no_delete`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_expiry_no_update`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_expiry_no_delete`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_band_no_update`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_band_no_delete`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_detail_no_update`;
DROP TRIGGER IF EXISTS `trg_option_contract_series_detail_no_delete`;

DELIMITER $$
CREATE TRIGGER `trg_option_contract_series_guard_update`
BEFORE UPDATE ON `t_option_contract_series`
FOR EACH ROW
BEGIN
  IF NOT (
    OLD.tenant_id <=> NEW.tenant_id AND OLD.request_key <=> NEW.request_key
    AND OLD.series_code <=> NEW.series_code AND OLD.version <=> NEW.version
    AND OLD.supersedes_id <=> NEW.supersedes_id
    AND OLD.template_contract_id <=> NEW.template_contract_id
    AND OLD.template_snapshot <=> NEW.template_snapshot
    AND OLD.underlying_symbol <=> NEW.underlying_symbol
    AND OLD.reference_price <=> NEW.reference_price
    AND OLD.reference_source <=> NEW.reference_source
    AND OLD.reference_time <=> NEW.reference_time
    AND OLD.evidence_ref <=> NEW.evidence_ref
    AND OLD.change_reason <=> NEW.change_reason
    AND OLD.payload_hash <=> NEW.payload_hash
    AND OLD.expected_contract_count <=> NEW.expected_contract_count
    AND OLD.created_by <=> NEW.created_by
    AND OLD.create_times <=> NEW.create_times
  ) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series inputs are immutable';
  END IF;
  IF NOT (
    (OLD.status = 1 AND NEW.status IN (1,2,3))
    OR (OLD.status = 2 AND NEW.status = 2)
    OR (OLD.status = 3 AND NEW.status = 3)
  ) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'illegal contract series status transition';
  END IF;
  IF NEW.generated_contract_count < OLD.generated_contract_count THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series generated count cannot decrease';
  END IF;
  IF NOT (
    (OLD.launch_status = 0 AND NEW.launch_status IN (0,1))
    OR (OLD.launch_status = 1 AND NEW.launch_status IN (1,2,3))
    OR (OLD.launch_status = 2 AND NEW.launch_status = 2)
    OR (OLD.launch_status = 3 AND NEW.launch_status = 3)
  ) THEN
    SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'illegal contract series launch transition';
  END IF;
END$$

CREATE TRIGGER `trg_option_contract_series_no_delete`
BEFORE DELETE ON `t_option_contract_series`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series history cannot be deleted';
END$$

CREATE TRIGGER `trg_option_contract_series_expiry_no_update`
BEFORE UPDATE ON `t_option_contract_series_expiry`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series expiry is immutable';
END$$

CREATE TRIGGER `trg_option_contract_series_expiry_no_delete`
BEFORE DELETE ON `t_option_contract_series_expiry`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series expiry cannot be deleted';
END$$

CREATE TRIGGER `trg_option_contract_series_band_no_update`
BEFORE UPDATE ON `t_option_contract_series_strike_band`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series strike band is immutable';
END$$

CREATE TRIGGER `trg_option_contract_series_band_no_delete`
BEFORE DELETE ON `t_option_contract_series_strike_band`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series strike band cannot be deleted';
END$$

CREATE TRIGGER `trg_option_contract_series_detail_no_update`
BEFORE UPDATE ON `t_option_contract_series_detail`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series generation detail is immutable';
END$$

CREATE TRIGGER `trg_option_contract_series_detail_no_delete`
BEFORE DELETE ON `t_option_contract_series_detail`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'contract series generation detail cannot be deleted';
END$$
DELIMITER ;
