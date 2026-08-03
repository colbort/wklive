-- Staking uses the delay queue as an acceleration path, while this cron job is
-- the authoritative catch-up and recovery path. Keep the migration replayable.
INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '质押收益发放与到期结算', 'STAKING',
  'staking.ProcessRewardsAndSettleOrders', '0 * * * * *', 1,
  '每分钟补发到期收益、结算到期订单并恢复未完成资金操作；处理逻辑具备租户级锁和业务幂等',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job
  WHERE invoke_target = 'staking.ProcessRewardsAndSettleOrders'
);

UPDATE sys_job
SET job_name = '质押收益发放与到期结算',
    job_group = 'STAKING',
    cron_expression = '0 * * * * *',
    status = 1,
    remark = '每分钟补发到期收益、结算到期订单并恢复未完成资金操作；处理逻辑具备租户级锁和业务幂等',
    update_by = 'system',
    update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE invoke_target = 'staking.ProcessRewardsAndSettleOrders';
