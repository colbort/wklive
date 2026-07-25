INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (20104, 20100, 2, '一键创建内部做市账户', 3, 'POST',
   '/admin/liquidity/providers/provision', 'liquidity:provider:add', '', '', 20104, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id),
  app_scope = VALUES(app_scope),
  name = VALUES(name),
  menu_type = VALUES(menu_type),
  method = VALUES(method),
  path = VALUES(path),
  perms = VALUES(perms),
  enabled = VALUES(enabled),
  update_times = VALUES(update_times);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, 20104
FROM sys_role r
WHERE r.code = 'liquidity_admin'
  AND r.app_scope = 2;
