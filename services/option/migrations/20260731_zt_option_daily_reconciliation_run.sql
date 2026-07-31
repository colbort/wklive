-- Option 日终对账运行记录。每次尝试只追加一行，既保存钱包镜像零差额成功心跳，
-- 也为后续 Asset/财务完整资金守恒生产者预留独立 scope=2。
-- Asset 钱包使用 DECIMAL(36,18)；镜像和权威流水摘要必须保持同等精度，扩列不会截断历史值。
ALTER TABLE `t_option_account`
  MODIFY COLUMN `balance` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset Option钱包总额镜像',
  MODIFY COLUMN `available_balance` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset Option钱包可用额镜像',
  MODIFY COLUMN `frozen_balance` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset Option钱包冻结额镜像';

ALTER TABLE `t_option_bill`
  MODIFY COLUMN `change_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset权威变动金额，正负都有可能',
  MODIFY COLUMN `balance_before` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset权威变动前余额',
  MODIFY COLUMN `balance_after` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Asset权威变动后余额';

CREATE TABLE IF NOT EXISTS `t_option_reconciliation_run` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL DEFAULT 0 COMMENT '租户ID',
  `business_date` CHAR(10) NOT NULL DEFAULT '' COMMENT 'UTC业务日期YYYY-MM-DD',
  `scope` TINYINT NOT NULL DEFAULT 1 COMMENT '范围：1钱包镜像 2完整资金守恒',
  `attempt_no` INT NOT NULL DEFAULT 1 COMMENT '当日同范围执行序号',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1成功 2有差异 3执行失败',
  `snapshot_time` BIGINT NOT NULL DEFAULT 0 COMMENT '一致性快照Unix秒',
  `snapshot_ref` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '快照或证据引用',
  `coin_count` BIGINT NOT NULL DEFAULT 0 COMMENT '检查币种数',
  `account_count` BIGINT NOT NULL DEFAULT 0 COMMENT '检查用户钱包数',
  `mismatch_count` BIGINT NOT NULL DEFAULT 0 COMMENT '差异用户钱包数',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '执行摘要或失败原因',
  `completed_at` BIGINT NOT NULL DEFAULT 0 COMMENT '完成Unix秒',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_reconciliation_run_attempt` (`tenant_id`,`business_date`,`scope`,`attempt_no`),
  UNIQUE KEY `uk_option_reconciliation_run_identity` (`id`,`tenant_id`,`business_date`,`scope`),
  KEY `idx_option_reconciliation_run_latest` (`tenant_id`,`scope`,`completed_at`,`status`,`id`),
  KEY `idx_option_reconciliation_run_monitor` (`scope`,`status`,`completed_at`,`tenant_id`,`id`),
  CONSTRAINT `chk_option_reconciliation_run` CHECK (
    `business_date` REGEXP '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'
    AND `scope` IN (1,2)
    AND `attempt_no` > 0
    AND `status` IN (1,2,3)
    AND `snapshot_time` > 0
    AND `coin_count` >= 0
    AND `account_count` >= 0
    AND `mismatch_count` >= 0
    AND `mismatch_count` <= `account_count`
    AND `completed_at` >= `snapshot_time`
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option日终对账运行与成功心跳';

CREATE TABLE IF NOT EXISTS `t_option_reconciliation_run_detail` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `run_id` BIGINT NOT NULL COMMENT '不可变对账运行ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID，冗余用于租户隔离查询',
  `business_date` CHAR(10) NOT NULL COMMENT 'UTC业务日期YYYY-MM-DD',
  `scope` TINYINT NOT NULL DEFAULT 2 COMMENT '范围，明细当前仅允许完整资金守恒scope=2',
  `dimension_type` TINYINT NOT NULL COMMENT '维度：1用户钱包逐币 2平台账户逐币 3Option子账逐币',
  `dimension_key` VARCHAR(96) NOT NULL COMMENT 'coin或account_type:coin稳定键',
  `opening_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '业务日期期初总额',
  `external_net` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '充值提现及跨钱包划转净额',
  `option_net` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT 'Option业务Asset流水净额',
  `manual_net` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '人工增减净额',
  `expected_closing` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '期初加三类净额',
  `actual_closing` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '由一致性截止点反推的实际日终',
  `difference_amount` DECIMAL(36,18) NOT NULL DEFAULT 0 COMMENT '实际日终减预期日终',
  `flow_count` BIGINT NOT NULL DEFAULT 0 COMMENT '纳入计算的权威流水数',
  `mismatch_count` BIGINT NOT NULL DEFAULT 0 COMMENT '余额链、字段恒等式或子账关联异常数',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1一致 2有差异 3数据不完整',
  `evidence_ref` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '游标、查询或快照证据引用',
  `detail` VARCHAR(1000) NOT NULL DEFAULT '' COMMENT '分类和完整性摘要',
  `create_times` BIGINT NOT NULL DEFAULT 0 COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_option_reconciliation_detail_dimension` (`run_id`,`dimension_type`,`dimension_key`),
  KEY `idx_option_reconciliation_detail_lookup` (`tenant_id`,`business_date`,`scope`,`status`,`id`),
  CONSTRAINT `fk_option_reconciliation_detail_run`
    FOREIGN KEY (`run_id`,`tenant_id`,`business_date`,`scope`)
    REFERENCES `t_option_reconciliation_run` (`id`,`tenant_id`,`business_date`,`scope`)
    ON DELETE RESTRICT ON UPDATE RESTRICT,
  CONSTRAINT `chk_option_reconciliation_detail` CHECK (
    `business_date` REGEXP '^[0-9]{4}-[0-9]{2}-[0-9]{2}$'
    AND `scope`=2
    AND `dimension_type` IN (1,2,3)
    AND `dimension_key`<>''
    AND `flow_count`>=0
    AND `mismatch_count`>=0
    AND `status` IN (1,2,3)
    AND `expected_closing`=`opening_amount`+`external_net`+`option_net`+`manual_net`
    AND `difference_amount`=`actual_closing`-`expected_closing`
    AND ((`status`=1 AND `difference_amount`=0 AND `mismatch_count`=0) OR `status`<>1)
  )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Option完整资金守恒逐维度不可变证据';

DROP TRIGGER IF EXISTS `trg_option_reconciliation_run_no_update`;
CREATE TRIGGER `trg_option_reconciliation_run_no_update`
BEFORE UPDATE ON `t_option_reconciliation_run`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation run is immutable';

DROP TRIGGER IF EXISTS `trg_option_reconciliation_run_no_delete`;
CREATE TRIGGER `trg_option_reconciliation_run_no_delete`
BEFORE DELETE ON `t_option_reconciliation_run`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation run cannot be deleted';

DROP TRIGGER IF EXISTS `trg_option_reconciliation_detail_no_update`;
CREATE TRIGGER `trg_option_reconciliation_detail_no_update`
BEFORE UPDATE ON `t_option_reconciliation_run_detail`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation detail is immutable';

DROP TRIGGER IF EXISTS `trg_option_reconciliation_detail_no_delete`;
CREATE TRIGGER `trg_option_reconciliation_detail_no_delete`
BEFORE DELETE ON `t_option_reconciliation_run_detail`
FOR EACH ROW
SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='option reconciliation detail cannot be deleted';
