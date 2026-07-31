INSERT INTO sys_job
(job_name, job_group, invoke_target, cron_expression, status, remark,
 create_by, create_times, update_by, update_times)
SELECT
  '期权日终钱包镜像对账', 'OPTION', 'option.ProcessDailyReconciliation', '15 5 * * * *', 1,
  '每小时触发，仅UTC 00点窗口自动关闭前一业务日；成功后跳过，显式租户可受控重跑',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target='option.ProcessDailyReconciliation'
);

UPDATE sys_job
SET cron_expression='15 5 * * * *',
    remark='每小时触发，仅UTC 00点窗口自动关闭前一业务日；成功后跳过，显式租户可受控重跑',
    update_by='system',update_times=UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000
WHERE invoke_target='option.ProcessDailyReconciliation';
