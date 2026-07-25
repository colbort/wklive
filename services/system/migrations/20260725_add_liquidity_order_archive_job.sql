INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
'做市零成交撤单归档', 'TRADE', 'trade.ArchiveLiquidityOrders', '0 20 3 * * *', 1,
'每日归档超过保留期且不存在成交明细的做市撤单',
'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ArchiveLiquidityOrders'
);
