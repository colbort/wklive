-- dbinit:baseline-safe
-- LockAsset now uses t_asset_idempotent. The business lock itself also needs a
-- unique identity so a retry can never create two lock rows.

SET @dbinit_exists := (
  SELECT COUNT(*) FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 't_asset_lock'
    AND index_name = 'uk_tenant_biz_lock'
);
SET @dbinit_sql := IF(
  @dbinit_exists = 0,
  'ALTER TABLE `t_asset_lock` ADD UNIQUE KEY `uk_tenant_biz_lock` (`tenant_id`,`biz_type`,`biz_no`)',
  'SELECT 1'
);
PREPARE dbinit_stmt FROM @dbinit_sql;
EXECUTE dbinit_stmt;
DEALLOCATE PREPARE dbinit_stmt;
