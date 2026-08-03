-- Staking 持久化资金操作可观测与人工恢复入口。
INSERT INTO sys_menu
  (id, parent_id, app_scope, name, menu_type, method, path, perms, component, icon, sort,
   visible, enabled, create_times, update_times)
VALUES
  (850, 800, 1, '资金操作与恢复', 2, 'GET', '/staking/operations',
   'staking:operation:list', 'staking/operations', 'Warning', 850, 1, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (851, 850, 1, '查询质押资金操作', 3, 'GET', '/staking/operations',
   'staking:operation:list', '', '', 851, 2, 1,
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (852, 850, 1, '重试质押资金操作', 3, 'POST', '/staking/operations/retry',
   'staking:operation:retry', '', '', 852, 2, 1,
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
JOIN sys_menu m ON m.id IN (850, 851, 852)
WHERE r.app_scope = 1 AND r.code IN ('super_admin', 'tenant_super_admin');
