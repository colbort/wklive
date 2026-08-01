-- dbinit:baseline-safe
-- Give Option's duplicate-freeze monitor a covering access path and adopt only
-- unambiguous legacy freezes into Asset's durable idempotency ledger. Duplicate
-- business keys are intentionally left untouched so the SEV-1 monitor can force
-- manual reconciliation instead of guessing which freeze is authoritative.

SET @option_freeze_index_exists := (
  SELECT COUNT(1)
  FROM information_schema.statistics
  WHERE table_schema=DATABASE()
    AND table_name='t_asset_freeze'
    AND index_name='idx_asset_freeze_option_business_key'
);
SET @option_freeze_index_sql := IF(
  @option_freeze_index_exists=0,
  'ALTER TABLE `t_asset_freeze` ADD INDEX `idx_asset_freeze_option_business_key` (`biz_type`,`tenant_id`,`scene_type`,`biz_no`,`create_times`,`id`)',
  'SELECT 1'
);
PREPARE option_freeze_index_stmt FROM @option_freeze_index_sql;
EXECUTE option_freeze_index_stmt;
DEALLOCATE PREPARE option_freeze_index_stmt;

INSERT INTO t_asset_idempotent (
  tenant_id,biz_type,biz_no,scene_type,status,remark,create_times,update_times
)
SELECT
  tenant_id,biz_type,biz_no,scene_type,2,
  'backfilled from unique option freeze evidence',
  MIN(create_times),MAX(update_times)
FROM t_asset_freeze
WHERE biz_type='option'
  AND TRIM(biz_no)<>''
GROUP BY tenant_id,biz_type,biz_no,scene_type
HAVING COUNT(*)=1
ON DUPLICATE KEY UPDATE id=id;
