-- liquidity-admin-api 接口权限。
-- 页面路由由 liquidity-admin-ui 自己维护；sys_menu 在这里作为统一 RBAC 权限资源使用。
INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (20000, 0, 2, '做市管理', 1, '', '/liquidity', '', '', '', 20000, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20001, 20000, 2, '做市看板', 2, 'GET', '/admin/liquidity/dashboard',
   'liquidity:dashboard:view', '', '', 20001, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20100, 20000, 2, '流动性提供方', 2, 'GET', '/admin/liquidity/providers',
   'liquidity:provider:list', '', '', 20100, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20101, 20100, 2, '新增提供方', 3, 'POST', '/admin/liquidity/providers',
   'liquidity:provider:add', '', '', 20101, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20102, 20100, 2, '测试连接', 3, 'POST', '/admin/liquidity/providers/{id}/test',
   'liquidity:provider:test', '', '', 20102, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20103, 20100, 2, '修改状态', 3, 'PUT', '/admin/liquidity/providers/{id}/status',
   'liquidity:provider:status', '', '', 20103, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20200, 20000, 2, '做市策略', 2, 'GET', '/admin/liquidity/symbol-configs',
   'liquidity:strategy:list', '', '', 20200, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20201, 20200, 2, '新增策略', 3, 'POST', '/admin/liquidity/symbol-configs',
   'liquidity:strategy:add', '', '', 20201, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20202, 20200, 2, '启动策略', 3, 'POST', '/admin/liquidity/symbol-configs/{id}/start',
   'liquidity:strategy:start', '', '', 20202, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20203, 20200, 2, '暂停策略', 3, 'POST', '/admin/liquidity/symbol-configs/{id}/pause',
   'liquidity:strategy:pause', '', '', 20203, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20204, 20200, 2, '停止策略', 3, 'POST', '/admin/liquidity/symbol-configs/{id}/stop',
   'liquidity:strategy:stop', '', '', 20204, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20300, 20000, 2, '内部报价单', 2, 'GET', '/admin/liquidity/quote-orders',
   'liquidity:quote:list', '', '', 20300, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20400, 20000, 2, '外部订单', 2, 'GET', '/admin/liquidity/external-orders',
   'liquidity:external-order:list', '', '', 20400, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20500, 20000, 2, '对冲任务', 2, 'GET', '/admin/liquidity/hedge-tasks',
   'liquidity:hedge:list', '', '', 20500, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20600, 20000, 2, '风险事件', 2, 'GET', '/admin/liquidity/risk-events',
   'liquidity:risk:list', '', '', 20600, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20700, 20000, 2, '对账批次', 2, 'GET', '/admin/liquidity/reconcile-batches',
   'liquidity:reconcile:list', '', '', 20700, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  name = VALUES(name),
  menu_type = VALUES(menu_type),
  method = VALUES(method),
  path = VALUES(path),
  perms = VALUES(perms),
  enabled = VALUES(enabled),
  update_times = VALUES(update_times);

-- 给所有租户中 code=liquidity_admin 的角色授予做市后台全部权限。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (
  20000, 20001, 20100, 20101, 20102, 20103,
  20200, 20201, 20202, 20203, 20204,
  20300, 20400, 20500, 20600, 20700
)
WHERE r.code = 'liquidity_admin'
  AND r.app_scope = 2;
