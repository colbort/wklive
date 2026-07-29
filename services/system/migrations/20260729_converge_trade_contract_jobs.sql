-- Converge legacy rows for the five system-owned Trade recovery jobs.
-- Older installations may already contain the invoke_target with a disabled
-- status, DEFAULT group or five-field cron expression. The earlier
-- insert-if-missing migration intentionally cannot repair those rows.

UPDATE sys_job
SET job_name = '交易撮合恢复',
    job_group = 'TRADE',
    cron_expression = '*/5 * * * * *',
    status = 1,
    remark = '每5秒恢复尚未完成的订单撮合事件',
    update_by = 'system',
    update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE invoke_target = 'trade.ProcessOrderMatching';

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '交易撮合恢复', 'TRADE', 'trade.ProcessOrderMatching', '*/5 * * * * *', 1,
  '每5秒恢复尚未完成的订单撮合事件', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessOrderMatching'
);

UPDATE sys_job
SET job_name = '合约仓位与风险处理',
    job_group = 'TRADE',
    cron_expression = '*/1 * * * * *',
    status = 1,
    remark = '每秒恢复合约仓位投影并执行风险扫描',
    update_by = 'system',
    update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE invoke_target = 'trade.ProcessPositions';

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '合约仓位与风险处理', 'TRADE', 'trade.ProcessPositions', '*/1 * * * * *', 1,
  '每秒恢复合约仓位投影并执行风险扫描', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessPositions'
);

UPDATE sys_job
SET job_name = '合约资金费与交割结算',
    job_group = 'TRADE',
    cron_expression = '*/1 * * * * *',
    status = 1,
    remark = '每秒处理永续资金费、交割生命周期及未完成结算',
    update_by = 'system',
    update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE invoke_target = 'trade.ProcessContractSettlements';

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '合约资金费与交割结算', 'TRADE', 'trade.ProcessContractSettlements',
  '*/1 * * * * *', 1, '每秒处理永续资金费、交割生命周期及未完成结算',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessContractSettlements'
);

UPDATE sys_job
SET job_name = '交易事件恢复',
    job_group = 'TRADE',
    cron_expression = '*/1 * * * * *',
    status = 1,
    remark = '每秒恢复未投递或未完成的交易领域事件',
    update_by = 'system',
    update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE invoke_target = 'trade.ProcessTradeEvents';

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '交易事件恢复', 'TRADE', 'trade.ProcessTradeEvents', '*/1 * * * * *', 1,
  '每秒恢复未投递或未完成的交易领域事件', 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessTradeEvents'
);

UPDATE sys_job
SET job_name = '秒合约结算兜底',
    job_group = 'TRADE',
    cron_expression = '0 * * * * *',
    status = 1,
    remark = '每分钟扫描秒合约激活、到期结算及退款漏单',
    update_by = 'system',
    update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE invoke_target = 'trade.ProcessSecondsSettlements';

INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '秒合约结算兜底', 'TRADE', 'trade.ProcessSecondsSettlements',
  '0 * * * * *', 1, '每分钟扫描秒合约激活、到期结算及退款漏单',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, 'system',
  UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'trade.ProcessSecondsSettlements'
);
