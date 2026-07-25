INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
'做市报价刷新', 'LIQUIDITY', 'liquidity.RefreshQuotes', '*/1 * * * * *', 1,
'扫描运行中的做市配置并按各配置刷新间隔撤旧挂新',
'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'liquidity.RefreshQuotes'
);

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`, `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
'做市报价恢复', 'LIQUIDITY', 'liquidity.RecoverQuoteOrders', '*/10 * * * * *', 1,
'恢复提交结果不确定的内部做市报价状态',
'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'liquidity.RecoverQuoteOrders'
);
