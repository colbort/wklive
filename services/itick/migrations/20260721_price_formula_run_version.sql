ALTER TABLE `t_itick_price_formula`
  ADD COLUMN `run_version` BIGINT NOT NULL DEFAULT 0 COMMENT 'Worker CAS version' AFTER `version`;
