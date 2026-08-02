-- Admin RPC 补齐后的接口权限。
-- liquidity-admin-api 对未登记到 sys_menu 的接口采用 fail-closed（405），
-- 因此所有新增路由都必须在这里注册，并授予做市管理员角色。

INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (20105, 20100, 2, '查看提供方详情', 3, 'GET', '/admin/liquidity/providers/{id}',
   'liquidity:provider:detail', '', '', 20105, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20106, 20100, 2, '编辑提供方', 3, 'PUT', '/admin/liquidity/providers/{id}',
   'liquidity:provider:update', '', '', 20106, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20208, 20200, 2, '撤销策略全部报价', 3, 'POST',
   '/admin/liquidity/symbol-configs/{id}/cancel-quotes', 'liquidity:quote:cancel-all', '', '',
   20208, 2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20301, 20300, 2, '查询报价周期', 3, 'GET', '/admin/liquidity/quote-cycles',
   'liquidity:quote-cycle:list', '', '', 20301, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20401, 20400, 2, '查询外部成交', 3, 'GET', '/admin/liquidity/external-fills',
   'liquidity:external-fill:list', '', '', 20401, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20402, 20400, 2, '撤销外部订单', 3, 'POST',
   '/admin/liquidity/external-orders/{id}/cancel', 'liquidity:external-order:cancel', '', '',
   20402, 2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20501, 20500, 2, '创建人工对冲', 3, 'POST', '/admin/liquidity/hedge-tasks/manual',
   'liquidity:hedge:create', '', '', 20501, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20502, 20500, 2, '取消对冲任务', 3, 'POST',
   '/admin/liquidity/hedge-tasks/{id}/cancel', 'liquidity:hedge:cancel', '', '', 20502, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20503, 20500, 2, '重试对冲任务', 3, 'POST',
   '/admin/liquidity/hedge-tasks/{id}/retry', 'liquidity:hedge:retry', '', '', 20503, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20504, 20500, 2, '查询库存快照', 3, 'GET', '/admin/liquidity/inventory-snapshots',
   'liquidity:inventory:list', '', '', 20504, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20505, 20500, 2, '查询最新库存', 3, 'GET',
   '/admin/liquidity/inventory-snapshots/latest', 'liquidity:inventory:latest', '', '',
   20505, 2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20601, 20600, 2, '处置风险事件', 3, 'POST',
   '/admin/liquidity/risk-events/{id}/resolve', 'liquidity:risk:resolve', '', '', 20601, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20701, 20700, 2, '发起对账', 3, 'POST', '/admin/liquidity/reconcile-batches/run',
   'liquidity:reconcile:run', '', '', 20701, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20702, 20700, 2, '查询对账明细', 3, 'GET',
   '/admin/liquidity/reconcile-batches/{batchId}/details', 'liquidity:reconcile:detail', '', '',
   20702, 2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20703, 20700, 2, '处置对账差异', 3, 'POST',
   '/admin/liquidity/reconcile-differences/{id}/resolve', 'liquidity:reconcile:resolve', '', '',
   20703, 2, 1, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  app_scope = VALUES(app_scope),
  name = VALUES(name),
  menu_type = VALUES(menu_type),
  method = VALUES(method),
  path = VALUES(path),
  perms = VALUES(perms),
  visible = VALUES(visible),
  enabled = VALUES(enabled),
  update_times = VALUES(update_times);

-- 保险覆盖证据查询是平台兜底政策创建、复核的共同只读能力。
INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (759, 751, 1, '查询保险覆盖证据', 3, 'GET', '/asset/insurance-covers',
   'asset:insurance-cover:detail', '', '', 759, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  app_scope = VALUES(app_scope),
  name = VALUES(name),
  menu_type = VALUES(menu_type),
  method = VALUES(method),
  path = VALUES(path),
  perms = VALUES(perms),
  visible = VALUES(visible),
  enabled = VALUES(enabled),
  update_times = VALUES(update_times);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (
  20105, 20106, 20208, 20301, 20401, 20402,
  20501, 20502, 20503, 20504, 20505,
  20601, 20701, 20702, 20703
)
WHERE r.code = 'liquidity_admin'
  AND r.app_scope = 2;

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, 759
FROM sys_role r
WHERE r.app_scope = 1
  AND r.code IN (
    'super_admin',
    'platform_backstop_policy_creator',
    'platform_backstop_policy_reviewer'
  );
