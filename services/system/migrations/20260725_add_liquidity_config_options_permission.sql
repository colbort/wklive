INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (20205, 20200, 2, '做市配置选项', 3, 'GET', '/admin/liquidity/config-options',
   'liquidity:strategy:list', '', '', 20205, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  parent_id = VALUES(parent_id), app_scope = VALUES(app_scope), name = VALUES(name),
  menu_type = VALUES(menu_type), method = VALUES(method), path = VALUES(path),
  perms = VALUES(perms), enabled = VALUES(enabled), update_times = VALUES(update_times);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, 20205
FROM sys_role r
WHERE r.code = 'liquidity_admin'
  AND r.app_scope = 2;
