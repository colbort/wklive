-- Daily cumulative accounting reconciliation. The snapshot exposes principal,
-- reward and fee differences per tenant/coin for operational review.
INSERT INTO sys_job
(`job_name`, `job_group`, `invoke_target`, `cron_expression`, `status`, `remark`,
 `create_by`, `create_times`, `update_by`, `update_times`)
SELECT
  '质押每日账实对账', 'STAKING',
  'staking.ReconcileStaking', '0 5 0 * * *', 1,
  '每日00:05按租户和币种核对订单本金、产品汇总、用户持仓、Asset锁仓、奖励支出及手续费收入',
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
  'system', UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE NOT EXISTS (
  SELECT 1 FROM sys_job WHERE invoke_target = 'staking.ReconcileStaking'
);

UPDATE sys_job
SET job_name = '质押每日账实对账', job_group = 'STAKING',
    cron_expression = '0 5 0 * * *', status = 1,
    remark = '每日00:05按租户和币种核对订单本金、产品汇总、用户持仓、Asset锁仓、奖励支出及手续费收入',
    update_by = 'system', update_times = UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000
WHERE invoke_target = 'staking.ReconcileStaking';

INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (860, 800, 1, '质押账实对账', 2, 'GET', '/staking/reconciliations',
   'staking:reconciliation:list', 'staking/reconciliations', 'DataAnalysis', 860, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (861, 860, 1, '查询质押账实对账', 3, 'GET', '/staking/reconciliations',
   'staking:reconciliation:list', '', '', 861, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id), app_scope = VALUES(app_scope), name = VALUES(name),
  menu_type = VALUES(menu_type), method = VALUES(method), path = VALUES(path),
  perms = VALUES(perms), component = VALUES(component), icon = VALUES(icon),
  sort = VALUES(sort), visible = VALUES(visible), enabled = VALUES(enabled),
  update_times = VALUES(update_times);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (860, 861)
WHERE r.app_scope = 1 AND r.code IN ('super_admin', 'tenant_super_admin');
