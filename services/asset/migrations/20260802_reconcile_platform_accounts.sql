-- dbinit:baseline-safe
-- 历史迁移 20260721_platform_accounts.sql 已发布后不可修改。
-- 在独立 reconciliation migration 中兼容旧保险覆盖表和已完成改造的当前基线。

SET @asset_platform_upgrade_sql = (
  SELECT IF(
    EXISTS (
      SELECT 1
      FROM information_schema.columns
      WHERE table_schema = DATABASE()
        AND table_name = 't_asset_insurance_cover'
        AND column_name = 'fund_user_id'
    ),
    'ALTER TABLE `t_asset_insurance_cover`
       ADD COLUMN `platform_account_id` BIGINT NOT NULL DEFAULT 0 AFTER `tenant_id`,
       DROP INDEX `idx_fund_asset_time`,
       ADD KEY `idx_fund_asset_time` (`tenant_id`,`platform_account_id`,`coin`,`create_times`),
       DROP COLUMN `fund_user_id`,
       DROP COLUMN `wallet_type`',
    'SELECT 1'
  )
);
PREPARE asset_platform_upgrade_stmt FROM @asset_platform_upgrade_sql;
EXECUTE asset_platform_upgrade_stmt;
DEALLOCATE PREPARE asset_platform_upgrade_stmt;
