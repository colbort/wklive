ALTER TABLE `t_contract_funding_settlement`
  ADD COLUMN `position_version` BIGINT NOT NULL DEFAULT 0 COMMENT '批次锁定时持仓版本' AFTER `position_qty`;
