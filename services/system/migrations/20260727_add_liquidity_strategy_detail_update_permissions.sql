-- 做市策略详情与编辑权限。
-- 使用独立迁移，避免修改已经执行过的 20260724 权限迁移。
INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (20206, 20200, 2, '策略详情', 3, 'GET', '/admin/liquidity/symbol-configs/{id}',
   'liquidity:strategy:detail', '', '', 20206, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (20207, 20200, 2, '编辑策略', 3, 'PUT', '/admin/liquidity/symbol-configs/{id}',
   'liquidity:strategy:update', '', '', 20207, 2, 1,
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

-- 已有做市管理员角色直接获得新增权限。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT r.tenant_id, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (20206, 20207)
WHERE r.code = 'liquidity_admin'
  AND r.app_scope = 2;
