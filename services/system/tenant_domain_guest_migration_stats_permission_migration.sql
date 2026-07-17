INSERT INTO sys_menu
  (id, parent_id, name, menu_type, method, path, perms, sort)
VALUES
  (11304, 11300, '游客迁移统计', 3, 'GET',
   '/system/tenant-domains/guest-migration-stats',
   'sys:tenant-domain:guest-migration-stats', 11304)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  method = VALUES(method),
  path = VALUES(path),
  perms = VALUES(perms),
  sort = VALUES(sort);

INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id) VALUES
  (0, 1, 11304),
  (0, 2, 11304);
