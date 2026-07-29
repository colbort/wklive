-- dbinit:baseline-safe
-- 旧库的保险配置仍保留已废弃的 fund_user_id/wallet_type 必填列。
-- 当前协议已将这两个字段保留（reserved），实际资金统一由 Asset 平台账户承载。
-- 保留旧列和已有值，仅为新配置提供兼容默认值，并把约束收敛到当前有效字段。

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_insurance_fund_account'
    AND column_name = 'fund_user_id'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_insurance_fund_account` MODIFY COLUMN `fund_user_id` BIGINT NOT NULL DEFAULT 0 COMMENT ''已废弃：保险资金由Asset平台账户承载''',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 't_contract_insurance_fund_account'
    AND column_name = 'wallet_type'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_insurance_fund_account` MODIFY COLUMN `wallet_type` TINYINT NOT NULL DEFAULT 0 COMMENT ''已废弃：保险资金由Asset平台账户承载''',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.table_constraints
  WHERE constraint_schema = DATABASE()
    AND table_name = 't_contract_insurance_fund_account'
    AND constraint_name = 'chk_insurance_fund_account'
    AND constraint_type = 'CHECK'
);
SET @dbinit_sql := IF(
  @dbinit_exists > 0,
  'ALTER TABLE `t_contract_insurance_fund_account` DROP CHECK `chk_insurance_fund_account`',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;

ALTER TABLE `t_contract_insurance_fund_account`
  ADD CONSTRAINT `chk_insurance_fund_account`
  CHECK (`adl_enabled` IN (1,2) AND `status` IN (1,2));
