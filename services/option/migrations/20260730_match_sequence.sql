ALTER TABLE `t_option_trade`
  ADD COLUMN `match_sequence` BIGINT NOT NULL DEFAULT 0 COMMENT '合约内严格递增撮合序号' AFTER `maker_side`;

UPDATE `t_option_trade` AS target
JOIN (
  SELECT `id`, ROW_NUMBER() OVER (
    PARTITION BY `tenant_id`, `contract_id`
    ORDER BY `trade_time`, `id`
  ) AS `sequence_no`
  FROM `t_option_trade`
) AS ranked ON ranked.id = target.id
SET target.match_sequence = ranked.sequence_no;

ALTER TABLE `t_option_trade`
  ADD UNIQUE KEY `uk_tenant_contract_match_sequence` (`tenant_id`, `contract_id`, `match_sequence`);

CREATE TABLE IF NOT EXISTS `t_option_match_sequence` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `contract_id` BIGINT NOT NULL COMMENT '期权合约ID',
  `next_sequence` BIGINT NOT NULL DEFAULT 1 COMMENT '下一个撮合序号',
  `update_times` BIGINT NOT NULL DEFAULT 0 COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_contract` (`tenant_id`, `contract_id`),
  CONSTRAINT `chk_option_match_sequence` CHECK (`next_sequence` > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='期权合约撮合序列';

INSERT INTO `t_option_match_sequence` (`tenant_id`, `contract_id`, `next_sequence`, `update_times`)
SELECT `tenant_id`, `contract_id`, MAX(`match_sequence`) + 1, UNIX_TIMESTAMP()
FROM `t_option_trade`
GROUP BY `tenant_id`, `contract_id`
ON DUPLICATE KEY UPDATE
  `next_sequence` = GREATEST(`next_sequence`, VALUES(`next_sequence`)),
  `update_times` = VALUES(`update_times`);
