-- Option Scope 2 日终复算按租户、钱包类型和UTC截止时间读取流水，并在窗口内按ID稳定排序。
SET @option_recon_flow_index_exists := (
  SELECT COUNT(1) FROM information_schema.statistics
  WHERE table_schema=DATABASE() AND table_name='t_asset_flow'
    AND index_name='idx_asset_flow_option_reconciliation'
);
SET @option_recon_flow_index_sql := IF(
  @option_recon_flow_index_exists=0,
  'ALTER TABLE `t_asset_flow` ADD INDEX `idx_asset_flow_option_reconciliation` (`tenant_id`,`wallet_type`,`create_times`,`id`,`user_id`,`coin`)',
  'SELECT 1'
);
PREPARE option_recon_flow_index_stmt FROM @option_recon_flow_index_sql;
EXECUTE option_recon_flow_index_stmt;
DEALLOCATE PREPARE option_recon_flow_index_stmt;

SET @option_recon_platform_index_exists := (
  SELECT COUNT(1) FROM information_schema.statistics
  WHERE table_schema=DATABASE() AND table_name='t_asset_platform_flow'
    AND index_name='idx_platform_flow_option_reconciliation'
);
SET @option_recon_platform_index_sql := IF(
  @option_recon_platform_index_exists=0,
  'ALTER TABLE `t_asset_platform_flow` ADD INDEX `idx_platform_flow_option_reconciliation` (`tenant_id`,`account_type`,`create_times`,`id`,`platform_account_id`,`coin`)',
  'SELECT 1'
);
PREPARE option_recon_platform_index_stmt FROM @option_recon_platform_index_sql;
EXECUTE option_recon_platform_index_stmt;
DEALLOCATE PREPARE option_recon_platform_index_stmt;
